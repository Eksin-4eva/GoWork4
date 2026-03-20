package response

const (
	CodeSuccess       = 10000
	CodeBadRequest    = 10001
	CodeUnauthorized  = 10002
	CodeForbidden     = 10003
	CodeNotFound      = 10004
	CodeInternalError = 10005
	CodeTokenExpired  = 401001
	CodeTokenInvalid  = 401002
	CodeAccessExpired = 401001
)

type BaseResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Response struct {
	Base BaseResp    `json:"base"`
	Data interface{} `json:"data,omitempty"`
}

func Success(data interface{}) *Response {
	return &Response{
		Base: BaseResp{
			Code: CodeSuccess,
			Msg:  "Success",
		},
		Data: data,
	}
}

func Fail(code int, msg string) *Response {
	return &Response{
		Base: BaseResp{
			Code: code,
			Msg:  msg,
		},
	}
}

func SuccessWithMsg(msg string, data interface{}) *Response {
	return &Response{
		Base: BaseResp{
			Code: CodeSuccess,
			Msg:  msg,
		},
		Data: data,
	}
}
