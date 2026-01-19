package main

import (
	"testing"
)

func TestConstants(t *testing.T) {
	// Test directory constants are set
	t.Run("directory constants", func(t *testing.T) {
		if WorktreesDir != ".worktrees" {
			t.Errorf("WorktreesDir = %q, want %q", WorktreesDir, ".worktrees")
		}
		if ClaudeDir != ".claude" {
			t.Errorf("ClaudeDir = %q, want %q", ClaudeDir, ".claude")
		}
		if DefaultHook != ".worktree-hook" {
			t.Errorf("DefaultHook = %q, want %q", DefaultHook, ".worktree-hook")
		}
	})
}
