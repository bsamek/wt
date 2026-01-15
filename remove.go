package main

import (
	"fmt"
	"path/filepath"
)

func remove(name string) error {
	// Find git root
	root, err := gitRoot()
	if err != nil {
		return err
	}

	worktreePath := filepath.Join(root, ".worktrees", name)

	// Remove worktree
	fmt.Printf("Removing worktree .worktrees/%s\n", name)
	if err := gitCmd(root, "worktree", "remove", worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete branch
	fmt.Printf("Deleting branch %s\n", name)
	if err := gitCmd(root, "branch", "-D", name); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	fmt.Printf("Done! Worktree and branch removed\n")
	return nil
}
