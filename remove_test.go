package main

import (
	"errors"
	"strings"
	"testing"
)

func TestRemove(t *testing.T) {
	// Save original functions and restore after test
	origGitRoot := gitRootFn
	origGitCmd := gitCmdFn
	defer func() {
		gitRootFn = origGitRoot
		gitCmdFn = origGitCmd
	}()

	t.Run("git root error", func(t *testing.T) {
		gitRootFn = func() (string, error) {
			return "", errors.New("not in a git repository")
		}

		err := remove("test-branch")
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("remove() error = %v, want 'not in a git repository'", err)
		}
	})

	t.Run("worktree remove fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" && args[1] == "remove" {
				return errors.New("worktree remove failed")
			}
			return nil
		}

		err := remove("test-branch")
		if err == nil || !strings.Contains(err.Error(), "failed to remove worktree") {
			t.Errorf("remove() error = %v, want error about failed to remove worktree", err)
		}
	})

	t.Run("branch delete fails", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "branch" && args[1] == "-D" {
				return errors.New("branch delete failed")
			}
			return nil
		}

		err := remove("test-branch")
		if err == nil || !strings.Contains(err.Error(), "failed to delete branch") {
			t.Errorf("remove() error = %v, want error about failed to delete branch", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}

		err := remove("test-branch")
		if err != nil {
			t.Errorf("remove() unexpected error: %v", err)
		}
	})
}
