package main

import (
	"fmt"
	"os"
)

// jump outputs a worktree path for the shell wrapper to cd into.
// If name is empty, it navigates to the repository root (when inside a worktree).
// If name is provided, it navigates to that specific worktree.
func jump(name string) error {
	wm, err := NewWorktreeManager()
	if err != nil {
		return err
	}

	// No name = go to root
	if name == "" {
		currentName, _ := wm.CurrentWorktreeName()
		if currentName != "" {
			fmt.Println(wm.Root())
		}
		return nil
	}

	// Jump to specific worktree
	worktreePath := wm.WorktreePath(name)
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree %q does not exist", name)
	}
	fmt.Println(worktreePath)
	return nil
}
