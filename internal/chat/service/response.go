package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// writeJSON 输出统一响应信封，与 HTTP 网关保持同一种结构。
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Errorf("chat.writeJSON: %v", err)
	}
}

// writeData 输出成功响应。
func writeData(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{
		"code":    strconv.FormatInt(errno.SuccessCode, 10),
		"message": errno.SuccessMsg,
		"data":    data,
	})
}

// writeError 输出错误响应。未识别的内部错误只回通用文案，原文进日志。
func writeError(w http.ResponseWriter, status int, err error) {
	e := errno.ConvertErr(err)
	if e.ErrorCode >= errno.InternalErrorCode {
		logger.Errorf("chat internal error: %v", err)
	}
	writeJSON(w, status, map[string]any{
		"code":    strconv.FormatInt(e.ErrorCode, 10),
		"message": e.ErrorMsg,
	})
}
