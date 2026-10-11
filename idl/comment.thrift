// 评论服务的 RPC 契约。
namespace go comment

include "rpc_model.thrift"

struct PublishCommentRequest {
    1: required i64 user_id
    2: required i64 video_id
    // parent_id 为 0 表示顶层评论
    3: optional i64 parent_id
    4: required string content
}

struct PublishCommentResponse {
    1: required rpc_model.BaseResp base
}

struct GetCommentListRequest {
    1: required i64 video_id
    // parent_id 非 0 表示查询该评论下的回复
    2: optional i64 parent_id
    3: optional i64 page_num
    4: optional i64 page_size
}

struct GetCommentListResponse {
    1: required rpc_model.BaseResp base
    2: optional list<rpc_model.Comment> items
    3: optional i64 total
}

struct DeleteCommentRequest {
    1: required i64 user_id
    2: required i64 comment_id
}

struct DeleteCommentResponse {
    1: required rpc_model.BaseResp base
}

service CommentService {
    // 发表评论或回复
    PublishCommentResponse PublishComment(1: PublishCommentRequest req)
    // 获取评论列表
    GetCommentListResponse GetCommentList(1: GetCommentListRequest req)
    // 删除自己的评论
    DeleteCommentResponse DeleteComment(1: DeleteCommentRequest req)
}
