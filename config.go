package main

import "time"

// Version is set at build time via ldflags
var Version = "dev"

// Directory structure constants
const (
	WorktreesDir = ".worktrees"
	ClaudeDir    = ".claude"
	DefaultHook  = ".worktree-hook"
)

// GitHub Actions check statuses
const (
	CheckStatusQueued     = "QUEUED"
	CheckStatusInProgress = "IN_PROGRESS"
	CheckStatusCompleted  = "COMPLETED"
)

// GitHub Actions check conclusions
const (
	CheckConclusionSuccess   = "SUCCESS"
	CheckConclusionNeutral   = "NEUTRAL"
	CheckConclusionSkipped   = "SKIPPED"
	CheckConclusionFailure   = "FAILURE"
	CheckConclusionCancelled = "CANCELLED"
)

// Check detail markers
const (
	MarkerSuccess = "+"
	MarkerFailure = "x"
	MarkerPending = " "
)

// GitHub Actions polling configuration
const (
	DefaultPollInterval = 30 * time.Second
	DefaultGHATimeout   = 60 * time.Minute
)

// GHATimeout is the timeout for GitHub Actions monitoring (configurable for testing)
var GHATimeout = DefaultGHATimeout
