package slurmer

import (
	"errors"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

type AppsContainer map[string]*Application

type Application struct {
	AccessToken string
	Directory   string
	// Jobs        AppJobs
	ID string
}

type Job struct {
	Name      string                       `json:"name"`
	Status    jobStatus                    `json:"status"`
	Id        string                       `json:"id" gorm:"primaryKey"`
	SlurmId   int                          `json:"slurm_id"`
	SlurmJob  *slurm.JobResponseProperties `json:"slurm_job" gorm:"foreignKey:SlurmId"`
	Directory string                       `json:"-"`
	AppId     string                       `json:"-"`
}

type JobsContainer interface {
	GetAllJobs() ([]*Job, error)
	GetJob(jobID string) (*Job, error)
	DeleteJob(jobID string) error
	UpdateJob(job *Job) error

	GetAllAppJobs(appID string) ([]*Job, error)
	GetAppJob(appID string, jobID string) (*Job, error)
	AddAppJob(appID string, job *Job) error
	DeleteAppJob(appID string, jobID string) error
	UpdateAppJob(appID string, job *Job) error
}

type jobStatus string

const (
	started jobStatus = "started"
	stopped jobStatus = "stopped"
)

var JobStatus = struct {
	Started jobStatus
	Stopped jobStatus
}{
	Started: started,
	Stopped: stopped,
}

func NewAppsContainer() *AppsContainer {
	c := make(AppsContainer)
	return &c
}

func (c *AppsContainer) GetApp(id string) (*Application, error) {
	app := (*c)[id]
	if app == nil {
		return nil, errors.New("Cannot find app with id " + id)
	}
	return app, nil
}

func (c *AppsContainer) AddApp(id string, app *Application) {
	(*c)[id] = app
}

func (c *AppsContainer) DeleteApp(id string) {
	delete(*c, id)
}

func (c *AppsContainer) MarshalJSON() ([]byte, error) { return MapToJSONArray(*c) }

// func NewJobsContainer() *JobsContainer {
// 	c := make(JobsContainer)
// 	return &c
// }

// func (c *JobsContainer) GetJob(id string) (*Job, error) {
// 	job := (*c)[id]
// 	if job == nil {
// 		return nil, errors.New("Cannot find job with id " + id)
// 	}
// 	return job, nil
// }

// func (c *JobsContainer) AddJob(id string, job *Job) {
// 	(*c)[id] = job
// }

// func (c *JobsContainer) DeleteJob(id string) {
// 	delete(*c, "string")
// }

// func (c *JobsContainer) MarshalJSON() ([]byte, error) { return SerializeMapAsArray(*c) }
