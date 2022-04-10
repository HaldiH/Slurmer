package rest

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
		// Will create app and jobs directory under /applications/{uuid}/jobs/
		err = os.MkdirAll(jobsDir, os.ModePerm)
		if err != nil {
			return nil, err
		}

		apps.AddApp(appUUID, &model.Application{
			AccessToken: appCfg.Token,
			Directory:   appDir,
			Id:          appUUID})
	}

	jobService := service.NewJobServiceImpl(slurmClient, slurmCache, jobs)
	appService := service.NewAppServiceImpl(apps)

	srv := Server{
		config: config,
		services: Services{
			app: appService,
			job: jobService,
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

func (srv *Server) Listen() error {
	if err := srv.updateJobs(); err != nil {
		return err
	}
	go srv.heartBeat(10 * time.Second)

	addr := fmt.Sprintf("%s:%d", srv.config.Slurmer.IP, srv.config.Slurmer.Port)
	log.Infof("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, srv.router())
}

func (srv *Server) heartBeat(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		if err := srv.updateJobs(); err != nil {
			panic(err)
		}
	}
}

var firstUpdate = true

func (s *Server) updateJobs() error {
	log.Debug("Update jobs")
	defer func() { firstUpdate = false }()
	jobs, err := s.services.job.GetAll()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Status == model.JobStarted || job.SlurmJob != nil {
			slurmJob, err := s.slurmClient.GetJob(job.SlurmId)
			if err != nil {
				if err == slurm.ErrJobNotFound {
					if err := s.slurmCache.DeleteSlurmJob(job.SlurmId); err != nil {
						return err
					}
					job.SlurmId = 0
					job.SlurmJob = nil
					job.Status = model.JobStopped
				} else {
					return err
				}
			} else {
				job.SlurmJob = slurmJob
				if firstUpdate && slurmJob.JobState == "CANCELLED" {
					job.Status = model.JobStopped
				}
			}

			if err := s.jobs.UpdateJob(job); err != nil {
				return err
			}
		}
	}
	return nil
}
