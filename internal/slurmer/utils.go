package slurmer

import (
	"archive/tar"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/slurmer"
	"github.com/ShinoYasx/Slurmer/pkg/utils"
)

func Response(w http.ResponseWriter, v interface{}) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}
	_, err = w.Write(jsonData)
	if err != nil {
		panic(err)
	}
}

func Error(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func WriteBatch(out io.Writer, batch *slurm.BatchProperties) error {
	tmpl, err := template.ParseFiles(filepath.Join("templates", "batch.tmpl"))
	if err != nil {
		return err
	}
	return tmpl.Execute(out, batch)
}

func SetContentType(contentType string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

func (srv *Server) handleStartJob(job *slurmer.Job) error {
	cmd := exec.Command("sbatch", "--wait", "batch.sh")
	cmd.Dir = job.Directory
	jobStdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		// Goroutine will get slurm job id and wait for the job to end, so it can change its status
		// Read the first line of sbatch to get the slurm job id
		words := strings.Split(utils.FirstLine(jobStdout), " ")
		slurmID, err := strconv.Atoi(words[len(words)-1])
		if err != nil {
			panic(err)
		}

		slurmJob, err := srv.slurmClient.GetJob(slurmID)
		if err != nil {
			panic(err)
		}

		job.CurrentSlurmID = slurmID
		job.SlurmJob = slurmJob

		if err := cmd.Wait(); err != nil {
			panic(err)
		}
		// When the job is terminated, mark the job as stopped
		job.Status = slurmer.JobStatus.Stopped

		slurmJob, err = srv.slurmClient.GetJob(slurmID)
		if err != nil {
			panic(err)
		}

		job.SlurmJob = slurmJob
	}()

	job.Status = slurmer.JobStatus.Started
	return nil
}

func (srv *Server) updateJobs() {
	fmt.Println("Updating jobs status")
	for _, app := range *srv.apps {
		for _, job := range *app.Jobs {
			if job.Status == slurmer.JobStatus.Started {
				slurmJob, err := srv.slurmClient.GetJob(job.CurrentSlurmID)
				if err != nil {
					panic(err)
				}

				job.SlurmJob = slurmJob
			}
		}
	}
}

func UntarDirectory(r io.Reader, dstDir string) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if header == nil {
			continue
		}

		target := filepath.Join(dstDir, header.Name)
		fileInfo := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, fileInfo.Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileInfo.Mode())
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}
		}
	}
	return nil
}

func UnzipFile(r io.ReaderAt, size int64, destDir string) error {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}
	for _, zipFile := range zipReader.File {
		target := filepath.Join(destDir, zipFile.Name)
		if zipFile.FileInfo().IsDir() {
			if err = os.MkdirAll(target, zipFile.Mode()); err != nil {
				return err
			}
			continue
		}
		f, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		destFile, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, zipFile.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()
		if _, err = io.Copy(destFile, f); err != nil {
			return err
		}
	}
	return nil
}

// func TarFile(srcDir string, w io.Writer) error {
// 	tarWriter := tar.NewWriter(w)
// 	defer tarWriter.Close()

// 	oldwd, err := os.Getwd()
// 	if err != nil {
// 		return err
// 	}

// 	os.Chdir(srcDir)
// 	defer os.Chdir(oldwd)

// 	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if path == "." {
// 			return nil
// 		}

// 		fileInfo, err := d.Info()
// 		if err != nil {
// 			return err
// 		}

// 		header, err := tar.FileInfoHeader(fileInfo, "")
// 		if err != nil {
// 			return err
// 		}
// 		header.Name = path

// 		f, err := tarWriter.CreateHeader(header)
// 		if err != nil {
// 			return err
// 		}

// 		if d.IsDir() {
// 			return nil
// 		}

// 		src, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer src.Close()

// 		_, err = io.Copy(f, src)
// 		if err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

func ZipFile(srcDir string, w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	oldwd, err := os.Getwd()
	if err != nil {
		return err
	}

	os.Chdir(srcDir)
	defer os.Chdir(oldwd)

	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = path

		f, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(f, src)
		if err != nil {
			return err
		}

		return nil
	})
}
