package rest

import (
	"net/http"

	"github.com/ShinoYasx/Slurmer/pkg/utils"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func filesRouter(r chi.Router) {
	r.Post("/", uploadZip)
	r.Get("/", downloadZip)
}

func uploadZip(w http.ResponseWriter, r *http.Request) {
	// Warning! A user can extract files to parents directories when zip contains ..
	// TODO: Fix jail escaping from a job
	job := getCtxJob(r.Context())

	file, header, err := r.FormFile("job_dir")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if header.Header.Get("Content-Type") == "application/zip" {
		err := utils.UnzipFile(file, header.Size, job.Directory)
		if err != nil {
			Error(w, http.StatusBadRequest)
			log.Warn(err.Error())
		}
	}
}

func downloadZip(w http.ResponseWriter, r *http.Request) {
	job := getCtxJob(r.Context())
	w.Header().Set("Content-Type", "application/zip")
	err := utils.ZipFile(job.Directory, w)
	if err != nil {
		panic(err)
	}
}
