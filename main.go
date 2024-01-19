// Copyright 2024 D2iQ, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrNotFound          = errors.New("objects not found")
	ErrGettingObjects    = errors.New("error getting objects")
	ErrWaitingForObjects = errors.New("error waiting for objects")
)

func main() {
	// can either be called directly as a binary or as a kubectl plugin, handle both cases
	// kubectl-betterwait wait ...
	// kubectl betterwait ...
	args := os.Args[2:]
	// when called as a kubectl plugin, the args will not have the wait command
	if strings.HasSuffix(os.Args[0], "kubectl-betterwait") {
		args = os.Args[1:]
	}

	// kubectl wait will error if the objects don't exist https://github.com/kubernetes/kubectl/issues/1516
	// first wait for objects to actually exist
	if err := waitForObjectsToExist(args...); err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("%v\n", err))
		os.Exit(1)
	}

	// wait for the condition only after the objects exist
	if err := waitForCondition(args...); err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("%v\n", err))
		os.Exit(1)
	}

}

func waitForObjectsToExist(args ...string) error {
	// use the same args
	getArgs := kubectlGetArgsFromArgs(args...)

	// always try it at-least once
	err := getObjects(getArgs...)
	if err == nil {
		return nil
	}

	// setup a context with the passed in timeout
	effectiveTimeout, err := effectiveTimeoutFromArgs(args...)
	if err != nil {
		return err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), effectiveTimeout)
	defer cancelFunc()

	for {
		// retry every 10 seconds
		timer := time.NewTimer(time.Second * 10)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("timed out waiting for objects to exist")
		case <-timer.C:
			getErr := getObjects(getArgs...)
			if getErr == nil {
				return nil
			} else if errors.Is(getErr, ErrNotFound) {
				// retry on next timer tick
			} else {
				// exit with an error on all other errors
				return getErr
			}
		}
	}
}

func getObjects(args ...string) error {
	cmd := exec.Command(kubectl(), args...)
	cmd.Stdout = os.Stdout
	// the error message will be sent to cmd.Stderr and no the returned as err
	var errBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &errBuffer)
	cmd.Stderr = mw
	cmd.Stdin = os.Stdin
	logCmd(os.Stdout, cmd)
	err := cmd.Run()

	if isNotFound(errBuffer.String()) {
		return ErrNotFound
	} else if err != nil {
		// the error returned by kubectl is not meaningful, replace it with a better one
		return ErrGettingObjects
	}

	return nil
}

func waitForCondition(args ...string) error {
	// build the wait command using the original args
	waitArgs := []string{"wait"}
	waitArgs = append(waitArgs, args...)
	cmd := exec.Command(kubectl(), waitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	logCmd(os.Stdout, cmd)
	if err := cmd.Run(); err != nil {
		return ErrWaitingForObjects
	}

	return nil
}

func kubectlGetArgsFromArgs(args ...string) []string {
	getArgs := []string{"get"}
	for _, arg := range args {
		// pass all valid flags to kubectl get
		if isValidGetFlag(arg) {
			getArgs = append(getArgs, arg)
		}
	}

	return getArgs
}

func effectiveTimeoutFromArgs(args ...string) (time.Duration, error) {
	var timeout []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--timeout") {
			timeout = strings.FieldsFunc(arg, flagSplitter)
		}
	}

	// the default kubectl wait timeout is 30 seconds, use that if not set
	effectiveTimeout := time.Second * 30
	if len(timeout) == 2 {
		var err error
		effectiveTimeout, err = time.ParseDuration(timeout[1])
		if err != nil {
			return -1, fmt.Errorf("error parsing --timeout value: %v", err)
		}
	}

	// comes from kubectl wait code
	if effectiveTimeout < 0 {
		effectiveTimeout = 168 * time.Hour
	}

	return effectiveTimeout, nil
}

var (
	kubectlExecutableEnv = os.Getenv("KUBECTL_EXECUTABLE")
)

// kubectl expects the kubectl executable to either be in the PATH or set with KUBECTL_EXECUTABLE env
func kubectl() string {
	executable := "/bin/kubectl"
	if kubectlExecutableEnv != "" {
		executable = kubectlExecutableEnv
	}
	return executable
}

func logCmd(out io.Writer, cmd *exec.Cmd) {
	fmt.Fprintf(out, "$ %s\n", strings.Join(cmd.Args, " "))
}

func flagSplitter(r rune) bool {
	return r == '=' || r == ' '
}

func isNotFound(err string) bool {
	formattedErr := strings.ToLower(err)
	return strings.Contains(formattedErr, "not found") ||
		strings.Contains(formattedErr, "no matching resources found") ||
		strings.Contains(formattedErr, "no resources found ")
}

// isValidGetFlag returns true if arg is a valid kubectl get arg
// kubectl wait flags are a subset of kubectl get flags plus "--for" and "--timeout" flags
func isValidGetFlag(arg string) bool {
	return !strings.HasPrefix(arg, "--for") &&
		!strings.HasPrefix(arg, "--timeout")
}