package slurm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) request(path string, method string, body io.Reader, v interface{}) error {
	url := fmt.Sprintf("http://%s/slurm/%s/%s", c.slurmrestHost, slurmrestVersion, path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "applcation/json")
	response, err := c.httpc.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	jsonData, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

func (c *Client) get(path string, v interface{}) error {
	return c.request(path, http.MethodGet, nil, v)
}
