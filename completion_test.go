package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListWorktrees(t *testing.T) {
	// Save original functions and restore after test
	origGitRoot := gitMainRootFn
	origListWorktrees := listWorktreesFn
	defer func() {
		gitMainRootFn = origGitRoot
		listWorktreesFn = origListWorktrees
	}()

	t.Run("git root error", func(t *testing.T) {
		listWorktreesFn = defaultListWorktrees
		gitMainRootFn = func() (string, error) {
			return "", errors.New("not in a git repository")
		}

		_, err := listWorktrees()
		if err == nil || err.Error() != "not in a git repository" {
			t.Errorf("listWorktrees() error = %v, want 'not in a git repository'", err)
		}
	})

	t.Run("no worktrees directory", func(t *testing.T) {
		listWorktreesFn = defaultListWorktrees
		tmpDir := t.TempDir()
		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		worktrees, err := listWorktrees()
		if err != nil {
			t.Errorf("listWorktrees() unexpected error: %v", err)
		}
		if len(worktrees) != 0 {
			t.Errorf("listWorktrees() = %v, want empty slice", worktrees)
		}
	})

	t.Run("empty worktrees directory", func(t *testing.T) {
		listWorktreesFn = defaultListWorktrees
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, ".worktrees"), 0755)
		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		worktrees, err := listWorktrees()
		if err != nil {
			t.Errorf("listWorktrees() unexpected error: %v", err)
		}
		if len(worktrees) != 0 {
			t.Errorf("listWorktrees() = %v, want empty slice", worktrees)
		}
	})

	t.Run("with worktrees", func(t *testing.T) {
		listWorktreesFn = defaultListWorktrees
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		os.MkdirAll(filepath.Join(worktreesDir, "feature-a"), 0755)
		os.MkdirAll(filepath.Join(worktreesDir, "feature-b"), 0755)
		// Create a file (should be ignored)
		os.WriteFile(filepath.Join(worktreesDir, "not-a-worktree"), []byte{}, 0644)

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		worktrees, err := listWorktrees()
		if err != nil {
			t.Errorf("listWorktrees() unexpected error: %v", err)
		}
		if len(worktrees) != 2 {
			t.Errorf("listWorktrees() returned %d worktrees, want 2", len(worktrees))
		}
		// Check both are present (order may vary)
		found := make(map[string]bool)
		for _, wt := range worktrees {
			found[wt] = true
		}
		if !found["feature-a"] || !found["feature-b"] {
			t.Errorf("listWorktrees() = %v, want [feature-a, feature-b]", worktrees)
		}
	})

	t.Run("readdir error", func(t *testing.T) {
		listWorktreesFn = defaultListWorktrees
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, ".worktrees")
		// Create as file instead of directory to cause ReadDir error
		os.WriteFile(worktreesDir, []byte{}, 0644)

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		_, err := listWorktrees()
		if err == nil {
			t.Error("listWorktrees() expected error for invalid directory")
		}
	})
}

func TestCompletion(t *testing.T) {
	t.Run("bash completion", func(t *testing.T) {
		var buf bytes.Buffer
		err := completion("bash", &buf)
		if err != nil {
			t.Errorf("completion(bash) unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "_wt_completions") {
			t.Error("bash completion missing _wt_completions function")
		}
		if !strings.Contains(output, "complete -F _wt_completions wt") {
			t.Error("bash completion missing complete command")
		}
		if !strings.Contains(output, "__complete jump") {
			t.Error("bash completion missing dynamic worktree completion for jump")
		}
	})

	t.Run("zsh completion", func(t *testing.T) {
		var buf bytes.Buffer
		err := completion("zsh", &buf)
		if err != nil {
			t.Errorf("completion(zsh) unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "#compdef wt") {
			t.Error("zsh completion missing #compdef")
		}
		if !strings.Contains(output, "_wt_worktrees") {
			t.Error("zsh completion missing _wt_worktrees function")
		}
		if !strings.Contains(output, "__complete jump") {
			t.Error("zsh completion missing dynamic worktree completion for jump")
		}
	})

	t.Run("fish completion", func(t *testing.T) {
		var buf bytes.Buffer
		err := completion("fish", &buf)
		if err != nil {
			t.Errorf("completion(fish) unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "__wt_worktrees") {
			t.Error("fish completion missing __wt_worktrees function")
		}
		if !strings.Contains(output, "complete -c wt") {
			t.Error("fish completion missing complete command")
		}
		if !strings.Contains(output, "__complete jump") {
			t.Error("fish completion missing dynamic worktree completion for jump")
		}
	})

	t.Run("unsupported shell", func(t *testing.T) {
		var buf bytes.Buffer
		err := completion("powershell", &buf)
		if err == nil {
			t.Error("completion(powershell) expected error")
		}
		if !strings.Contains(err.Error(), "unsupported shell") {
			t.Errorf("completion(powershell) error = %v, want 'unsupported shell'", err)
		}
	})
}

func TestBashCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := bashCompletion(&buf)
	if err != nil {
		t.Errorf("bashCompletion() unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("bashCompletion() wrote nothing")
	}
}

func TestZshCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := zshCompletion(&buf)
	if err != nil {
		t.Errorf("zshCompletion() unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("zshCompletion() wrote nothing")
	}
}

func TestFishCompletion(t *testing.T) {
	var buf bytes.Buffer
	err := fishCompletion(&buf)
	if err != nil {
		t.Errorf("fishCompletion() unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("fishCompletion() wrote nothing")
	}
}

func TestCompleteWorktrees(t *testing.T) {
	// Save original function and restore after test
	origListWorktrees := listWorktreesFn
	defer func() {
		listWorktreesFn = origListWorktrees
	}()

	t.Run("success", func(t *testing.T) {
		listWorktreesFn = func() ([]string, error) {
			return []string{"feature-a", "feature-b", "bugfix-c"}, nil
		}

		var buf bytes.Buffer
		err := completeWorktrees(&buf)
		if err != nil {
			t.Errorf("completeWorktrees() unexpected error: %v", err)
		}

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 3 {
			t.Errorf("completeWorktrees() output %d lines, want 3", len(lines))
		}
		expected := map[string]bool{"feature-a": true, "feature-b": true, "bugfix-c": true}
		for _, line := range lines {
			if !expected[line] {
				t.Errorf("completeWorktrees() unexpected line: %q", line)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		listWorktreesFn = func() ([]string, error) {
			return nil, errors.New("mock error")
		}

		var buf bytes.Buffer
		err := completeWorktrees(&buf)
		if err == nil || err.Error() != "mock error" {
			t.Errorf("completeWorktrees() error = %v, want 'mock error'", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		listWorktreesFn = func() ([]string, error) {
			return []string{}, nil
		}

		var buf bytes.Buffer
		err := completeWorktrees(&buf)
		if err != nil {
			t.Errorf("completeWorktrees() unexpected error: %v", err)
		}
		if buf.Len() != 0 {
			t.Errorf("completeWorktrees() wrote output for empty list: %q", buf.String())
		}
	})
}
