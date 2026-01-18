package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewWorktreeManager(t *testing.T) {
	// Save original function and restore after test
	origGitMainRoot := gitMainRootFn
	defer func() {
		gitMainRootFn = origGitMainRoot
	}()

	t.Run("success", func(t *testing.T) {
		gitMainRootFn = func() (string, error) {
			return "/test/repo", nil
		}

		wm, err := NewWorktreeManager()
		if err != nil {
			t.Errorf("NewWorktreeManager() unexpected error: %v", err)
		}
		if wm == nil {
			t.Fatal("NewWorktreeManager() returned nil")
		}
		if wm.Root() != "/test/repo" {
			t.Errorf("wm.Root() = %q, want %q", wm.Root(), "/test/repo")
		}
	})

	t.Run("error", func(t *testing.T) {
		gitMainRootFn = func() (string, error) {
			return "", errors.New("not in a git repository")
		}

		wm, err := NewWorktreeManager()
		if err == nil {
			t.Error("NewWorktreeManager() expected error")
		}
		if wm != nil {
			t.Error("NewWorktreeManager() should return nil on error")
		}
	})
}

func TestWorktreeManagerPaths(t *testing.T) {
	wm := &WorktreeManager{root: "/test/repo"}

	t.Run("Root", func(t *testing.T) {
		if wm.Root() != "/test/repo" {
			t.Errorf("Root() = %q, want %q", wm.Root(), "/test/repo")
		}
	})

	t.Run("WorktreesPath", func(t *testing.T) {
		expected := filepath.Join("/test/repo", WorktreesDir)
		if wm.WorktreesPath() != expected {
			t.Errorf("WorktreesPath() = %q, want %q", wm.WorktreesPath(), expected)
		}
	})

	t.Run("WorktreePath", func(t *testing.T) {
		expected := filepath.Join("/test/repo", WorktreesDir, "my-branch")
		if wm.WorktreePath("my-branch") != expected {
			t.Errorf("WorktreePath() = %q, want %q", wm.WorktreePath("my-branch"), expected)
		}
	})

	t.Run("ClaudePath", func(t *testing.T) {
		expected := filepath.Join("/test/repo", ClaudeDir)
		if wm.ClaudePath() != expected {
			t.Errorf("ClaudePath() = %q, want %q", wm.ClaudePath(), expected)
		}
	})

	t.Run("HookPath", func(t *testing.T) {
		expected := filepath.Join("/test/repo", "custom-hook.sh")
		if wm.HookPath("custom-hook.sh") != expected {
			t.Errorf("HookPath() = %q, want %q", wm.HookPath("custom-hook.sh"), expected)
		}
	})
}

func TestWorktreeManagerValidateWorktreesDir(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, WorktreesDir), 0755)

		wm := &WorktreeManager{root: tmpDir}
		err := wm.ValidateWorktreesDir()
		if err != nil {
			t.Errorf("ValidateWorktreesDir() unexpected error: %v", err)
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		wm := &WorktreeManager{root: tmpDir}
		err := wm.ValidateWorktreesDir()
		if err == nil {
			t.Error("ValidateWorktreesDir() expected error")
		}
	})
}

func TestWorktreeManagerClaudeDirExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, ClaudeDir), 0755)

		wm := &WorktreeManager{root: tmpDir}
		if !wm.ClaudeDirExists() {
			t.Error("ClaudeDirExists() = false, want true")
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		wm := &WorktreeManager{root: tmpDir}
		if wm.ClaudeDirExists() {
			t.Error("ClaudeDirExists() = true, want false")
		}
	})
}

func TestWorktreeManagerHookExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		hookPath := filepath.Join(tmpDir, "hook.sh")
		os.WriteFile(hookPath, []byte("#!/bin/sh\n"), 0755)

		wm := &WorktreeManager{root: tmpDir}
		if !wm.HookExists("hook.sh") {
			t.Error("HookExists() = false, want true")
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		wm := &WorktreeManager{root: tmpDir}
		if wm.HookExists("nonexistent.sh") {
			t.Error("HookExists() = true, want false")
		}
	})
}

func TestCurrentWorktreeName(t *testing.T) {
	// Save original function and restore after test
	origGetwd := getwdFn
	defer func() {
		getwdFn = origGetwd
	}()

	t.Run("inside worktree root", func(t *testing.T) {
		tmpDir := "/test/repo"
		wm := &WorktreeManager{root: tmpDir}

		getwdFn = func() (string, error) {
			return filepath.Join(tmpDir, WorktreesDir, "my-feature"), nil
		}

		name, err := wm.CurrentWorktreeName()
		if err != nil {
			t.Errorf("CurrentWorktreeName() unexpected error: %v", err)
		}
		if name != "my-feature" {
			t.Errorf("CurrentWorktreeName() = %q, want %q", name, "my-feature")
		}
	})

	t.Run("inside worktree subdirectory", func(t *testing.T) {
		tmpDir := "/test/repo"
		wm := &WorktreeManager{root: tmpDir}

		getwdFn = func() (string, error) {
			return filepath.Join(tmpDir, WorktreesDir, "my-feature", "src", "components"), nil
		}

		name, err := wm.CurrentWorktreeName()
		if err != nil {
			t.Errorf("CurrentWorktreeName() unexpected error: %v", err)
		}
		if name != "my-feature" {
			t.Errorf("CurrentWorktreeName() = %q, want %q", name, "my-feature")
		}
	})

	t.Run("not inside worktree - in repo root", func(t *testing.T) {
		tmpDir := "/test/repo"
		wm := &WorktreeManager{root: tmpDir}

		getwdFn = func() (string, error) {
			return tmpDir, nil
		}

		name, err := wm.CurrentWorktreeName()
		if err != nil {
			t.Errorf("CurrentWorktreeName() unexpected error: %v", err)
		}
		if name != "" {
			t.Errorf("CurrentWorktreeName() = %q, want empty string", name)
		}
	})

	t.Run("not inside worktree - different directory", func(t *testing.T) {
		tmpDir := "/test/repo"
		wm := &WorktreeManager{root: tmpDir}

		getwdFn = func() (string, error) {
			return "/some/other/directory", nil
		}

		name, err := wm.CurrentWorktreeName()
		if err != nil {
			t.Errorf("CurrentWorktreeName() unexpected error: %v", err)
		}
		if name != "" {
			t.Errorf("CurrentWorktreeName() = %q, want empty string", name)
		}
	})

	t.Run("getwd fails", func(t *testing.T) {
		tmpDir := "/test/repo"
		wm := &WorktreeManager{root: tmpDir}

		getwdFn = func() (string, error) {
			return "", errors.New("getwd failed")
		}

		name, err := wm.CurrentWorktreeName()
		if err != nil {
			t.Errorf("CurrentWorktreeName() unexpected error: %v", err)
		}
		if name != "" {
			t.Errorf("CurrentWorktreeName() = %q, want empty string", name)
		}
	})
}
