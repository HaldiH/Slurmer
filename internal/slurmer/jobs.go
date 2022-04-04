package slurmer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (srv *Server) jobsRouter(r chi.Router) {
	r.Get("/", srv.listJobs)
	r.Post("/", srv.createJob)
	r.Route("/{jobId}", func(r chi.Router) {
		r.Use(srv.JobCtx)
		r.Get("/", srv.getJob)
		r.Put("/status", srv.updateJobStatus)
		r.Route("/files", filesRouter)
	})
}

func (srv *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*Application)

	jobs, err := srv.jobs.GetAllAppJobs(app.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		log.Panic(err)
	}

	Response(w, jobs)
}

func (srv *Server) getJob(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*Job)

	// if job.Status == slurmer.JobStatus.Started {
	// 	jobProp, err := srv.slurmClient.GetJob(job.CurrentSlurmID)
	// 	if err != nil {
	// 		Error(w, http.StatusInternalServerError)
	// 		panic(err)
	// 	}
	// 	job.SlurmJob = jobProp
	// }

	Response(w, job)
	// job.SlurmJob = nil
}

func (srv *Server) createJob(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value("app").(*Application)

	var jobId string
	// Debug purposes
	if app.ID == "debug" {
		jobId = "debug"
	} else {
		jobId = uuid.NewString()
	}

	reqBody, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	jobDir := filepath.Join(app.Directory, "jobs", jobId)

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

	var batchProperties slurm.BatchProperties
	err = json.Unmarshal(reqBody, &batchProperties)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	err = writeBatch(batchFile, &batchProperties)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	job := Job{
		Name:      batchProperties.JobName,
		Status:    JobStatus.Stopped,
		Id:        jobId,
		Directory: jobDir,
	}

	if err := srv.jobs.AddAppJob(app.ID, &job); err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	Response(w, &job)
}

func (srv *Server) updateJobStatus(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*Job)

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest)
		panic(err)
	}
	defer r.Body.Close()

	var status string
	if err = json.Unmarshal(reqBody, &status); err != nil {
		Error(w, http.StatusBadRequest)
		panic(err)
	}

	switch status {
	case "started":
		if job.Status == JobStatus.Stopped {
			if err := srv.handleStartJob(job); err != nil {
				Error(w, http.StatusInternalServerError)
				panic(err)
			}
		}
	case "stopped":
		if job.Status == JobStatus.Started {
			if err := srv.slurmClient.CancelJob(job.SlurmId); err != nil {
				Error(w, http.StatusInternalServerError)
				panic(err)
			}
		}
	}

	Response(w, status)
}

func (srv *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := ctx.Value("app").(*Application)
	job := ctx.Value("job").(*Job)

	// First we need to stop pending/running job
	if job.Status == JobStatus.Started {
		err := srv.slurmClient.CancelJob(job.SlurmId)
		if err != nil {
			Error(w, http.StatusInternalServerError)
			panic(err)
		}
	}

	if err := srv.jobs.DeleteAppJob(app.ID, job.Id); err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	Error(w, http.StatusOK)
}

func (srv *Server) JobCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := r.Context().Value("app").(*Application)
		jobId := chi.URLParam(r, "jobId")
		job, err := srv.jobs.GetAppJob(app.ID, jobId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				Error(w, http.StatusNotFound)
				return
			}
			log.Panic(err)
		}
		ctx := context.WithValue(r.Context(), "job", job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
