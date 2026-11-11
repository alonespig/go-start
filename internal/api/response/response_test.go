package response

import (
	"errors"
	"testing"

	"github.com/alonespig/go-start/pkg/errcode"
)

// 非业务错误只能暴露通用文案，不能把底层错误原文返回给客户端
func TestErrorBizHidesInternal(t *testing.T) {
	got := ErrorBiz(errors.New("Error 1062: Duplicate entry 'a@b.c' for key 'users.email'"))

	if got.Code != errcode.CodeInternal {
		t.Fatalf("code = %d, want %d", got.Code, errcode.CodeInternal)
	}
	if got.Message != "服务器内部错误" {
		t.Fatalf("message = %q, 内部细节泄漏", got.Message)
	}
}

// 业务错误需要原样保留 code 与 message
func TestErrorBizKeepsBizError(t *testing.T) {
	got := ErrorBiz(errcode.New(errcode.CodeNotFound, "用户不存在"))

	if got.Code != errcode.CodeNotFound || got.Message != "用户不存在" {
		t.Fatalf("got = %+v", got)
	}
}
