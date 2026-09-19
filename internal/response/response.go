// Package response 定义统一 HTTP 响应体与写出辅助函数。
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
)

// Body 是所有接口的统一响应体。
type Body struct {
	Code    errs.Code `json:"code"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
}

// JSON 按给定 HTTP 状态码写出响应体。
func JSON(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write response", "error", err)
	}
}

// OK 写出成功响应。
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Body{Code: errs.CodeOK, Message: "ok", Data: data})
}

// Fail 将任意 error 归一化为业务错误并写出；未知错误按 500 处理且不泄漏内部信息。
func Fail(w http.ResponseWriter, err error) {
	appErr := errs.From(err)
	JSON(w, appErr.HTTPStatus, Body{Code: appErr.Code, Message: appErr.Message})
}
