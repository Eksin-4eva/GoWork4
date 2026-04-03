package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gobili/gateway/internal/types"
)

func (r *Repository) CreateUser(ctx context.Context, user *UserRecord) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, username, password_hash, avatar_url) VALUES (?, ?, ?, ?)`, user.ID, user.Username, user.PasswordHash, user.AvatarURL)
	return err
}

func (r *Repository) FindUserByUsername(ctx context.Context, username string) (*UserRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, username, password_hash, avatar_url, mfa_secret, mfa_enabled, created_at, updated_at, deleted_at FROM users WHERE username = ? AND deleted_at IS NULL LIMIT 1`, username)
	return scanUser(row)
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (*UserRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, username, password_hash, avatar_url, mfa_secret, mfa_enabled, created_at, updated_at, deleted_at FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1`, id)
	return scanUser(row)
}

func (r *Repository) UpdateUserAvatar(ctx context.Context, id string, avatarURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET avatar_url = ? WHERE id = ?`, avatarURL, id)
	return err
}

func (r *Repository) EnableMFA(ctx context.Context, id string, secret string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET mfa_secret = ?, mfa_enabled = 1 WHERE id = ?`, secret, id)
	return err
}

func (r *Repository) StoreRefreshToken(ctx context.Context, id, userID string, token string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO refresh_tokens (id, user_id, token, expires_at) VALUES (?, ?, ?, ?)`, id, userID, token, expiresAt)
	return err
}

func (r *Repository) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM refresh_tokens WHERE token = ? AND revoked = 0 AND deleted_at IS NULL AND expires_at > NOW()`, token).Scan(&count)
	return count > 0, err
}

func (r *Repository) CreateVideo(ctx context.Context, video *VideoRecord) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO videos (id, user_id, video_url, cover_url, title, description) VALUES (?, ?, ?, ?, ?, ?)`, video.ID, video.UserID, video.VideoURL, video.CoverURL, video.Title, video.Description)
	return err
}

func (r *Repository) ListFeedVideos(ctx context.Context, latestTime string) ([]VideoRecord, error) {
	query := `SELECT id, user_id, video_url, cover_url, title, description, visit_count, like_count, comment_count, created_at, updated_at, deleted_at FROM videos WHERE deleted_at IS NULL`
	args := make([]any, 0)
	if strings.TrimSpace(latestTime) != "" {
		query += ` AND UNIX_TIMESTAMP(created_at) * 1000 >= ?`
		args = append(args, latestTime)
	}
	query += ` ORDER BY created_at DESC LIMIT 30`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVideos(rows)
}

