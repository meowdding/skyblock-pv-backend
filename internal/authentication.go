package internal

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const guestAuthenticationKey = "00000000000000000000000000000000"

type AuthenticationContext struct {
	Requester   string
	BypassCache bool
	IsGuest     bool
}

func CreateGuestAuthenticationKey(ctx RouteContext, bypassCache bool) (string, error) {
	return CreateAuthenticationKey(ctx, guestAuthenticationKey, bypassCache)
}

func CreateAuthenticationKey(ctx RouteContext, subject string, bypassCache bool) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":    subject,
			"exp":    jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			"bypass": bypassCache,
		},
	)
	return token.SignedString([]byte(ctx.Config.JwtToken))
}

func GetAuthenticatedContext(ctx RouteContext, data string) *AuthenticationContext {
	token, err := jwt.Parse(data, func(token *jwt.Token) (interface{}, error) {
		return []byte(ctx.Config.JwtToken), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil || !token.Valid {
		return nil
	}

	sub, err := token.Claims.GetSubject()
	bypass, ok := token.Claims.(jwt.MapClaims)["bypass"].(bool)

	if err != nil {
		return nil
	}

	return &AuthenticationContext{
		Requester:   sub,
		BypassCache: bypass && ok,
		IsGuest:     sub == guestAuthenticationKey,
	}
}
