package auth

import (
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal"

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
	config := internal.NewConfig()

	token := jwt.NewWithClaims(algorithm, claims)
	return token.SignedString(config.JWTSecretKey)
}

func TokenDecoder(tokenString string) (*AuthClaims, error) {
	config := internal.NewConfig()
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AuthClaims{},
		func(t *jwt.Token) (any, error) { return config.JWTSecretKey, nil },
		jwt.WithStrictDecoding(),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{algorithm.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
