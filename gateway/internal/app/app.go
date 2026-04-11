package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gobili/gateway/internal/config"
	"gobili/gateway/internal/types"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db    *sql.DB
	redis *redis.Client
}

type App struct {
	cfg  config.Config
	repo *Repository
}

type AuthUser struct {
	UserID string
	Tokens map[string]string
}

type UserRecord struct {
	ID           string
	Username     string
	PasswordHash string
	AvatarURL    string
	MFASecret    string
	MFAEnabled   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    sql.NullTime
}

type VideoRecord struct {
	ID           string
	UserID       string
	VideoURL     string
	CoverURL     string
	Title        string
	Description  string
	VisitCount   int64
	LikeCount    int64
	CommentCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    sql.NullTime
}

type CommentRecord struct {
	ID         string
	UserID     string
	VideoID    string
	ParentID   string
	LikeCount  int64
	ChildCount int64
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  sql.NullTime
}

type SocialRecord struct {
	ID        string
	Username  string
	AvatarURL string
}

func NewApp(cfg config.Config) (*App, error) {
	db, err := sql.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})
	return &App{cfg: cfg, repo: &Repository{db: db, redis: rdb}}, nil
}

func (a *App) Close() error {
	if a == nil || a.repo == nil || a.repo.db == nil {
		return nil
	}
	return a.repo.db.Close()
}

func (a *App) MustAuth(ctx context.Context) (AuthUser, error) {
	value := ctx.Value("authUser")
	if value == nil {
		return AuthUser{}, errors.New("unauthorized")
	}
	user, ok := value.(AuthUser)
	if !ok || user.UserID == "" {
		return AuthUser{}, errors.New("unauthorized")
	}
	return user, nil
}

func (a *App) Register(ctx context.Context, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return errors.New("username and password are required")
	}
	if _, err := a.repo.FindUserByUsername(ctx, username); err == nil {
		return errors.New("username already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := &UserRecord{ID: fmt.Sprintf("%d", time.Now().UnixNano()), Username: username, PasswordHash: string(hash)}
	return a.repo.CreateUser(ctx, user)
}

func (a *App) Login(ctx context.Context, username, password, code string) (*types.UserResp, map[string]string, error) {
	user, err := a.repo.FindUserByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, errors.New("invalid username or password")
		}
		return nil, nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, nil, errors.New("invalid username or password")
	}
	if user.MFAEnabled && !a.verifyMFACode(user.MFASecret, code) {
		return nil, nil, errors.New("invalid mfa code")
	}
	accessToken, err := a.signToken(user.ID, a.cfg.JWT.AccessSecret, a.cfg.JWT.AccessExpireSeconds)
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := a.signToken(user.ID, a.cfg.JWT.RefreshSecret, a.cfg.JWT.RefreshExpireSeconds)
	if err != nil {
		return nil, nil, err
	}
	if err := a.repo.StoreRefreshToken(ctx, fmt.Sprintf("%d", time.Now().UnixNano()), user.ID, refreshToken, time.Now().Add(time.Duration(a.cfg.JWT.RefreshExpireSeconds)*time.Second)); err != nil {
		return nil, nil, err
	}
	return &types.UserResp{Base: successBase(), Data: toUser(user)}, map[string]string{
		a.cfg.Auth.AccessHeader:  accessToken,
		a.cfg.Auth.RefreshHeader: refreshToken,
	}, nil
}

func (a *App) ParseUserFromTokens(ctx context.Context, accessToken, refreshToken string) (AuthUser, error) {
	if accessToken == "" && refreshToken == "" {
		return AuthUser{}, errors.New("missing token")
	}
	if accessToken != "" {
		userID, err := a.parseToken(accessToken, a.cfg.JWT.AccessSecret)
		if err == nil {
			return AuthUser{UserID: userID, Tokens: map[string]string{a.cfg.Auth.AccessHeader: accessToken, a.cfg.Auth.RefreshHeader: refreshToken}}, nil
		}
	}
	if refreshToken == "" {
		return AuthUser{}, errors.New("invalid token")
	}
	userID, err := a.parseToken(refreshToken, a.cfg.JWT.RefreshSecret)
	if err != nil {
		return AuthUser{}, err
	}
	ok, err := a.repo.IsRefreshTokenValid(ctx, refreshToken)
	if err != nil {
		return AuthUser{}, err
	}
	if !ok {
		return AuthUser{}, errors.New("refresh token revoked")
	}
	newAccess, err := a.signToken(userID, a.cfg.JWT.AccessSecret, a.cfg.JWT.AccessExpireSeconds)
	if err != nil {
		return AuthUser{}, err
	}
	return AuthUser{UserID: userID, Tokens: map[string]string{a.cfg.Auth.AccessHeader: newAccess, a.cfg.Auth.RefreshHeader: refreshToken}}, nil
}

