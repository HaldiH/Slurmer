package rest

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/containers"
	oidcmiddleware "github.com/ShinoYasx/Slurmer/internal/oidc-middleware"
	"github.com/ShinoYasx/Slurmer/internal/persistent"
	"github.com/ShinoYasx/Slurmer/internal/service"
	"github.com/ShinoYasx/Slurmer/pkg/cliexecutor"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/slurmcli"
	"github.com/google/uuid"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Services struct {
	app service.AppService
	job service.JobService
}

type requestCtxKey uint

const (
	AppKey requestCtxKey = iota
	JobKey
	UserKey
	ClientInfoKey
)

type Server struct {
	config         *appconfig.Config
	services       Services
	slurmClient    slurm.Client
	slurmCache     containers.SlurmCache
	jobs           containers.JobsContainer
	authMiddleware *oidcmiddleware.OIDCMiddleware
}

func NewServer(config *appconfig.Config) (*Server, error) {
	if config.Slurmer.WorkingDir == "" {
		config.Slurmer.WorkingDir = "."
	}

	executor := cliexecutor.NewExecutor(config.Slurmer.ExecutorPath)

	var slurmClient slurm.Client
	switch config.Slurmer.Connector {
	// rest client not implemented yet
	// case "slurmrest":
	// 	sc, err = slurmrest.NewRestClient(config.Slurmrest.URL)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	case "slurmcli":
		slurmClient = slurmcli.NewCliClient(executor)
	default:
		log.Fatal("Unimplemented slurm controller: ", config.Slurmer.Connector)
	}

	cfgTemplatesDir, err := filepath.Abs(config.Slurmer.TemplatesDir)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(config.Slurmer.WorkingDir, os.ModePerm); err != nil {
		return nil, err
	}

	if err := os.Chdir(config.Slurmer.WorkingDir); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(appconfig.SlurmerDB), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		return nil, err
	}

	slurmCache, err := persistent.NewSlurmCache(db)
	if err != nil {
		return nil, err
	}

	apps := NewAppsMap()

	jobs, err := persistent.NewPersistentJobs(db)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(appconfig.AppsDir, os.ModePerm); err != nil {
		return nil, err
	}

	for _, appCfg := range config.Slurmer.Applications {
		app := model.Application{
			Name:        appCfg.Name,
			AccessToken: appCfg.Token,
			Id:          uuid.MustParse(appCfg.UUID),
		}

		if err := service.InitAppDir(&app, cfgTemplatesDir); err != nil {
			return nil, err
		}
		apps.AddApp(app.Id, &app)
	}

	var authMiddleware *oidcmiddleware.OIDCMiddleware
	if config.OIDC.Enabled {
		authMiddleware, err = oidcmiddleware.New(config)
		if err != nil {
			return nil, err
		}
	}

	appService, err := service.NewAppService(apps, config.Slurmer.ConfigPath, cfgTemplatesDir)
	if err != nil {
		return nil, err
	}

	srv := Server{
		config: config,
		services: Services{
			app: appService,
			job: service.NewJobService(slurmClient, slurmCache, jobs, executor),
		},
		slurmClient:    slurmClient,
		slurmCache:     slurmCache,
		jobs:           jobs,
		authMiddleware: authMiddleware,
	}

	return &srv, nil
}

func (s *Server) router() http.Handler {
	r := chi.NewRouter()
	r.Use(SetContentType("application/json"))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infof("%s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
	})
	if s.config.OIDC.Enabled {
		r.Use(s.authMiddleware.Authentication)
	}
	r.Route("/apps", s.appsRouter)
	return r
}

func (s *Server) Listen() error {
	if err := s.services.job.PollJobsStatus(); err != nil {
		return err
	}
	go s.heartBeat(time.Duration(s.config.Slurmer.PollInterval) * time.Second)

	addr := fmt.Sprintf("%s:%s", s.config.Slurmer.IP, s.config.Slurmer.Port)
	log.Infof("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, s.router())
}

func (s *Server) heartBeat(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		if err := s.services.job.PollJobsStatus(); err != nil {
			panic(err)
		}
	}
}
