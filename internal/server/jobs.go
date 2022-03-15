package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/slurmer"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
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
	// TODO: list only app specific jobs

	app := r.Context().Value("app").(*slurmer.Application)

	jsonData, err := json.Marshal(app.Jobs)

	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	w.Write(jsonData)
}

func (srv *Server) getJob(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented)
}

func (srv *Server) createJob(w http.ResponseWriter, r *http.Request) {
	// TODO: Generate UUID for new jobs and store batch file in per job separated directory

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

	err = slurmer.WriteBatch(batchFile, &batchProperties)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	job := slurmer.Job{
		Name:   batchProperties.JobName,
		Status: slurmer.Stopped,
		ID:     jobID,
	}

	app.Jobs.AddJob(jobID, &job)

	jsonData, err := json.Marshal(&job)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	w.Write(jsonData)
}

func (srv *Server) updateJobStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := ctx.Value("app").(*slurmer.Application)
	job := ctx.Value("job").(*slurmer.Job)

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
		if job.Status == slurmer.Stopped {
			fmt.Println("Starting job")
			cmd := exec.Command("sbatch", filepath.Join(app.Directory, "jobs", job.ID, "batch.sh"))
			err := cmd.Run()
			if err != nil {
				Error(w, http.StatusInternalServerError)
				panic(err)
			}
			job.Status = slurmer.Started
			// TODO: save job pid and set job.status stopped when the job has terminated
		}
	case "stopped":
		if job.Status == slurmer.Started {
			// TODO: cancel the job with slurm id job.PID
			job.Status = slurmer.Stopped
		}
	}

	res, err := json.Marshal(status)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}
	w.Write(res)
	w.Header().Set("Content-Type", "application/json")
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
