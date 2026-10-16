package errcode

import "fmt"

// 业务错误码（0 表示成功）
const (
	CodeSuccess      = 0
	CodeInvalidParam = 40000
	CodeUnauthorized = 40100
	CodeNotFound     = 40400
	CodeInternal     = 50000
)

// BizError 业务错误：service/repository 层返回它，handler 层翻译成响应。
// 放在底层包，让 service 可以依赖它，而不会反向依赖 api 层。
type BizError struct {
	Code    int
	Message string
	Err     error
}

func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *BizError) Unwrap() error { return e.Err }

func New(code int, msg string) *BizError {
	return &BizError{Code: code, Message: msg}
}

func Wrap(err error, code int, msg string) *BizError {
	return &BizError{Code: code, Message: msg, Err: err}
}
