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
	"gorm.io/gorm"
)

const (
	userCacheKey = "user:%d"
	userCacheTTL = 10 * time.Minute
)

// UpdateUserInput 用指针区分“字段未传”和“显式传零值”
type UpdateUserInput struct {
	Name  *string
	Email *string
	Age   *int
}

type UserService interface {
	Create(ctx context.Context, u *model.User) (int64, error)
	GetByID(ctx context.Context, id int64) (model.User, error)
	List(ctx context.Context, page, size int) ([]model.User, int64, error)
	Update(ctx context.Context, id int64, in UpdateUserInput) error
	Delete(ctx context.Context, id int64) error
}

type userService struct {
	repo  repository.UserRepository
	cache cache.Cache
}

func NewUserService(repo repository.UserRepository, c cache.Cache) UserService {
	return &userService{repo: repo, cache: c}
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

// Update 按字段更新用户，成功后失效缓存
func (s *userService) Update(ctx context.Context, id int64, in UpdateUserInput) error {
	fields := make(map[string]any, 3)
	if in.Name != nil {
		fields["name"] = *in.Name
	}
	if in.Email != nil {
		fields["email"] = *in.Email
	}
	if in.Age != nil {
		fields["age"] = *in.Age
	}
	if len(fields) == 0 {
		return errcode.New(errcode.CodeInvalidParam, "没有需要更新的字段")
	}

	n, err := s.repo.UpdateFields(ctx, id, fields)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errcode.New(errcode.CodeConflict, "邮箱已存在")
		}
		return err
	}
	if n == 0 {
		// MySQL 在值未变化时 RowsAffected 也可能为 0，再确认一次记录是否存在
		if _, err := s.repo.GetByID(ctx, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.New(errcode.CodeNotFound, "用户不存在")
			}
			return err
		}
	}

	s.invalidate(ctx, id)
	return nil
}

// Delete 删除用户并失效缓存
func (s *userService) Delete(ctx context.Context, id int64) error {
	n, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return errcode.New(errcode.CodeNotFound, "用户不存在")
	}
	s.invalidate(ctx, id)
	return nil
}

// invalidate 写操作后删除缓存，避免读到旧数据
func (s *userService) invalidate(ctx context.Context, id int64) {
	if s.cache != nil {
		_ = s.cache.Delete(ctx, fmt.Sprintf(userCacheKey, id))
	}
}
