package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func create(name, hookPath string) error {
	wm, err := NewWorktreeManager()
	if err != nil {
		return err
	}

	if err := wm.ValidateWorktreesDir(); err != nil {
		return err
	}

	worktreePath := wm.WorktreePath(name)

	// Create worktree with new branch
	fmt.Fprintf(os.Stderr, "Creating worktree at %s/%s with branch %s\n", WorktreesDir, name, name)
	if err := gitCmd(wm.Root(), "worktree", "add", worktreePath, "-b", name); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Create symlink to .claude/ directory if it exists
	if wm.ClaudeDirExists() {
		fmt.Fprintf(os.Stderr, "Creating symlink to %s/ directory...\n", ClaudeDir)
		dstClaudeDir := filepath.Join(worktreePath, ClaudeDir)
		if err := os.Symlink(wm.ClaudePath(), dstClaudeDir); err != nil {
			return fmt.Errorf("failed to create %s/ symlink: %w", ClaudeDir, err)
		}
	}

	// Run hook if it exists
	if wm.HookExists(hookPath) {
		fmt.Fprintf(os.Stderr, "Running hook: %s\n", hookPath)
		if err := runHook(wm.HookPath(hookPath), worktreePath); err != nil {
			return fmt.Errorf("hook failed: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Done! Worktree ready at %s/%s\n", WorktreesDir, name)
	// Output path to stdout for shell wrapper to cd into
	fmt.Println(worktreePath)
	return nil
}

func runHook(hookPath, worktreePath string) error {
	cmd := exec.Command(hookPath)
	cmd.Dir = worktreePath
	cmd.Stdout = os.Stderr // Redirect to stderr to keep stdout clean for worktree path
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
