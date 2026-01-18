package main

import (
	"fmt"
)

func remove(name string) error {
	wm, err := NewWorktreeManager()
	if err != nil {
		return err
	}

	worktreePath := wm.WorktreePath(name)

	// Remove worktree
	fmt.Printf("Removing worktree %s/%s\n", WorktreesDir, name)
	if err := gitCmd(wm.Root(), "worktree", "remove", worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete branch
	fmt.Printf("Deleting branch %s\n", name)
	if err := gitCmd(wm.Root(), "branch", "-D", name); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	fmt.Printf("Done! Worktree and branch removed\n")
	return nil
}
