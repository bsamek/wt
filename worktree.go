package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// WorktreeManager provides centralized worktree path management
type WorktreeManager struct {
	root string
}

// NewWorktreeManager creates a WorktreeManager after finding the git root
func NewWorktreeManager() (*WorktreeManager, error) {
	root, err := gitRoot()
	if err != nil {
		return nil, err
	}
	return &WorktreeManager{root: root}, nil
}

// Root returns the git repository root path
func (wm *WorktreeManager) Root() string {
	return wm.root
}

// WorktreesPath returns the path to the .worktrees directory
func (wm *WorktreeManager) WorktreesPath() string {
	return filepath.Join(wm.root, WorktreesDir)
}

// WorktreePath returns the path to a specific worktree
func (wm *WorktreeManager) WorktreePath(name string) string {
	return filepath.Join(wm.WorktreesPath(), name)
}

// ClaudePath returns the path to the .claude directory in the root
func (wm *WorktreeManager) ClaudePath() string {
	return filepath.Join(wm.root, ClaudeDir)
}

// HookPath returns the full path to a hook script
func (wm *WorktreeManager) HookPath(hookRelPath string) string {
	return filepath.Join(wm.root, hookRelPath)
}

// ValidateWorktreesDir checks that the .worktrees directory exists
func (wm *WorktreeManager) ValidateWorktreesDir() error {
	if _, err := os.Stat(wm.WorktreesPath()); os.IsNotExist(err) {
		return fmt.Errorf("%s directory does not exist (create it first)", WorktreesDir)
	}
	return nil
}

// ClaudeDirExists returns true if the .claude directory exists
func (wm *WorktreeManager) ClaudeDirExists() bool {
	_, err := os.Stat(wm.ClaudePath())
	return err == nil
}

// HookExists returns true if the hook file exists
func (wm *WorktreeManager) HookExists(hookRelPath string) bool {
	_, err := os.Stat(wm.HookPath(hookRelPath))
	return err == nil
}
