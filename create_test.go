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

		err := create("test-branch", ".worktree-hook")
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("create() error = %v, want 'not in a git repository'", err)
		}
	})

	t.Run("worktrees dir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}

		err := create("test-branch", ".worktree-hook")
		if err == nil || !strings.Contains(err.Error(), ".worktrees directory does not exist") {
			t.Errorf("create() error = %v, want error about .worktrees not existing", err)
		}
	})

	t.Run("git worktree add fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, ".worktrees"), 0755)

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			if len(args) > 0 && args[0] == "worktree" {
				return errors.New("git worktree failed")
			}
			return nil
		}

		err := create("test-branch", ".worktree-hook")
		if err == nil || !strings.Contains(err.Error(), "failed to create worktree") {
			t.Errorf("create() error = %v, want error about failed to create worktree", err)
		}
	})

	t.Run("success without hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, ".worktrees"), 0755)

		gitRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}

		err := create("test-branch", ".worktree-hook")
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}
	})

	t.Run("success with hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		os.MkdirAll(worktreesDir, 0755)

		// Create a hook script that succeeds
		hookPath := filepath.Join(tmpDir, ".worktree-hook")
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

		err = create("test-branch", ".worktree-hook")
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}
	})

	t.Run("hook fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		os.MkdirAll(worktreesDir, 0755)

		// Create a hook script that fails
		hookPath := filepath.Join(tmpDir, ".worktree-hook")
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

		err = create("test-branch", ".worktree-hook")
		if err == nil || !strings.Contains(err.Error(), "hook failed") {
			t.Errorf("create() error = %v, want error about hook failed", err)
		}
	})

	t.Run("custom hook path", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
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
}

func TestCopyDir(t *testing.T) {
	t.Run("copies files successfully", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		// Create a file in source
		err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644)
		if err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		err = copyDir(srcDir, dstDir)
		if err != nil {
			t.Errorf("copyDir() unexpected error: %v", err)
		}

		// Verify file was copied
		data, err := os.ReadFile(filepath.Join(dstDir, "file.txt"))
		if err != nil {
			t.Errorf("failed to read copied file: %v", err)
		}
		if string(data) != "hello" {
			t.Errorf("copied file content = %q, want %q", string(data), "hello")
		}
	})

	t.Run("copies nested directories", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		// Create nested structure
		nestedDir := filepath.Join(srcDir, "subdir")
		os.MkdirAll(nestedDir, 0755)
		err := os.WriteFile(filepath.Join(nestedDir, "nested.txt"), []byte("nested content"), 0644)
		if err != nil {
			t.Fatalf("failed to create nested file: %v", err)
		}

		err = copyDir(srcDir, dstDir)
		if err != nil {
			t.Errorf("copyDir() unexpected error: %v", err)
		}

		// Verify nested file was copied
		data, err := os.ReadFile(filepath.Join(dstDir, "subdir", "nested.txt"))
		if err != nil {
			t.Errorf("failed to read nested file: %v", err)
		}
		if string(data) != "nested content" {
			t.Errorf("nested file content = %q, want %q", string(data), "nested content")
		}
	})

	t.Run("source directory does not exist", func(t *testing.T) {
		dstDir := filepath.Join(t.TempDir(), "dst")
		err := copyDir("/nonexistent/path", dstDir)
		if err == nil {
			t.Error("copyDir() expected error for non-existent source")
		}
	})

	t.Run("empty source directory", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		err := copyDir(srcDir, dstDir)
		if err != nil {
			t.Errorf("copyDir() unexpected error: %v", err)
		}

		// Verify destination was created
		info, err := os.Stat(dstDir)
		if err != nil {
			t.Errorf("destination directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("destination is not a directory")
		}
	})

	t.Run("unreadable file in source", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		// Create a file and make it unreadable
		filePath := filepath.Join(srcDir, "unreadable.txt")
		err := os.WriteFile(filePath, []byte("secret"), 0000)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		defer os.Chmod(filePath, 0644) // Restore for cleanup

		err = copyDir(srcDir, dstDir)
		if err == nil {
			t.Error("copyDir() expected error for unreadable file")
		}
	})

	t.Run("filepath.Rel error", func(t *testing.T) {
		// Save original function
		origFilepathRel := filepathRel
		defer func() { filepathRel = origFilepathRel }()

		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		// Create a file in source
		err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644)
		if err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		// Mock filepathRel to return an error
		filepathRel = func(basepath, targpath string) (string, error) {
			return "", errors.New("cannot make path relative")
		}

		err = copyDir(srcDir, dstDir)
		if err == nil || !strings.Contains(err.Error(), "cannot make path relative") {
			t.Errorf("copyDir() error = %v, want error about cannot make path relative", err)
		}
	})
}

func TestCreateWithClaudeDir(t *testing.T) {
	// Save original functions and restore after test
	origGitRoot := gitRootFn
	origGitCmd := gitCmdFn
	defer func() {
		gitRootFn = origGitRoot
		gitCmdFn = origGitCmd
	}()

	t.Run("copies .claude directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		os.MkdirAll(worktreesDir, 0755)

		// Create .claude directory with content
		claudeDir := filepath.Join(tmpDir, ".claude")
		os.MkdirAll(claudeDir, 0755)
		err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"key": "value"}`), 0644)
		if err != nil {
			t.Fatalf("failed to create claude settings: %v", err)
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

		err = create("test-branch", ".worktree-hook")
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}

		// Verify .claude directory was copied
		data, err := os.ReadFile(filepath.Join(worktreePath, ".claude", "settings.json"))
		if err != nil {
			t.Errorf("failed to read copied claude settings: %v", err)
		}
		if string(data) != `{"key": "value"}` {
			t.Errorf("copied settings content = %q, want %q", string(data), `{"key": "value"}`)
		}
	})

	t.Run("copies .claude directory with hook", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		os.MkdirAll(worktreesDir, 0755)

		// Create .claude directory
		claudeDir := filepath.Join(tmpDir, ".claude")
		os.MkdirAll(claudeDir, 0755)
		os.WriteFile(filepath.Join(claudeDir, "config.txt"), []byte("config"), 0644)

		// Create hook
		hookPath := filepath.Join(tmpDir, ".worktree-hook")
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

		err = create("test-branch", ".worktree-hook")
		if err != nil {
			t.Errorf("create() unexpected error: %v", err)
		}

		// Verify .claude directory was copied
		_, err = os.Stat(filepath.Join(worktreePath, ".claude", "config.txt"))
		if err != nil {
			t.Errorf(".claude directory was not copied: %v", err)
		}
	})

	t.Run("copyDir failure is handled", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		os.MkdirAll(worktreesDir, 0755)

		// Create .claude directory with unreadable file
		claudeDir := filepath.Join(tmpDir, ".claude")
		os.MkdirAll(claudeDir, 0755)
		unreadablePath := filepath.Join(claudeDir, "unreadable.txt")
		os.WriteFile(unreadablePath, []byte("secret"), 0000)
		defer os.Chmod(unreadablePath, 0644)

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

		err := create("test-branch", ".worktree-hook")
		if err == nil || !strings.Contains(err.Error(), "failed to copy .claude/ directory") {
			t.Errorf("create() error = %v, want error about failed to copy .claude/ directory", err)
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
