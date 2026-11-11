package response

import (
	"errors"
	"net/http"

	"github.com/alonespig/go-start/pkg/errcode"
	"github.com/alonespig/go-start/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

// ErrorBiz 从任意 error 里提取业务错误码；非 BizError 一律按内部错误处理。
// 对外只返回通用文案，真实错误由调用方记日志，避免泄漏表名、SQL 等内部细节。
func ErrorBiz(err error) *Response {
	var be *errcode.BizError
	if errors.As(err, &be) {
		return &Response{Code: be.Code, Message: be.Message}
	}
	return &Response{Code: errcode.CodeInternal, Message: "服务器内部错误"}
}

// OK 返回成功响应（HTTP 200 + 业务 code）
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Success(data))
}

// Fail 返回失败响应（HTTP 200 + 业务错误码）
func Fail(c *gin.Context, err error) {
	logIfInternal(c, err)
	c.JSON(http.StatusOK, ErrorBiz(err))
}

// FailWithData 与 Fail 相同，额外把字段级明细放进 data（如参数校验错误）
func FailWithData(c *gin.Context, err error, data any) {
	logIfInternal(c, err)
	resp := ErrorBiz(err)
	resp.Data = data
	c.JSON(http.StatusOK, resp)
}

// logIfInternal 只记录非业务错误，用户输入错误不打 error 日志
func logIfInternal(c *gin.Context, err error) {
	var be *errcode.BizError
	if errors.As(err, &be) {
		return
	}
	log.Error("request failed",
		zap.Error(err),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)
}
