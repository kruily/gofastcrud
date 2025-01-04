package fast_jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	userID := uint(1)
	username := "test_user"
	duration := time.Minute

	// 测试创建token
	token, err := maker.CreateToken(userID, username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// 测试验证token
	claims, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	require.Equal(t, userID, claims.UserID)
	require.Equal(t, username, claims.Username)
	require.NotZero(t, claims.ExpiresAt)
}

func TestExpiredToken(t *testing.T) {
	maker, err := NewJWTMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	token, err := maker.CreateToken(1, "test_user", -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, claims)
}

func TestInvalidToken(t *testing.T) {
	maker, err := NewJWTMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	// 测试无效的token
	token := "invalid_token"
	claims, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, claims)
}
