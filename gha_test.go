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
					{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
					{Name: "test", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
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
					{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
					{Name: "test", Status: CheckStatusCompleted, Conclusion: CheckConclusionFailure},
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
						{Name: "build", Status: CheckStatusInProgress, Conclusion: ""},
					},
				}, nil
			}
			return &PRStatus{
				Number: 123,
				State:  "OPEN",
				StatusCheckRollup: []CheckStatus{
					{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
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
		origTimeout := GHATimeout
		defer func() { GHATimeout = origTimeout }()

		// Set a very short timeout
		GHATimeout = 1 * time.Nanosecond

		ghPRViewFn = func() (*PRStatus, error) {
			// Simulate time passing
			time.Sleep(10 * time.Millisecond)
			return &PRStatus{
				Number: 123,
				State:  "OPEN",
				StatusCheckRollup: []CheckStatus{
					{Name: "build", Status: CheckStatusInProgress, Conclusion: ""},
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
				{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
				{Name: "test", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
			},
			wantResult: CheckResultSuccess,
		},
		{
			name: "one failure",
			checks: []CheckStatus{
				{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
				{Name: "test", Status: CheckStatusCompleted, Conclusion: CheckConclusionFailure},
			},
			wantResult: CheckResultFailure,
		},
		{
			name: "still pending",
			checks: []CheckStatus{
				{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
				{Name: "test", Status: CheckStatusInProgress, Conclusion: ""},
			},
			wantResult: CheckResultPending,
		},
		{
			name: "skipped counts as success",
			checks: []CheckStatus{
				{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSkipped},
			},
			wantResult: CheckResultSuccess,
		},
		{
			name: "neutral counts as success",
			checks: []CheckStatus{
				{Name: "lint", Status: CheckStatusCompleted, Conclusion: CheckConclusionNeutral},
			},
			wantResult: CheckResultSuccess,
		},
		{
			name: "cancelled counts as failure",
			checks: []CheckStatus{
				{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionCancelled},
			},
			wantResult: CheckResultFailure,
		},
		{
			name: "queued status is pending",
			checks: []CheckStatus{
				{Name: "build", Status: CheckStatusQueued, Conclusion: ""},
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

func TestCheckStats(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		stats := CheckStats{Passed: 2, Failed: 1, Pending: 1, Total: 4}
		expected := "Checks: 3/4 completed (2 passed, 1 failed, 1 pending)"
		if stats.String() != expected {
			t.Errorf("CheckStats.String() = %q, want %q", stats.String(), expected)
		}
	})

	t.Run("Result pending", func(t *testing.T) {
		stats := CheckStats{Passed: 2, Failed: 0, Pending: 1, Total: 3}
		if stats.Result() != CheckResultPending {
			t.Errorf("CheckStats.Result() = %v, want CheckResultPending", stats.Result())
		}
	})

	t.Run("Result failure", func(t *testing.T) {
		stats := CheckStats{Passed: 2, Failed: 1, Pending: 0, Total: 3}
		if stats.Result() != CheckResultFailure {
			t.Errorf("CheckStats.Result() = %v, want CheckResultFailure", stats.Result())
		}
	})

	t.Run("Result success", func(t *testing.T) {
		stats := CheckStats{Passed: 3, Failed: 0, Pending: 0, Total: 3}
		if stats.Result() != CheckResultSuccess {
			t.Errorf("CheckStats.Result() = %v, want CheckResultSuccess", stats.Result())
		}
	})
}

func TestIsCheckComplete(t *testing.T) {
	tests := []struct {
		name  string
		check CheckStatus
		want  bool
	}{
		{"completed", CheckStatus{Status: CheckStatusCompleted}, true},
		{"in progress", CheckStatus{Status: CheckStatusInProgress}, false},
		{"queued", CheckStatus{Status: CheckStatusQueued}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCheckComplete(tt.check); got != tt.want {
				t.Errorf("isCheckComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCheckSuccess(t *testing.T) {
	tests := []struct {
		name  string
		check CheckStatus
		want  bool
	}{
		{"success", CheckStatus{Conclusion: CheckConclusionSuccess}, true},
		{"neutral", CheckStatus{Conclusion: CheckConclusionNeutral}, true},
		{"skipped", CheckStatus{Conclusion: CheckConclusionSkipped}, true},
		{"failure", CheckStatus{Conclusion: CheckConclusionFailure}, false},
		{"cancelled", CheckStatus{Conclusion: CheckConclusionCancelled}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCheckSuccess(tt.check); got != tt.want {
				t.Errorf("isCheckSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountCheckStatuses(t *testing.T) {
	checks := []CheckStatus{
		{Name: "a", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
		{Name: "b", Status: CheckStatusCompleted, Conclusion: CheckConclusionFailure},
		{Name: "c", Status: CheckStatusInProgress, Conclusion: ""},
	}

	stats := countCheckStatuses(checks)

	if stats.Total != 3 {
		t.Errorf("stats.Total = %d, want 3", stats.Total)
	}
	if stats.Passed != 1 {
		t.Errorf("stats.Passed = %d, want 1", stats.Passed)
	}
	if stats.Failed != 1 {
		t.Errorf("stats.Failed = %d, want 1", stats.Failed)
	}
	if stats.Pending != 1 {
		t.Errorf("stats.Pending = %d, want 1", stats.Pending)
	}
}

func TestGetCheckMarker(t *testing.T) {
	tests := []struct {
		name  string
		check CheckStatus
		want  string
	}{
		{"success", CheckStatus{Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess}, MarkerSuccess},
		{"failure", CheckStatus{Status: CheckStatusCompleted, Conclusion: CheckConclusionFailure}, MarkerFailure},
		{"neutral", CheckStatus{Status: CheckStatusCompleted, Conclusion: CheckConclusionNeutral}, MarkerPending},
		{"in progress", CheckStatus{Status: CheckStatusInProgress}, MarkerPending},
		{"queued", CheckStatus{Status: CheckStatusQueued}, MarkerPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCheckMarker(tt.check); got != tt.want {
				t.Errorf("getCheckMarker() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCheckStatusDisplay(t *testing.T) {
	tests := []struct {
		name  string
		check CheckStatus
		want  string
	}{
		{"completed shows conclusion", CheckStatus{Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess}, CheckConclusionSuccess},
		{"in progress shows status", CheckStatus{Status: CheckStatusInProgress, Conclusion: ""}, CheckStatusInProgress},
		{"queued shows status", CheckStatus{Status: CheckStatusQueued, Conclusion: ""}, CheckStatusQueued},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCheckStatusDisplay(tt.check); got != tt.want {
				t.Errorf("getCheckStatusDisplay() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNoPRError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   bool
	}{
		{"singular", "no pull request found for branch", true},
		{"plural", "no pull requests found for branch", true},
		{"other error", "some other error message", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNoPRError(tt.stderr); got != tt.want {
				t.Errorf("isNoPRError(%q) = %v, want %v", tt.stderr, got, tt.want)
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
			{Name: "build", Status: CheckStatusCompleted, Conclusion: CheckConclusionSuccess},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "["+MarkerSuccess+"]") {
			t.Errorf("printCheckDetails() output missing [%s] marker: %s", MarkerSuccess, output)
		}
		if !strings.Contains(output, "build") {
			t.Errorf("printCheckDetails() output missing check name: %s", output)
		}
	})

	t.Run("failure marker", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "test", Status: CheckStatusCompleted, Conclusion: CheckConclusionFailure},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "["+MarkerFailure+"]") {
			t.Errorf("printCheckDetails() output missing [%s] marker: %s", MarkerFailure, output)
		}
	})

	t.Run("in progress status", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "build", Status: CheckStatusInProgress, Conclusion: ""},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, CheckStatusInProgress) {
			t.Errorf("printCheckDetails() output missing %s status: %s", CheckStatusInProgress, output)
		}
		if !strings.Contains(output, "["+MarkerPending+"]") {
			t.Errorf("printCheckDetails() output missing [%s] marker: %s", MarkerPending, output)
		}
	})

	t.Run("queued status", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "deploy", Status: CheckStatusQueued, Conclusion: ""},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, CheckStatusQueued) {
			t.Errorf("printCheckDetails() output missing %s status: %s", CheckStatusQueued, output)
		}
	})

	t.Run("neutral conclusion", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdout = w

		checks := []CheckStatus{
			{Name: "optional", Status: CheckStatusCompleted, Conclusion: CheckConclusionNeutral},
		}
		printCheckDetails(checks)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// NEUTRAL is neither SUCCESS nor FAILURE, so marker should be space
		if !strings.Contains(output, "["+MarkerPending+"]") {
			t.Errorf("printCheckDetails() output missing [%s] marker for NEUTRAL: %s", MarkerPending, output)
		}
	})
}
