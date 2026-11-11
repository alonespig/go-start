package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type memoryItem struct {
	data    []byte
	expires time.Time
}

// Memory 进程内缓存，适合本地开发或单元测试。
// 多实例部署请使用 Redis，避免各实例数据不一致。
type Memory struct {
	mu    sync.RWMutex
	items map[string]memoryItem
}

func NewMemory() *Memory {
	return &Memory{items: make(map[string]memoryItem)}
}

func (m *Memory) Get(_ context.Context, key string, dest any) (bool, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()
	if !ok {
		return false, nil
	}
	if !item.expires.IsZero() && time.Now().After(item.expires) {
		_ = m.Delete(context.Background(), key)
		return false, nil
	}
	return true, json.Unmarshal(item.data, dest)
}

func (m *Memory) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	m.mu.Lock()
	m.items[key] = memoryItem{data: data, expires: expires}
	m.mu.Unlock()
	return nil
}

func (m *Memory) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, k := range keys {
		delete(m.items, k)
	}
	m.mu.Unlock()
	return nil
}
