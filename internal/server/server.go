package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ShinoYasx/Slurmer/internal/appconfig"
	"github.com/ShinoYasx/Slurmer/internal/slurmer"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/go-chi/chi"
)

type Server struct {
	config      *appconfig.Config
	slurmClient *slurm.Client
	router      chi.Router
	appsMap     map[string]slurmer.Application
}

func New(config *appconfig.Config) (*Server, error) {
	sc, err := slurm.NewClient(config.Slurmrest.URL)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(config.Slurmer.WorkingDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	os.Chdir(config.Slurmer.WorkingDir)

	appsDir := filepath.Join(config.Slurmer.WorkingDir, "applications")
	appsMap := make(map[string]slurmer.Application)
	err = os.MkdirAll(appsDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	for _, app := range config.Slurmer.Applications {
		appDir := filepath.Join(appsDir, app.UUID)
		jobsDir := filepath.Join(appDir, "jobs")
		err = os.MkdirAll(appDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		err = os.MkdirAll(jobsDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		appsMap[app.Token] = slurmer.Application{Directory: appDir, Jobs: make(map[string]slurmer.Job)}
	}

	srv := Server{
		config:      config,
		slurmClient: sc,
		router:      chi.NewRouter(),
		appsMap:     appsMap,
	}

	return &srv, nil
}

func (srv *Server) Listen() error {
	srv.router.Get("/jobs", func(w http.ResponseWriter, r *http.Request) {
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
	})

	srv.router.Post("/jobs", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Generate UUID for new jobs and store batch file in per job separated directory

		// TODO: Write in response the created job

		token := r.Header.Get("X-Auth-Token")
		app := srv.appsMap[token]

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
	})

	srv.router.Put("/jobs/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		app := srv.appsMap[token]
		jobId := chi.URLParam(r, "id")
		job := app.Jobs[jobId]

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
				cmd := exec.Command("sbatch", filepath.Join(app.Directory, "jobs", jobId, "batch.sh"))
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
	})

	addr := fmt.Sprintf("%s:%d", srv.config.Slurmer.IP, srv.config.Slurmer.Port)
	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, srv.router)
}
