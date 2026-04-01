package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TransactionManager 事务管理器，封装事务控制逻辑
type TransactionManager struct {
	pool *Pool
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(pool *Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// WithTransaction 在事务中执行回调函数，自动处理提交和回滚
// 如果回调返回错误，事务会自动回滚；否则提交事务
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// defer 回滚，如果事务已提交则回滚无害
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err // 回滚会在 defer 中执行
	}

	return tx.Commit(ctx)
}

// Pool 返回底层连接池
func (tm *TransactionManager) Pool() *Pool {
	return tm.pool
}