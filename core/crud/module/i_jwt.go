package module

import "github.com/golang-jwt/jwt/v5"

type IJwt interface {
	IModule
	CreateToken(any, string) (string, error)
	VerifyToken(string) (jwt.Claims, error)
	RefreshToken(string) (string, error)
}
