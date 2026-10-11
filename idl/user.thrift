// 用户服务的 RPC 契约。
//
// 设计要点：
//   - 所有响应都带 model.BaseResp，业务错误通过它回传（而不是抛异常），
//     这样网关能拿到准确的业务码，而不是一个笼统的 RPC 错误
//   - **文件不进 RPC**：头像由网关直接传 MinIO，这里只收 object key。
//     否则一个头像要在内存里序列化两次，视频就更不用说了
namespace go user

include "rpc_model.thrift"

struct RegisterRequest {
    1: required string username
    2: required string password
}

struct RegisterResponse {
    1: required rpc_model.BaseResp base
}

struct LoginRequest {
    1: required string username
    2: required string password
    // code 为 MFA 验证码，账号未开启 MFA 时可不传
    3: optional string code
}

struct LoginResponse {
    1: required rpc_model.BaseResp base
    2: optional rpc_model.User user
    // 双令牌随响应返回，由网关写进 HTTP 响应头
    3: optional string access_token
    4: optional string refresh_token
}

struct GetUserInfoRequest {
    // target_user_id 为 0 表示查当前登录用户
    1: required i64 current_user_id
    2: optional i64 target_user_id
}

struct GetUserInfoResponse {
    1: required rpc_model.BaseResp base
    2: optional rpc_model.User user
}

struct UpdateAvatarRequest {
    1: required i64 user_id
    // avatar_key 是网关上传到对象存储后拿到的 key
    2: required string avatar_key
}

struct UpdateAvatarResponse {
    1: required rpc_model.BaseResp base
    2: optional rpc_model.User user
}

struct GetMfaQRCodeRequest {
    1: required i64 user_id
}

struct GetMfaQRCodeResponse {
    1: required rpc_model.BaseResp base
    2: optional string secret
    3: optional string qrcode
}

struct BindMfaRequest {
    1: required i64 user_id
    2: required string secret
    3: required string code
}

struct BindMfaResponse {
    1: required rpc_model.BaseResp base
}

// VerifyRefreshToken 给网关的鉴权中间件用。
//
// 拆成微服务后网关不再直连数据库，所以「刷新令牌是否已被吊销」
// 这个判断只能通过 RPC 问用户服务。
struct VerifyRefreshTokenRequest {
    1: required string token
}

struct VerifyRefreshTokenResponse {
    1: required rpc_model.BaseResp base
    2: optional bool valid
}

service UserService {
    // 用户注册
    RegisterResponse Register(1: RegisterRequest req)
    // 用户登录，返回双令牌
    LoginResponse Login(1: LoginRequest req)
    // 获取用户信息
    GetUserInfoResponse GetUserInfo(1: GetUserInfoRequest req)
    // 更新头像（key 由网关上传后传入）
    UpdateAvatarResponse UpdateAvatar(1: UpdateAvatarRequest req)
    // 获取 MFA 绑定二维码
    GetMfaQRCodeResponse GetMfaQRCode(1: GetMfaQRCodeRequest req)
    // 绑定 MFA
    BindMfaResponse BindMfa(1: BindMfaRequest req)
    // 校验刷新令牌是否仍然有效
    VerifyRefreshTokenResponse VerifyRefreshToken(1: VerifyRefreshTokenRequest req)
}
