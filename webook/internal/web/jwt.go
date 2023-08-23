package web

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	Id        int64
	UserAgent string
	jwt.RegisteredClaims
}

var JWTKey = []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm")
