// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含数据库连接池管理的核心逻辑，包括：
//   - 连接池的创建和配置
//   - 连接池的生命周期管理
//   - 连接统计和健康检查
//
// 主要用途：
//
//	用于管理 PostgreSQL 数据库连接，提供高效的连接池复用机制。
//
// 注意事项：
//   - 使用前需确保配置正确加载
//   - 连接池是并发安全的，可在多个 goroutine 中使用
//   - 应用退出时应调用 Close() 释放资源
//
// 作者：xfy
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 数据库连接池，封装 pgxpool.Pool 提供统一的连接管理接口。
//
// 连接池维护一组可复用的数据库连接，减少连接创建开销，
// 提高数据库访问性能。支持连接生命周期管理和统计监控。
type Pool struct {
	*pgxpool.Pool
}

// NewPool 创建新的数据库连接池
func NewPool(ctx context.Context, cfg Config) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// 配置连接池参数
	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 验证连接
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{pool}, nil
}

// Close 关闭连接池
func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

// Stats 返回连接池统计信息
func (p *Pool) Stats() *pgxpool.Stat {
	return p.Pool.Stat()
}

// WaitForConnection 等待连接可用
func (p *Pool) WaitForConnection(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if int64(p.Pool.Stat().MaxConns()) > p.Pool.Stat().AcquireCount() {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}