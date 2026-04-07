package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nekoimi/go-project-template/internal/pkg/errcode"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    errcode.OK.Value,
		Message: errcode.OK.Message,
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus int, code *errcode.Code) {
	c.JSON(httpStatus, APIResponse{
		Code:    code.Value,
		Message: code.Message,
	})
}

func ErrorWithMsg(c *gin.Context, httpStatus int, code *errcode.Code, msg string) {
	c.JSON(httpStatus, APIResponse{
		Code:    code.Value,
		Message: msg,
	})
}
