package slurmer

import (
	"context"
	"encoding/json"
	"github.com/ShinoYasx/Slurmer/pkg/slurmer"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func (srv *Server) jobsRouter(r chi.Router) {
	r.Get("/", srv.listJobs)
	r.Post("/", srv.createJob)
	r.Route("/{jobID}", func(r chi.Router) {
		r.Use(srv.JobCtx)
		r.Get("/", srv.getJob)
		r.Put("/status", srv.updateJobStatus)
	})
}

func (srv *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*slurmer.Application)

	Response(w, app.Jobs)
}

func (srv *Server) getJob(w http.ResponseWriter, r *http.Request) {
	// TODO: add slurm job properties in response
	job := r.Context().Value("job").(*slurmer.Job)
	Response(w, job)
}

func (srv *Server) createJob(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*slurmer.Application)

	jobID := uuid.New().String()

	reqBody, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	jobDir := filepath.Join(app.Directory, "jobs", jobID)

	err = os.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	batchFile, err := os.Create(filepath.Join(jobDir, "batch.sh"))
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}
	defer batchFile.Close()

	var batchProperties slurmer.BatchProperties
	err = json.Unmarshal(reqBody, &batchProperties)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	err = WriteBatch(batchFile, &batchProperties)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	job := slurmer.Job{
		Name:      batchProperties.JobName,
		Status:    slurmer.JobStatus.Stopped,
		ID:        jobID,
		Directory: jobDir,
	}

	app.Jobs.AddJob(jobID, &job)

	w.WriteHeader(http.StatusCreated)
	Response(w, &job)
}

func (srv *Server) updateJobStatus(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*slurmer.Job)

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest)
		panic(err)
	}
	defer r.Body.Close()

	var status string
	err = json.Unmarshal(reqBody, &status)
	if err != nil {
		Error(w, http.StatusBadRequest)
		panic(err)
	}

	switch status {
	case "started":
		if job.Status == slurmer.JobStatus.Stopped {
			err := handleStartJob(job)
			if err != nil {
				Error(w, http.StatusInternalServerError)
				log.Panic(err)
			}
		}
	case "stopped":
		if job.Status == slurmer.JobStatus.Started {
			err := handleStopJob(job)
			if err != nil {
				Error(w, http.StatusInternalServerError)
				log.Panic(err)
			}
		}
	}

	Response(w, status)
}

func (srv *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := ctx.Value("app").(*slurmer.Application)
	job := ctx.Value("job").(*slurmer.Job)

	// First we need to stop pending/running job
	if job.Status == slurmer.JobStatus.Started {
		err := handleStopJob(job)
		if err != nil {
			Error(w, http.StatusInternalServerError)
			log.Panic(err)
		}
	}

	app.Jobs.DeleteJob(job.ID)

	Error(w, http.StatusOK)
}

func (srv *Server) JobCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := r.Context().Value("app").(*slurmer.Application)
		jobID := chi.URLParam(r, "jobID")
		job, err := app.Jobs.GetJob(jobID)
		if err != nil {
			Error(w, http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "job", job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
