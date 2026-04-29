package handler

import (
	"errors"
	"net/url"

	"ai_ad_platform_recall_process/internal/service"
	"ai_ad_platform_recall_process/pkg/response"

	"github.com/gin-gonic/gin"
)

type NotifyHandler struct {
	notifyService *service.NotifyService
}

func NewNotifyHandler(notifyService *service.NotifyService) *NotifyHandler {
	return &NotifyHandler{
		notifyService: notifyService,
	}
}

type SetNotifyRequest struct {
	NotifyURL string `json:"notify_url" binding:"required"`
}

func (h *NotifyHandler) SetNotifyURL(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	var req SetNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	_, err := url.ParseRequestURI(req.NotifyURL)
	if err != nil {
		response.BadRequest(c, response.InvalidNotifyURLCode, "通知URL格式错误", nil)
		return
	}

	if err := h.notifyService.SetNotifyURL(user.ID, req.NotifyURL); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.BadRequest(c, response.InvalidNotifyURLCode, "通知URL格式错误", nil)
			return
		}
		response.InternalError(c, "设置通知URL失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"notify_url": req.NotifyURL,
	})
}

func (h *NotifyHandler) GetNotifyURL(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	notifyURL, err := h.notifyService.GetNotifyURL(user.ID)
	if err != nil {
		response.InternalError(c, "获取通知URL失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"notify_url": notifyURL,
	})
}

func (h *NotifyHandler) CancelNotifyURL(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	if err := h.notifyService.SetNotifyURL(user.ID, ""); err != nil {
		response.InternalError(c, "取消通知URL失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"notify_url": "",
	})
}
