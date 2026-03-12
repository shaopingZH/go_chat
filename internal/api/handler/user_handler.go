package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-chat/internal/api/middleware"
	"go-chat/internal/model"
	"go-chat/internal/service"
)

type UserHandler struct {
	profileService *service.ProfileService
	onlineService  *service.OnlineService
}

func NewUserHandler(profileService *service.ProfileService, onlineService *service.OnlineService) *UserHandler {
	return &UserHandler{profileService: profileService, onlineService: onlineService}
}

type updateMyProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Avatar      *string `json:"avatar"`
	Bio         *string `json:"bio"`
}

type onlineBatchRequest struct {
	UserIDs []uint64 `json:"user_ids"`
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.profileService.GetMyProfile(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	h.writeProfile(c, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateMyProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.profileService.UpdateMyProfile(c.Request.Context(), userID, service.UpdateProfileInput{
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
		Bio:         req.Bio,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProfilePayload),
			errors.Is(err, service.ErrInvalidDisplayName),
			errors.Is(err, service.ErrInvalidAvatar),
			errors.Is(err, service.ErrInvalidBio):
			writeError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	h.writeProfile(c, updated)
}

func (h *UserHandler) PublicProfile(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		writeError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.profileService.GetPublicProfile(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	h.writeProfile(c, user)
}

func (h *UserHandler) Online(c *gin.Context) {
	_, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		writeError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	status, err := h.onlineService.GetUserOnlineStatus(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *UserHandler) OnlineBatch(c *gin.Context) {
	_, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req onlineBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.UserIDs) == 0 || len(req.UserIDs) > 100 {
		writeError(c, http.StatusBadRequest, service.ErrInvalidUserIDs.Error())
		return
	}
	for _, userID := range req.UserIDs {
		if userID == 0 {
			writeError(c, http.StatusBadRequest, service.ErrInvalidUserIDs.Error())
			return
		}
	}

	items, err := h.onlineService.BatchGetUserOnlineStatus(c.Request.Context(), req.UserIDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserIDs):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *UserHandler) writeProfile(c *gin.Context, user *model.User) {
	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"bio":          user.Bio,
	})
}
