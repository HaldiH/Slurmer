package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func SetContentType(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

func Response(w http.ResponseWriter, v interface{}) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}
	_, err = w.Write(jsonData)
	if err != nil {
		panic(err)
	}
}

func Error(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func Must[T any](v T, err error) func(http.ResponseWriter) T {
	return func(w http.ResponseWriter) T {
		if err != nil {
			Error(w, http.StatusInternalServerError)
			log.Error(err)
		}
		return v
	}
}

func MustNone(err error, w http.ResponseWriter) {
	Must[any](nil, err)(w)
}
