package slurmcli

import (
	"encoding/json"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ShinoYasx/Slurmer/pkg/cliexecutor"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/ShinoYasx/Slurmer/pkg/utils"
)

type cliClient struct {
	executor cliexecutor.Executor
}

func NewCliClient(executor cliexecutor.Executor) slurm.Client {
	return &cliClient{
		executor: executor,
	}
}

func (c *cliClient) getAllJobs() ([]slurm.JobResponseProperties, error) {
	var res slurm.JobsResponse
	cmd := exec.Command("squeue", "--json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer stderr.Close()

	decoder := json.NewDecoder(stdout)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	return res.Jobs, cliexecutor.ReadStderr(stderr, cmd.Wait())
}

// GetJobs gives a slice of all slurm jobs.
func (c *cliClient) GetJobs(ids ...int) ([]slurm.JobResponseProperties, error) {
	allJobs, err := c.getAllJobs()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return allJobs, nil
	}

	var jobs []slurm.JobResponseProperties
	for _, job := range allJobs {
		if contains(ids, job.JobId) {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (c *cliClient) GetJob(id int) (*slurm.JobResponseProperties, error) {
	allJobs, err := c.getAllJobs()
	if err != nil {
		return nil, err
	}
	for _, job := range allJobs {
		if job.JobId == id {
			return &job, nil
		}
	}
	return nil, slurm.ErrJobNotFound
}

func (c *cliClient) SubmitJob(user string, o *slurm.SBatchOptions, script string, workingDir string) (slurmId int, err error) {
	cmdCtx := c.prepareBatch(o, script, user)
	cmdCtx.Dir = workingDir

	if err := c.executor.ExecCommand(cmdCtx, func(r io.Reader, waitErr error) error {
		if waitErr != nil {
			return nil
		}

		firstLine, err := utils.FirstLine(r)
		if err != nil {
			return err
		}

		words := strings.Split(firstLine, " ")
		slurmId, err = strconv.Atoi(words[len(words)-1])
		if err != nil {
			return err
		}
		return nil
	}, cliexecutor.ReadStderr); err != nil {
		return 0, err
	}

	return slurmId, nil
}

func (c *cliClient) CancelJob(user string, ids ...int) error {
	return c.executor.ExecCommand(&cliexecutor.CommandContext{
		User:    user,
		Command: "scancel",
		Args:    map2(ids, strconv.Itoa),
	}, nil, cliexecutor.ReadStderr)
}
