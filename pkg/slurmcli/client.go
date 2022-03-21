package slurmcli

import (
	"errors"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"os/exec"
	"strconv"
)

type CliClient struct{}

func NewCliClient() *CliClient {
	return &CliClient{}
}

// GetJobs gives a slice of all slurm jobs.
func (c *CliClient) GetJobs(ids ...int) (*slurm.JobsResponse, error) {
	var res *slurm.JobsResponse
	res, err := execCommand[slurm.JobsResponse](exec.Command("squeue", "--json"))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return res, nil
	}

	var jobs []slurm.JobResponseProperties
	for _, job := range res.Jobs {
		if contains[int](ids, *job.JobId) {
			jobs = append(jobs, job)
		}
	}

	res.Jobs = jobs
	return res, nil
}

func (c *CliClient) GetJob(id int) (*slurm.JobResponseProperties, error) {
	jobsResponse, err := c.GetJobs()
	if err != nil {
		return nil, err
	}
	for _, job := range jobsResponse.Jobs {
		if *job.JobId == id {
			return &job, nil
		}
	}
	return nil, errors.New("invalid job id: " + strconv.Itoa(id))
}
