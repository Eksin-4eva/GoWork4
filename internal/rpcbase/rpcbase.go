// Package rpcbase 封装 RPC 响应的统一格式与错误传递约定。
//
// **为什么业务错误要通过 BaseResp 回传，而不是直接返回 Go error？**
//
// RPC 框架会把服务端返回的 error 包成它自己的错误类型，
// 网关拿到的只是一个笼统的「调用失败」，**具体的业务码丢了**——
// 客户端就无法区分「用户名已占用」和「数据库炸了」。
//
// 所以约定：
//   - **系统级错误**（数据库连不上、代码 bug）→ 返回 Go error，
//     网关统一兜底成「内部服务错误」
//   - **业务级错误**（用户名占用、视频不存在）→ 塞进 BaseResp 返回，
//     网关还原成 errno，HTTP 层拿到的错误跟本地调用时一模一样
package rpcbase

import (
	"strconv"

	"gobili/kitex_gen/rpcmodel"
	"gobili/pkg/errno"
)

// OK 返回成功响应。
func OK() *rpcmodel.BaseResp {
	return &rpcmodel.BaseResp{
		Code:    strconv.FormatInt(errno.SuccessCode, 10),
		Message: errno.SuccessMsg,
	}
}

// Fail 把业务错误转成 BaseResp。
func Fail(err error) *rpcmodel.BaseResp {
	e := errno.ConvertErr(err)
	return &rpcmodel.BaseResp{
		Code:    strconv.FormatInt(e.ErrorCode, 10),
		Message: e.ErrorMsg,
	}
}

// Check 检查响应中的业务码，非成功时还原成 errno。
//
// 网关侧在每个 RPC 调用后都要先 Check，把 BaseResp 还原成错误返回。
func Check(base *rpcmodel.BaseResp) error {
	if base == nil {
		return errno.InternalError
	}

	code, err := strconv.ParseInt(base.Code, 10, 64)
	if err != nil {
		return errno.InternalError
	}
	if code == errno.SuccessCode {
		return nil
	}
	return errno.NewErrNo(code, base.Message)
}
