package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-chat/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// 数据格式
type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 注册逻辑
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	// 1. ShouldBindJSON：把前端传来的 JSON 数据“塞”进 req 结构体里
	// 如果数据格式不对（比如没填密码），直接报错
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 2. 调用Service注册
	// c.Request.Context() 是传递这次请求的生命周期上下文
	user, err := h.authService.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRegister):
			writeError(c, http.StatusBadRequest, err.Error()) // 返回 409 冲突
		case errors.Is(err, service.ErrUsernameTaken):
			writeError(c, http.StatusConflict, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	// 注册成功，返回 201 Created 和用户信息
	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"avatar":   user.Avatar,
	})
}

// 登录逻辑
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(c, http.StatusUnauthorized, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"avatar":   user.Avatar,
		},
	})
}
