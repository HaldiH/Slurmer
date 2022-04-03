package slurmer

import (
	"net/http"

	"github.com/go-chi/chi"
)

func filesRouter(r chi.Router) {
	r.Post("/", uploadZip)
	r.Get("/", downloadZip)
}

func uploadZip(w http.ResponseWriter, r *http.Request) {
	// Warning! A user can extract files to parents directories when zip contains ..
	// TODO: Fix jail escaping from a job
	job := r.Context().Value("job").(*Job)
	file, header, err := r.FormFile("job_dir")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if header.Header.Get("Content-Type") == "application/zip" {
		err := unzipFile(file, header.Size, job.Directory)
		if err != nil {
			panic(err)
		}
	}
}

func downloadZip(w http.ResponseWriter, r *http.Request) {
	job := r.Context().Value("job").(*Job)
	w.Header().Set("Content-Type", "application/zip")
	err := zipFile(job.Directory, w)
	if err != nil {
		panic(err)
	}
}
