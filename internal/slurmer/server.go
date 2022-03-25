package slurmer

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/slurmcli"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/pkg/slurmer"
	"github.com/go-chi/chi"
)

type Server struct {
	config      *appconfig.Config
	slurmClient slurm.Client
	router      chi.Router
	apps        *slurmer.AppsContainer
}

func New(config *appconfig.Config) (*Server, error) {
	if config.Slurmer.WorkingDir == "" {
		config.Slurmer.WorkingDir = "."
	}

	var sc slurm.Client
	var err error

	switch config.Slurmer.Connector {
	// rest client not implemented yet
	// case "slurmrest":
	// 	sc, err = slurmrest.NewRestClient(config.Slurmrest.URL)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	case "slurmcli":
		sc = slurmcli.NewCliClient()
	}

	err = os.MkdirAll(config.Slurmer.WorkingDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(config.Slurmer.WorkingDir)
	if err != nil {
		return nil, err
	}

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
			Jobs:        slurmer.NewJobsContainer(),
			ID:          appCfg.UUID})
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
	srv.router.Use(SetContentType("application/json"))
	srv.router.Route("/apps", srv.appsRouter)
}

func (srv *Server) Listen() error {
	addr := fmt.Sprintf("%s:%d", srv.config.Slurmer.IP, srv.config.Slurmer.Port)
	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, srv.router)
}
