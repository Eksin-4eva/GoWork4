package errno

import (
	"errors"
	"fmt"
	"testing"
)

func TestConvertErrNilReturnsSuccess(t *testing.T) {
	if got := ConvertErr(nil); got != Success {
		t.Fatalf("ConvertErr(nil) = %v, want %v", got, Success)
	}
}

func TestConvertErrKeepsErrNo(t *testing.T) {
	err := BizNotExist.WithMessage("视频不存在")
	if got := ConvertErr(err); got != err {
		t.Fatalf("ConvertErr kept %v, want %v", got, err)
	}
}

func TestConvertErrUnwrapsWrappedErrNo(t *testing.T) {
	wrapped := fmt.Errorf("query user: %w", BizNotExist)
	if got := ConvertErr(wrapped); got.ErrorCode != BizErrorCode {
		t.Fatalf("ConvertErr code = %d, want %d", got.ErrorCode, BizErrorCode)
	}
}

func TestConvertErrHidesUnknownError(t *testing.T) {
	got := ConvertErr(errors.New("Error 1062: Duplicate entry for key 'users.username'"))

	if got.ErrorCode != InternalErrorCode {
		t.Fatalf("code = %d, want %d", got.ErrorCode, InternalErrorCode)
	}
	// 原始错误内容绝不能透给客户端
	if got.ErrorMsg != InternalError.ErrorMsg {
		t.Fatalf("message = %q, want %q", got.ErrorMsg, InternalError.ErrorMsg)
	}
}

func TestWithErrorAppendsCause(t *testing.T) {
	got := ParamError.WithError(errors.New("missing username"))
	if got.ErrorCode != ParamErrorCode {
		t.Fatalf("code = %d, want %d", got.ErrorCode, ParamErrorCode)
	}
	if got.ErrorMsg != "参数错误, missing username" {
		t.Fatalf("message = %q", got.ErrorMsg)
	}
}

func TestErrorFormat(t *testing.T) {
	if got := BizForbidden.Error(); got != "[40001] 没有操作权限" {
		t.Fatalf("Error() = %q", got)
	}
}

// 同一大类下的细分场景必须复用错误码，只靠文案区分。
// 这条约束保证错误码总量不会随业务增长而膨胀。
func TestCodesAreReusedWithinCategory(t *testing.T) {
	cases := []struct {
		name string
		errs []ErrNo
		code int64
	}{
		{"参数类", []ErrNo{ParamError, ParamEmpty}, ParamErrorCode},
		{"鉴权类", []ErrNo{AuthMissing, AuthInvalid}, AuthErrorCode},
		{"业务类", []ErrNo{BizError, BizNotExist, BizForbidden, BizDuplicated}, BizErrorCode},
	}
	for _, c := range cases {
		for _, e := range c.errs {
			if e.ErrorCode != c.code {
				t.Errorf("%s: %q 的码为 %d, 期望 %d", c.name, e.ErrorMsg, e.ErrorCode, c.code)
			}
		}
	}
}
