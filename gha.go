package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Function variables for testing (following git.go pattern)
var (
	ghPRViewFn = defaultGhPRView
	sleepFn    = time.Sleep
	execCommand = exec.Command
)

// Default polling configuration
const defaultPollInterval = 30 * time.Second

var ghaTimeout = 60 * time.Minute

// PRStatus represents the GitHub PR status response
type PRStatus struct {
	Number            int           `json:"number"`
	State             string        `json:"state"`
	StatusCheckRollup []CheckStatus `json:"statusCheckRollup"`
}

// CheckStatus represents a single CI check
type CheckStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"`     // QUEUED, IN_PROGRESS, COMPLETED
	Conclusion string `json:"conclusion"` // SUCCESS, FAILURE, CANCELLED, etc.
}

// CheckResult represents the final outcome
type CheckResult int

const (
	CheckResultPending CheckResult = iota
	CheckResultSuccess
	CheckResultFailure
)

func gha() error {
	startTime := time.Now()

	fmt.Println("Monitoring GitHub Actions for current branch's PR...")

	for {
		// Check timeout
		if time.Since(startTime) > ghaTimeout {
			return fmt.Errorf("timeout: checks did not complete within %v", ghaTimeout)
		}

		// Get PR status
		status, err := ghPRViewFn()
		if err != nil {
			return err
		}

		// Analyze check results
		result, summary := analyzeChecks(status.StatusCheckRollup)

		// Print status update
		fmt.Printf("\r%s", summary)

		switch result {
		case CheckResultSuccess:
			fmt.Println("\nAll checks passed!")
			return nil
		case CheckResultFailure:
			fmt.Println("\nSome checks failed!")
			printCheckDetails(status.StatusCheckRollup)
			return fmt.Errorf("checks failed")
		case CheckResultPending:
			// Continue polling
			sleepFn(defaultPollInterval)
		}
	}
}

func defaultGhPRView() (*PRStatus, error) {
	cmd := execCommand("gh", "pr", "view", "--json", "number,state,statusCheckRollup")
	out, err := cmd.Output()
	if err != nil {
		// Check if it's because there's no PR
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no pull request") || strings.Contains(stderr, "no pull requests") {
				return nil, fmt.Errorf("no PR found for current branch")
			}
		}
		return nil, fmt.Errorf("failed to get PR status: %w", err)
	}

	var status PRStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("failed to parse PR status: %w", err)
	}

	return &status, nil
}

func analyzeChecks(checks []CheckStatus) (CheckResult, string) {
	if len(checks) == 0 {
		return CheckResultPending, "No checks found yet..."
	}

	var pending, passed, failed int

	for _, check := range checks {
		switch check.Status {
		case "COMPLETED":
			switch check.Conclusion {
			case "SUCCESS", "NEUTRAL", "SKIPPED":
				passed++
			default:
				failed++
			}
		default:
			pending++
		}
	}

	total := len(checks)
	summary := fmt.Sprintf("Checks: %d/%d completed (%d passed, %d failed, %d pending)",
		passed+failed, total, passed, failed, pending)

	if pending > 0 {
		return CheckResultPending, summary
	}
	if failed > 0 {
		return CheckResultFailure, summary
	}
	return CheckResultSuccess, summary
}

func printCheckDetails(checks []CheckStatus) {
	fmt.Println("\nCheck details:")
	for _, check := range checks {
		var status string
		if check.Status == "COMPLETED" {
			status = check.Conclusion
		} else {
			status = check.Status
		}

		// Use simple markers for pass/fail
		marker := " "
		if check.Conclusion == "SUCCESS" {
			marker = "+"
		} else if check.Conclusion == "FAILURE" {
			marker = "x"
		}

		fmt.Printf("  [%s] %s: %s\n", marker, check.Name, status)
	}
}
