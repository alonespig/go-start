package repository

import (
	"context"

	"github.com/alonespig/go-start/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (model.User, error)
	List(ctx context.Context, page, size int) ([]model.User, int64, error)
	UpdateFields(ctx context.Context, id int64, fields map[string]any) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
}

type userRepository struct {
	*BaseRepo[model.User]
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{BaseRepo: NewBaseRepo[model.User](db)}
}
