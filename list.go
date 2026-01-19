package main

import (
	"fmt"
	"io"
)

// list outputs all worktree names, one per line.
func list(w io.Writer) error {
	worktrees, err := listWorktrees()
	if err != nil {
		return err
	}
	for _, wt := range worktrees {
		fmt.Fprintln(w, wt)
	}
	return nil
}
