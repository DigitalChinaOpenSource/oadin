package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiResponse[T any] struct {
	Code int    `json:"business_code"`
	Msg  string `json:"message"`
	Data T      `json:"data"`
}

// Resp 统一响应函数
func Resp[T any](c *gin.Context, code int, msg string, data T) {
	c.JSON(200, ApiResponse[T]{Code: code, Msg: msg, Data: data})
}

// Success 成功响应
func Success[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, ApiResponse[T]{Code: 200, Msg: "success", Data: data})
}

func ValidFailure(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ApiResponse[any]{Code: 400, Msg: message, Data: nil})
}
