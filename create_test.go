package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
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

		err := create("test-branch", DefaultHook)
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("create() error = %v, want 'not in a git repository'", err)
		}
	})

	t.Run("worktrees dir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}

		err := create("test-branch", DefaultHook)
		if err == nil || !strings.Contains(err.Error(), WorktreesDir+" directory does not exist") {
			t.Errorf("create() error = %v, want error about %s not existing", err, WorktreesDir)
		}
	})

	t.Run("git worktree add fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, WorktreesDir), 0755)

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" {
				return errors.New("git worktree failed")
			}
			return nil
		}

		err := create("test-branch", DefaultHook)
		if err == nil || !strings.Contains(err.Error(), "failed to create worktree") {
			t.Errorf("create() error = %v, want error about failed to create worktree", err)
		}
	})

	t.Run("success without hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, WorktreesDir), 0755)

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}

		err := create("test-branch", DefaultHook)
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}
	})

	t.Run("success with hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		os.MkdirAll(worktreesDir, 0755)

		// Create a hook script that succeeds
		hookPath := filepath.Join(tmpDir, DefaultHook)
		err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 0\n"), 0755)
		if err != nil {
			t.Fatalf("failed to create hook: %v", err)
		}

		// Create the worktree directory (simulating git worktree add)
		worktreePath := filepath.Join(worktreesDir, "test-branch")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			// Simulate git worktree add by creating the directory
			if len(args) > 0 && args[0] == "worktree" {
				os.MkdirAll(worktreePath, 0755)
			}
			return nil
		}

		err = create("test-branch", DefaultHook)
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}
	})

	t.Run("hook fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		os.MkdirAll(worktreesDir, 0755)

		// Create a hook script that fails
		hookPath := filepath.Join(tmpDir, DefaultHook)
		err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 1\n"), 0755)
		if err != nil {
			t.Fatalf("failed to create hook: %v", err)
		}

		// Create the worktree directory
		worktreePath := filepath.Join(worktreesDir, "test-branch")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" {
				os.MkdirAll(worktreePath, 0755)
			}
			return nil
		}

		err = create("test-branch", DefaultHook)
		if err == nil || !strings.Contains(err.Error(), "hook failed") {
			t.Errorf("create() error = %v, want error about hook failed", err)
		}
	})

	t.Run("custom hook path", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		os.MkdirAll(worktreesDir, 0755)

		// Create a custom hook script
		hookPath := filepath.Join(tmpDir, "custom-hook.sh")
		err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 0\n"), 0755)
		if err != nil {
			t.Fatalf("failed to create hook: %v", err)
		}

		worktreePath := filepath.Join(worktreesDir, "test-branch")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" {
				os.MkdirAll(worktreePath, 0755)
			}
			return nil
		}

		err = create("test-branch", "custom-hook.sh")
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}
	})

	t.Run("creates symlink to .claude directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		os.MkdirAll(worktreesDir, 0755)

		// Create .claude directory with a file
		claudeDir := filepath.Join(tmpDir, ClaudeDir)
		os.MkdirAll(claudeDir, 0755)
		os.WriteFile(filepath.Join(claudeDir, "test.md"), []byte("test content"), 0644)

		worktreePath := filepath.Join(worktreesDir, "test-branch")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" {
				os.MkdirAll(worktreePath, 0755)
			}
			return nil
		}

		err := create("test-branch", DefaultHook)
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}

		// Verify symlink was created
		symlinkPath := filepath.Join(worktreePath, ClaudeDir)
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("failed to stat symlink: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be a symlink", ClaudeDir)
		}

		// Verify symlink points to correct location
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}
		if target != claudeDir {
			t.Errorf("symlink target = %v, want %v", target, claudeDir)
		}
	})

	t.Run("symlink creation fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		os.MkdirAll(worktreesDir, 0755)

		// Create .claude directory
		claudeDir := filepath.Join(tmpDir, ClaudeDir)
		os.MkdirAll(claudeDir, 0755)

		worktreePath := filepath.Join(worktreesDir, "test-branch")

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" {
				os.MkdirAll(worktreePath, 0755)
				// Create a file at .claude path to make symlink fail
				os.WriteFile(filepath.Join(worktreePath, ClaudeDir), []byte("block"), 0644)
			}
			return nil
		}

		err := create("test-branch", DefaultHook)
		if err == nil || !strings.Contains(err.Error(), "failed to create "+ClaudeDir+"/ symlink") {
			t.Errorf("create() error = %v, want error about failed to create symlink", err)
		}
	})
}

func TestRunHook(t *testing.T) {
	t.Run("successful hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		hookPath := filepath.Join(tmpDir, "hook.sh")
		err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 0\n"), 0755)
		if err != nil {
			t.Fatalf("failed to create hook: %v", err)
		}

		err = runHook(hookPath, tmpDir)
		if err != nil {
			t.Errorf("runHook() unexpected error: %v", err)
		}
	})

	t.Run("failing hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		hookPath := filepath.Join(tmpDir, "hook.sh")
		err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 42\n"), 0755)
		if err != nil {
			t.Fatalf("failed to create hook: %v", err)
		}

		err = runHook(hookPath, tmpDir)
		if err == nil {
			t.Error("runHook() expected error for failing hook")
		}
	})

	t.Run("non-existent hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := runHook(filepath.Join(tmpDir, "nonexistent.sh"), tmpDir)
		if err == nil {
			t.Error("runHook() expected error for non-existent hook")
		}
	})
}
