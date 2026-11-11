package cache

import (
	"context"
	"time"

	"golang.org/x/sync/singleflight"
)

// Cache 缓存抽象：便于在 Redis、进程内内存之间切换，也方便单测替换。
type Cache interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

var flight singleflight.Group

// GetOrSet 泛型 cache-aside：先查缓存，未命中则回源加载并回填。
// 同一 key 的并发回源通过 singleflight 合并，避免缓存击穿。
// c 为 nil 时退化为直接回源。
func GetOrSet[T any](ctx context.Context, c Cache, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	if c != nil {
		var v T
		if ok, err := c.Get(ctx, key, &v); err == nil && ok {
			return v, nil
		}
	}

	loaded, err, _ := flight.Do(key, func() (any, error) {
		return loader()
	})
	if err != nil {
		var zero T
		return zero, err
	}
	v := loaded.(T)

	if c != nil {
		_ = c.Set(ctx, key, v, ttl)
	}
	return v, nil
}
