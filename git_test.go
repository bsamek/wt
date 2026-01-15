package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitRoot(t *testing.T) {
	// Save original function and restore after test
	origGitRoot := gitRootFn
	defer func() {
		gitRootFn = origGitRoot
	}()

	t.Run("delegates to gitRootFn", func(t *testing.T) {
		called := false
		gitRootFn = func() (string, error) {
			called = true
			return "/test/path", nil
		}

		result, err := gitRoot()
		if !called {
			t.Error("gitRoot() did not call gitRootFn")
		}
		if err != nil {
			t.Errorf("gitRoot() unexpected error: %v", err)
		}
		if result != "/test/path" {
			t.Errorf("gitRoot() = %q, want %q", result, "/test/path")
		}
	})
}

func TestGitCmd(t *testing.T) {
	// Save original function and restore after test
	origGitCmd := gitCmdFn
	defer func() {
		gitCmdFn = origGitCmd
	}()

	t.Run("delegates to gitCmdFn", func(t *testing.T) {
		var capturedDir string
		var capturedArgs []string
		gitCmdFn = func(dir string, args ...string) error {
			capturedDir = dir
			capturedArgs = args
			return nil
		}

		err := gitCmd("/test/dir", "status", "-s")
		if err != nil {
			t.Errorf("gitCmd() unexpected error: %v", err)
		}
		if capturedDir != "/test/dir" {
			t.Errorf("gitCmd() dir = %q, want %q", capturedDir, "/test/dir")
		}
		if len(capturedArgs) != 2 || capturedArgs[0] != "status" || capturedArgs[1] != "-s" {
			t.Errorf("gitCmd() args = %v, want [status -s]", capturedArgs)
		}
	})
}

func TestDefaultGitRoot(t *testing.T) {
	t.Run("not in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Change to temp dir that's not a git repo
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		_, err := defaultGitRoot()
		if err == nil {
			t.Error("defaultGitRoot() expected error when not in git repo")
		}
		if err.Error() != "not in a git repository" {
			t.Errorf("defaultGitRoot() error = %q, want %q", err.Error(), "not in a git repository")
		}
	})

	t.Run("in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize a git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Skipf("git init failed: %v", err)
		}

		// Change to the git repo
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		root, err := defaultGitRoot()
		if err != nil {
			t.Errorf("defaultGitRoot() unexpected error: %v", err)
		}

		// The root should be the tmpDir (accounting for symlinks)
		expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
		actualRoot, _ := filepath.EvalSymlinks(root)
		if actualRoot != expectedRoot {
			t.Errorf("defaultGitRoot() = %q, want %q", actualRoot, expectedRoot)
		}
	})
}

func TestDefaultGitCmd(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize a git repo
		initCmd := exec.Command("git", "init")
		initCmd.Dir = tmpDir
		if err := initCmd.Run(); err != nil {
			t.Skipf("git init failed: %v", err)
		}

		// Run a simple git command
		err := defaultGitCmd(tmpDir, "status")
		if err != nil {
			t.Errorf("defaultGitCmd() unexpected error: %v", err)
		}
	})

	t.Run("failing command", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Run a git command that should fail (not a git repo, invalid command)
		err := defaultGitCmd(tmpDir, "invalid-command-xyz")
		if err == nil {
			t.Error("defaultGitCmd() expected error for invalid command")
		}
	})
}
