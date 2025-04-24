package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func CreateAuthenticationKey(ctx RouteContext, subject string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": subject,
			"exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	)
	return token.SignedString([]byte(ctx.Config.JwtToken))
}

func IsAuthenticated(ctx RouteContext, data string) bool {
	token, err := jwt.Parse(data, func(token *jwt.Token) (interface{}, error) {
		return []byte(ctx.Config.JwtToken), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return false
	}

	return token.Valid
}
