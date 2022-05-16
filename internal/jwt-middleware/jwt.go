package jwtmiddleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

type ClientInfo struct {
	Name string
}

type CustomClaims struct {
	*jwt.RegisteredClaims
	*ClientInfo
}

type JWTMiddleware struct {
	publicKey []byte
}

func New(publicKey string) *JWTMiddleware {
	return &JWTMiddleware{
		publicKey: []byte(publicKey),
	}
}

func (jm *JWTMiddleware) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorization) != 2 ||
			strings.ToLower(authorization[0]) != "bearer" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		tokenString := authorization[1]
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			return jm.publicKey, nil
		})

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(*CustomClaims)
		ctx := context.WithValue(r.Context(), "clientInfo", claims.ClientInfo)
		log.Debug("Client authentified with name " + claims.ClientInfo.Name)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
