package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (s *Server) appsRouter(r chi.Router) {
	r.Get("/", s.listApps)
	r.Post("/", s.createApp)
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
	app := getCtxApp(r.Context())
	Response(w, app)
}

func (s *Server) createApp(w http.ResponseWriter, r *http.Request) {
	var app model.Application

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&app); err != nil {
		http.Error(w, "Error in application body", http.StatusBadRequest)
		return
	}

	if err := s.services.app.Add(&app); err != nil {
		http.Error(w, "Cannot create application", http.StatusInternalServerError)
		log.Error(err)
		return
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&app); err != nil {
		log.Error(err)
		http.Error(w, "Cannot show application", http.StatusInternalServerError)
		return
	}
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
			if err == containers.ErrAppNotFound {
				Error(w, http.StatusNotFound)
			} else {
				Error(w, http.StatusInternalServerError)
				log.Error(err)
			}
			return
		}

		if app.AccessToken != token {
			Error(w, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), AppKey, app)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getCtxApp(ctx context.Context) *model.Application {
	return ctx.Value(AppKey).(*model.Application)
}
