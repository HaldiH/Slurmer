package slurmcli

import (
	"strconv"

	"github.com/ShinoYasx/Slurmer/pkg/cliexecutor"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
)

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

func (c *cliClient) prepareBatch(o *slurm.SBatchOptions, script string, user string) *cliexecutor.CommandContext {
	var opts []string
	if o != nil {
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
	}

	return &cliexecutor.CommandContext{
		User:    user,
		Command: "sbatch",
		Args:    append(opts, script),
	}
}

func map2[T any, R any](v []T, mapper func(T) R) []R {
	mapped := make([]R, len(v))
	for i, e := range v {
		mapped[i] = mapper(e)
	}
	return mapped
}