func (a *App) GetUserInfo(ctx context.Context, currentUserID, targetUserID string) (*types.UserResp, error) {
	lookup := strings.TrimSpace(targetUserID)
	if lookup == "" {
		lookup = currentUserID
	}
	user, err := a.repo.FindUserByID(ctx, lookup)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &types.UserResp{Base: successBase(), Data: toUser(user)}, nil
}

func (a *App) GenerateMFA(ctx context.Context, currentUserID string) (*types.MfaQRCodeResp, error) {
	secret := generateMFASecret()
	issuer := "Gobili"
	accountName := currentUserID
	otpauthURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, accountName, secret, issuer)
	png, err := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate qr code: %w", err)
	}
	qrcodeBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return &types.MfaQRCodeResp{Base: successBase(), Data: types.MfaQRCodeRespData{Secret: secret, Qrcode: qrcodeBase64}}, nil
}

func generateMFASecret() string {
	bytes := make([]byte, 20)
	_, _ = rand.Read(bytes)
	return base32Encode(bytes)
}

func base32Encode(data []byte) string {
	const base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	result := make([]byte, 0, (len(data)*8+4)/5)
	bits := 0
	value := 0
	for _, b := range data {
		value = (value << 8) | int(b)
		bits += 8
		for bits >= 5 {
			bits -= 5
			result = append(result, base32Chars[(value>>bits)&0x1F])
		}
	}
	if bits > 0 {
		result = append(result, base32Chars[(value<<(5-bits))&0x1F])
	}
	return string(result)
}

func (a *App) BindMFA(ctx context.Context, currentUserID, secret, code string) error {
	if !a.verifyMFACode(secret, code) {
		return errors.New("invalid mfa code")
	}
	return a.repo.EnableMFA(ctx, currentUserID, secret)
}

func (a *App) UploadAvatar(ctx context.Context, currentUserID string, file multipart.File, header *multipart.FileHeader) (*types.UserResp, error) {
	path, err := a.saveUploadedFile(file, header, a.cfg.Upload.AvatarDir)
	if err != nil {
		return nil, err
	}
	url := a.cfg.Upload.BaseURL + "/avatars/" + filepath.Base(path)
	if err := a.repo.UpdateUserAvatar(ctx, currentUserID, url); err != nil {
		return nil, err
	}
	user, err := a.repo.FindUserByID(ctx, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found, please register first")
		}
		return nil, err
	}
	return &types.UserResp{Base: successBase(), Data: toUser(user)}, nil
}

func (a *App) ImageSearch(ctx context.Context, currentUserID string, file multipart.File, header *multipart.FileHeader) (*types.ImageSearchResp, error) {
	path, err := a.saveUploadedFile(file, header, a.cfg.Upload.ImageDir)
	if err != nil {
		return nil, err
	}
	url := a.cfg.Upload.BaseURL + "/images/" + filepath.Base(path)
	_ = a.repo.CreateImageSearch(ctx, fmt.Sprintf("%d", time.Now().UnixNano()), currentUserID, url)
	return &types.ImageSearchResp{Base: successBase(), Data: url}, nil
}

func (a *App) PublishVideo(ctx context.Context, currentUserID, title, description string, file multipart.File, header *multipart.FileHeader) error {
	path, err := a.saveUploadedFile(file, header, a.cfg.Upload.VideoDir)
	if err != nil {
		return err
	}
	video := &VideoRecord{ID: fmt.Sprintf("%d", time.Now().UnixNano()), UserID: currentUserID, VideoURL: a.cfg.Upload.BaseURL + "/videos/" + filepath.Base(path), CoverURL: a.cfg.Upload.BaseURL + "/videos/cover-" + filepath.Base(path) + ".jpg", Title: title, Description: description}
	return a.repo.CreateVideo(ctx, video)
}

