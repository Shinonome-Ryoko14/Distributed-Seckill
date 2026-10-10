// Package errs 定义业务错误码与可携带 HTTP 状态的 AppError。
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Code 是业务错误码。0 表示成功；1xxxx 为用户域,2xxxx 为商品域,5xxxx 为系统域。
type Code int

// 业务错误码定义。
const (
	CodeOK Code = 0

	CodeInvalidParam  Code = 10001
	CodeUserExists    Code = 10002
	CodeUserNotFound  Code = 10003
	CodePasswordWrong Code = 10004

	CodeProductNotFound Code = 20001
	CodeProductInvalid  Code = 20002

	CodeInternal Code = 50000
)

// AppError 是携带业务码与 HTTP 状态的错误，用于 handler 层统一转换为响应。
type AppError struct {
	Code       Code
	Message    string
	HTTPStatus int
	cause      error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.cause }

// New 构造一个业务错误。
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

// Wrap 在业务错误上保留底层 cause，便于日志排查。
func Wrap(cause error, code Code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus, cause: cause}
}

// From 将任意 error 归一化为 *AppError：未知错误统一按内部错误（500）处理。
func From(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(err, CodeInternal, "internal server error", http.StatusInternalServerError)
}
