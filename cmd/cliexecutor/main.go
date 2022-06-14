package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/ShinoYasx/Slurmer/pkg/cliexecutor"
)

const shell = "/bin/sh"

var (
	slurmerUid = "1000"
	minUid     = "1000"
	maxUid     = "65535"
)

func main() {
	uidPtr := flag.Int("uid", 0, "UID to run the program with (only if root or SlurmUser)")
	userPtr := flag.String("user", "", "User to run the program with (only if root or SlurmUser)")
	cmdPtr := flag.String("command", "", "Command to execute")
	stdinPtr := flag.Bool("stdin", false, "Read command from stdin")
	flag.Parse()

	slurmerUid, err := strconv.Atoi(slurmerUid)
	if err != nil {
		panic(err)
	}
	minUid, err := strconv.Atoi(minUid)
	if err != nil {
		panic(err)
	}
	maxUid, err := strconv.Atoi(maxUid)
	if err != nil {
		panic(err)
	}

	die := func(reason string, code int) {
		fmt.Fprintln(flag.CommandLine.Output(), reason)
		if !*stdinPtr {
			fmt.Fprintf(flag.CommandLine.Output(), "Try '%s --help' for more information.\n", os.Args[0])
		}
		os.Exit(code)
	}

	if *stdinPtr {
		var cmdCtx cliexecutor.CommandContext
		decoder := json.NewDecoder(os.Stdin)
		if err := decoder.Decode(&cmdCtx); err != nil {
			panic(err)
		}

		u, err := user.Lookup(cmdCtx.User)
		if err != nil {
			die("user not found", 2)
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			panic(err)
		}

		if !(minUid <= uid && uid <= maxUid) {
			die(fmt.Sprintf("user id must be between %d and %d", minUid, maxUid), 3)
		}

		ruid := syscall.Getuid()
		if ruid != slurmerUid && ruid != 0 {
			die("bad calling user", 4)
		}

		if err := syscall.Setuid(uid); err != nil {
			die(err.Error(), 5)
		}

		if len(cmdCtx.Dir) > 0 {
			os.Chdir(cmdCtx.Dir)
		}

		cmd := exec.Command(cmdCtx.Command, cmdCtx.Args...)

		stdin, err := cmd.StdinPipe()
		if err != nil {
			panic(err)
		}
		stdin.Write([]byte(cmdCtx.Stdin))
		stdin.Close()

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			die(err.Error(), 6)
		}
		return
	}

	isUidPassed := false
	isUserPassed := false

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "uid":
			isUidPassed = true
		case "user":
			isUserPassed = true
		}
	})

	if isUidPassed || isUserPassed {
		if isUidPassed && isUserPassed {
			die("cannot set both user and uid", 10)
		}

		if isUserPassed {
			u, err := user.Lookup(*userPtr)
			if err != nil {
				die(err.Error(), 2)
			}
			uid, err := strconv.Atoi(u.Uid)
			if err != nil {
				panic(err)
			}
			*uidPtr = uid
		}

		ruid := os.Getuid()
		if ruid != slurmerUid && ruid != 0 {
			die("bad calling user", 11)
		}

		if os.Geteuid() != 0 {
			die("setuid not set or not owned by root", 12)
		}

		wantedUid := *uidPtr
		if !(minUid <= wantedUid && wantedUid <= maxUid) {
			die(fmt.Sprintf("user id must be between %d and %d", minUid, maxUid), 13)
		}

		if err := syscall.Setuid(wantedUid); err != nil {
			panic(err)
		}
	}

	cmdString := *cmdPtr
	if len(cmdString) == 0 {
		die("no command given", 1)
	}

	if err := syscall.Exec(shell, []string{shell, "-c", cmdString}, os.Environ()); err != nil {
		panic(err)
	}
}
