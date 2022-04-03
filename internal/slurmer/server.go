package slurmer

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/slurmcli"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/go-chi/chi"
)

type Server struct {
	config      *appconfig.Config
	slurmClient slurm.Client
	router      chi.Router
	apps        *AppsContainer
	jobs        JobsContainer
	slurmCache  *SlurmCache
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
	default:
		log.Fatal("Unimplemented slurm controller: ", config.Slurmer.Connector)
	}

	err = os.MkdirAll(config.Slurmer.WorkingDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(config.Slurmer.WorkingDir)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open("slurmer.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	appsDir := "applications"
	apps := NewAppsContainer()
	err = os.MkdirAll(appsDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	for _, appCfg := range config.Slurmer.Applications {
		appUUID := appCfg.UUID
		appDir := filepath.Join(appsDir, appUUID)
		jobsDir := filepath.Join(appDir, "jobs")
		// Will create app and jobs directory under /applications/{uuid}/jobs/
		err = os.MkdirAll(jobsDir, os.ModePerm)
		if err != nil {
			return nil, err
		}

		apps.AddApp(appCfg.UUID, &Application{
			AccessToken: appCfg.Token,
			Directory:   appDir,
			ID:          appUUID})
	}

	persistentJobs, err := NewPersistentJobs(db)
	if err != nil {
		return nil, err
	}

	slurmCache, err := NewSlurmCache(db)
	if err != nil {
		return nil, err
	}

	srv := Server{
		config:      config,
		slurmClient: sc,
		router:      chi.NewRouter(),
		apps:        apps,
		jobs:        persistentJobs,
		slurmCache:  slurmCache,
	}

	srv.registerRoutes()

	return &srv, nil
}

func (srv *Server) registerRoutes() {
	srv.router.Use(SetContentType("application/json"))
	srv.router.Route("/apps", srv.appsRouter)
}

func (srv *Server) Listen() error {
	srv.updateJobs()
	go srv.heartBeat(10 * time.Second)

	addr := fmt.Sprintf("%s:%d", srv.config.Slurmer.IP, srv.config.Slurmer.Port)
	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, srv.router)
}

func (srv *Server) heartBeat(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		srv.updateJobs()
	}
}
