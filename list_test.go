package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	// Save original function and restore after test
	origListWorktrees := listWorktreesFn
	defer func() {
		listWorktreesFn = origListWorktrees
	}()

	t.Run("success with worktrees", func(t *testing.T) {
		listWorktreesFn = func() ([]string, error) {
			return []string{"feature-a", "feature-b", "bugfix-c"}, nil
		}

		var buf bytes.Buffer
		err := list(&buf)
		if err != nil {
			t.Errorf("list() unexpected error: %v", err)
		}

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 3 {
			t.Errorf("list() output %d lines, want 3", len(lines))
		}
		expected := map[string]bool{"feature-a": true, "feature-b": true, "bugfix-c": true}
		for _, line := range lines {
			if !expected[line] {
				t.Errorf("list() unexpected line: %q", line)
			}
		}
	})

	t.Run("success with no worktrees", func(t *testing.T) {
		listWorktreesFn = func() ([]string, error) {
			return []string{}, nil
		}

		var buf bytes.Buffer
		err := list(&buf)
		if err != nil {
			t.Errorf("list() unexpected error: %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("list() wrote output for empty list: %q", buf.String())
		}
	})

	t.Run("error from listWorktrees", func(t *testing.T) {
		listWorktreesFn = func() ([]string, error) {
			return nil, errors.New("not in a git repository")
		}

		var buf bytes.Buffer
		err := list(&buf)
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("list() error = %v, want 'not in a git repository'", err)
		}
	})
}
