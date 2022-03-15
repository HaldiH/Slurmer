package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/slurmer"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/go-chi/chi"
)

type AppsContainer map[string]*slurmer.Application

func NewAppsContainer() *AppsContainer {
	c := make(AppsContainer)
	return &c
}

func (c *AppsContainer) GetApp(id string) (*slurmer.Application, error) {
	app := (*c)[id]
	if app == nil {
		return nil, errors.New("Cannot find app with id " + id)
	}
	return app, nil
}

func (c *AppsContainer) AddApp(id string, app *slurmer.Application) {
	(*c)[id] = app
}

type Server struct {
	config      *appconfig.Config
	slurmClient *slurm.Client
	router      chi.Router
	apps        *AppsContainer
}

func New(config *appconfig.Config) (*Server, error) {
	sc, err := slurm.NewClient(config.Slurmrest.URL)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(config.Slurmer.WorkingDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	os.Chdir(config.Slurmer.WorkingDir)

	appsDir := filepath.Join(config.Slurmer.WorkingDir, "applications")
	apps := NewAppsContainer()
	err = os.MkdirAll(appsDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	for _, appCfg := range config.Slurmer.Applications {
		appDir := filepath.Join(appsDir, appCfg.UUID)
		jobsDir := filepath.Join(appDir, "jobs")
		// Will create app and jobs directory under /applications/{uuid}/jobs/
		err = os.MkdirAll(jobsDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		apps.AddApp(appCfg.UUID, &slurmer.Application{
			Token:     appCfg.Token,
			Directory: appDir,
			Jobs:      slurmer.NewJobsContainer()})
	}

	srv := Server{
		config:      config,
		slurmClient: sc,
		router:      chi.NewRouter(),
		apps:        apps,
	}

	srv.registerRoutes()

	return &srv, nil
}

func (srv *Server) registerRoutes() {
	srv.router.Route("/apps", srv.appsRouter)
}

func (srv *Server) Listen() error {
	addr := fmt.Sprintf("%s:%d", srv.config.Slurmer.IP, srv.config.Slurmer.Port)
	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, srv.router)
}
