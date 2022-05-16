package rest

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ShinoYasx/Slurmer/pkg/utils"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/internal/persistent"
	"github.com/ShinoYasx/Slurmer/internal/service"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/slurmcli"
	"github.com/go-chi/chi"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const appsDir = "applications"

type Services struct {
	app service.AppService
	job service.JobService
}

type requestContextKey uint

const (
	appKey requestContextKey = iota
	jobKey
)

type Server struct {
	config      *appconfig.Config
	services    Services
	slurmClient slurm.Client
	slurmCache  containers.SlurmCache
	jobs        containers.JobsContainer
}

func NewServer(config *appconfig.Config) (*Server, error) {
	if config.Slurmer.WorkingDir == "" {
		config.Slurmer.WorkingDir = "."
	}

	var slurmClient slurm.Client

	switch config.Slurmer.Connector {
	// rest client not implemented yet
	// case "slurmrest":
	// 	sc, err = slurmrest.NewRestClient(config.Slurmrest.URL)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	case "slurmcli":
		slurmClient = slurmcli.NewCliClient()
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

	db, err := gorm.Open(sqlite.Open("slurmer.db"), &gorm.Config{
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

	if err := os.MkdirAll(appsDir, os.ModePerm); err != nil {
		return nil, err
	}

	for _, appCfg := range config.Slurmer.Applications {
		appUUID, err := uuid.Parse(appCfg.UUID)
		if err != nil {
			return nil, err
		}
		appDir := filepath.Join(appsDir, appCfg.UUID)
		jobsDir := filepath.Join(appDir, "jobs")
		templatesDir := path.Join(appDir, "templates")

		// Will create app and jobs directory under /applications/{uuid}/jobs/
		if err = os.MkdirAll(jobsDir, os.ModePerm); err != nil {
			return nil, err
		}

		if err := os.MkdirAll(templatesDir, os.ModePerm); err != nil {
			return nil, err
		}

		if err := utils.CopyDirectory(cfgTemplatesDir, templatesDir, false); err != nil {
			return nil, err
		}

		apps.AddApp(appUUID, &model.Application{
			AccessToken: appCfg.Token,
			Directory:   appDir,
			Id:          appUUID})
	}

	srv := Server{
		config: config,
		services: Services{
			app: service.NewAppService(apps),
			job: service.NewJobService(slurmClient, slurmCache, jobs),
		},
		slurmClient: slurmClient,
		slurmCache:  slurmCache,
		jobs:        jobs,
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
	r.Route("/apps", s.appsRouter)
	return r
}

func (s *Server) Listen() error {
	if err := s.services.job.PollJobsStatus(); err != nil {
		return err
	}
	go s.heartBeat(10 * time.Second)

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
