package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

func (srv *Server) appsRouter(r chi.Router) {
	r.Get("/", srv.listApps)
	r.Route("/{appID}", func(r chi.Router) {
		r.Use(srv.AppCtx)
		r.Route("/jobs", srv.jobsRouter)
	})
}

func (srv *Server) listApps(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented)
}

func (srv *Server) AppCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appID := chi.URLParam(r, "appID")
		app := srv.apps[appID]
		ctx := context.WithValue(r.Context(), "app", app)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
