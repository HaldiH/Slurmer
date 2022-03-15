package slurmer

import (
	"encoding/json"
	"errors"
)

type JobsContainer map[string]*Job

func NewJobsContainer() *JobsContainer {
	c := make(JobsContainer)
	return &c
}

func (c *JobsContainer) MarshalJSON() ([]byte, error) {
	jsonData := []byte{'['}

	first := true
	for _, v := range *c {
		if first {
			first = false
		} else {
			jsonData = append(jsonData, ',')
		}
		jobJsonData, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, jobJsonData...)
	}

	jsonData = append(jsonData, ']')

	return jsonData, nil
}

type Application struct {
	Token     string
	Directory string
	Jobs      *JobsContainer
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

const (
	Stopped = "stopped"
	Started = "started"
)

type Job struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	ID     string `json:"id"`
}
