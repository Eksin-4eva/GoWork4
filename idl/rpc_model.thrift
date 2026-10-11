// RPC 层专用的数据模型。
//
// 为什么不复用 model.thrift（HTTP 用的那套）：
//
//  1. **HTTP 要 URL，RPC 要 key**。HTTP 响应里客户端需要能直接打开的地址，
//     而 RPC 之间传 object key 就够了——「怎么拼成对外可访问的 URL」
//     是网关的职责（取决于客户端从哪个域名访问），下游服务不该知道。
//  2. **ID 类型不同**。HTTP 层用 string（避免 JS 精度丢失），
//     RPC 层用 i64（不做无意义的字符串转换，也省带宽）。
//
// 两套模型各管一段，转换只发生在网关的 api/pack 里。
namespace go rpcmodel

struct BaseResp {
    1: required string code
    2: required string message
}

struct User {
    1: required i64 id
    2: required string username
    // avatar_key 是对象存储的 key，不是完整 URL
    3: required string avatar_key
    4: required string created_at
    5: required string updated_at
}

struct Video {
    1: required i64 id
    2: required i64 user_id
    3: required string video_key
    4: required string cover_key
    5: required string title
    6: required string description
    7: required i64 visit_count
    8: required i64 comment_count
    9: required string created_at
    10: required string updated_at
}

struct Comment {
    1: required i64 id
    2: required i64 user_id
    3: required i64 video_id
    4: required i64 parent_id
    5: required i64 child_count
    6: required string content
    7: required string created_at
    8: required string updated_at
}
