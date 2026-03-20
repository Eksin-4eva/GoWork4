namespace go api.auth

include "../common.thrift"

// ==================== MFA 模块请求/响应 ====================

struct MFAQrcodeReq {
}

struct MFAQrcodeResp {
    1: common.BaseResp base
    2: string secret
    3: string qrcode
}

struct BindMFAReq {
    1: string code (api.form="code")
    2: string secret (api.form="secret")
}

struct BindMFAResp {
    1: common.BaseResp base
}

// ==================== 认证服务定义 ====================

service AuthService {
    MFAQrcodeResp GetMFAQrcode(1: MFAQrcodeReq req) (api.get="/auth/mfa/qrcode")
    BindMFAResp BindMFA(1: BindMFAReq req) (api.post="/auth/mfa/bind")
}
