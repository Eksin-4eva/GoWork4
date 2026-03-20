namespace go api.interact

include "../common.thrift"

// ==================== 互动模块请求/响应 ====================

// 点赞操作
struct LikeActionReq {
    1: i64 user_id (api.json="user_id")
    2: i64 video_id (api.json="video_id")
    3: i8 action_type (api.json="action_type") // 1=点赞, 2=取消
}

struct LikeActionResp {
    1: common.BaseResp base
}

// 点赞列表
struct LikeListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct LikeListResp {
    1: common.BaseResp base
    2: list<common.Video> videos
    3: i64 total
}

// 发布评论
struct PublishCommentReq {
    1: i64 user_id (api.json="user_id")
    2: i64 video_id (api.json="video_id")
    3: string content (api.json="content")
    4: optional i64 parent_id (api.json="parent_id")
}

struct PublishCommentResp {
    1: common.BaseResp base
    2: common.Comment comment
}

// 评论列表
struct CommentListReq {
    1: i64 video_id (api.query="video_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct CommentListResp {
    1: common.BaseResp base
    2: list<common.Comment> comments
    3: i64 total
}

// 删除评论
struct DeleteCommentReq {
    1: i64 user_id (api.query="user_id")
    2: i64 comment_id (api.query="comment_id")
}

struct DeleteCommentResp {
    1: common.BaseResp base
}

// ==================== 互动服务定义 ====================

service InteractService {
    LikeActionResp LikeAction(1: LikeActionReq req) (api.post="/like/action")
    LikeListResp GetLikeList(1: LikeListReq req) (api.get="/like/list")
    PublishCommentResp PublishComment(1: PublishCommentReq req) (api.post="/comment/publish")
    CommentListResp GetCommentList(1: CommentListReq req) (api.get="/comment/list")
    DeleteCommentResp DeleteComment(1: DeleteCommentReq req) (api.delete="/comment/delete")
}
