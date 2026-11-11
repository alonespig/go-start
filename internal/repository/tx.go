package repository

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// TxManager 提供事务边界。事务内通过 context 传递连接，
// 各个 repository 会自动复用同一个 *gorm.DB，无需改动方法签名。
//
//	err := tx.WithTx(ctx, func(ctx context.Context) error {
//	    if err := userRepo.Create(ctx, u); err != nil { return err }
//	    return orderRepo.Create(ctx, o)
//	})
type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTx 在一个事务中执行 fn；fn 返回 error 时自动回滚，返回 nil 时提交。
func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFromContext 优先使用事务连接，不在事务里时回退到仓库自身连接。
func dbFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}
