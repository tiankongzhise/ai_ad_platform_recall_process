package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	SuccessCode            = 0
	UsernameExistsCode     = 1001
	InvalidCredentialsCode = 1002
	InvalidTokenCode       = 1003
	MissingTokenCode       = 1004
	InvalidTokenFormatCode = 1005
	MissingParamsCode      = 2001
	StatusFormatErrorCode  = 2002
	ParamsParseErrorCode  = 2003
	InvalidNotifyURLCode   = 3001
	InternalErrorCode      = 5001
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: "操作成功",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    SuccessCode,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Unauthorized(c *gin.Context, code int, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    InternalErrorCode,
		Message: message,
		Data:    nil,
	})
}
