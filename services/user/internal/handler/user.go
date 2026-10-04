// Package handler 提供 user 服务的 HTTP 接入层。
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/errs"
	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/response"
)

type userService interface {
	Register(ctx context.Context, username, password string) (uint64, error)
	Login(ctx context.Context, username, password string) (string, error)
}

// UserHandler 处理用户注册与登录的 HTTP 请求。
type UserHandler struct {
	svc userService
}

// NewUserHandler 构造 UserHandler。
func NewUserHandler(svc userService) *UserHandler {
	return &UserHandler{svc: svc}
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register 处理 POST /api/v1/users/register。
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid param",
			http.StatusBadRequest))
		return
	}

	id, err := h.svc.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}
	response.OK(c.Writer, gin.H{"id": id})
}

// Login 处理 POST /api/v1/users/login。
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Fail(c.Writer, errs.New(errs.CodeInvalidParam, "invalid param",
			http.StatusBadRequest))
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c.Writer, err)
		return
	}
	response.OK(c.Writer, gin.H{"token": token})
}
