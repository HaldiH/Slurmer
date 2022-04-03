package slurmcli

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/utils"
)

type CliClient struct{}

func NewCliClient() *CliClient {
	return &CliClient{}
}

// GetJobs gives a slice of all slurm jobs.
func (c *CliClient) GetJobs(ids ...int) (*slurm.JobsResponse, error) {
	var res *slurm.JobsResponse
	// We cannot specify wanted jobs in the request since the json flag will ignore the -j option
	res, err := execCommand[slurm.JobsResponse](exec.Command("squeue", "--json"))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return res, nil
	}

	var jobs []slurm.JobResponseProperties
	for _, job := range res.Jobs {
		if contains(ids, job.JobId) {
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
		if job.JobId == id {
			return &job, nil
		}
	}
	return nil, errors.New("invalid job id: " + strconv.Itoa(id))
}

func (c *CliClient) SubmitBatch(o slurm.SBatchOptions) (jobID int, err error) {
	cmd := c.prepareBatch(o)

	jobStdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}

	err = cmd.Start()
	if err != nil {
		return 0, err
	}

	words := strings.Split(utils.FirstLine(jobStdout), " ")
	return strconv.Atoi(words[len(words)-1])
}

func (c *CliClient) CancelJob(id int) error {
	return exec.Command("scancel", strconv.Itoa(id)).Start()
}
