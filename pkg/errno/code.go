package errno

// 错误码只按大类分段，不按场景细分。
//
//	0     成功
//	200xx 参数错误 (Param)
//	300xx 鉴权错误 (Auth)
//	400xx 业务错误 (Biz)
//	500xx 内部错误 (Internal)
//
// 同一大类下的具体场景（视频不存在 / 用户名已占用 / 验证码错误…）
// 复用同一个码，靠文案区分，见 default.go。
const (
	SuccessCode = 0
	SuccessMsg  = "Success"

	ParamErrorCode    int64 = 20001
	AuthErrorCode     int64 = 30001
	BizErrorCode      int64 = 40001
	InternalErrorCode int64 = 50001
)
