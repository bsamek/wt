package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Function variables for testing
var (
	gitRootFn = defaultGitRoot
	gitCmdFn  = defaultGitCmd
)

func gitRoot() (string, error) {
	return gitRootFn()
}

func gitCmd(dir string, args ...string) error {
	return gitCmdFn(dir, args...)
}

func defaultGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func defaultGitCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
