package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

const shell = "/bin/sh"
const slurmUid int = 1000
const (
	minUid = 1000
	maxUid = 65535
)

func main() {
	uidPtr := flag.Int("uid", 0, "UID to run the program with (only if root or SlurmUser)")
	userPtr := flag.String("user", "", "User to run the program with (only if root or SlurmUser)")
	cmdPtr := flag.String("command", "", "command to execute")
	flag.Parse()

	die := func(reason string, code int) {
		fmt.Fprintln(flag.CommandLine.Output(), reason)
		fmt.Fprintf(flag.CommandLine.Output(), "Try '%s --help' for more information.\n", os.Args[0])
		os.Exit(code)
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
		if ruid != slurmUid && ruid != 0 {
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
