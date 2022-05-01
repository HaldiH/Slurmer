package rest

import (
	"context"
	"encoding/json"
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
	r.Get("/", s.listJobs)
	r.Post("/", s.createJob)
	r.Route("/{jobId}", func(r chi.Router) {
		r.Use(s.jobCtx)
		r.Get("/", s.getJob)
		r.Delete("/", s.deleteJob)
		r.Patch("/", s.patchJob)
		r.Get("/batch", s.getBatch)
		r.Put("/status", s.updateJobStatus)
		r.Route("/files", filesRouter)
	})
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*model.Application)
	jobs := Must(s.services.job.GetAppAll(app))(w)
	Response(w, jobs)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*model.Job)
	Response(w, job)
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*model.Application)

	reqBody := Must(io.ReadAll(r.Body))(w)
	defer r.Body.Close()

	var prop slurm.BatchProperties
	MustNone(json.Unmarshal(reqBody, &prop), w)

	validate := validator.New()
	if err := validate.Struct(&prop); err != nil {
		Error(w, http.StatusBadRequest)
		return
	}

	job := Must(s.services.job.Create(app, &prop))(w)
	w.WriteHeader(http.StatusCreated)
	Response(w, job)
}

var updateJobMutex sync.Mutex

func (s *Server) updateJobStatus(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*model.Job)

	reqBody := Must(io.ReadAll(r.Body))(w)
	defer r.Body.Close()

	var status model.JobStatus
	if err := json.Unmarshal(reqBody, &status); err != nil {
		log.Warn(err)
		Error(w, http.StatusBadRequest)
		return
	}

	if !updateJobMutex.TryLock() {
		Error(w, http.StatusConflict)
		return
	}
	MustNone(s.services.job.UpdateStatus(job, status), w)
	updateJobMutex.Unlock()

	Response(w, status)
}

func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := ctx.Value("app").(*model.Application)
	job := ctx.Value("job").(*model.Job)

	MustNone(s.services.job.Delete(app, job), w)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) patchJob(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*model.Job)

	reqBody := Must(io.ReadAll(r.Body))(w)
	defer r.Body.Close()

	var patchReq model.JobPatchRequest
	if err := json.Unmarshal(reqBody, &patchReq); err != nil {
		Error(w, http.StatusBadRequest)
		return
	}

	if patchReq.Action != nil {
		switch *patchReq.Action {
		case model.JobPrune:
			MustNone(s.services.job.PruneJob(job), w)
		case model.JobStart:
			MustNone(s.services.job.Start(job), w)
		case model.JobStop:
			MustNone(s.services.job.Stop(job), w)
		}
	}
}

func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*model.Job)

	batchFile := Must(os.Open(filepath.Join(job.Directory, "batch.sh")))(w)
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
		app := r.Context().Value("app").(*model.Application)
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
			panic(err)
		}

		ctx := context.WithValue(r.Context(), "job", job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
