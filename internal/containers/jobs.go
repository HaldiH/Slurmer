package containers

import (
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
)

type JobsContainer interface {
	GetAllJobs() ([]*model.Job, error)
	GetJob(jobId uuid.UUID) (*model.Job, error)
	DeleteJob(jobId uuid.UUID) error
	UpdateJob(job *model.Job) error

	GetAllAppJobs(appId uuid.UUID) ([]*model.Job, error)
	GetAppJob(appId uuid.UUID, jobId uuid.UUID) (*model.Job, error)
	AddAppJob(appId uuid.UUID, job *model.Job) error
	DeleteAppJob(appId uuid.UUID, jobId uuid.UUID) error
	UpdateAppJob(appId uuid.UUID, job *model.Job) error
}
