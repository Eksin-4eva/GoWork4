namespace go api

include "model.thrift"

## ----------------------------------------------------------------------------
## 用户 / MFA
## ----------------------------------------------------------------------------

struct RegisterRequest {
    1: required string username
    2: required string password
}

struct RegisterResponse {
}

struct LoginRequest {
    1: required string username
    2: required string password
    // code 为 MFA 验证码，账号未开启 MFA 时可不传
    3: optional string code
}

struct LoginResponse {
    1: optional model.User data
}

struct GetUserInfoRequest {
    // user_id 为空时查询当前登录用户
    1: optional string user_id
}

struct GetUserInfoResponse {
    1: optional model.User data
}

// UploadAvatar 与 PublishVideo 为 multipart/form-data 接口，
// 文件通过 c.FormFile("file") 读取，因此请求体本身没有字段。
// 注意：hz 生成 model 后需要手工调整，详见 Makefile 注释。
struct UploadAvatarRequest {
}

struct UploadAvatarResponse {
    1: optional model.User data
}

struct GetMfaQRCodeRequest {
}

struct GetMfaQRCodeResponse {
    1: required string secret
    2: required string qrcode
}

struct BindMfaRequest {
    1: required string code
    2: required string secret
}

struct BindMfaResponse {
}

service UserService {
    // 用户注册
    RegisterResponse Register(1: RegisterRequest request)(api.post="/user/register")
    // 用户登录，成功后 Access-Token / Refresh-Token 通过响应头下发
    LoginResponse Login(1: LoginRequest request)(api.post="/user/login")
    // 获取用户信息
    GetUserInfoResponse GetUserInfo(1: GetUserInfoRequest request)(api.get="/user/info")
    // 上传头像
    UploadAvatarResponse UploadAvatar(1: UploadAvatarRequest request)(api.put="/user/avatar/upload")
    // 获取 MFA 绑定二维码
    GetMfaQRCodeResponse GetMfaQRCode(1: GetMfaQRCodeRequest request)(api.get="/auth/mfa/qrcode")
    // 绑定 MFA
    BindMfaResponse BindMfa(1: BindMfaRequest request)(api.post="/auth/mfa/bind")
}

## ----------------------------------------------------------------------------
## 视频
## ----------------------------------------------------------------------------

struct FeedRequest {
    // latest_time 为毫秒时间戳，返回该时间之后的最新视频
    1: optional string latest_time
}

struct FeedResponse {
    1: optional list<model.Video> items
}

struct PublishVideoRequest {
}

struct PublishVideoResponse {
}

struct GetVideoListRequest {
    1: required string user_id
    2: optional i64 page_num
    3: optional i64 page_size
}

struct GetVideoListResponse {
    1: optional list<model.Video> items
    2: optional i64 total
}

struct GetPopularVideosRequest {
    1: optional i64 page_num
    2: optional i64 page_size
}

struct GetPopularVideosResponse {
    1: optional list<model.Video> items
    2: optional i64 total
}

struct SearchVideosRequest {
    1: optional string keywords
    // from_date / to_date 为毫秒时间戳，用于限定投稿时间区间
    2: optional i64 from_date
    3: optional i64 to_date
    4: optional string username
    5: optional i64 page_num
    6: optional i64 page_size
}

struct SearchVideosResponse {
    1: optional list<model.Video> items
    2: optional i64 total
}

struct VisitVideoRequest {
    1: required string video_id
}

struct VisitVideoResponse {
}

service VideoService {
    // 获取最新视频流
    FeedResponse GetFeed(1: FeedRequest request)(api.get="/video/feed")
    // 投稿视频
    PublishVideoResponse PublishVideo(1: PublishVideoRequest request)(api.post="/video/publish")
    // 获取指定用户的投稿列表
    GetVideoListResponse GetVideoList(1: GetVideoListRequest request)(api.get="/video/list")
    // 热门排行榜
    GetPopularVideosResponse GetPopularVideos(1: GetPopularVideosRequest request)(api.get="/video/popular")
    // 多条件搜索视频
    SearchVideosResponse SearchVideos(1: SearchVideosRequest request)(api.post="/video/search")
    // 记录视频访问
    VisitVideoResponse VisitVideo(1: VisitVideoRequest request)(api.post="/video/visit")
}

## ----------------------------------------------------------------------------
## 评论
## ----------------------------------------------------------------------------

struct PublishCommentRequest {
    1: required string video_id
    // comment_id 非空时表示回复该评论
    2: optional string comment_id
    3: required string content
}

struct PublishCommentResponse {
}

struct GetCommentListRequest {
    1: required string video_id
    // comment_id 非空时查询该评论下的子评论
    2: optional string comment_id
    3: optional i64 page_num
    4: optional i64 page_size
}

struct GetCommentListResponse {
    1: optional list<model.Comment> items
    2: optional i64 total
}

struct DeleteCommentRequest {
    1: required string comment_id
}

struct DeleteCommentResponse {
}

service CommentService {
    // 发表评论或回复评论
    PublishCommentResponse PublishComment(1: PublishCommentRequest request)(api.post="/comment/publish")
    // 获取评论列表
    GetCommentListResponse GetCommentList(1: GetCommentListRequest request)(api.get="/comment/list")
    // 删除自己的评论
    DeleteCommentResponse DeleteComment(1: DeleteCommentRequest request)(api.delete="/comment/delete")
}
