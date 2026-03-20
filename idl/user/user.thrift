namespace go api.user

include "../common.thrift"

// ==================== 用户模块请求/响应 ====================

struct RegisterReq {
    1: string username (api.form="username")
    2: string password (api.form="password")
}

struct RegisterResp {
    1: common.BaseResp base
    2: i64 user_id
    3: string access_token
    4: string refresh_token
}

struct LoginReq {
    1: string username (api.form="username")
    2: string password (api.form="password")
}

struct LoginResp {
    1: common.BaseResp base
    2: i64 user_id
    3: string access_token
    4: string refresh_token
    5: common.User user
}

struct UserInfoReq {
    1: i64 user_id (api.query="user_id")
}

struct UserInfoResp {
    1: common.BaseResp base
    2: common.User user
}

struct UploadAvatarReq {
    1: i64 user_id (api.form="user_id")
    2: binary avatar_data (api.form="avatar_data")
    3: string file_name (api.form="file_name")
}

struct UploadAvatarResp {
    1: common.BaseResp base
    2: string avatar_url
}

struct RefreshTokenReq {
    1: string refresh_token (api.header="X-Refresh-Token")
}

struct RefreshTokenResp {
    1: common.BaseResp base
    2: i64 user_id
    3: string access_token
    4: string refresh_token
}

// ==================== 用户服务定义 ====================

service UserService {
    RegisterResp Register(1: RegisterReq req) (api.post="/user/register")
    LoginResp Login(1: LoginReq req) (api.post="/user/login")
    UserInfoResp GetUserInfo(1: UserInfoReq req) (api.get="/user/info")
    UploadAvatarResp UploadAvatar(1: UploadAvatarReq req) (api.put="/user/avatar/upload")
    RefreshTokenResp RefreshToken(1: RefreshTokenReq req) (api.post="/user/refresh")
}
