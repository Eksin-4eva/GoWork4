namespace go api

// ==================== 基础类型定义 ====================

// 用户基础信息
struct User {
    1: i64 id
    2: string username
    3: string avatar_url
    4: i64 follower_count
    5: i64 following_count
    6: string created_at
    7: string updated_at
}

// 视频信息
struct Video {
    1: i64 id
    2: i64 user_id
    3: string title
    4: string description
    5: string video_url
    6: string cover_url
    7: i64 view_count
    8: i64 like_count
    9: i64 comment_count
    10: string created_at
    11: string updated_at
    12: User author
}

// 评论信息
struct Comment {
    1: i64 id
    2: i64 video_id
    3: i64 user_id
    4: string content
    5: i64 parent_id
    6: string created_at
    7: string updated_at
    8: User author
}

// 关注关系
struct Relation {
    1: i64 id
    2: i64 user_id
    3: i64 to_user_id
    4: i8 status
    5: string created_at
    6: string updated_at
}

// 通用响应
struct BaseResp {
    1: i32 code
    2: string message
}

// ==================== 用户模块请求/响应 ====================

struct RegisterReq {
    1: string username (api.form="username")
    2: string password (api.form="password")
}

struct RegisterResp {
    1: BaseResp base
    2: i64 user_id
    3: string token
}

struct LoginReq {
    1: string username (api.form="username")
    2: string password (api.form="password")
}

struct LoginResp {
    1: BaseResp base
    2: i64 user_id
    3: string token
    4: User user
}

struct UserInfoReq {
    1: i64 user_id (api.query="user_id")
}

struct UserInfoResp {
    1: BaseResp base
    2: User user
}

struct UploadAvatarReq {
    1: i64 user_id (api.form="user_id")
    2: binary avatar_data (api.form="avatar_data")
    3: string file_name (api.form="file_name")
}

struct UploadAvatarResp {
    1: BaseResp base
    2: string avatar_url
}

// ==================== 视频模块请求/响应 ====================

struct PublishVideoReq {
    1: i64 user_id (api.form="user_id")
    2: string title (api.form="title")
    3: string description (api.form="description")
    4: binary video_data (api.form="video_data")
    5: binary cover_data (api.form="cover_data")
    6: string video_file_name (api.form="video_file_name")
    7: string cover_file_name (api.form="cover_file_name")
}

struct PublishVideoResp {
    1: BaseResp base
    2: i64 video_id
}

struct VideoListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page (api.query="page")
    3: i32 page_size (api.query="page_size")
}

struct VideoListResp {
    1: BaseResp base
    2: list<Video> videos
    3: i32 total
    4: i32 page
    5: i32 page_size
}

struct PopularVideosReq {
    1: i32 limit (api.query="limit")
}

struct PopularVideosResp {
    1: BaseResp base
    2: list<Video> videos
}

struct SearchVideoReq {
    1: string keyword (api.form="keyword")
    2: i32 page (api.form="page")
    3: i32 page_size (api.form="page_size")
}

struct SearchVideoResp {
    1: BaseResp base
    2: list<Video> videos
    3: i32 total
    4: i32 page
    5: i32 page_size
}

// ==================== 互动模块请求/响应 ====================

struct LikeActionReq {
    1: i64 user_id (api.form="user_id")
    2: i64 video_id (api.form="video_id")
    3: i8 action_type (api.form="action_type")
}

struct LikeActionResp {
    1: BaseResp base
}

struct LikeListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page (api.query="page")
    3: i32 page_size (api.query="page_size")
}

struct LikeListResp {
    1: BaseResp base
    2: list<Video> videos
    3: i32 total
}

struct PublishCommentReq {
    1: i64 user_id (api.form="user_id")
    2: i64 video_id (api.form="video_id")
    3: string content (api.form="content")
    4: optional i64 parent_id (api.form="parent_id")
}

struct PublishCommentResp {
    1: BaseResp base
    2: Comment comment
}

struct CommentListReq {
    1: i64 video_id (api.query="video_id")
    2: i32 page (api.query="page")
    3: i32 page_size (api.query="page_size")
}

struct CommentListResp {
    1: BaseResp base
    2: list<Comment> comments
    3: i32 total
}

struct DeleteCommentReq {
    1: i64 user_id (api.query="user_id")
    2: i64 comment_id (api.query="comment_id")
}

struct DeleteCommentResp {
    1: BaseResp base
}

// ==================== 社交模块请求/响应 ====================

struct RelationActionReq {
    1: i64 user_id (api.form="user_id")
    2: i64 to_user_id (api.form="to_user_id")
    3: i8 action_type (api.form="action_type")
}

struct RelationActionResp {
    1: BaseResp base
}

struct FollowingListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page (api.query="page")
    3: i32 page_size (api.query="page_size")
}

struct FollowingListResp {
    1: BaseResp base
    2: list<User> users
    3: i32 total
}

struct FollowerListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page (api.query="page")
    3: i32 page_size (api.query="page_size")
}

struct FollowerListResp {
    1: BaseResp base
    2: list<User> users
    3: i32 total
}

struct FriendsListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page (api.query="page")
    3: i32 page_size (api.query="page_size")
}

struct FriendsListResp {
    1: BaseResp base
    2: list<User> users
    3: i32 total
}

// ==================== 服务定义 ====================

service BiliGoService {
    // 用户模块
    RegisterResp Register(1: RegisterReq req) (api.post="/user/register")
    LoginResp Login(1: LoginReq req) (api.post="/user/login")
    UserInfoResp GetUserInfo(1: UserInfoReq req) (api.get="/user/info")
    UploadAvatarResp UploadAvatar(1: UploadAvatarReq req) (api.put="/user/avatar/upload")

    // 视频模块
    PublishVideoResp PublishVideo(1: PublishVideoReq req) (api.post="/video/publish")
    VideoListResp GetVideoList(1: VideoListReq req) (api.get="/video/list")
    PopularVideosResp GetPopularVideos(1: PopularVideosReq req) (api.get="/video/popular")
    SearchVideoResp SearchVideo(1: SearchVideoReq req) (api.post="/video/search")

    // 互动模块
    LikeActionResp LikeAction(1: LikeActionReq req) (api.post="/like/action")
    LikeListResp GetLikeList(1: LikeListReq req) (api.get="/like/list")
    PublishCommentResp PublishComment(1: PublishCommentReq req) (api.post="/comment/publish")
    CommentListResp GetCommentList(1: CommentListReq req) (api.get="/comment/list")
    DeleteCommentResp DeleteComment(1: DeleteCommentReq req) (api.delete="/comment/delete")

    // 社交模块
    RelationActionResp RelationAction(1: RelationActionReq req) (api.post="/relation/action")
    FollowingListResp GetFollowingList(1: FollowingListReq req) (api.get="/following/list")
    FollowerListResp GetFollowerList(1: FollowerListReq req) (api.get="/follower/list")
    FriendsListResp GetFriendsList(1: FriendsListReq req) (api.get="/friends/list")
}
