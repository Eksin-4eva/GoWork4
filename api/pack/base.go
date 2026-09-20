// Package pack 负责统一响应格式与模型转换。
//
// handler 不直接写 JSON，一律通过本包输出，保证所有接口的响应结构一致：
//
//	{"code": "0", "message": "Success", "data": {...}}
package pack

import (
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Base 是无数据响应体。
type Base struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespWithData 是带数据响应体。
type RespWithData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// RespSuccess 返回成功且无数据。
func RespSuccess(c *app.RequestContext) {
	c.JSON(consts.StatusOK, Base{
		Code:    strconv.FormatInt(errno.SuccessCode, 10),
		Message: errno.SuccessMsg,
	})
}

// RespData 返回成功并携带数据。
func RespData(c *app.RequestContext, data any) {
	c.JSON(consts.StatusOK, RespWithData{
		Code:    strconv.FormatInt(errno.SuccessCode, 10),
		Message: errno.SuccessMsg,
		Data:    data,
	})
}

// RespList 返回成功并携带列表，nil 切片会被规整为空数组，
// 避免客户端拿到 {"data": null}。
func RespList[T any](c *app.RequestContext, items []T) {
	if items == nil {
		items = []T{}
	}
	RespData(c, items)
}

// RespError 返回错误。
//
// 未识别的内部错误只回一个通用文案，原始错误写入日志，
// 避免把 SQL 语句、表名、文件路径等实现细节透给客户端。
func RespError(c *app.RequestContext, err error) {
	e := errno.ConvertErr(err)
	if e.ErrorCode >= errno.InternalErrorCode {
		logger.Errorf("internal error: %v", err)
	}
	c.JSON(consts.StatusOK, Base{
		Code:    strconv.FormatInt(e.ErrorCode, 10),
		Message: e.ErrorMsg,
	})
}
