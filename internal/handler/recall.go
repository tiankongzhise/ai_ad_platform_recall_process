package handler

import (
	"errors"
	"strconv"
	"strings"

	"ai_ad_platform_recall_process/internal/model"
	"ai_ad_platform_recall_process/internal/repository"
	"ai_ad_platform_recall_process/internal/service"
	"ai_ad_platform_recall_process/pkg/response"

	"github.com/gin-gonic/gin"
)

const TokenContextKey = "token"

type RecallHandler struct {
	recallService *service.RecallService
	notifyService *service.NotifyService
	authService   *service.AuthService
}

func NewRecallHandler(recallService *service.RecallService, notifyService *service.NotifyService, authService *service.AuthService) *RecallHandler {
	return &RecallHandler{
		recallService: recallService,
		notifyService: notifyService,
		authService:   authService,
	}
}

// getCurrentUserFromRequest 从请求头解析 token 并验证用户身份
func (h *RecallHandler) getCurrentUserFromRequest(c *gin.Context) (*model.User, string) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, "未提供Authorization请求头"
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, "Token格式错误，请使用Bearer <token>"
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		return nil, "Token不能为空"
	}

	user, err := h.authService.ValidateToken(tokenStr)
	if err != nil {
		return nil, "Token无效或已过期"
	}

	return user, ""
}

func (h *RecallHandler) RecallInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 0,
		"message": "Recall接口使用说明",
		"data": gin.H{
			"description": "回调接口使用说明",
			"format":       "/recall?state=RecallServiceUserUid_PlatformNumber_UserNumber",
			"format_example": "RecallServiceUserUid_PlatformNumber_UserNumber",
			"params": gin.H{
				"RecallServiceUserUid": "32位十六进制字符串，用户注册时生成，用于标识用户",
				"PlatformNumber":       "不超过13位的纯数字，由用户自行管理对应关系",
				"UserNumber":           "不超过13位的纯数字，由用户自行管理对应关系",
			},
			"example": "e8b5f1a2c3d4e5f6a7b8c9d0e1f2a3b4_12345_67890",
		},
	})
}

func (h *RecallHandler) HandleRecall(c *gin.Context) {
	state := c.Query("state")

	// 无参数访问时返回接口使用说明
	if state == "" {
		h.RecallInfo(c)
		return
	}

	params, missingParams, err := h.recallService.ProcessRecallWithParams(state)
	if err != nil {
		if errors.Is(err, service.ErrStateFormatError) {
			response.BadRequest(c, response.StateFormatErrorCode, "state参数格式错误，格式应为：RecallServiceUserUid_PlatformNumber_UserNumber", nil)
			return
		}
		if errors.Is(err, service.ErrInvalidUid) {
			response.BadRequest(c, response.StateFormatErrorCode, "RecallServiceUserUid格式错误，应为32位十六进制字符串", nil)
			return
		}
		if errors.Is(err, service.ErrInvalidPlatformNumber) {
			response.BadRequest(c, response.StateFormatErrorCode, "PlatformNumber格式错误，应为不超过13位的纯数字", nil)
			return
		}
		if errors.Is(err, service.ErrInvalidUserNumber) {
			response.BadRequest(c, response.StateFormatErrorCode, "UserNumber格式错误，应为不超过13位的纯数字", nil)
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			response.BadRequest(c, response.InvalidCredentialsCode, "用户不存在或RecallServiceUserUid无效", nil)
			return
		}
		if errors.Is(err, service.ErrMissingRequiredParam) {
			msg := "缺少必填参数: "
			for i, p := range missingParams {
				if i > 0 {
					msg += ", "
				}
				msg += p
			}
			response.BadRequest(c, response.MissingParamsCode, msg, gin.H{
				"missing_params": missingParams,
			})
			return
		}
		response.InternalError(c, "回调处理失败: "+err.Error())
		return
	}

	// 获取 state 之外的所有查询参数作为额外参数
	extraParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if key == "state" {
			continue
		}
		// 只取第一个值
		if len(values) > 0 {
			extraParams[key] = values[0]
		}
	}

	resp, err := h.recallService.SaveRecall(params, extraParams)
	if err != nil {
		response.InternalError(c, "回调处理失败: "+err.Error())
		return
	}

	if params.RecallServiceName != "" && params.PlatformNumber != "" && params.UserNumber != "" {
		h.notifyService.TriggerNotify(params.RecallServiceName, params.PlatformNumber, params.UserNumber)
	}

	response.SuccessWithMessage(c, "回调处理成功", resp)
}

func (h *RecallHandler) Query(c *gin.Context) {
	// 必须通过 Token 鉴权，强制只能查询自己的记录
	currentUser, errMsg := h.getCurrentUserFromRequest(c)
	if errMsg != "" {
		response.Unauthorized(c, response.InvalidTokenCode, errMsg+"，无权访问")
		return
	}

	params := repository.QueryParams{
		RecallServiceName: currentUser.RecallServiceName, // 强制使用当前用户，不允许查询其他用户
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			params.PageSize = pageSize
		}
	}

	result, err := h.recallService.Query(params)
	if err != nil {
		response.InternalError(c, "查询失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "查询成功", result)
}

func (h *RecallHandler) QueryLatest(c *gin.Context) {
	// 必须通过 Token 鉴权，强制只能查询自己的记录
	currentUser, errMsg := h.getCurrentUserFromRequest(c)
	if errMsg != "" {
		response.Unauthorized(c, response.InvalidTokenCode, errMsg+"，无权访问")
		return
	}

	params := repository.QueryParams{
		RecallServiceName: currentUser.RecallServiceName, // 强制使用当前用户，不允许查询其他用户
	}

	record, err := h.recallService.QueryLatest(params)
	if err != nil {
		response.InternalError(c, "查询失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "查询成功", record)
}

func (h *RecallHandler) QueryHistory(c *gin.Context) {
	// 必须通过 Token 鉴权，强制只能查询自己的记录
	currentUser, errMsg := h.getCurrentUserFromRequest(c)
	if errMsg != "" {
		response.Unauthorized(c, response.InvalidTokenCode, errMsg+"，无权访问")
		return
	}

	params := repository.QueryParams{
		RecallServiceName: currentUser.RecallServiceName, // 强制使用当前用户，不允许查询其他用户
	}

	records, err := h.recallService.QueryAll(params)
	if err != nil {
		response.InternalError(c, "查询失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "查询成功", gin.H{
		"total":   len(records),
		"records": records,
	})
}

func GetUserFromContext(c *gin.Context) *model.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*model.User)
}