func (r *Repository) ListVideosByUser(ctx context.Context, userID string, pageNum, pageSize int64) ([]VideoRecord, int64, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, video_url, cover_url, title, description, visit_count, like_count, comment_count, created_at, updated_at, deleted_at FROM videos WHERE user_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanVideos(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM videos WHERE user_id = ? AND deleted_at IS NULL`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) ListPopularVideos(ctx context.Context, pageNum, pageSize int64) ([]VideoRecord, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, video_url, cover_url, title, description, visit_count, like_count, comment_count, created_at, updated_at, deleted_at FROM videos WHERE deleted_at IS NULL ORDER BY visit_count DESC, created_at DESC LIMIT ? OFFSET ?`, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVideos(rows)
}

func (r *Repository) SearchVideos(ctx context.Context, req *types.SearchVideoReq) ([]VideoRecord, int64, error) {
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	where := []string{"v.deleted_at IS NULL"}
	args := make([]any, 0)
	joinUsers := false
	if strings.TrimSpace(req.Keywords) != "" {
		where = append(where, `(v.title LIKE ? OR v.description LIKE ?)`)
		kw := "%" + strings.TrimSpace(req.Keywords) + "%"
		args = append(args, kw, kw)
	}
	if req.FromDate > 0 {
		where = append(where, `v.created_at >= FROM_UNIXTIME(? / 1000)`)
		args = append(args, req.FromDate)
	}
	if req.ToDate > 0 {
		where = append(where, `v.created_at <= FROM_UNIXTIME(? / 1000)`)
		args = append(args, req.ToDate)
	}
	if strings.TrimSpace(req.Username) != "" {
		joinUsers = true
		where = append(where, `u.username LIKE ?`)
		args = append(args, "%"+strings.TrimSpace(req.Username)+"%")
	}
	base := ` FROM videos v `
	if joinUsers {
		base += `JOIN users u ON u.id = v.user_id `
	}
	condition := ` WHERE ` + strings.Join(where, ` AND `)
	query := `SELECT v.id, v.user_id, v.video_url, v.cover_url, v.title, v.description, v.visit_count, v.like_count, v.comment_count, v.created_at, v.updated_at, v.deleted_at` + base + condition + ` ORDER BY v.created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, append(args, req.PageSize, (req.PageNum-1)*req.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanVideos(rows)
	if err != nil {
		return nil, 0, err
	}
	countQuery := `SELECT COUNT(1)` + base + condition
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) AddLike(ctx context.Context, userID, videoID, commentID string) error {
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := r.db.ExecContext(ctx, `INSERT INTO likes (id, user_id, video_id, comment_id) VALUES (?, ?, ?, ?)`, id, userID, videoID, commentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil
		}
		return err
	}
	if videoID != "" {
		_, err = r.db.ExecContext(ctx, `UPDATE videos SET like_count = like_count + 1 WHERE id = ?`, videoID)
		return err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE comments SET like_count = like_count + 1 WHERE id = ?`, commentID)
	return err
}

func (r *Repository) RemoveLike(ctx context.Context, userID, videoID, commentID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM likes WHERE user_id = ? AND video_id = ? AND comment_id = ?`, userID, videoID, commentID)
	if err != nil {
		return err
	}
	if videoID != "" {
		_, err = r.db.ExecContext(ctx, `UPDATE videos SET like_count = IF(like_count > 0, like_count - 1, 0) WHERE id = ?`, videoID)
		return err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE comments SET like_count = IF(like_count > 0, like_count - 1, 0) WHERE id = ?`, commentID)
	return err
}

func (r *Repository) ListLikedVideos(ctx context.Context, userID string, pageNum, pageSize int64) ([]VideoRecord, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	rows, err := r.db.QueryContext(ctx, `SELECT v.id, v.user_id, v.video_url, v.cover_url, v.title, v.description, v.visit_count, v.like_count, v.comment_count, v.created_at, v.updated_at, v.deleted_at FROM likes l JOIN videos v ON v.id = l.video_id WHERE l.user_id = ? AND l.video_id != '' AND v.deleted_at IS NULL ORDER BY l.created_at DESC LIMIT ? OFFSET ?`, userID, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVideos(rows)
}

func (r *Repository) CreateComment(ctx context.Context, comment *CommentRecord) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO comments (id, user_id, video_id, parent_id, content) VALUES (?, ?, ?, ?, ?)`, comment.ID, comment.UserID, comment.VideoID, comment.ParentID, comment.Content)
	if err != nil {
		return err
	}
	_, _ = r.db.ExecContext(ctx, `UPDATE videos SET comment_count = comment_count + 1 WHERE id = ?`, comment.VideoID)
	if comment.ParentID != "" {
		_, _ = r.db.ExecContext(ctx, `UPDATE comments SET child_count = child_count + 1 WHERE id = ?`, comment.ParentID)
	}
	return nil
}

func (r *Repository) ListComments(ctx context.Context, videoID, parentID string, pageNum, pageSize int64) ([]CommentRecord, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	query := `SELECT id, user_id, video_id, parent_id, like_count, child_count, content, created_at, updated_at, deleted_at FROM comments WHERE deleted_at IS NULL`
	args := make([]any, 0)
	if videoID != "" {
		query += ` AND video_id = ?`
		args = append(args, videoID)
	}
	query += ` AND parent_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, parentID, pageSize, (pageNum-1)*pageSize)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanComments(rows)
}

