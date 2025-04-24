package fast_jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (maker *JWTMaker) CreateTokenUUID(uuid, username string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   uuid,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}
