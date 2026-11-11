package validation

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

type testReq struct {
	Name  string `json:"name" binding:"required,max=3"`
	Email string `json:"email" binding:"required,email"`
}

// 校验错误应使用 JSON 字段名，并翻译成中文提示
func TestFormatError(t *testing.T) {
	Init()

	err := binding.Validator.ValidateStruct(testReq{Name: "abcd", Email: "bad"})
	got := FormatError(err)

	if got["name"] != "长度或数值不能大于3" {
		t.Fatalf("name 错误信息不符: %q", got["name"])
	}
	if got["email"] != "邮箱格式不正确" {
		t.Fatalf("email 错误信息不符: %q", got["email"])
	}
}
