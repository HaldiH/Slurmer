package slurmer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ShinoYasx/Slurmer/pkg/slurmer"

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
	app := r.Context().Value("app").(*slurmer.Application)

	Ok(w, app.Jobs)
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
	Response(w, &job, http.StatusCreated)
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
			cmd := exec.Command("sbatch", "--wait", "batch.sh")
			cmd.Dir = job.Directory
			jobStdout, err := cmd.StdoutPipe()
			if err != nil {
				Error(w, http.StatusInternalServerError)
				log.Panic(err)
			}

			err = cmd.Start()
			if err != nil {
				Error(w, http.StatusInternalServerError)
				log.Panic(err)
			}

			go func() {
				// Goroutine will get slurm job id and wait for the job to end, so it can change its status
				scanner := bufio.NewScanner(jobStdout)
				// Read the first line of sbatch to get the slurm job id
				if scanner.Scan() {
					submitLine := scanner.Text()
					words := strings.Split(submitLine, " ")
					job.CurrentSlurmID, err = strconv.Atoi(words[len(words)-1])
					if err != nil {
						log.Panic(err)
					}
				}
				err = cmd.Wait()
				if err != nil {
					log.Panic(err)
				}
				// When the job is terminated, mark the job as stopped
				job.Status = slurmer.JobStatus.Stopped
				job.CurrentSlurmID = 0 // 0 is job not active
			}()

			job.Status = slurmer.JobStatus.Started
		}
	case "stopped":
		if job.Status == slurmer.JobStatus.Started {
			// TODO: cancel the job with slurm id job.PID
			job.Status = slurmer.JobStatus.Stopped
		}
	}

	Ok(w, status)
}

func (srv *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := ctx.Value("app").(*slurmer.Application)
	job := ctx.Value("job").(*slurmer.Job)

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
