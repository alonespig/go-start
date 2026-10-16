package model

import "time"

// User 用户模型（MySQL，GORM）
type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Email     string    `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Age       int       `gorm:"default:0" json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
