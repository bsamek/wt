package main

import (
	"errors"
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
