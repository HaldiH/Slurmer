package slurmcli

import (
	"encoding/json"
	"errors"
	"os/exec"
)

func execCommand[T any](cmd *exec.Cmd) (*T, error) {
	jsonData, err := cmd.Output()
	if err != nil {
		return nil, errors.New("invalid slurm command")
	}

	var res T
	err = json.Unmarshal(jsonData, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
