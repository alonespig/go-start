package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go-starter/internal/model"
	"go-starter/internal/repository"
	"go-starter/pkg/cache"
	"go-starter/pkg/errcode"
	"gorm.io/gorm"
)

const (
	userCacheKey = "user:%d"
	userCacheTTL = 10 * time.Minute
)

type UserService interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (model.User, error)
	List(ctx context.Context, page, size int) ([]model.User, int64, error)
}

type userService struct {
	repo  repository.UserRepository
	cache *redis.Client
}

func NewUserService(repo repository.UserRepository, cache *redis.Client) UserService {
	return &userService{repo: repo, cache: cache}
}

func (s *userService) Create(ctx context.Context, u *model.User) error {
	return s.repo.Create(ctx, u)
}

// GetByID 用泛型 cache-aside 工具：缓存命中直接返回，未命中查库并回填
func (s *userService) GetByID(ctx context.Context, id int64) (model.User, error) {
	key := fmt.Sprintf(userCacheKey, id)
	return cache.GetOrSet(ctx, s.cache, key, userCacheTTL, func() (model.User, error) {
		u, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.User{}, errcode.New(errcode.CodeNotFound, "用户不存在")
			}
			return model.User{}, err
		}
		return u, nil
	})
}

func (s *userService) List(ctx context.Context, page, size int) ([]model.User, int64, error) {
	return s.repo.List(ctx, page, size)
}
