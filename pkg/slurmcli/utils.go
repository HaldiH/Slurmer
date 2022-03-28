package slurmcli

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

func execCommand[T any](cmd *exec.Cmd) (filledStruct *T, err error) {
	jsonData, err := cmd.Output()
	if err != nil {
		return nil, errors.New("invalid slurm command")
	}

	filledStruct = new(T)
	err = json.Unmarshal(jsonData, filledStruct)
	if err != nil {
		return nil, err
	}

	return filledStruct, nil
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
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

func firstLine(reader io.Reader) string {
	bio := bufio.NewReader(reader)
	line, _ := bio.ReadBytes('\n')
	return strings.TrimRight(string(line), "\n")
}

func (c *CliClient) prepareBatch(o slurm.SBatchOptions) (cmd *exec.Cmd) {
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

	cmd = exec.Command("sbatch", opts...)
	cmd.Dir = o.Chdir

	return cmd
}
