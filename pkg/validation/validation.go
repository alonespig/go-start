package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Init 注册字段名称映射，应用启动时执行一次
func Init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		if name == "" {
			return field.Name
		}

		return name
	})
}

// FormatError 将绑定和校验错误转换为前端友好的格式
func FormatError(err error) map[string]string {
	result := make(map[string]string)

	var validationErr validator.ValidationErrors

	if errors.As(err, &validationErr) {
		for _, e := range validationErr {
			// 使用 Namespace 保留嵌套结构体路径
			field := e.Namespace()
			if index := strings.Index(field, "."); index >= 0 {
				field = field[index+1:]
			}

			// 一个字段只返回第一条错误
			if _, exists := result[field]; exists {
				continue
			}

			switch e.Tag() {
			case "required":
				result[field] = "不能为空"
			case "email":
				result[field] = "邮箱格式不正确"
			case "min":
				result[field] = fmt.Sprintf("长度或数值不能小于%s", e.Param())
			case "max":
				result[field] = fmt.Sprintf("长度或数值不能大于%s", e.Param())
			case "gte":
				result[field] = fmt.Sprintf("必须大于或等于%s", e.Param())
			case "lte":
				result[field] = fmt.Sprintf("必须小于或等于%s", e.Param())
			case "len":
				result[field] = fmt.Sprintf("长度必须为%s", e.Param())
			case "oneof":
				result[field] = fmt.Sprintf("必须是以下值之一：%s", e.Param())
			case "url":
				result[field] = "URL格式不正确"
			default:
				result[field] = fmt.Sprintf("参数校验失败（%s）", e.Tag())
			}
		}
		return result
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		if field == "" {
			field = "_body"
		}
		result[field] = "参数类型错误，期望类型：" + typeErr.Type.String()
		return result
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		result["_body"] = "json格式错误"
		return result
	}

	if errors.Is(err, io.EOF) {
		result["_body"] = "请求体不能为空"
		return result
	}

	result["_body"] = "请求参数不合法"
	return result
}

func BindJson[T any](ctx *gin.Context, req T) bool {
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数校验失败",
			"errors":  FormatError(err),
		})
		return false
	}
	return true
}