func (a *App) Feed(ctx context.Context, latestTime string) (*types.VideoListResp, error) {
	items, err := a.repo.ListFeedVideos(ctx, latestTime)
	if err != nil {
		return nil, err
	}
	return &types.VideoListResp{Base: successBase(), Data: types.VideoListData{Items: toVideos(items)}}, nil
}
func (a *App) VideoList(ctx context.Context, userID string, pageNum, pageSize int64) (*types.PagedVideoListResp, error) {
	items, total, err := a.repo.ListVideosByUser(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	return &types.PagedVideoListResp{Base: successBase(), Data: types.PagedVideoListData{Items: toVideos(items), Total: total}}, nil
}
func (a *App) Popular(ctx context.Context, pageNum, pageSize int64) (*types.VideoListResp, error) {
	items, err := a.repo.ListPopularVideos(ctx, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	return &types.VideoListResp{Base: successBase(), Data: types.VideoListData{Items: toVideos(items)}}, nil
}
func (a *App) SearchVideos(ctx context.Context, req *types.SearchVideoReq) (*types.PagedVideoListResp, error) {
	items, total, err := a.repo.SearchVideos(ctx, req)
	if err != nil {
		return nil, err
	}
	return &types.PagedVideoListResp{Base: successBase(), Data: types.PagedVideoListData{Items: toVideos(items), Total: total}}, nil
}
func (a *App) RecordVideoVisit(ctx context.Context, currentUserID, videoID string) error {
	_ = currentUserID
	videoID = strings.TrimSpace(videoID)
	if videoID == "" {
		return errors.New("video_id is required")
	}
	return a.repo.RecordVideoVisit(ctx, videoID)
}
func (a *App) LikeAction(ctx context.Context, currentUserID string, req *types.LikeActionReq) error {
	videoID := defaultEmpty(req.VideoId)
	commentID := defaultEmpty(req.CommentId)
	if req.ActionType == "1" {
		return a.repo.AddLike(ctx, currentUserID, videoID, commentID)
	}
	return a.repo.RemoveLike(ctx, currentUserID, videoID, commentID)
}
func (a *App) LikeList(ctx context.Context, userID string, pageNum, pageSize int64) (*types.VideoListResp, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	items, err := a.repo.ListLikedVideos(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	return &types.VideoListResp{Base: successBase(), Data: types.VideoListData{Items: toVideos(items)}}, nil
}
func (a *App) PublishComment(ctx context.Context, currentUserID string, req *types.CommentPublishReq) error {
	videoID := defaultEmpty(req.VideoId)
	parentID := defaultEmpty(req.CommentId)
	return a.repo.CreateComment(ctx, &CommentRecord{ID: fmt.Sprintf("%d", time.Now().UnixNano()), UserID: currentUserID, VideoID: videoID, ParentID: parentID, Content: strings.TrimSpace(req.Content)})
}
func (a *App) CommentList(ctx context.Context, req *types.CommentListReq) (*types.CommentListResp, error) {
	videoID := defaultEmpty(req.VideoId)
	parentID := defaultEmpty(req.CommentId)
	items, err := a.repo.ListComments(ctx, videoID, parentID, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &types.CommentListResp{Base: successBase(), Data: types.CommentListData{Items: toComments(items)}}, nil
}
func (a *App) DeleteComment(ctx context.Context, currentUserID string, commentID string) error {
	return a.repo.DeleteComment(ctx, currentUserID, commentID)
}
func (a *App) RelationAction(ctx context.Context, currentUserID string, req *types.RelationActionReq) error {
	if req.ActionType == 0 {
		return a.repo.Follow(ctx, currentUserID, req.ToUserId)
	}
	return a.repo.Unfollow(ctx, currentUserID, req.ToUserId)
}
func (a *App) FollowingList(ctx context.Context, userID string, pageNum, pageSize int64) (*types.SocialListResp, error) {
	items, total, err := a.repo.ListFollowing(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	return &types.SocialListResp{Base: successBase(), Data: types.SocialListData{Items: toSocialUsers(items), Total: total}}, nil
}
func (a *App) FollowerList(ctx context.Context, userID string, pageNum, pageSize int64) (*types.SocialListResp, error) {
	items, total, err := a.repo.ListFollowers(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	return &types.SocialListResp{Base: successBase(), Data: types.SocialListData{Items: toSocialUsers(items), Total: total}}, nil
}
func (a *App) FriendList(ctx context.Context, currentUserID string, pageNum, pageSize int64) (*types.SocialListResp, error) {
	items, total, err := a.repo.ListFriends(ctx, currentUserID, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	return &types.SocialListResp{Base: successBase(), Data: types.SocialListData{Items: toSocialUsers(items), Total: total}}, nil
}

func (a *App) signToken(userID string, secret string, expireSeconds int64) (string, error) {
	claims := jwt.MapClaims{"user_id": userID, "exp": time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (a *App) parseToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	value, ok := claims["user_id"]
	if !ok {
		return "", errors.New("missing user_id")
	}
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprint(v), nil
	}
}

func (a *App) verifyMFACode(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	secretBytes, err := base32Decode(secret)
	if err != nil {
		return false
	}
	now := time.Now().Unix() / 30
	for offset := -1; offset <= 1; offset++ {
		expectedCode := generateTOTP(secretBytes, now+int64(offset))
		if code == expectedCode {
			return true
		}
	}
	return false
}

func generateTOTP(secret []byte, timestamp int64) string {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(timestamp))
	hash := hmac.New(sha1.New, secret)
	hash.Write(data)
	h := hash.Sum(nil)
	offset := h[len(h)-1] & 0x0F
	code := (int(h[offset]&0x7F) << 24) | (int(h[offset+1]) << 16) | (int(h[offset+2]) << 8) | int(h[offset+3])
	return fmt.Sprintf("%06d", code%1000000)
}

func base32Decode(s string) ([]byte, error) {
	s = strings.ToUpper(strings.ReplaceAll(s, "=", ""))
	const base32Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	decodeMap := make(map[byte]int)
	for i := 0; i < len(base32Chars); i++ {
		decodeMap[base32Chars[i]] = i
	}
	result := make([]byte, 0, len(s)*5/8)
	bits := 0
	value := 0
	for i := 0; i < len(s); i++ {
		v, ok := decodeMap[s[i]]
		if !ok {
			return nil, fmt.Errorf("invalid base32 character: %c", s[i])
		}
		value = (value << 5) | v
		bits += 5
		for bits >= 8 {
			bits -= 8
			result = append(result, byte((value>>bits)&0xFF))
		}
	}
	return result, nil
}

func (a *App) saveUploadedFile(file multipart.File, header *multipart.FileHeader, dir string) (string, error) {
	if header == nil {
		return "", errors.New("file is required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	target := filepath.Join(dir, name)
	out, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}
	return target, nil
}

func successBase() types.Base { return types.Base{Code: 10000, Msg: "success"} }
func toUser(user *UserRecord) types.User {
	return types.User{Id: user.ID, Username: user.Username, AvatarUrl: user.AvatarURL, CreatedAt: formatTime(user.CreatedAt), UpdatedAt: formatTime(user.UpdatedAt), DeletedAt: formatNullTime(user.DeletedAt)}
}
func toVideos(items []VideoRecord) []types.Video {
	result := make([]types.Video, 0, len(items))
	for _, item := range items {
		result = append(result, types.Video{Id: item.ID, UserId: item.UserID, VideoUrl: item.VideoURL, CoverUrl: item.CoverURL, Title: item.Title, Description: item.Description, VisitCount: item.VisitCount, LikeCount: item.LikeCount, CommentCount: item.CommentCount, CreatedAt: formatTime(item.CreatedAt), UpdatedAt: formatTime(item.UpdatedAt), DeletedAt: formatNullTime(item.DeletedAt)})
	}
	return result
}
func toComments(items []CommentRecord) []types.Comment {
	result := make([]types.Comment, 0, len(items))
	for _, item := range items {
		result = append(result, types.Comment{Id: item.ID, UserId: item.UserID, VideoId: item.VideoID, ParentId: item.ParentID, LikeCount: item.LikeCount, ChildCount: item.ChildCount, Content: item.Content, CreatedAt: formatTime(item.CreatedAt), UpdatedAt: formatTime(item.UpdatedAt), DeletedAt: formatNullTime(item.DeletedAt)})
	}
	return result
}
func toSocialUsers(items []SocialRecord) []types.SocialUser {
	result := make([]types.SocialUser, 0, len(items))
	for _, item := range items {
		result = append(result, types.SocialUser{Id: item.ID, Username: item.Username, AvatarUrl: item.AvatarURL})
	}
	return result
}
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return formatTime(t.Time)
}
func defaultEmpty(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, `"`)
	return v
}
