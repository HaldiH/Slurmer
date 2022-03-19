package slurmer

import (
	"context"
	"github.com/ShinoYasx/Slurmer/pkg/slurmer"
	"net/http"

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
	Ok(w, srv.apps)
}

func (srv *Server) getApp(w http.ResponseWriter, r *http.Request) {
	// Debug route
	app, ok := r.Context().Value("app").(*slurmer.Application)
	if !ok {
		panic("Requested resource is not an Application")
	}

	Ok(w, app)
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
		if app.AccessToken != token {
			if app.AccessToken == "" {
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
