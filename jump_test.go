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

func TestJump(t *testing.T) {
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

		err := jump("")
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("jump() error = %v, want 'not in a git repository'", err)
		}
	})

	t.Run("no name inside worktree outputs root path", func(t *testing.T) {
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

		err := jump("")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("jump() unexpected error: %v", err)
		}
		if output != tmpDir {
			t.Errorf("jump() stdout = %q, want %q", output, tmpDir)
		}
	})

	t.Run("no name inside worktree subdirectory outputs root path", func(t *testing.T) {
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

		err := jump("")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("jump() unexpected error: %v", err)
		}
		if output != tmpDir {
			t.Errorf("jump() stdout = %q, want %q", output, tmpDir)
		}
	})

	t.Run("no name not inside worktree outputs nothing", func(t *testing.T) {
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

		err := jump("")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("jump() unexpected error: %v", err)
		}
		if output != "" {
			t.Errorf("jump() stdout = %q, want empty", output)
		}
	})

	t.Run("no name at repository root outputs nothing", func(t *testing.T) {
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

		err := jump("")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("jump() unexpected error: %v", err)
		}
		if output != "" {
			t.Errorf("jump() stdout = %q, want empty", output)
		}
	})

	t.Run("no name getwd error is handled gracefully", func(t *testing.T) {
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

		err := jump("")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("jump() unexpected error: %v", err)
		}
		// Should not output anything when getwd fails
		if output != "" {
			t.Errorf("jump() stdout = %q, want empty", output)
		}
	})

	t.Run("with name to existing worktree outputs path", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		worktreePath := filepath.Join(worktreesDir, "my-feature")
		os.MkdirAll(worktreePath, 0755)

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := jump("my-feature")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Errorf("jump() unexpected error: %v", err)
		}
		if output != worktreePath {
			t.Errorf("jump() stdout = %q, want %q", output, worktreePath)
		}
	})

	t.Run("with name to non-existent worktree returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		err := jump("non-existent")
		if err == nil {
			t.Error("jump() expected error for non-existent worktree")
		}
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("jump() error = %v, want error containing 'does not exist'", err)
		}
	})
}
