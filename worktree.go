package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getwdFn is replaceable for testing
var getwdFn = os.Getwd

// WorktreeManager provides centralized worktree path management
type WorktreeManager struct {
	root string
}

// NewWorktreeManager creates a WorktreeManager after finding the main git root
// Uses gitMainRoot() to always get the main repository root, even when run from a worktree
func NewWorktreeManager() (*WorktreeManager, error) {
	root, err := gitMainRoot()
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

// CurrentWorktreeName returns the worktree name if cwd is inside a worktree, empty string otherwise
func (wm *WorktreeManager) CurrentWorktreeName() (string, error) {
	cwd, err := getwdFn()
	if err != nil {
		return "", nil // Not an error, just can't detect
	}

	worktreesPath := wm.WorktreesPath()
	if !strings.HasPrefix(cwd, worktreesPath+string(filepath.Separator)) {
		return "", nil // Not inside .worktrees
	}

	// Extract worktree name: cwd is like /repo/.worktrees/foo or /repo/.worktrees/foo/subdir
	// Since we've verified cwd starts with worktreesPath, Rel cannot fail
	rel, _ := filepath.Rel(worktreesPath, cwd)

	// Get first path component (the worktree name)
	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	return parts[0], nil
}
