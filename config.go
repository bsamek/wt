package main

// Version is set at build time via ldflags
var Version = "dev"

// Directory structure constants
const (
	WorktreesDir = ".worktrees"
	ClaudeDir    = ".claude"
	DefaultHook  = ".worktree-hook"
)
