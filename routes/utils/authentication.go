package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

func CreateAuthenticationKey(subject string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": subject,
			"exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

func IsAuthenticated(data string) bool {
	token, err := jwt.Parse(data, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return false
	}

	return token.Valid
}