func (r *Repository) DeleteComment(ctx context.Context, userID, commentID string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE comments SET deleted_at = NOW() WHERE id = ? AND user_id = ? AND deleted_at IS NULL`, commentID, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("comment not found")
	}
	return nil
}

func (r *Repository) Follow(ctx context.Context, userID, toUserID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO relations (id, user_id, to_user_id) VALUES (?, ?, ?)`, fmt.Sprintf("%d", time.Now().UnixNano()), userID, toUserID)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return err
	}
	return nil
}

func (r *Repository) Unfollow(ctx context.Context, userID, toUserID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM relations WHERE user_id = ? AND to_user_id = ?`, userID, toUserID)
	return err
}

func (r *Repository) ListFollowing(ctx context.Context, userID string, pageNum, pageSize int64) ([]SocialRecord, int64, error) {
	return r.listSocial(ctx, `JOIN relations rel ON rel.to_user_id = u.id WHERE rel.user_id = ?`, userID, pageNum, pageSize)
}

func (r *Repository) ListFollowers(ctx context.Context, userID string, pageNum, pageSize int64) ([]SocialRecord, int64, error) {
	return r.listSocial(ctx, `JOIN relations rel ON rel.user_id = u.id WHERE rel.to_user_id = ?`, userID, pageNum, pageSize)
}

func (r *Repository) ListFriends(ctx context.Context, userID string, pageNum, pageSize int64) ([]SocialRecord, int64, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	query := `SELECT u.id, u.username, u.avatar_url FROM users u JOIN relations r1 ON r1.to_user_id = u.id AND r1.user_id = ? JOIN relations r2 ON r2.user_id = u.id AND r2.to_user_id = ? WHERE u.deleted_at IS NULL LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, userID, userID, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanSocial(rows)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	countQuery := `SELECT COUNT(1) FROM users u JOIN relations r1 ON r1.to_user_id = u.id AND r1.user_id = ? JOIN relations r2 ON r2.user_id = u.id AND r2.to_user_id = ? WHERE u.deleted_at IS NULL`
	if err := r.db.QueryRowContext(ctx, countQuery, userID, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) CreateImageSearch(ctx context.Context, id, userID string, imageURL string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO image_searches (id, user_id, image_url) VALUES (?, ?, ?)`, id, userID, imageURL)
	return err
}

func (r *Repository) listSocial(ctx context.Context, join string, userID string, pageNum, pageSize int64) ([]SocialRecord, int64, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	query := `SELECT u.id, u.username, u.avatar_url FROM users u ` + join + ` AND u.deleted_at IS NULL LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanSocial(rows)
	if err != nil {
		return nil, 0, err
	}
	countQuery := `SELECT COUNT(1) FROM users u ` + join + ` AND u.deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func scanUser(row *sql.Row) (*UserRecord, error) {
	var user UserRecord
	var enabled int
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.AvatarURL, &user.MFASecret, &enabled, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	user.MFAEnabled = enabled == 1
	return &user, err
}

func scanVideos(rows *sql.Rows) ([]VideoRecord, error) {
	items := make([]VideoRecord, 0)
	for rows.Next() {
		var item VideoRecord
		if err := rows.Scan(&item.ID, &item.UserID, &item.VideoURL, &item.CoverURL, &item.Title, &item.Description, &item.VisitCount, &item.LikeCount, &item.CommentCount, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanComments(rows *sql.Rows) ([]CommentRecord, error) {
	items := make([]CommentRecord, 0)
	for rows.Next() {
		var item CommentRecord
		if err := rows.Scan(&item.ID, &item.UserID, &item.VideoID, &item.ParentID, &item.LikeCount, &item.ChildCount, &item.Content, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanSocial(rows *sql.Rows) ([]SocialRecord, error) {
	items := make([]SocialRecord, 0)
	for rows.Next() {
		var item SocialRecord
		if err := rows.Scan(&item.ID, &item.Username, &item.AvatarURL); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
