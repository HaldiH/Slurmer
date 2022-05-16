package slurmcli

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

func execCommand[T any](cmd *exec.Cmd) (filledStruct *T, err error) {
	jsonData, err := cmd.Output()
	if err != nil {
		return nil, errors.New("invalid slurm command")
	}

	filledStruct = new(T)
	if err = json.Unmarshal(jsonData, filledStruct); err != nil {
		return nil, err
	}

	return filledStruct, nil
}

// contains return true if `slice` contains `element` otherwise return false
func contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func commaSeparatedArray[T any](a []T, stringify func(T) string) (str string) {
	first := true
	for _, v := range a {
		if first {
			first = false
		} else {
			str += ","
		}
		str += stringify(v)
	}
	return str
}

func (c *CliClient) prepareBatch(o *slurm.SBatchOptions, script string) *exec.Cmd {
	if o == nil {
		return exec.Command("sbatch", script)
	}

	var opts []string
	if len(o.Account) > 0 {
		opts = append(opts, "--account", o.Account)
	}
	if len(o.Array) > 0 {
		opts = append(opts, "--array", commaSeparatedArray(o.Array, strconv.Itoa))
	}
	if len(o.Begin) > 0 {
		opts = append(opts, "--begin", o.Begin)
	}
	if o.Wait {
		opts = append(opts, "--wait")
	}
	if len(o.Chdir) > 0 {
		opts = append(opts, "--chdir", o.Chdir)
	}
	if len(o.Gid) > 0 {
		opts = append(opts, "--gid", o.Gid)
	}
	if len(o.Uid) > 0 {
		opts = append(opts, "--uid", o.Uid)
	}

	return exec.Command("sbatch", append(opts, script)...)
}
