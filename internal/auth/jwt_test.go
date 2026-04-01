package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试用的固定密钥（至少 32 字符）
const testSecret = "test-secret-key-for-testing-32-chars"

func TestClaims_GetJWTID(t *testing.T) {
	t.Run("returns ID when set", func(t *testing.T) {
		jti := "test-jti-123"
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ID: jti,
			},
		}
		assert.Equal(t, jti, claims.GetJWTID())
	})

	t.Run("returns empty string when ID is empty", func(t *testing.T) {
		claims := &Claims{}
		assert.Equal(t, "", claims.GetJWTID())
	})
}

func TestNewJWTService(t *testing.T) {
	config := JWTConfig{
		Secret:        testSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test",
	}

	svc := NewJWTService(config)
	assert.NotNil(t, svc)
	assert.Equal(t, config, svc.Config)
}

func TestJWTService_Generate(t *testing.T) {
	svc := NewJWTService(JWTConfig{
		Secret:        testSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test",
	})

	userID := uuid.New()
	roleID := uuid.New()

	t.Run("generates valid token", func(t *testing.T) {
		tokenString, jti, err := svc.Generate(userID, roleID)

		assert.NoError(t, err)
		assert.NotEmpty(t, tokenString)
		assert.NotEmpty(t, jti)

		// 验证 token 可以被解析
		claims, err := svc.Validate(tokenString)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, roleID, claims.RoleID)
		assert.Equal(t, jti, claims.GetJWTID())
		assert.Equal(t, "test", claims.Issuer)
	})

	t.Run("generates unique jti for each token", func(t *testing.T) {
		token1, jti1, _ := svc.Generate(userID, roleID)
		token2, jti2, _ := svc.Generate(userID, roleID)

		assert.NotEqual(t, token1, token2)
		assert.NotEqual(t, jti1, jti2)
	})
}

func TestJWTService_Validate(t *testing.T) {
	svc := NewJWTService(JWTConfig{
		Secret:        testSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test",
	})

	userID := uuid.New()
	roleID := uuid.New()

	t.Run("valid token returns claims", func(t *testing.T) {
		tokenString, _, err := svc.Generate(userID, roleID)
		require.NoError(t, err)

		claims, err := svc.Validate(tokenString)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, roleID, claims.RoleID)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		// 构造一个已过期的 token
		claims := &Claims{
			UserID: userID,
			RoleID: roleID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(testSecret))
		require.NoError(t, err)

		_, err = svc.Validate(tokenString)
		assert.Error(t, err)
	})

	t.Run("invalid signature returns error", func(t *testing.T) {
		tokenString, _, err := svc.Generate(userID, roleID)
		require.NoError(t, err)

		// 使用错误密钥验证
		wrongSvc := NewJWTService(JWTConfig{
			Secret:        "wrong-secret-key-for-testing-32-chars",
			Expiry:        time.Hour,
			RefreshExpiry: 24 * time.Hour,
			Issuer:        "test",
		})

		_, err = wrongSvc.Validate(tokenString)
		assert.Error(t, err)
	})

	t.Run("malformed token returns error", func(t *testing.T) {
		_, err := svc.Validate("not-a-valid-token")
		assert.Error(t, err)
	})

	t.Run("empty token returns error", func(t *testing.T) {
		_, err := svc.Validate("")
		assert.Error(t, err)
	})

	t.Run("wrong signing algorithm returns error", func(t *testing.T) {
		// 使用不同的签名算法构造 token
		claims := &Claims{
			UserID: userID,
			RoleID: roleID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		// 空字符串作为 none 算法的密钥
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = svc.Validate(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})
}

func TestJWTService_Refresh(t *testing.T) {
	svc := NewJWTService(JWTConfig{
		Secret:        testSecret,
		Expiry:        time.Hour,
		RefreshExpiry: 24 * time.Hour,
		Issuer:        "test",
	})

	userID := uuid.New()
	roleID := uuid.New()

	t.Run("refresh valid token returns new token", func(t *testing.T) {
		tokenString, oldJti, err := svc.Generate(userID, roleID)
		require.NoError(t, err)

		newToken, newJti, err := svc.Refresh(tokenString)

		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		assert.NotEqual(t, oldJti, newJti)

		// 验证新 token 有效
		claims, err := svc.Validate(newToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("refresh expired token within window - requires special parsing", func(t *testing.T) {
		// 注意：当前的 Refresh 实现调用 Validate，而 Validate 会拒绝过期 token
		// 所以这个测试验证的是当前行为：过期 token 无法刷新
		// 如果需要支持过期 token 刷新，需要修改 Refresh 使用不同的解析方法

		// 构造一个刚过期的 token
		claims := &Claims{
			UserID: userID,
			RoleID: roleID,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-30 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    "test",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(testSecret))
		require.NoError(t, err)

		// 当前行为：过期 token 被 Validate 拒绝
		_, _, err = svc.Refresh(tokenString)
		assert.Error(t, err)
	})

	t.Run("refresh token outside window - requires special parsing", func(t *testing.T) {
		// 构造一个超出刷新窗口的 token
		claims := &Claims{
			UserID: userID,
			RoleID: roleID,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-30 * time.Hour)),
				Issuer:    "test",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(testSecret))
		require.NoError(t, err)

		// 当前行为：过期 token 被 Validate 拒绝
		_, _, err = svc.Refresh(tokenString)
		assert.Error(t, err)
	})

	t.Run("refresh invalid token fails", func(t *testing.T) {
		_, _, err := svc.Refresh("invalid-token")
		assert.Error(t, err)
	})
}