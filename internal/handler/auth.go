package handler

import (
	"errors"

	"ai_ad_platform_recall_process/internal/service"
	"ai_ad_platform_recall_process/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.BadRequest(c, response.UsernameExistsCode, "用户名已存在", nil)
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			response.BadRequest(c, response.UsernameExistsCode, "手机号已被注册", nil)
			return
		}
		response.InternalError(c, "注册失败: "+err.Error())
		return
	}

	response.Created(c, "注册成功", resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.BadRequest(c, response.InvalidCredentialsCode, "用户名或密码错误", nil)
			return
		}
		response.InternalError(c, "登录失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// CreateJWTToken 用ApiToken换取JWT和RefreshToken
func (h *AuthHandler) CreateJWTToken(c *gin.Context) {
	var req service.CreateJWTTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.CreateJWTToken(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidApiToken) {
			response.Unauthorized(c, response.InvalidTokenCode, "ApiToken无效")
			return
		}
		response.InternalError(c, "创建JWT失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// RefreshJWTByRefreshToken 使用RefreshToken刷新JWT
func (h *AuthHandler) RefreshJWTByRefreshToken(c *gin.Context) {
	var req service.RefreshJWTByRefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.RefreshJWTByRefreshToken(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			response.Unauthorized(c, response.InvalidTokenCode, "RefreshToken无效或已过期")
			return
		}
		response.InternalError(c, "刷新JWT失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// RefreshJWTByApiToken 使用ApiToken刷新JWT
func (h *AuthHandler) RefreshJWTByApiToken(c *gin.Context) {
	var req service.RefreshJWTByApiTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.RefreshJWTByApiToken(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidApiToken) {
			response.Unauthorized(c, response.InvalidTokenCode, "ApiToken无效")
			return
		}
		response.InternalError(c, "刷新JWT失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// GetJWTTokenInfo 获取JWT和RefreshToken信息（需认证）
func (h *AuthHandler) GetJWTTokenInfo(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	resp, err := h.authService.GetJWTTokenInfo(user.ID)
	if err != nil {
		response.InternalError(c, "获取Token信息失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// UpdateApiToken 更换ApiToken（需认证）
func (h *AuthHandler) UpdateApiToken(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	var req service.UpdateApiTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.UpdateApiToken(user.ID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidApiToken) {
			response.BadRequest(c, response.InvalidCredentialsCode, "原ApiToken不正确", nil)
			return
		}
		response.InternalError(c, "更换ApiToken失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// GetApiToken 获取当前用户的ApiToken（需认证）
func (h *AuthHandler) GetApiToken(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	resp, err := h.authService.GetApiToken(user.ID)
	if err != nil {
		response.InternalError(c, "获取ApiToken失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

// GetAccountInfo 获取当前用户的账户信息（需认证）
func (h *AuthHandler) GetAccountInfo(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	resp, err := h.authService.GetAccountInfo(user.ID)
	if err != nil {
		response.InternalError(c, "获取账户信息失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString(TokenContextKey)
	if token == "" {
		response.BadRequest(c, response.InvalidTokenFormatCode, "Token不存在", nil)
		return
	}

	if err := h.authService.Logout(token); err != nil {
		response.InternalError(c, "注销失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "注销成功", nil)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	resp, err := h.authService.RefreshToken(user.ID)
	if err != nil {
		response.InternalError(c, "Token刷新失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "Token刷新成功", resp)
}

func (h *AuthHandler) SendRegisterCode(c *gin.Context) {
	var req service.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.SendRegisterCode(req)
	if err != nil {
		response.InternalError(c, "发送验证码失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) SendResetCode(c *gin.Context) {
	var req service.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.SendResetCode(req)
	if err != nil {
		if errors.Is(err, service.ErrPhoneNotFound) {
			response.BadRequest(c, response.InvalidCredentialsCode, "手机号未注册", nil)
			return
		}
		response.InternalError(c, "发送验证码失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.ResetPassword(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			response.BadRequest(c, response.InvalidCredentialsCode, "验证码错误", nil)
			return
		}
		if errors.Is(err, service.ErrPhoneNotFound) {
			response.BadRequest(c, response.InvalidCredentialsCode, "手机号未注册", nil)
			return
		}
		response.InternalError(c, "重置密码失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.ChangePassword(user.ID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.BadRequest(c, response.InvalidCredentialsCode, "原密码错误", nil)
			return
		}
		response.InternalError(c, "修改密码失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	user := GetUserFromContext(c)
	if user == nil {
		response.Unauthorized(c, response.InvalidTokenCode, "用户未认证")
		return
	}

	var req service.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.DeleteAccount(user.ID, req)
	if err != nil {
		response.BadRequest(c, response.InternalErrorCode, err.Error(), nil)
		return
	}

	response.Success(c, resp)
}

// GetRecallServiceUserUidByUsername 通过用户名查询 RecallServiceUserUid
func (h *AuthHandler) GetRecallServiceUserUidByUsername(c *gin.Context) {
	var req service.GetRecallServiceUserUidByUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.InternalErrorCode, "请求参数错误: "+err.Error(), nil)
		return
	}

	resp, err := h.authService.GetRecallServiceUserUidByUsername(req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.BadRequest(c, response.InvalidCredentialsCode, "用户不存在", nil)
			return
		}
		response.InternalError(c, "查询失败: "+err.Error())
		return
	}

	response.Success(c, resp)
}
