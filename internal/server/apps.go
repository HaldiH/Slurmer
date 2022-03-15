package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ShinoYasx/Slurmer/internal/slurmer"
	"github.com/go-chi/chi"
)

func (srv *Server) appsRouter(r chi.Router) {
	r.Get("/", srv.listApps)
	r.Route("/{appID}", func(r chi.Router) {
		r.Use(srv.AppCtx)
		r.Get("/", srv.getApp)
		r.Route("/jobs", srv.jobsRouter)
	})
}

func (srv *Server) listApps(w http.ResponseWriter, r *http.Request) {
	// Debug route
	jsonData, err := json.Marshal(srv.apps)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	w.Write(jsonData)
}

func (srv *Server) getApp(w http.ResponseWriter, r *http.Request) {
	// Debug route
	app := r.Context().Value("app").(*slurmer.Application)

	jsonData, err := json.Marshal(app)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	w.Write(jsonData)
}

func (srv *Server) AppCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appID := chi.URLParam(r, "appID")
		app, err := srv.apps.GetApp(appID)
		if err != nil {
			Error(w, http.StatusNotFound)
			return
		}
		token := r.Header.Get("X-Auth-Token")
		if app.Token != token {
			if app.Token == "" {
				Error(w, http.StatusUnauthorized)
			} else {
				Error(w, http.StatusForbidden)
			}
			return
		}

		ctx := context.WithValue(r.Context(), "app", app)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
