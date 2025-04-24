package fast_jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTMaker struct {
	secretKey string
	Expire    string
	duration  time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID   any    `json:"user_id"`
	Username string `json:"username"`
}

// NewJWTMaker 创建一个新的JWT maker
func NewJWTMaker(secretKey string, expire string) (*JWTMaker, error) {
	if len(secretKey) > 32 {
		return nil, errors.New("secret key must be less than 32 characters")
	}
	if expire == "" {
		return nil, errors.New("expire must be greater than 0")
	}
	duration, err := time.ParseDuration(expire)
	if err != nil {
		return nil, err
	}
	return &JWTMaker{secretKey: secretKey, Expire: expire, duration: duration}, nil
}

// CreateToken 创建一个新的token
func (maker *JWTMaker) CreateToken(userID any, username string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   userID,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}

// VerifyToken 验证token并返回claims
func (maker *JWTMaker) VerifyToken(tokenString string) (jwt.Claims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken 刷新token
func (maker *JWTMaker) RefreshToken(tokenString string) (string, error) {
	claims, err := maker.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}
	c := claims.(*Claims)

	expirationTime := time.Now().Add(maker.duration)
	c.ExpiresAt = jwt.NewNumericDate(expirationTime)
	c.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(maker.secretKey))
}
