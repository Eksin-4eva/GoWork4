package errno

// 预定义错误。
//
// 约定：错误码只表达大类，细分场景复用同一个码、用文案区分。
// 新增场景优先复用现有条目，确实无法表达时再用 WithMessage 覆盖文案，
// 不要轻易新增错误码。
var (
	Success = NewErrNo(SuccessCode, SuccessMsg)

	// 参数类
	ParamError = NewErrNo(ParamErrorCode, "参数错误")
	ParamEmpty = NewErrNo(ParamErrorCode, "参数不能为空")

	// 鉴权类。两种情况对客户端都意味着「请重新登录」：
	// 令牌缺失，以及令牌过期/非法/被吊销。
	AuthMissing = NewErrNo(AuthErrorCode, "未携带鉴权信息")
	AuthInvalid = NewErrNo(AuthErrorCode, "登录状态无效，请重新登录")

	// 业务类
	BizError      = NewErrNo(BizErrorCode, "请求处理失败")
	BizNotExist   = NewErrNo(BizErrorCode, "资源不存在")
	BizForbidden  = NewErrNo(BizErrorCode, "没有操作权限")
	BizDuplicated = NewErrNo(BizErrorCode, "资源已存在")

	// 内部类。具体是数据库、缓存还是存储出了问题，记在日志里，
	// 不回传给客户端。
	InternalError = NewErrNo(InternalErrorCode, "内部服务错误")
)
