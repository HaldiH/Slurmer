package slurmer

import (
	"archive/zip"
	"bufio"
	"encoding/json"
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

func handleStartJob(job *slurmer.Job) error {
	cmd := exec.Command("sbatch", "--wait", "batch.sh")
	cmd.Dir = job.Directory
	jobStdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
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
				panic(err)
			}
		}
		err = cmd.Wait()
		if err != nil {
			panic(err)
		}
		// When the job is terminated, mark the job as stopped
		job.Status = slurmer.JobStatus.Stopped
		job.CurrentSlurmID = 0 // 0 is job not active
	}()

	job.Status = slurmer.JobStatus.Started
	return nil
}

func UnzipFile(r io.ReaderAt, size int64, destDir string) error {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}
	for _, zipFile := range zipReader.File {
		if zipFile.FileInfo().IsDir() {
			err = os.MkdirAll(filepath.Join(destDir, zipFile.Name), zipFile.Mode())
			if err != nil {
				return err
			}
			continue
		}
		f, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		destFile, err := os.OpenFile(filepath.Join(destDir, zipFile.Name), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, zipFile.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()
		_, err = io.Copy(destFile, f)
		if err != nil {
			return err
		}
	}
	return nil
}

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
