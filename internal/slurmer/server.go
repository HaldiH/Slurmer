package slurmer

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/slurmer"
	"github.com/go-chi/chi"
)

type Server struct {
	config      *appconfig.Config
	slurmClient *slurm.Client
	router      chi.Router
	apps        *slurmer.AppsContainer
}

func New(config *appconfig.Config) (*Server, error) {
	if config.Slurmer.WorkingDir == "" {
		config.Slurmer.WorkingDir = "."
	}

	sc, err := slurm.NewClient(config.Slurmrest.URL)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(config.Slurmer.WorkingDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	os.Chdir(config.Slurmer.WorkingDir)

	appsDir := "applications"
	apps := slurmer.NewAppsContainer()
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
			AccessToken: appCfg.Token,
			Directory:   appDir,
			Jobs:        slurmer.NewJobsContainer()})
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
