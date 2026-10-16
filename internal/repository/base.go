package repository

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepo 泛型基础仓库：基于 GORM 官方泛型 API gorm.G[T]。
// T 为结构体类型（非指针），查询方法返回 T 值；未找到时返回 gorm.ErrRecordNotFound。
type BaseRepo[T any] struct {
	db *gorm.DB
}

func NewBaseRepo[T any](db *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{db: db}
}

// Create 插入一条记录
func (r *BaseRepo[T]) Create(ctx context.Context, v *T) error {
	return gorm.G[T](r.db).Create(ctx, v)
}

// GetByID 按主键查询，未找到时返回 (零值, gorm.ErrRecordNotFound)
func (r *BaseRepo[T]) GetByID(ctx context.Context, id int64) (T, error) {
	return gorm.G[T](r.db).Where("id = ?", id).First(ctx)
}

// GetByIDs 批量按主键查询
func (r *BaseRepo[T]) GetByIDs(ctx context.Context, ids []int64) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}
	return gorm.G[T](r.db).Where("id IN ?", ids).Find(ctx)
}

// List 分页查询（按 id 倒序）
func (r *BaseRepo[T]) List(ctx context.Context, page, size int) ([]T, int64, error) {
	total, err := gorm.G[T](r.db).Count(ctx, "*")
	if err != nil {
		return nil, 0, err
	}
	list, err := gorm.G[T](r.db).
		Order("id desc").
		Limit(size).
		Offset((page - 1) * size).
		Find(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update 按主键更新单列
func (r *BaseRepo[T]) Update(ctx context.Context, id int64, column string, value any) error {
	_, err := gorm.G[T](r.db).Where("id = ?", id).Update(ctx, column, value)
	return err
}

// Updates 按主键更新（用 T 的非零字段；零值字段请用 Update 单列）
func (r *BaseRepo[T]) Updates(ctx context.Context, id int64, t T) error {
	_, err := gorm.G[T](r.db).Where("id = ?", id).Updates(ctx, t)
	return err
}

// UpdateBy 条件更新单列，返回影响行数
func (r *BaseRepo[T]) UpdateBy(ctx context.Context, column string, value any, query string, args ...any) (int64, error) {
	n, err := gorm.G[T](r.db).Where(query, args...).Update(ctx, column, value)
	return int64(n), err
}

// Delete 按主键删除
func (r *BaseRepo[T]) Delete(ctx context.Context, id int64) error {
	_, err := gorm.G[T](r.db).Where("id = ?", id).Delete(ctx)
	return err
}

// DeleteBy 条件删除，返回影响行数
func (r *BaseRepo[T]) DeleteBy(ctx context.Context, query string, args ...any) (int64, error) {
	n, err := gorm.G[T](r.db).Where(query, args...).Delete(ctx)
	return int64(n), err
}
