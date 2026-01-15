package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Function variable for testing
var filepathRel = filepath.Rel

func create(name, hookPath string) error {
	// Find git root
	root, err := gitRoot()
	if err != nil {
		return err
	}

	// Check .worktrees directory exists
	worktreesDir := filepath.Join(root, ".worktrees")
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		return fmt.Errorf(".worktrees directory does not exist (create it first)")
	}

	worktreePath := filepath.Join(worktreesDir, name)

	// Create worktree with new branch
	fmt.Printf("Creating worktree at .worktrees/%s with branch %s\n", name, name)
	if err := gitCmd(root, "worktree", "add", worktreePath, "-b", name); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Copy .claude/ directory if it exists
	claudeDir := filepath.Join(root, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		fmt.Println("Copying .claude/ directory...")
		dstClaudeDir := filepath.Join(worktreePath, ".claude")
		if err := copyDir(claudeDir, dstClaudeDir); err != nil {
			return fmt.Errorf("failed to copy .claude/ directory: %w", err)
		}
	}

	// Run hook if it exists
	hookFullPath := filepath.Join(root, hookPath)
	if _, err := os.Stat(hookFullPath); err == nil {
		fmt.Printf("Running hook: %s\n", hookPath)
		if err := runHook(hookFullPath, worktreePath); err != nil {
			return fmt.Errorf("hook failed: %w", err)
		}
	}

	fmt.Printf("Done! Worktree ready at .worktrees/%s\n", name)
	return nil
}

func runHook(hookPath, worktreePath string) error {
	cmd := exec.Command(hookPath)
	cmd.Dir = worktreePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepathRel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}
