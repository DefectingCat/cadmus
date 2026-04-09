// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含数据库扫描的通用工具，用于简化 Repository 层的数据扫描操作。
//
// 主要功能：
//   - RowScanner 接口：统一 pgx.Row 和 pgx.Rows 的扫描操作
//   - ScanAll 泛型函数：简化多行数据扫描
//
// 作者：xfy
package database

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

// RowScanner 通用行扫描接口。
//
// 该接口抽象了 pgx.Row 和 pgx.Rows 的共同方法，
// 使得可以用同一个扫描函数处理单行和多行查询结果。
//
// pgx.Row 和 pgx.Rows 都实现了 Scan 方法，因此都满足此接口。
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// ScanAll 扫描多行数据到切片。
//
// 该泛型函数简化了从 pgx.Rows 扫描多个实体到切片的操作。
// 它自动处理 rows 的关闭和迭代。
//
// 类型参数：
//   - T: 实体类型
//
// 参数：
//   - rows: pgx 查询返回的行对象
//   - scanFunc: 单行扫描函数，接收 RowScanner 返回 *T
//
// 返回值：
//   - []*T: 扫描结果切片
//   - error: 扫描过程中的错误
//
// 使用示例：
//
//	users, err := ScanAll(rows, func(row RowScanner) (*user.User, error) {
//	    u := &user.User{}
//	    err := row.Scan(&u.ID, &u.Username, &u.Email)
//	    return u, err
//	})
func ScanAll[T any](rows pgx.Rows, scanFunc func(RowScanner) (*T, error)) ([]*T, error) {
	defer rows.Close()

	result := make([]*T, 0)
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, item)
	}

	// 检查迭代过程中是否有错误
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}

// ScanOne 扫描单行数据。
//
// 该函数用于处理单行查询结果，统一错误处理。
//
// 参数：
//   - row: pgx 查询返回的单行对象
//   - scanFunc: 扫描函数，接收 RowScanner 返回 *T
//
// 返回值：
//   - *T: 扫描结果
//   - error: 扫描过程中的错误（包括 pgx.ErrNoRows）
//
// 使用示例：
//
//	user, err := ScanOne(row, func(r RowScanner) (*user.User, error) {
//	    u := &user.User{}
//	    err := r.Scan(&u.ID, &u.Username, &u.Email)
//	    return u, err
//	})
func ScanOne[T any](row pgx.Row, scanFunc func(RowScanner) (*T, error)) (*T, error) {
	result, err := scanFunc(row)
	if err != nil {
		return nil, err
	}
	return result, nil
}
