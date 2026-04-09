package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultJWTConfig(t *testing.T) {
	// 保存原始环境变量
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	t.Run("returns error when JWT_SECRET not set", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET")

		config, err := DefaultJWTConfig()

		assert.Error(t, err)
		assert.Equal(t, "JWT_SECRET environment variable is required", err.Error())
		assert.Empty(t, config.Secret)
	})

	t.Run("returns error when secret is too short", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "too-short")

		config, err := DefaultJWTConfig()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 32 characters")
		assert.Empty(t, config.Secret)
	})

	t.Run("returns config when secret is exactly 32 characters", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "12345678901234567890123456789012") // exactly 32 chars

		config, err := DefaultJWTConfig()

		assert.NoError(t, err)
		assert.Equal(t, "12345678901234567890123456789012", config.Secret)
		assert.Equal(t, "cadmus", config.Issuer)
	})

	t.Run("returns config when secret is longer than 32 characters", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-more-than-32-chars")

		config, err := DefaultJWTConfig()

		assert.NoError(t, err)
		assert.Equal(t, "this-is-a-very-long-secret-key-more-than-32-chars", config.Secret)
		assert.Equal(t, "cadmus", config.Issuer)
	})

	t.Run("uses default values", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "test-secret-with-at-least-32-characters")

		config, err := DefaultJWTConfig()

		assert.NoError(t, err)
		assert.Equal(t, "cadmus", config.Issuer)
		assert.Equal(t, 24*60*60*1000*1000*1000, int(config.Expiry))          // 24 hours in nanoseconds
		assert.Equal(t, 7*24*60*60*1000*1000*1000, int(config.RefreshExpiry)) // 7 days in nanoseconds
	})
}

func TestMustJWTConfig(t *testing.T) {
	// 保存原始环境变量
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	t.Run("panics when JWT_SECRET not set", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET")

		defer func() {
			if r := recover(); r != nil {
				assert.Contains(t, r.(error).Error(), "JWT_SECRET")
			}
		}()

		MustJWTConfig()
		t.Error("expected panic but didn't get one")
	})

	t.Run("returns config when secret is valid", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "valid-secret-key-with-at-least-32-characters")

		config := MustJWTConfig()

		assert.Equal(t, "valid-secret-key-with-at-least-32-characters", config.Secret)
	})
}
