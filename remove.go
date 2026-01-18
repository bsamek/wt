package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getwdFn is replaceable for testing
var getwdFn = os.Getwd

func remove(name string) error {
	wm, err := NewWorktreeManager()
	if err != nil {
		return err
	}

	worktreePath := wm.WorktreePath(name)

	// Check if we're currently inside the worktree being removed
	cwd, err := getwdFn()
	insideWorktree := err == nil && (cwd == worktreePath || strings.HasPrefix(cwd, worktreePath+string(filepath.Separator)))

	// Remove worktree
	fmt.Fprintf(os.Stderr, "Removing worktree %s/%s\n", WorktreesDir, name)
	if err := gitCmd(wm.Root(), "worktree", "remove", worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete branch
	fmt.Fprintf(os.Stderr, "Deleting branch %s\n", name)
	if err := gitCmd(wm.Root(), "branch", "-D", name); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Done! Worktree and branch removed")

	// Output path to stdout for shell wrapper to cd into
	// If we were inside the worktree, output root so shell can cd there
	// Otherwise, output empty line (no directory change needed)
	if insideWorktree {
		fmt.Println(wm.Root())
	}
	return nil
}
