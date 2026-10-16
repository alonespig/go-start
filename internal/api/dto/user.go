package dto

import "go-starter/internal/model"

type CreateUserReq struct {
	Name  string `json:"name" binding:"required,max=64"`
	Email string `json:"email" binding:"required,email,max=128"`
	Age   int    `json:"age" binding:"omitempty,gte=0,lte=200"`
}

// ToModel 转成数据模型
func (r *CreateUserReq) ToModel() *model.User {
	return &model.User{
		Name:  r.Name,
		Email: r.Email,
		Age:   r.Age,
	}
}

type UserResp struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// ToUserResp 从数据模型转成响应 DTO
func ToUserResp(u model.User) *UserResp {
	return &UserResp{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Age:   u.Age,
	}
}
