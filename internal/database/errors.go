// Package database 提供了 Cadmus 数据库访问层的实现。
//
// 该文件包含数据库错误处理的共享工具函数。
//
// 主要用途：
//
//	统一数据库错误判断逻辑，避免代码重复。
//
// 作者：xfy
package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation 检查是否为指定约束的唯一性冲突错误。
//
// 用于判断 INSERT/UPDATE 操作是否违反了唯一约束，
// 如用户名重复、邮箱重复、Slug 重复等场景。
//
// 参数：
//   - err: 原始错误对象
//   - constraintName: 约束名称（如 "users_username_key", "posts_slug_key"）
//
// 返回值：
//   - true: 错误是指定约束的唯一性冲突
//   - false: 错误不是该约束的唯一性冲突，或错误为 nil
//
// 使用示例：
//
//	if IsUniqueViolation(err, "users_username_key") {
//	    return user.ErrUserAlreadyExists
//	}
func IsUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
	}
	return false
}

// IsForeignKeyViolation 检查是否为外键约束冲突错误。
//
// 用于判断 INSERT/UPDATE 操作是否违反了外键约束，
// 如引用不存在的用户、文章等场景。
//
// 参数：
//   - err: 原始错误对象
//   - constraintName: 约束名称（如 "comments_post_id_fkey"）
//
// 返回值：
//   - true: 错误是指定约束的外键冲突
//   - false: 错误不是该约束的外键冲突，或错误为 nil
//
// 使用示例：
//
//	if IsForeignKeyViolation(err, "comments_post_id_fkey") {
//	    return comment.ErrPostNotFound
//	}
func IsForeignKeyViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503" && pgErr.ConstraintName == constraintName
	}
	return false
}