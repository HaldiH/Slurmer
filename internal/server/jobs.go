package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/slurmer"
	"github.com/go-chi/chi"
)

func (srv *Server) jobsRouter(r chi.Router) {
	r.Get("/", srv.listJobs)
	r.Post("/", srv.addJob)
	r.Route("/{jobID}", func(r chi.Router) {
		r.Use(srv.JobCtx)
		r.Get("/", srv.getJob)
		r.Put("/status", srv.updateJobStatus)
	})
}

func (srv *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	// TODO: list only app specific jobs

	// token := r.Header.Get("X-Auth-Token")
	// app := srv.appsMap[token]

	jobs, err := srv.slurmClient.GetJobs()
	if err != nil {
		panic(err)
	}

	jsonData, err := json.Marshal(jobs)
	if err != nil {
		return
	}
	w.Write(jsonData)
}

func (srv *Server) getJob(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented)
}

func (srv *Server) addJob(w http.ResponseWriter, r *http.Request) {
	// TODO: Generate UUID for new jobs and store batch file in per job separated directory

	// TODO: Write in response the created job

	app := r.Context().Value("app").(*slurmer.Application)

	reqBody, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	batchFile, err := os.Create(filepath.Join(app.Directory, "jobs", "batch.sh"))
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
			cmd := exec.Command("sbatch", filepath.Join(app.Directory, "jobs", "batch.sh"))
			err := cmd.Run()
			if err != nil {
				Error(w, http.StatusInternalServerError)
				panic(err)
			}
			// TODO: save job pid and set job.status stopped when the job has terminated
		}
	case "stopped":
		if job.Status == slurmer.Started {
			// TODO: cancel the job with slurm id job.PID
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
		job := app.Jobs[jobID]
		ctx := context.WithValue(r.Context(), "job", job)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
