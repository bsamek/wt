package main

import "fmt"

// root outputs the repository root path if inside a worktree.
// If already at the root or not in a worktree, it's a no-op (outputs nothing).
// The shell wrapper will cd to the output path if one is printed.
func root() error {
	wm, err := NewWorktreeManager()
	if err != nil {
		return err
	}

	// Check if we're inside a worktree
	name, _ := wm.CurrentWorktreeName()
	if name != "" {
		// Inside a worktree - output root path for shell wrapper to cd
		fmt.Println(wm.Root())
	}
	// If not in a worktree, do nothing (no-op)

	return nil
}
