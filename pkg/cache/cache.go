package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetOrSet 泛型 cache-aside：先查缓存，未命中则回源加载并回填缓存。
// loader 返回 (T, error)，T 为值类型；rdb 为 nil 时退化为直接回源（不缓存）。
func GetOrSet[T any](ctx context.Context, rdb *redis.Client, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	// 1. 查缓存
	if rdb != nil {
		if data, err := rdb.Get(ctx, key).Bytes(); err == nil {
			var v T
			if json.Unmarshal(data, &v) == nil {
				return v, nil
			}
		}
	}

	// 2. 回源加载
	v, err := loader()
	if err != nil {
		return v, err
	}

	// 3. 回填缓存
	if rdb != nil {
		if data, err := json.Marshal(v); err == nil {
			_ = rdb.Set(ctx, key, data, ttl).Err()
		}
	}
	return v, nil
}
