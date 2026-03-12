package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-chat/internal/api/middleware"
	"go-chat/internal/service"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

type createGroupRequest struct {
	Name string `json:"name" binding:"required,max=100"` // 群名必填，最长100字符
}

type updateGroupProfileRequest struct {
	Name   *string `json:"name"`
	Avatar *string `json:"avatar"`
}

// 创建群
func (h *GroupHandler) Create(c *gin.Context) {
	// 从“上下文”里拿出当前登录的人是谁
	// 登录中间件会把用户 ID 塞进 c 里
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// 3. 调用 Service 开始建群，传过去：上下文、建群人的 ID、群名
	group, err := h.groupService.CreateGroup(c.Request.Context(), userID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidGroupName):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusCreated, group)
}

func (h *GroupHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		writeError(c, http.StatusBadRequest, "invalid group id")
		return
	}

	var req updateGroupProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Name == nil && req.Avatar == nil) {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.groupService.UpdateGroupProfile(c.Request.Context(), userID, groupID, req.Name, req.Avatar)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotGroupOwner):
			writeError(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidGroupName), errors.Is(err, service.ErrInvalidAvatar):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

// 加入群组
func (h *GroupHandler) Join(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	//解析路径参数
	// 比如网址是 /groups/123/join，c.Param("id") 拿到的就是字符串 "123"
	// strconv.ParseUint 的作用是把字符串 "123" 转成数字 123
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid group id")
		return
	}

	// 调用 Service 逻辑：把用户和群在数据库里关联起来
	err = h.groupService.JoinGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error()) // 进一个不存在的群，返回 404
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "joined"})
}

func (h *GroupHandler) ListMine(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	//帮我查查这个 userID 都加了哪些群？
	groups, err := h.groupService.ListMyGroups(c.Request.Context(), userID)
	if err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": groups})
}

func (h *GroupHandler) ListMembers(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		writeError(c, http.StatusBadRequest, "invalid group id")
		return
	}

	members, err := h.groupService.ListMembers(c.Request.Context(), userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotGroupMember):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": members})
}

func (h *GroupHandler) Leave(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		writeError(c, http.StatusBadRequest, "invalid group id")
		return
	}

	err = h.groupService.LeaveGroup(c.Request.Context(), userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotGroupMember), errors.Is(err, service.ErrGroupOwnerCannotLeave):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		writeError(c, http.StatusBadRequest, "invalid group id")
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil || targetUserID == 0 {
		writeError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.groupService.RemoveMember(c.Request.Context(), userID, groupID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGroupNotFound):
			writeError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotGroupOwner):
			writeError(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrCannotRemoveGroupOwner), errors.Is(err, service.ErrTargetNotGroupMember):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}
