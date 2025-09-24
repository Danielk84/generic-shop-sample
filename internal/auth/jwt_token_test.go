package auth_test

import (
	"generic-shop-sample/db/queries"
	"generic-shop-sample/internal/auth"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTToken(t *testing.T) {
	t.Run("validate claims data", func(st *testing.T) {
		claims := auth.AuthClaims{
			ID:             123,
			Username:       "adminUser",
			PermissionType: queries.Admin,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		tokenString, err := auth.TokenEncoder(claims)
		if err != nil {
			st.Fatalf("failed to encoding claims in JWTEncoder, %s", err)
		}
		decodedClaims, err := auth.TokenDecoder(tokenString)
		if err != nil {
			st.Fatalf("failed to decoding tokenString in JWTDecoder, %s", err)
		}

		if !reflect.DeepEqual(claims, *decodedClaims) {
			st.Fatalf("not equals values")
		}
	})
	t.Run("check expiration", func(st *testing.T) {
		claims := auth.AuthClaims{
			ID:             123,
			Username:       "adminUser",
			PermissionType: queries.Admin,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Second)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		tokenString, err := auth.TokenEncoder(claims)
		if err != nil {
			st.Fatalf("failed to encoding claims in JWTEncoder, %s", err)
		}
		time.Sleep(2 * time.Second)
		_, err = auth.TokenDecoder(tokenString)
		if err == nil {
			st.Fatalf("failed to return error on expiration")
		}
	})
}
