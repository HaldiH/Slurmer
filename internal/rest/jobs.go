package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"net/http"

	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (s *Server) jobsRouter(r chi.Router) {
	r.Use(s.userCtx)
	r.Get("/", s.listJobs)
	r.Post("/", s.createJob)
	r.Route("/{jobId}", func(r chi.Router) {
		r.Use(s.jobCtx)
		r.Get("/", s.getJob)
		r.Delete("/", s.deleteJob)
		r.Patch("/", s.patchJob)
		r.Get("/batch", s.getBatch)
		r.Get("/out", s.getOut)
		r.Put("/status", s.updateJobStatus)
		r.Route("/files", filesRouter)
	})
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	app := getCtxApp(r.Context())
	jobs, err := s.services.job.GetAppAll(app)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		log.Error(err)
		return
	}
	Response(w, jobs)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	job := getCtxJob(r.Context())
	Response(w, job)
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := getCtxApp(ctx)
	user := ctx.Value(UserKey).(string)

	var prop slurm.BatchProperties
	if err := json.NewDecoder(r.Body).Decode(&prop); err != nil {
		Error(w, http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(&prop); err != nil {
		Error(w, http.StatusBadRequest)
		return
	}

	job, err := s.services.job.Create(user, app, &prop)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		log.Error(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	Response(w, job)
}

var updateJobMutex sync.Mutex

func (s *Server) updateJobStatus(w http.ResponseWriter, r *http.Request) {
	job := getCtxJob(r.Context())

	var status model.JobStatus
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		Error(w, http.StatusBadRequest)
		return
	}

	if !updateJobMutex.TryLock() {
		Error(w, http.StatusConflict)
		return
	}
	if err := s.services.job.UpdateStatus(job, status); err != nil {
		Error(w, http.StatusInternalServerError)
		log.Error(err)
		return
	}
	updateJobMutex.Unlock()

	Response(w, status)
}

func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := getCtxApp(ctx)
	job := getCtxJob(ctx)

	if err := s.services.job.Delete(app, job); err != nil {
		Error(w, http.StatusInternalServerError)
		log.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getOut(w http.ResponseWriter, r *http.Request) {
	job := getCtxJob(r.Context())

	file, err := os.Open(filepath.Join(job.Directory, fmt.Sprintf("slurm-%d.out", job.SlurmId)))
	if err != nil {
		log.Debug(err)
		Error(w, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, file)
}

func (s *Server) patchJob(w http.ResponseWriter, r *http.Request) {
	job := getCtxJob(r.Context())

	var patchReq model.JobPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&patchReq); err != nil {
		Error(w, http.StatusBadRequest)
		return
	}

	if patchReq.Action != nil {
		var err error
		switch *patchReq.Action {
		case model.JobPrune:
			err = s.services.job.PruneJob(job)
		case model.JobStart:
			err = s.services.job.Start(job)
		case model.JobStop:
			err = s.services.job.Stop(job)
		default:
			http.Error(w, "Unknown action: "+string(*patchReq.Action), http.StatusBadRequest)
			return
		}
		if err != nil {
			Error(w, http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	job := getCtxJob(r.Context())

	batchFile, err := os.Open(filepath.Join(job.Directory, "batch.sh"))
	if err != nil {
		log.Error(err)
		Error(w, http.StatusInternalServerError)
		return
	}
	defer batchFile.Close()

	w.Header().Set("Content-Type", "text/plain")
	if _, err := io.Copy(w, batchFile); err != nil {
		log.Error(err)
		Error(w, http.StatusInternalServerError)
		return
	}
}

func (s *Server) jobCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := getCtxApp(ctx)
		jobId, err := uuid.Parse(chi.URLParam(r, "jobId"))
		if err != nil {
			Error(w, http.StatusNotFound)
			return
		}

		job, err := s.services.job.GetApp(app, jobId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				Error(w, http.StatusNotFound)
				return
			}
			log.Panic(err)
		}

		ctx = context.WithValue(ctx, JobKey, job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("User")

		if len(user) == 0 {
			Error(w, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getCtxJob(ctx context.Context) *model.Job {
	return ctx.Value(JobKey).(*model.Job)
}
