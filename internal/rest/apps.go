package rest

import (
	"context"
	"net/http"

	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (s *Server) appsRouter(r chi.Router) {
	r.Get("/", s.listApps)
	r.Route("/{appId}", func(r chi.Router) {
		r.Use(s.appCtx)
		r.Get("/", s.getApp)
		r.Route("/jobs", s.jobsRouter)
	})
}

func (s *Server) listApps(w http.ResponseWriter, r *http.Request) {
	// Debug route
	Response(w, s.services.app)
}

func (s *Server) getApp(w http.ResponseWriter, r *http.Request) {
	// Debug route
	app := r.Context().Value("app").(*model.Application)
	Response(w, app)
}

func (s *Server) appCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token == "" {
			Error(w, http.StatusUnauthorized)
			return
		}

		appId := chi.URLParam(r, "appId")
		id, err := uuid.Parse(appId)
		if err != nil {
			Error(w, http.StatusNotFound)
			return
		}

		app, err := s.services.app.Get(id)
		if err != nil {
			Error(w, http.StatusNotFound)
			return
		}

		if app.AccessToken != token {
			Error(w, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "app", app)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
