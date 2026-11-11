package handler

import (
	"strconv"

	"github.com/alonespig/go-start/internal/api/dto"
	"github.com/alonespig/go-start/internal/api/response"
	"github.com/alonespig/go-start/internal/api/validation"
	"github.com/alonespig/go-start/internal/service"
	"github.com/alonespig/go-start/pkg/errcode"
	"github.com/alonespig/go-start/pkg/slice"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithData(c,
			errcode.Wrap(err, errcode.CodeInvalidParam, "参数校验失败"),
			validation.FormatError(err))
		return
	}
	id, err := h.svc.Create(c.Request.Context(), req.ToModel())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToUserResp(u))
}

func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithData(c,
			errcode.Wrap(err, errcode.CodeInvalidParam, "参数校验失败"),
			validation.FormatError(err))
		return
	}

	in := service.UpdateUserInput{Name: req.Name, Email: req.Email, Age: req.Age}
	if err := h.svc.Update(c.Request.Context(), id, in); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// parseUserID 解析并校验路径参数，失败时已写入响应
func parseUserID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errcode.New(errcode.CodeInvalidParam, "无效的用户 id"))
		return 0, false
	}
	return id, true
}

func (h *UserHandler) List(c *gin.Context) {
	var req dto.PageReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithData(c,
			errcode.Wrap(err, errcode.CodeInvalidParam, "参数校验失败"),
			validation.FormatError(err))
		return
	}
	page, size := req.Normalize()

	list, total, err := h.svc.List(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}

	resp := slice.Map(list, dto.ToUserResp)
	response.OK(c, response.PageData[*dto.UserResp]{List: resp, Total: total})
}
