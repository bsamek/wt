package main

import (
	"testing"
	"time"
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

	// Test check status constants
	t.Run("check status constants", func(t *testing.T) {
		if CheckStatusQueued != "QUEUED" {
			t.Errorf("CheckStatusQueued = %q, want %q", CheckStatusQueued, "QUEUED")
		}
		if CheckStatusInProgress != "IN_PROGRESS" {
			t.Errorf("CheckStatusInProgress = %q, want %q", CheckStatusInProgress, "IN_PROGRESS")
		}
		if CheckStatusCompleted != "COMPLETED" {
			t.Errorf("CheckStatusCompleted = %q, want %q", CheckStatusCompleted, "COMPLETED")
		}
	})

	// Test check conclusion constants
	t.Run("check conclusion constants", func(t *testing.T) {
		if CheckConclusionSuccess != "SUCCESS" {
			t.Errorf("CheckConclusionSuccess = %q, want %q", CheckConclusionSuccess, "SUCCESS")
		}
		if CheckConclusionNeutral != "NEUTRAL" {
			t.Errorf("CheckConclusionNeutral = %q, want %q", CheckConclusionNeutral, "NEUTRAL")
		}
		if CheckConclusionSkipped != "SKIPPED" {
			t.Errorf("CheckConclusionSkipped = %q, want %q", CheckConclusionSkipped, "SKIPPED")
		}
		if CheckConclusionFailure != "FAILURE" {
			t.Errorf("CheckConclusionFailure = %q, want %q", CheckConclusionFailure, "FAILURE")
		}
		if CheckConclusionCancelled != "CANCELLED" {
			t.Errorf("CheckConclusionCancelled = %q, want %q", CheckConclusionCancelled, "CANCELLED")
		}
	})

	// Test marker constants
	t.Run("marker constants", func(t *testing.T) {
		if MarkerSuccess != "+" {
			t.Errorf("MarkerSuccess = %q, want %q", MarkerSuccess, "+")
		}
		if MarkerFailure != "x" {
			t.Errorf("MarkerFailure = %q, want %q", MarkerFailure, "x")
		}
		if MarkerPending != " " {
			t.Errorf("MarkerPending = %q, want %q", MarkerPending, " ")
		}
	})

	// Test timing constants
	t.Run("timing constants", func(t *testing.T) {
		if DefaultPollInterval != 30*time.Second {
			t.Errorf("DefaultPollInterval = %v, want %v", DefaultPollInterval, 30*time.Second)
		}
		if DefaultGHATimeout != 60*time.Minute {
			t.Errorf("DefaultGHATimeout = %v, want %v", DefaultGHATimeout, 60*time.Minute)
		}
	})

	// Test that GHATimeout defaults to DefaultGHATimeout
	t.Run("GHATimeout default", func(t *testing.T) {
		// Reset to default if modified
		origTimeout := GHATimeout
		defer func() { GHATimeout = origTimeout }()

		GHATimeout = DefaultGHATimeout
		if GHATimeout != DefaultGHATimeout {
			t.Errorf("GHATimeout = %v, want %v", GHATimeout, DefaultGHATimeout)
		}
	})
}
