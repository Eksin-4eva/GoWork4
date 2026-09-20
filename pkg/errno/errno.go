// Package errno 提供全局错误码与统一的错误类型。
//
// 使用约定：
//   - 业务层返回本包定义的错误（可用 WithMessage / WithError 补充上下文）
//   - 网关层统一用 ConvertErr 归一后返回给客户端，绝不回显原始 error
package errno

import (
	"errors"
	"fmt"
)

// ErrNo 是带错误码的错误类型。
type ErrNo struct {
	ErrorCode int64
	ErrorMsg  string
}

func (e ErrNo) Error() string {
	return fmt.Sprintf("[%d] %s", e.ErrorCode, e.ErrorMsg)
}

// NewErrNo 创建一个错误码错误。
func NewErrNo(code int64, msg string) ErrNo {
	return ErrNo{ErrorCode: code, ErrorMsg: msg}
}

// Errorf 创建一个错误码错误，消息按模板格式化。
func Errorf(code int64, format string, args ...any) ErrNo {
	return ErrNo{ErrorCode: code, ErrorMsg: fmt.Sprintf(format, args...)}
}

// WithMessage 替换错误消息，保留错误码。
func (e ErrNo) WithMessage(msg string) ErrNo {
	e.ErrorMsg = msg
	return e
}

// WithError 在原有消息后追加原始错误信息，保留错误码。
func (e ErrNo) WithError(err error) ErrNo {
	if err != nil {
		e.ErrorMsg = e.ErrorMsg + ", " + err.Error()
	}
	return e
}

// ConvertErr 把任意 error 归一成 ErrNo。
//
// 已经是 ErrNo（含包装）的原样返回；其余一律落到 InternalError，
// 避免把 SQL、路径、字段名等实现细节透给客户端。
// 未识别的原始错误请由调用方（api/pack）记录日志。
func ConvertErr(err error) ErrNo {
	if err == nil {
		return Success
	}
	var e ErrNo
	if errors.As(err, &e) {
		return e
	}
	return InternalError
}
