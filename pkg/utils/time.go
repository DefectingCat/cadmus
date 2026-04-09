// Package utils 提供通用工具函数。
//
// 该文件包含时间处理的工具函数。
//
// 作者：xfy
package utils

import "time"

// InitTimestamps 初始化时间戳，如果为零值则使用当前时间。
//
// 用于创建实体时统一处理 CreatedAt 和 UpdatedAt 字段。
// 如果传入的时间已经设置，则保持不变；如果为零值，则设置为当前时间。
//
// 参数：
//   - created: 创建时间，如果为零值则设置为当前时间
//   - updated: 更新时间，如果为零值则设置为当前时间
//
// 返回值：
//   - created: 初始化后的创建时间
//   - updated: 初始化后的更新时间
//
// 使用示例：
//
//	entity.CreatedAt, entity.UpdatedAt = utils.InitTimestamps(entity.CreatedAt, entity.UpdatedAt)
func InitTimestamps(created, updated time.Time) (time.Time, time.Time) {
	now := time.Now()
	if created.IsZero() {
		created = now
	}
	if updated.IsZero() {
		updated = now
	}
	return created, updated
}

// NormalizeTime 规范化时间字段。
//
// 如果时间为零值，返回当前时间；否则返回原值。
//
// 参数：
//   - t: 待检查的时间值
//
// 返回值：
//   - 如果 t 为零值，返回当前时间；否则返回 t
func NormalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
