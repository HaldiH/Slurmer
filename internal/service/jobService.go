package service

import (
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/google/uuid"
)

type JobService interface {
	GetAll() ([]*model.Job, error)
	GetAppAll(app *model.Application) ([]*model.Job, error)
	GetApp(app *model.Application, jobId uuid.UUID) (*model.Job, error)

	// UpdateStatus is used to start or stop a job. If the job is stopped and UpdateStatus
	// is called with `JobStarted`, the job will be executed and vice-versa.
	// If UpdateStatus is called with the same status as the `job`, then it has no effect.
	UpdateStatus(job *model.Job, status model.JobStatus) error
	Create(app *model.Application, prop *slurm.BatchProperties) (*model.Job, error)
	Delete(app *model.Application, job *model.Job) error

	// PollJobsStatus retrieve the new job status from slurm for the running jobs
	// and update the database.
	PollJobsStatus() error

	// MarshalJSON should return the list of all registered jobs in JSON.
	MarshalJSON() ([]byte, error)
}
