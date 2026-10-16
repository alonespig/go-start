package dto

// PageReq 通用分页请求
type PageReq struct {
	Page int `form:"page" json:"page" binding:"omitempty,min=1"`
	Size int `form:"size" json:"size" binding:"omitempty,min=1,max=100"`
}

// Normalize 返回规范化后的分页参数
func (p *PageReq) Normalize() (page, size int) {
	page, size = p.Page, p.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return
}
