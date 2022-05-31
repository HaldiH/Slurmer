package cliexecutor

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
)

type CommandContext struct {
	User    string   `json:"user"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Dir     string   `json:"dir"`
	Stdin   string   `json:"stdin"`
}

func NewCommandContext(user, command string, args ...string) *CommandContext {
	return &CommandContext{User: user, Command: command, Args: args}
}

type Executor interface {
	ExecCommand(cmdCtx *CommandContext, handlerFunc HandlerFunc, handlerErrFunc HandlerFunc) error
}

type executor struct {
	executorPath string
}

func NewExecutor(executorPath string) Executor {
	return &executor{executorPath: executorPath}
}

type HandlerFunc func(r io.Reader, waitErr error) error

func ReadStderr(r io.Reader, waitErr error) error {
	if waitErr == nil {
		return nil
	}
	errStr, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return errors.New(string(errStr))
}

func (e *executor) ExecCommand(cmdCtx *CommandContext, handlerFunc HandlerFunc, handlerErrFunc HandlerFunc) error {
	cmd := exec.Command(e.executorPath, "--stdin")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(stdin)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := encoder.Encode(cmdCtx); err != nil {
		return err
	}
	stdin.Close()

	waitErr := cmd.Wait()

	if handlerFunc != nil {
		if err := handlerFunc(stdout, waitErr); err != nil {
			return err
		}
	}

	if handlerErrFunc != nil {
		if err := handlerErrFunc(stderr, waitErr); err != nil {
			return err
		}
	}

	if handlerFunc == nil && handlerErrFunc == nil {
		return waitErr
	} else {
		return nil
	}
}
