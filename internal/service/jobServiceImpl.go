package service

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type jobServiceImpl struct {
	jobs        containers.JobsContainer
	slurmCache  containers.SlurmCache
	slurmClient slurm.Client
}

func NewJobService(client slurm.Client, cache containers.SlurmCache, container containers.JobsContainer) JobService {
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
			slurmId, err := s.slurmClient.SubmitJob(nil, "batch.sh", job.Directory)
			if err != nil {
				return err
			}

			job.SlurmJob, err = s.slurmClient.GetJob(slurmId)
			if err != nil {
				return err
			}

			if err := s.slurmCache.DeleteSlurmJob(job.SlurmId); err != nil {
				return err
			}

			job.Status = model.JobStarted

			if err := s.jobs.UpdateJob(job); err != nil {
				return err
			}
		}
	case model.JobStopped:
		if job.Status == model.JobStarted {
			if err := s.slurmClient.CancelJob(job.SlurmId); err != nil {
				return err
			}
			var err error
			job.SlurmJob, err = s.slurmClient.GetJob(job.SlurmId)
			if err != nil {
				return err
			}
			if job.SlurmJob.JobState.IsStopped() {
				job.Status = model.JobStopped
				if err := s.jobs.UpdateJob(job); err != nil {
					return err
				}
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

	if err := os.MkdirAll(jobDir, 0777); err != nil {
		return nil, err
	}

	batchFile, err := os.Create(filepath.Join(jobDir, "batch.sh"))
	if err != nil {
		return nil, err
	}
	defer batchFile.Close()

	templateFile := filepath.Join(app.Directory, "templates", "batch.tmpl")

	if err := writeBatch(templateFile, batchFile, prop); err != nil {
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

	// TODO: JobsContainer.DeleteJob should also delete the association to
	// the SlurmJob, but for an unknown reason, GORM won't delete the
	// associated job. Need to fix it.
	if err := s.slurmCache.DeleteSlurmJob(job.SlurmId); err != nil {
		return err
	}

	if err := s.jobs.DeleteJob(job.Id); err != nil {
		return err
	}

	return nil
}

func writeBatch(templatePath string, out io.Writer, batch *slurm.BatchProperties) error {
	funcMap := template.FuncMap{
		"escapeBash": func(s string) string {
			return strings.ReplaceAll(s, "'", "'\\''")
		},
	}

	tmpl, err := template.New("batch.tmpl").
		Funcs(funcMap).
		ParseFiles(templatePath)
	if err != nil {
		return err
	}
	return tmpl.Execute(out, batch)
}

func (s *jobServiceImpl) PollJobsStatus() error {
	log.Debug("Polling jobs")
	jobs, err := s.GetAll()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Status == model.JobStarted || job.SlurmJob != nil {
			slurmJob, err := s.slurmClient.GetJob(job.SlurmId)
			if err != nil {
				if err == slurm.ErrJobNotFound {
					if err := s.slurmCache.DeleteSlurmJob(job.SlurmId); err != nil {
						return err
					}
					job.SlurmId = 0
					job.SlurmJob = nil
					job.Status = model.JobStopped
				} else {
					return err
				}
			} else {
				job.SlurmJob = slurmJob
				if slurmJob.JobState.IsStopped() {
					job.Status = model.JobStopped
				}
			}

			if err := s.jobs.UpdateJob(job); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *jobServiceImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.jobs)
}
