package service

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/cliexecutor"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var funcMap = template.FuncMap{
	"escapeBash": func(s string) string {
		return strings.ReplaceAll(s, "'", "'\\''")
	},
}

var batchTmpl = template.New("batch.tmpl").Funcs(funcMap)

type jobServiceImpl struct {
	jobs        containers.JobsContainer
	slurmCache  containers.SlurmCache
	slurmClient slurm.Client
	executor    cliexecutor.Executor
}

func NewJobService(client slurm.Client, cache containers.SlurmCache, container containers.JobsContainer, executor cliexecutor.Executor) JobService {
	return &jobServiceImpl{
		jobs:        container,
		slurmClient: client,
		slurmCache:  cache,
		executor:    executor,
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
			slurmId, err := s.slurmClient.SubmitJob(job.UserName, nil, "batch.sh", job.Directory)
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
			if err := s.slurmClient.CancelJob(job.UserName, job.SlurmId); err != nil {
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

func (s *jobServiceImpl) Start(job *model.Job) error {
	return s.UpdateStatus(job, model.JobStarted)
}

func (s *jobServiceImpl) Stop(job *model.Job) error {
	return s.UpdateStatus(job, model.JobStopped)
}

func (s *jobServiceImpl) Create(user string, app *model.Application, prop *slurm.BatchProperties) (*model.Job, error) {
	jobId := uuid.New()
	jobDir := filepath.Join(app.Directory, "jobs", jobId.String())

	cmdCtx := cliexecutor.NewCommandContext(user, "mkdir", "-p", jobDir)
	if err := s.executor.ExecCommand(cmdCtx, nil, cliexecutor.ReadStderr); err != nil {
		return nil, err
	}

	templateFile := filepath.Join(app.Directory, "templates", "batch.tmpl")
	buf := new(bytes.Buffer)

	if err := writeBatch(templateFile, buf, prop); err != nil {
		return nil, err
	}

	stdin, err := io.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	cmdCtx = &cliexecutor.CommandContext{
		User:    user,
		Command: "tee",
		Args:    []string{"batch.sh"},
		Dir:     jobDir,
		Stdin:   string(stdin),
	}
	s.executor.ExecCommand(cmdCtx, nil, cliexecutor.ReadStderr)

	job := model.Job{
		Name:      prop.JobName,
		Status:    model.JobStopped,
		Id:        jobId,
		Directory: jobDir,
		UserName:  user,
	}

	if err := s.jobs.AddAppJob(app.Id, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *jobServiceImpl) Delete(app *model.Application, job *model.Job) error {
	// First we need to stop pending/running job
	if job.Status == model.JobStarted {
		err := s.slurmClient.CancelJob(job.UserName, job.SlurmId)
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

	cmdCtx := cliexecutor.CommandContext{
		User:    job.UserName,
		Command: "rm",
		Args:    []string{"-rf", job.Directory},
	}
	if err := s.executor.ExecCommand(&cmdCtx, nil, cliexecutor.ReadStderr); err != nil {
		return err
	}

	return nil
}

func (s *jobServiceImpl) PruneJob(job *model.Job) error {
	files, err := os.ReadDir(job.Directory)
	if err != nil {
		return err
	}

	for _, dirEntry := range files {
		if dirEntry.Name() == "batch.sh" {
			continue
		}

		cmdCtx := cliexecutor.CommandContext{
			User:    job.UserName,
			Command: "rm",
			Args:    []string{"-rf", dirEntry.Name()},
			Dir:     job.Directory,
		}
		if err := s.executor.ExecCommand(&cmdCtx, nil, cliexecutor.ReadStderr); err != nil {
			return err
		}
	}
	return nil
}

func writeBatch(templatePath string, out io.Writer, batch *slurm.BatchProperties) error {
	tmpl, err := batchTmpl.ParseFiles(templatePath)
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
