namespace go api.interact

include "../common.thrift"

struct LikeActionReq {
    1: optional i64 video_id (api.json="video_id")
    2: optional i64 comment_id (api.json="comment_id")
    3: i8 action_type (api.json="action_type")
}

struct LikeActionResp {
    1: common.BaseResp base
}

struct LikeListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct LikeListResp {
    1: common.BaseResp base
    2: list<common.Video> items
    3: i64 total
}

struct PublishCommentReq {
    1: optional i64 video_id (api.json="video_id")
    2: optional i64 comment_id (api.json="comment_id")
    3: string content (api.json="content")
}

struct PublishCommentResp {
    1: common.BaseResp base
    2: common.Comment comment
}

struct CommentListReq {
    1: optional i64 video_id (api.query="video_id")
    2: optional i64 comment_id (api.query="comment_id")
    3: i32 page_num (api.query="page_num")
    4: i32 page_size (api.query="page_size")
}

struct CommentListResp {
    1: common.BaseResp base
    2: list<common.Comment> items
    3: i64 total
}

struct DeleteCommentReq {
    1: i64 comment_id (api.query="comment_id")
}

struct DeleteCommentResp {
    1: common.BaseResp base
}

service InteractService {
    LikeActionResp LikeAction(1: LikeActionReq req) (api.post="/like/action")
    LikeListResp GetLikeList(1: LikeListReq req) (api.get="/like/list")
    PublishCommentResp PublishComment(1: PublishCommentReq req) (api.post="/comment/publish")
    CommentListResp GetCommentList(1: CommentListReq req) (api.get="/comment/list")
    DeleteCommentResp DeleteComment(1: DeleteCommentReq req) (api.delete="/comment/delete")
}
