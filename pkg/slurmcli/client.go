package slurmcli

import (
	"io"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

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
	return nil, slurm.ErrJobNotFound
}

func (c *CliClient) SubmitJob(o *slurm.SBatchOptions, script string, cwd string) (slurmId int, err error) {
	cmd := c.prepareBatch(o, script)
	cmd.Dir = cwd
	jobStdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}

	jobStderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}

	err = cmd.Start()
	if err != nil {
		return 0, err
	}

	words := strings.Split(utils.FirstLine(jobStdout), " ")
	slurmId, err = strconv.Atoi(words[len(words)-1])
	if err != nil {
		log.Error(err)
		errStr, err := io.ReadAll(jobStderr)
		if err != nil {
			return 0, err
		}
		log.Error(string(errStr))
		return 0, err
	}

	return slurmId, nil
}

func (c *CliClient) CancelJob(id int) error {
	return exec.Command("scancel", strconv.Itoa(id)).Start()
}
