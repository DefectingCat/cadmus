// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含事务管理的核心逻辑，包括：
//   - 事务的开启、提交和回滚
//   - 回调式事务执行模式
//   - 自动错误处理和资源释放
//
// 主要用途：
//
//	用于需要多个数据库操作作为一个原子单元执行的场景，
//	保证数据一致性和完整性。
//
// 注意事项：
//   - 事务回调中不应执行耗时操作，避免长事务
//   - 回调返回错误时事务会自动回滚
//   - 事务不支持嵌套，请避免在回调中再次开启事务
//
// 作者：xfy
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TransactionManager 事务管理器，封装事务控制逻辑。
//
// 提供回调式事务接口，自动处理提交和回滚，
// 简化事务使用并减少错误处理代码。
type TransactionManager struct {
	// pool 底层连接池
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