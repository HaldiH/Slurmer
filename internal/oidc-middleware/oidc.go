package oidcmiddleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

type ClientInfo struct {
}

type CustomClaims struct {
	jwt.RegisteredClaims
	ClientInfo
}

type OIDCMiddleware struct {
	verifier *oidc.IDTokenVerifier
}

func New(config *appconfig.Config) (*OIDCMiddleware, error) {
	ctx := context.TODO()
	provider, err := oidc.NewProvider(ctx, config.OIDC.Issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.OIDC.Audience,
	})

	mw := &OIDCMiddleware{
		verifier: verifier,
	}

	return mw, nil
}

func (jm *OIDCMiddleware) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		error := func(error string, code int) {
			http.Error(w, error, code)
		}

		ctx := r.Context()
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer") {
			error(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authorization, "Bearer")
		idToken, err := jm.verifier.Verify(ctx, tokenString)
		if err != nil {
			error("Invalid token", http.StatusForbidden)
			return
		}

		var claims CustomClaims

		if err := idToken.Claims(&claims); err != nil {
			log.Error(err)
			error("Cannot extract JWT claims", http.StatusBadRequest)
			return
		}

		res, err := json.Marshal(&claims)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(res))

		ctx = context.WithValue(ctx, "customClaims", &claims.ClientInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
