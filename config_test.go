package main

import (
	"testing"
)

func TestVersion(t *testing.T) {
	// Test that Version has a default value
	if Version == "" {
		t.Error("Version should not be empty")
	}
	// Default value should be "dev" when not set via ldflags
	if Version != "dev" {
		t.Logf("Version = %q (expected 'dev' in test, may be set via ldflags)", Version)
	}
}

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
