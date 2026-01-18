package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Function variables for testing
var (
	gitRootFn     = defaultGitRoot
	gitMainRootFn = defaultGitMainRoot
	gitCmdFn      = defaultGitCmd
	filepathAbsFn = filepath.Abs
)

func gitRoot() (string, error) {
	return gitRootFn()
}

func gitMainRoot() (string, error) {
	return gitMainRootFn()
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

func defaultGitMainRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	gitDir := strings.TrimSpace(string(out))
	// gitDir is the .git directory (or path to it), parent is repo root
	absGitDir, err := filepathAbsFn(gitDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve git directory path")
	}
	return filepath.Dir(absGitDir), nil
}

func defaultGitCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr // Redirect to stderr to keep stdout clean for directory path
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
