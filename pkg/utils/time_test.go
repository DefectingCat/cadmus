package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitTimestamps(t *testing.T) {
	t.Run("both zero values", func(t *testing.T) {
		before := time.Now()
		created, updated := InitTimestamps(time.Time{}, time.Time{})
		after := time.Now()

		// 两个时间都应该被设置为当前时间
		assert.True(t, created.After(before) || created.Equal(before), "created should be now or after")
		assert.True(t, created.Before(after) || created.Equal(after), "created should be before or equal to after")
		assert.True(t, updated.After(before) || updated.Equal(before), "updated should be now or after")
		assert.True(t, updated.Before(after) || updated.Equal(after), "updated should be before or equal to after")
	})

	t.Run("created non-zero, updated zero", func(t *testing.T) {
		fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		before := time.Now()
		created, updated := InitTimestamps(fixedTime, time.Time{})
		after := time.Now()

		// created 应保持不变
		assert.Equal(t, fixedTime, created, "created should remain unchanged")
		// updated 应被设置为当前时间
		assert.True(t, updated.After(before) || updated.Equal(before), "updated should be now")
		assert.True(t, updated.Before(after) || updated.Equal(after), "updated should be around now")
	})

	t.Run("created zero, updated non-zero", func(t *testing.T) {
		fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		before := time.Now()
		created, updated := InitTimestamps(time.Time{}, fixedTime)
		after := time.Now()

		// created 应被设置为当前时间
		assert.True(t, created.After(before) || created.Equal(before), "created should be now")
		assert.True(t, created.Before(after) || created.Equal(after), "created should be around now")
		// updated 应保持不变
		assert.Equal(t, fixedTime, updated, "updated should remain unchanged")
	})

	t.Run("both non-zero", func(t *testing.T) {
		createdTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		updatedTime := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
		created, updated := InitTimestamps(createdTime, updatedTime)

		// 两个时间都应保持不变
		assert.Equal(t, createdTime, created, "created should remain unchanged")
		assert.Equal(t, updatedTime, updated, "updated should remain unchanged")
	})
}

func TestNormalizeTime(t *testing.T) {
	t.Run("zero value returns now", func(t *testing.T) {
		before := time.Now()
		result := NormalizeTime(time.Time{})
		after := time.Now()

		assert.True(t, result.After(before) || result.Equal(before), "result should be now or after")
		assert.True(t, result.Before(after) || result.Equal(after), "result should be before or equal to after")
	})

	t.Run("non-zero value returns original", func(t *testing.T) {
		fixedTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
		result := NormalizeTime(fixedTime)

		assert.Equal(t, fixedTime, result, "result should be the original time")
	})
}