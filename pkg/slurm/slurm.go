package slurm

import (
	"context"
	"net"
	"net/http"
	"net/url"
)

const slurmrestVersion = "v0.0.37"

type Client struct {
	httpc         *http.Client
	slurmrestHost string
}

func NewClient(slurmrestdURL string) (*Client, error) {
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

	slurmClient := &Client{
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

func (c *Client) GetJobs() (*JobsResponse, error) {
	var jobs JobsResponse
	err := c.get("/jobs", &jobs)
	return &jobs, err
}
