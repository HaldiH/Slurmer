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
	UpdateStatus(job *model.Job, status model.JobStatus) error
	Create(app *model.Application, prop *slurm.BatchProperties) (*model.Job, error)
	Delete(app *model.Application, job *model.Job) error
	MarshalJSON() ([]byte, error)
}
