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

// ==================== 认证服务定义 ====================

service AuthService {
    MFAQrcodeResp GetMFAQrcode(1: MFAQrcodeReq req) (api.get="/auth/mfa/qrcode")
}
