package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go-starter/internal/api/dto"
	"go-starter/internal/api/response"
	"go-starter/internal/service"
	"go-starter/pkg/errcode"
	"go-starter/pkg/slice"
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
		response.Fail(c, errcode.New(errcode.CodeInvalidParam, "参数错误: "+err.Error()))
		return
	}
	if err := h.svc.Create(c.Request.Context(), req.ToModel()); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"message": "created"})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.New(errcode.CodeInvalidParam, "无效的用户 id"))
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, dto.ToUserResp(u))
}

func (h *UserHandler) List(c *gin.Context) {
	var req dto.PageReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errcode.New(errcode.CodeInvalidParam, "参数错误: "+err.Error()))
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
