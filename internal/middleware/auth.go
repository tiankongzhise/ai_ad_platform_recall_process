package middleware

import (
	"strings"

	"ai_ad_platform_recall_process/internal/service"
	"ai_ad_platform_recall_process/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
	TokenContextKey     = "token"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.Unauthorized(c, response.MissingTokenCode, "未提供Token")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Unauthorized(c, response.InvalidTokenFormatCode, "Token格式错误，请使用Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenStr == "" {
			response.Unauthorized(c, response.InvalidTokenFormatCode, "Token不能为空")
			c.Abort()
			return
		}

		user, err := authService.ValidateToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, response.InvalidTokenCode, "Token无效或已过期")
			c.Abort()
			return
		}

		c.Set(UserContextKey, user)
		c.Set(TokenContextKey, tokenStr)
		c.Next()
	}
}
