package auth

import (
	"generic-shop-sample/internal/logger"
	"generic-shop-sample/storage/queries"

	"github.com/golang-jwt/jwt/v5"
)

var algorithm = jwt.SigningMethodHS512

type JWTToken struct {
	Log          logger.Logger
	JWTSecretKey []byte
}

type AuthClaims struct {
	ID             string                 `json:"id"`
	Email          string                 `json:"email"`
	Name           string                 `json:"name"`
	PermissionType queries.PermissionType `json:"permission_type"`
	jwt.RegisteredClaims
}

func (j *JWTToken) Encoder(claims AuthClaims) (string, error) {
	token := jwt.NewWithClaims(algorithm, claims)
	tokenString, err := token.SignedString(j.JWTSecretKey)
	if err != nil {
		j.Log.Warn("failed to encode claims", "error", err)
	}
	return tokenString, err
}

func (j *JWTToken) Decoder(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AuthClaims{},
		func(t *jwt.Token) (any, error) { return j.JWTSecretKey, nil },
		jwt.WithStrictDecoding(),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{algorithm.Alg()}),
	)
	if err != nil {
		if err != jwt.ErrTokenExpired {
			j.Log.Warn("failed to decode jwt tokem", "error", err)
		}
		return nil, err
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
