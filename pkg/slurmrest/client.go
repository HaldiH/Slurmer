package slurmrest

import (
	"context"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"net"
	"net/http"
	"net/url"
)

const slurmrestVersion = "v0.0.37"

type RestClient struct {
	httpc         *http.Client
	slurmrestHost string
}

func NewRestClient(slurmrestdURL string) (*RestClient, error) {
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

func (c *RestClient) GetJobs() (*slurm.JobsResponse, error) {
	var jobs slurm.JobsResponse
	err := c.get("/jobs", &jobs)
	return &jobs, err
}

func (c *RestClient) GetJob(id int) (*slurm.JobResponseProperties, error) {
	panic("RestClient.GetJob not implemented")
}
