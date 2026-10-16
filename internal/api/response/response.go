package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-starter/pkg/errcode"
)

// Response 统一响应信封
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageData 分页数据（泛型：List 元素类型安全）
type PageData[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

func Success(data any) *Response {
	return &Response{Code: errcode.CodeSuccess, Message: "ok", Data: data}
}

func Error(code int, msg string) *Response {
	return &Response{Code: code, Message: msg}
}

// ErrorBiz 从任意 error 里提取业务错误码；非 BizError 一律按内部错误处理
func ErrorBiz(err error) *Response {
	var be *errcode.BizError
	if errors.As(err, &be) {
		return &Response{Code: be.Code, Message: be.Message}
	}
	return &Response{Code: errcode.CodeInternal, Message: err.Error()}
}

// OK 返回成功响应（HTTP 200 + 业务 code）
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Success(data))
}

// Fail 返回失败响应（HTTP 200 + 业务错误码）
func Fail(c *gin.Context, err error) {
	c.JSON(http.StatusOK, ErrorBiz(err))
}
