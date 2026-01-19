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

func TestRoot(t *testing.T) {
	// Save original functions and restore after test
	origGitRoot := gitMainRootFn
	origGetwd := getwdFn
	defer func() {
		gitMainRootFn = origGitRoot
		getwdFn = origGetwd
	}()

	t.Run("git root error", func(t *testing.T) {
		gitMainRootFn = func() (string, error) {
			return "", errors.New("not in a git repository")
		}

		err := root()
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("root() error = %v, want 'not in a git repository'", err)
		}
	})

	t.Run("inside worktree outputs root path", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreePath := filepath.Join(tmpDir, WorktreesDir, "my-feature")

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		getwdFn = func() (string, error) {
			return worktreePath, nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("root() unexpected error: %v", err)
		}
		if output != tmpDir {
			t.Errorf("root() stdout = %q, want %q", output, tmpDir)
		}
	})

	t.Run("inside worktree subdirectory outputs root path", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreePath := filepath.Join(tmpDir, WorktreesDir, "my-feature", "src", "components")

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		getwdFn = func() (string, error) {
			return worktreePath, nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("root() unexpected error: %v", err)
		}
		if output != tmpDir {
			t.Errorf("root() stdout = %q, want %q", output, tmpDir)
		}
	})

	t.Run("not inside worktree outputs nothing", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		getwdFn = func() (string, error) {
			return "/some/other/dir", nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("root() unexpected error: %v", err)
		}
		if output != "" {
			t.Errorf("root() stdout = %q, want empty", output)
		}
	})

	t.Run("at repository root outputs nothing", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		getwdFn = func() (string, error) {
			return tmpDir, nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("root() unexpected error: %v", err)
		}
		if output != "" {
			t.Errorf("root() stdout = %q, want empty", output)
		}
	})

	t.Run("getwd error is handled gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		getwdFn = func() (string, error) {
			return "", errors.New("getwd failed")
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("root() unexpected error: %v", err)
		}
		// Should not output anything when getwd fails
		if output != "" {
			t.Errorf("root() stdout = %q, want empty", output)
		}
	})
}
