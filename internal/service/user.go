package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alonespig/go-start/internal/model"
	"github.com/alonespig/go-start/internal/repository"
	"github.com/alonespig/go-start/pkg/cache"
	"github.com/alonespig/go-start/pkg/errcode"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	userCacheKey = "user:%d"
	userCacheTTL = 10 * time.Minute
)

type UserService interface {
	Create(ctx context.Context, u *model.User) (int64, error)
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

// Create 创建用户，返回新记录 ID。
// 唯一键冲突（如邮箱重复）翻译成业务错误，避免把数据库异常透给调用方。
func (s *userService) Create(ctx context.Context, u *model.User) (int64, error) {
	if err := s.repo.Create(ctx, u); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, errcode.New(errcode.CodeConflict, "邮箱已存在")
		}
		return 0, err
	}
	return u.ID, nil
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
