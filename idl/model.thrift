namespace go model

// BaseResp 是统一响应信封的模型定义。
// 注意：api.thrift 的响应体只描述 data 部分，信封由 api/pack 统一拼装，
// 这里保留 BaseResp 仅用于声明信封形态，供后续 RPC 化时复用。
struct BaseResp {
    1: required string code
    2: required string message
}

struct User {
    1: required string id
    2: required string username
    3: required string avatar_url
    4: required string created_at
    5: required string updated_at
}

struct Video {
    1: required string id
    2: required string user_id
    3: required string video_url
    4: required string cover_url
    5: required string title
    6: required string description
    7: required i64 visit_count
    8: required i64 comment_count
    9: required string created_at
    10: required string updated_at
}

struct Comment {
    1: required string id
    2: required string user_id
    3: required string video_id
    4: required string parent_id
    5: required i64 child_count
    6: required string content
    7: required string created_at
    8: required string updated_at
}
