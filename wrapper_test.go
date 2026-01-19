package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildWtBinary builds the wt binary to a temp directory and returns the path.
func buildWtBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "wt")

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build wt binary: %v\n%s", err, output)
	}
	return binPath
}

func TestBashWrapperCompletionPassthrough(t *testing.T) {
	binPath := buildWtBinary(t)

	// Write wrapper script to temp file
	wrapperContent, err := os.ReadFile("wt.sh")
	if err != nil {
		t.Fatalf("failed to read wt.sh: %v", err)
	}
	wrapperPath := filepath.Join(t.TempDir(), "wt.sh")
	if err := os.WriteFile(wrapperPath, wrapperContent, 0644); err != nil {
		t.Fatalf("failed to write wrapper: %v", err)
	}

	// Execute via bash, setting PATH to include our binary
	binDir := filepath.Dir(binPath)
	script := "export PATH=" + binDir + ":$PATH && source " + wrapperPath + " && wt completion bash"
	cmd := exec.Command("bash", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wt completion bash failed: %v\n%s", err, output)
	}

	// Verify completion script content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "_wt_completions") {
		t.Errorf("bash wrapper did not pass through completion output.\nGot: %s", outputStr)
	}
	if !strings.Contains(outputStr, "complete -F _wt_completions wt") {
		t.Errorf("bash wrapper missing complete command in output.\nGot: %s", outputStr)
	}
}

func TestBashWrapperCompletePassthrough(t *testing.T) {
	binPath := buildWtBinary(t)

	// Write wrapper script to temp file
	wrapperContent, err := os.ReadFile("wt.sh")
	if err != nil {
		t.Fatalf("failed to read wt.sh: %v", err)
	}
	wrapperPath := filepath.Join(t.TempDir(), "wt.sh")
	if err := os.WriteFile(wrapperPath, wrapperContent, 0644); err != nil {
		t.Fatalf("failed to write wrapper: %v", err)
	}

	// Execute via bash, setting PATH to include our binary
	binDir := filepath.Dir(binPath)
	script := "export PATH=" + binDir + ":$PATH && source " + wrapperPath + " && wt __complete remove"
	cmd := exec.Command("bash", "-c", script)

	// __complete may return empty or worktree names, but should not error
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wt __complete remove failed: %v\n%s", err, output)
	}
	// Success - command executed without the wrapper swallowing output
}

func TestFishWrapperCompletionPassthrough(t *testing.T) {
	// Skip if fish is not installed
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish shell not installed, skipping test")
	}

	binPath := buildWtBinary(t)

	// Write wrapper script to temp file
	wrapperContent, err := os.ReadFile("wt.fish")
	if err != nil {
		t.Fatalf("failed to read wt.fish: %v", err)
	}
	wrapperPath := filepath.Join(t.TempDir(), "wt.fish")
	if err := os.WriteFile(wrapperPath, wrapperContent, 0644); err != nil {
		t.Fatalf("failed to write wrapper: %v", err)
	}

	// Execute via fish, setting PATH to include our binary
	binDir := filepath.Dir(binPath)
	script := "set -x PATH " + binDir + " $PATH; source " + wrapperPath + "; and wt completion fish"
	cmd := exec.Command("fish", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wt completion fish failed: %v\n%s", err, output)
	}

	// Verify completion script content is present
	outputStr := string(output)
	if !strings.Contains(outputStr, "__wt_worktrees") {
		t.Errorf("fish wrapper did not pass through completion output.\nGot: %s", outputStr)
	}
	if !strings.Contains(outputStr, "complete -c wt") {
		t.Errorf("fish wrapper missing complete command in output.\nGot: %s", outputStr)
	}
}

func TestFishWrapperCompletePassthrough(t *testing.T) {
	// Skip if fish is not installed
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish shell not installed, skipping test")
	}

	binPath := buildWtBinary(t)

	// Write wrapper script to temp file
	wrapperContent, err := os.ReadFile("wt.fish")
	if err != nil {
		t.Fatalf("failed to read wt.fish: %v", err)
	}
	wrapperPath := filepath.Join(t.TempDir(), "wt.fish")
	if err := os.WriteFile(wrapperPath, wrapperContent, 0644); err != nil {
		t.Fatalf("failed to write wrapper: %v", err)
	}

	// Execute via fish, setting PATH to include our binary
	binDir := filepath.Dir(binPath)
	script := "set -x PATH " + binDir + " $PATH; source " + wrapperPath + "; and wt __complete remove"
	cmd := exec.Command("fish", "-c", script)

	// __complete may return empty or worktree names, but should not error
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wt __complete remove failed: %v\n%s", err, output)
	}
	// Success - command executed without the wrapper swallowing output
}
