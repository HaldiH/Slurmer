package slurmrest

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

const slurmrestVersion = "v0.0.37"

type RestClient struct {
	httpc         *http.Client
	slurmrestHost string
}

func NewRestClient(slurmrestdURL string) (slurm.Client, error) {
	u, err := url.Parse(slurmrestdURL)
	if err != nil {
		return nil, err
	}

	var transport http.Transport
	switch u.Scheme {
	case "http+unix":
		var socket string
		query := u.Query()
		if query.Has("socket") {
			socket = query.Get("socket")
		} else {
			socket = "/var/run/slurmrestd.sock"
		}
		transport = http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		}
	}

	slurmClient := &RestClient{
		httpc: &http.Client{
			Transport: &transport,
		},
		slurmrestHost: u.Host,
	}

	if slurmClient.slurmrestHost == "" {
		slurmClient.slurmrestHost = "unix"
	}

	return slurmClient, nil
}

func (c *RestClient) GetJobs(ids ...int) ([]slurm.JobResponseProperties, error) {
	var jobs slurm.JobsResponse
	err := c.get("/jobs", &jobs)
	return jobs.Jobs, err
}

func (c *RestClient) GetJob(id int) (*slurm.JobResponseProperties, error)

func (c *RestClient) SubmitJob(o *slurm.SBatchOptions, script string, cwd string) (int, error)

func (c *RestClient) CancelJob(id int) error
