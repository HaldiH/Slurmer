package service

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type jobServiceImpl struct {
	jobs        containers.JobsContainer
	slurmCache  containers.SlurmCache
	slurmClient slurm.Client
}

func NewJobServiceImpl(client slurm.Client, cache containers.SlurmCache, container containers.JobsContainer) JobService {
	return &jobServiceImpl{
		jobs:        container,
		slurmClient: client,
		slurmCache:  cache,
	}
}

func (s *jobServiceImpl) GetAll() ([]*model.Job, error) {
	return s.jobs.GetAllJobs()
}

func (s *jobServiceImpl) GetAppAll(app *model.Application) ([]*model.Job, error) {
	return s.jobs.GetAllAppJobs(app.Id)
}

func (s *jobServiceImpl) GetApp(app *model.Application, jobId uuid.UUID) (*model.Job, error) {
	return s.jobs.GetAppJob(app.Id, jobId)
}

func (s *jobServiceImpl) UpdateStatus(job *model.Job, status model.JobStatus) error {
	switch status {
	case model.JobStarted:
		if job.Status == model.JobStopped {
			if err := s.handleStartJob(job); err != nil {
				return err
			}
		}
	case model.JobStopped:
		if job.Status == model.JobStarted {
			if err := s.slurmClient.CancelJob(job.SlurmId); err != nil {
				return err
			}
		}

	}
	return nil
}

func (s *jobServiceImpl) Create(app *model.Application, prop *slurm.BatchProperties) (*model.Job, error) {
	/*** Debug purposes ***/
	// if app.Id == uuid.NullUUID {
	// 	jobId = "debug"
	// } else {
	// 	jobId = uuid.New()
	// }
	jobId := uuid.New()
	jobDir := filepath.Join(app.Directory, "jobs", jobId.String())

	if err := os.MkdirAll(jobDir, 0770); err != nil {
		return nil, err
	}

	batchFile, err := os.Create(filepath.Join(jobDir, "batch.sh"))
	if err != nil {
		return nil, err
	}
	defer batchFile.Close()

	if err := writeBatch(batchFile, prop); err != nil {
		return nil, err
	}

	job := model.Job{
		Name:      prop.JobName,
		Status:    model.JobStopped,
		Id:        jobId,
		Directory: jobDir,
	}

	if err := s.jobs.AddAppJob(app.Id, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *jobServiceImpl) Delete(app *model.Application, job *model.Job) error {
	// First we need to stop pending/running job
	if job.Status == model.JobStarted {
		err := s.slurmClient.CancelJob(job.SlurmId)
		if err != nil {
			return err
		}
	}

	if err := s.jobs.DeleteJob(job.Id); err != nil {
		return err
	}

	return nil
}

func writeBatch(out io.Writer, batch *slurm.BatchProperties) error {
	funcMap := template.FuncMap{
		"escapeBash": func(s string) string {
			return strings.ReplaceAll(s, "'", "'\\''")
		},
	}

	tmpl, err := template.New("batch.tmpl").
		Funcs(funcMap).
		ParseFiles(filepath.Join("templates", "batch.tmpl"))
	if err != nil {
		return err
	}
	return tmpl.Execute(out, batch)
}

func (s *jobServiceImpl) handleStartJob(job *model.Job) error {
	cmd := exec.Command("sbatch", "--wait", "batch.sh")
	cmd.Dir = job.Directory
	jobStdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	jobStderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	words := strings.Split(utils.FirstLine(jobStdout), " ")
	slurmId, err := strconv.Atoi(words[len(words)-1])
	if err != nil {
		errStr, _ := io.ReadAll(jobStderr)
		log.Error(string(errStr))
		log.Error(err)
		return err
	}

	job.SlurmJob, err = s.slurmClient.GetJob(slurmId)
	if err != nil {
		return err
	}

	if oldSlurmId := job.SlurmId; oldSlurmId != 0 {
		if err := s.slurmCache.DeleteSlurmJob(job.SlurmId); err != nil {
			if err != slurm.ErrJobNotFound {
				return err
			}
		}
	}

	job.Status = model.JobStarted
	if err := s.jobs.UpdateJob(job); err != nil {
		return err
	}

	go func() {
		// Goroutine will get slurm job id and wait for the job to end, so it can change its status
		// Read the first line of sbatch to get the slurm job id
		if err := cmd.Wait(); err != nil {
			log.Panic(err)
		}
		log.Debugf("Job %d has terminated", job.SlurmId)
		// When the job is terminated, mark the job as stopped
		job.SlurmJob, err = s.slurmClient.GetJob(slurmId)
		if err != nil {
			log.Panic(err)
		}

		job.Status = model.JobStopped
		if err := s.jobs.UpdateJob(job); err != nil {
			log.Panic(err)
		}
	}()

	return nil
}

func (s *jobServiceImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.jobs)
}
