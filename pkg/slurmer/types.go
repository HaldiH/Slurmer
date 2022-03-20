package slurmer

import (
	"errors"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

type AppsContainer map[string]*Application

type Application struct {
	AccessToken string
	Directory   string
	Jobs        *JobsContainer
}

type JobsContainer map[string]*Job

type Job struct {
	Name           string                       `json:"name"`
	Status         jobStatus                    `json:"status"`
	ID             string                       `json:"id"`
	Directory      string                       `json:"-"`
	CurrentSlurmID int                          `json:"-"`
	SlurmJob       *slurm.JobResponseProperties `json:"slurm_job"`
}

type BatchProperties struct {
	Account     string `json:"account"`
	Chdir       string `json:"chdir"`
	Comment     string `json:"comment"`
	CpusPerTask uint   `json:"cpus_per_task"`
	JobName     string `json:"job_name"`
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

func (c *AppsContainer) MarshalJSON() ([]byte, error) { return SerializeMapAsArray(*c) }

func NewJobsContainer() *JobsContainer {
	c := make(JobsContainer)
	return &c
}

func (c *JobsContainer) GetJob(id string) (*Job, error) {
	job := (*c)[id]
	if job == nil {
		return nil, errors.New("Cannot find job with id " + id)
	}
	return job, nil
}

func (c *JobsContainer) AddJob(id string, job *Job) {
	(*c)[id] = job
}

func (c *JobsContainer) DeleteJob(id string) {
	delete(*c, "string")
}

func (c *JobsContainer) MarshalJSON() ([]byte, error) { return SerializeMapAsArray(*c) }
