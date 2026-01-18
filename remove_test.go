package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemove(t *testing.T) {
	// Save original functions and restore after test
	origGitRoot := gitRootFn
	origGitCmd := gitCmdFn
	origGetwd := getwdFn
	defer func() {
		gitRootFn = origGitRoot
		gitCmdFn = origGitCmd
		getwdFn = origGetwd
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
		getwdFn = func() (string, error) {
			return "/some/other/dir", nil
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
		getwdFn = func() (string, error) {
			return "/some/other/dir", nil
		}

		err := remove("test-branch")
		if err == nil || !strings.Contains(err.Error(), "failed to delete branch") {
			t.Errorf("remove() error = %v, want error about failed to delete branch", err)
		}
	})

	t.Run("success from outside worktree", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}
		getwdFn = func() (string, error) {
			return "/some/other/dir", nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := remove("test-branch")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("remove() unexpected error: %v", err)
		}
		// Should not output any path when not inside worktree
		if output != "" {
			t.Errorf("remove() stdout = %q, want empty", output)
		}
	})

	t.Run("success from inside worktree outputs root", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreePath := filepath.Join(tmpDir, WorktreesDir, "test-branch")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}
		getwdFn = func() (string, error) {
			return worktreePath, nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := remove("test-branch")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("remove() unexpected error: %v", err)
		}
		// Should output root path when inside worktree
		if output != tmpDir {
			t.Errorf("remove() stdout = %q, want %q", output, tmpDir)
		}
	})

	t.Run("success from inside worktree subdirectory outputs root", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreePath := filepath.Join(tmpDir, WorktreesDir, "test-branch")
		subDir := filepath.Join(worktreePath, "src", "components")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}
		getwdFn = func() (string, error) {
			return subDir, nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := remove("test-branch")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("remove() unexpected error: %v", err)
		}
		// Should output root path when inside worktree subdirectory
		if output != tmpDir {
			t.Errorf("remove() stdout = %q, want %q", output, tmpDir)
		}
	})

	t.Run("getwd error is handled gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}
		getwdFn = func() (string, error) {
			return "", errors.New("getwd failed")
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := remove("test-branch")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("remove() unexpected error: %v", err)
		}
		// Should not output any path when getwd fails
		if output != "" {
			t.Errorf("remove() stdout = %q, want empty", output)
		}
	})
}
