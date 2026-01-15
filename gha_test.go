package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestGha(t *testing.T) {
	// Save original functions
	origGhPRView := ghPRViewFn
	origSleep := sleepFn
	defer func() {
		ghPRViewFn = origGhPRView
		sleepFn = origSleep
	}()

	t.Run("no PR found", func(t *testing.T) {
		ghPRViewFn = func() (*PRStatus, error) {
			return nil, errors.New("no PR found for current branch")
		}

		err := gha()
		if err == nil || !strings.Contains(err.Error(), "no PR found") {
			t.Errorf("gha() error = %v, want error about no PR found", err)
		}
	})

	t.Run("all checks pass immediately", func(t *testing.T) {
		ghPRViewFn = func() (*PRStatus, error) {
			return &PRStatus{
				Number: 123,
				State:  "OPEN",
				StatusCheckRollup: []CheckStatus{
					{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
					{Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"},
				},
			}, nil
		}

		err := gha()
		if err != nil {
			t.Errorf("gha() unexpected error: %v", err)
		}
	})

	t.Run("check fails", func(t *testing.T) {
		ghPRViewFn = func() (*PRStatus, error) {
			return &PRStatus{
				Number: 123,
				State:  "OPEN",
				StatusCheckRollup: []CheckStatus{
					{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
					{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE"},
				},
			}, nil
		}

		err := gha()
		if err == nil || !strings.Contains(err.Error(), "checks failed") {
			t.Errorf("gha() error = %v, want error about checks failed", err)
		}
	})

	t.Run("polls until complete", func(t *testing.T) {
		callCount := 0
		ghPRViewFn = func() (*PRStatus, error) {
			callCount++
			if callCount < 3 {
				return &PRStatus{
					Number: 123,
					State:  "OPEN",
					StatusCheckRollup: []CheckStatus{
						{Name: "build", Status: "IN_PROGRESS", Conclusion: ""},
					},
				}, nil
			}
			return &PRStatus{
				Number: 123,
				State:  "OPEN",
				StatusCheckRollup: []CheckStatus{
					{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
				},
			}, nil
		}

		sleepFn = func(d time.Duration) {
			// Don't actually sleep in tests
		}

		err := gha()
		if err != nil {
			t.Errorf("gha() unexpected error: %v", err)
		}
		if callCount != 3 {
			t.Errorf("gha() called gh %d times, want 3", callCount)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		// Save original timeout and restore after test
		origTimeout := ghaTimeout
		defer func() { ghaTimeout = origTimeout }()

		// Set a very short timeout
		ghaTimeout = 1 * time.Nanosecond

		ghPRViewFn = func() (*PRStatus, error) {
			// Simulate time passing
			time.Sleep(10 * time.Millisecond)
			return &PRStatus{
				Number: 123,
				State:  "OPEN",
				StatusCheckRollup: []CheckStatus{
					{Name: "build", Status: "IN_PROGRESS", Conclusion: ""},
				},
			}, nil
		}

		sleepFn = func(d time.Duration) {
			// Don't actually sleep in tests
		}

		err := gha()
		if err == nil || !strings.Contains(err.Error(), "timeout") {
			t.Errorf("gha() error = %v, want timeout error", err)
		}
	})
}

func TestAnalyzeChecks(t *testing.T) {
	tests := []struct {
		name       string
		checks     []CheckStatus
		wantResult CheckResult
	}{
		{
			name:       "empty checks",
			checks:     []CheckStatus{},
			wantResult: CheckResultPending,
		},
		{
			name: "all success",
			checks: []CheckStatus{
				{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
				{Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"},
			},
			wantResult: CheckResultSuccess,
		},
		{
			name: "one failure",
			checks: []CheckStatus{
				{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
				{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE"},
			},
			wantResult: CheckResultFailure,
		},
		{
			name: "still pending",
			checks: []CheckStatus{
				{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
				{Name: "test", Status: "IN_PROGRESS", Conclusion: ""},
			},
			wantResult: CheckResultPending,
		},
		{
			name: "skipped counts as success",
			checks: []CheckStatus{
				{Name: "build", Status: "COMPLETED", Conclusion: "SKIPPED"},
			},
			wantResult: CheckResultSuccess,
		},
		{
			name: "neutral counts as success",
			checks: []CheckStatus{
				{Name: "lint", Status: "COMPLETED", Conclusion: "NEUTRAL"},
			},
			wantResult: CheckResultSuccess,
		},
		{
			name: "cancelled counts as failure",
			checks: []CheckStatus{
				{Name: "build", Status: "COMPLETED", Conclusion: "CANCELLED"},
			},
			wantResult: CheckResultFailure,
		},
		{
			name: "queued status is pending",
			checks: []CheckStatus{
				{Name: "build", Status: "QUEUED", Conclusion: ""},
			},
			wantResult: CheckResultPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := analyzeChecks(tt.checks)
			if result != tt.wantResult {
				t.Errorf("analyzeChecks() result = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestDefaultGhPRView(t *testing.T) {
	// Save original function
	origExecCommand := execCommand
	defer func() {
		execCommand = origExecCommand
	}()

	t.Run("successful PR view", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			// Return a command that outputs valid JSON
			return exec.Command("echo", `{"number":123,"state":"OPEN","statusCheckRollup":[{"name":"build","status":"COMPLETED","conclusion":"SUCCESS"}]}`)
		}

		status, err := defaultGhPRView()
		if err != nil {
			t.Errorf("defaultGhPRView() unexpected error: %v", err)
		}
		if status == nil {
			t.Fatal("defaultGhPRView() returned nil status")
		}
		if status.Number != 123 {
			t.Errorf("status.Number = %d, want 123", status.Number)
		}
		if len(status.StatusCheckRollup) != 1 {
			t.Errorf("status.StatusCheckRollup length = %d, want 1", len(status.StatusCheckRollup))
		}
	})

	t.Run("no pull request found (plural)", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			// Create a command that fails with "no pull requests" error
			cmd := exec.Command("sh", "-c", "echo 'no pull requests found for branch' >&2; exit 1")
			return cmd
		}

		_, err := defaultGhPRView()
		if err == nil || !strings.Contains(err.Error(), "no PR found") {
			t.Errorf("defaultGhPRView() error = %v, want error about no PR found", err)
		}
	})

	t.Run("no pull request found (singular)", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			cmd := exec.Command("sh", "-c", "echo 'no pull request found for branch' >&2; exit 1")
			return cmd
		}

		_, err := defaultGhPRView()
		if err == nil || !strings.Contains(err.Error(), "no PR found") {
			t.Errorf("defaultGhPRView() error = %v, want error about no PR found", err)
		}
	})

	t.Run("command execution error", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			// Command that fails with a different error
			cmd := exec.Command("sh", "-c", "echo 'some other error' >&2; exit 1")
			return cmd
		}

		_, err := defaultGhPRView()
		if err == nil || !strings.Contains(err.Error(), "failed to get PR status") {
			t.Errorf("defaultGhPRView() error = %v, want error about failed to get PR status", err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "not valid json")
		}

		_, err := defaultGhPRView()
		if err == nil || !strings.Contains(err.Error(), "failed to parse PR status") {
			t.Errorf("defaultGhPRView() error = %v, want error about failed to parse PR status", err)
		}
	})
}

func TestPrintCheckDetails(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	t.Run("success marker", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "build", Status: "COMPLETED", Conclusion: "SUCCESS"},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "[+]") {
			t.Errorf("printCheckDetails() output missing [+] marker: %s", output)
		}
		if !strings.Contains(output, "build") {
			t.Errorf("printCheckDetails() output missing check name: %s", output)
		}
	})

	t.Run("failure marker", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE"},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "[x]") {
			t.Errorf("printCheckDetails() output missing [x] marker: %s", output)
		}
	})

	t.Run("in progress status", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "build", Status: "IN_PROGRESS", Conclusion: ""},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "IN_PROGRESS") {
			t.Errorf("printCheckDetails() output missing IN_PROGRESS status: %s", output)
		}
		if !strings.Contains(output, "[ ]") {
			t.Errorf("printCheckDetails() output missing [ ] marker: %s", output)
		}
	})

	t.Run("queued status", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "deploy", Status: "QUEUED", Conclusion: ""},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "QUEUED") {
			t.Errorf("printCheckDetails() output missing QUEUED status: %s", output)
		}
	})

	t.Run("neutral conclusion", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "optional", Status: "COMPLETED", Conclusion: "NEUTRAL"},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// NEUTRAL is neither SUCCESS nor FAILURE, so marker should be space
		if !strings.Contains(output, "[ ]") {
			t.Errorf("printCheckDetails() output missing [ ] marker for NEUTRAL: %s", output)
		}
	})
}
