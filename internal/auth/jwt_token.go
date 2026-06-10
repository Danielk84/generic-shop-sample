package auth

import (
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
)

var algorithm = jwt.SigningMethodHS512

type AuthClaims struct {
	ID             int32                  `json:"id"`
	Username       string                 `json:"username"`
	PermissionType queries.PermissionType `json:"permission_type"`
	jwt.RegisteredClaims
}

func TokenEncoder(claims AuthClaims) (string, error) {
	config := internal.GetConfig()

	token := jwt.NewWithClaims(algorithm, claims)
	tokenString, err := token.SignedString(config.Opt.JWTSecretKey)
	if err != nil {
		slog.Warn("failed to encode claims", "error", err)
	}
	return tokenString, err
}

func TokenDecoder(tokenString string) (*AuthClaims, error) {
	config := internal.GetConfig()

	token, err := jwt.ParseWithClaims(
		tokenString,
		&AuthClaims{},
		func(t *jwt.Token) (any, error) { return config.Opt.JWTSecretKey, nil },
		jwt.WithStrictDecoding(),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{algorithm.Alg()}),
	)
	if err != nil {
		if err != jwt.ErrTokenExpired {
			slog.Warn("failed to decode jwt tokem", "error", err)
		}
		return nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
