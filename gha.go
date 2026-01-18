package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"time"
)

// Function variables for testing (following git.go pattern)
var (
	ghPRViewFn  = defaultGhPRView
	sleepFn     = time.Sleep
	execCommand = exec.Command
)

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

// CheckStats holds the counts of check statuses
type CheckStats struct {
	Passed  int
	Failed  int
	Pending int
	Total   int
}

// String returns a human-readable summary of the check stats
func (cs CheckStats) String() string {
	completed := cs.Passed + cs.Failed
	return fmt.Sprintf("Checks: %d/%d completed (%d passed, %d failed, %d pending)",
		completed, cs.Total, cs.Passed, cs.Failed, cs.Pending)
}

// Result returns the overall check result based on stats
func (cs CheckStats) Result() CheckResult {
	if cs.Pending > 0 {
		return CheckResultPending
	}
	if cs.Failed > 0 {
		return CheckResultFailure
	}
	return CheckResultSuccess
}

func gha() error {
	startTime := time.Now()

	fmt.Println("Monitoring GitHub Actions for current branch's PR...")

	for {
		// Check timeout
		if time.Since(startTime) > GHATimeout {
			return fmt.Errorf("timeout: checks did not complete within %v", GHATimeout)
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
			sleepFn(DefaultPollInterval)
		}
	}
}

// noPRRegex matches the "no pull request(s)" error message from gh CLI
var noPRRegex = regexp.MustCompile(`no pull requests?`)

// isNoPRError checks if the error message indicates no PR was found
func isNoPRError(stderr string) bool {
	return noPRRegex.MatchString(stderr)
}

func defaultGhPRView() (*PRStatus, error) {
	cmd := execCommand("gh", "pr", "view", "--json", "number,state,statusCheckRollup")
	out, err := cmd.Output()
	if err != nil {
		// Check if it's because there's no PR
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if isNoPRError(stderr) {
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

// isCheckComplete returns true if the check has completed
func isCheckComplete(check CheckStatus) bool {
	return check.Status == CheckStatusCompleted
}

// isCheckSuccess returns true if the check completed successfully
func isCheckSuccess(check CheckStatus) bool {
	switch check.Conclusion {
	case CheckConclusionSuccess, CheckConclusionNeutral, CheckConclusionSkipped:
		return true
	default:
		return false
	}
}

// countCheckStatuses counts check statuses and returns stats
func countCheckStatuses(checks []CheckStatus) CheckStats {
	stats := CheckStats{Total: len(checks)}
	for _, check := range checks {
		if isCheckComplete(check) {
			if isCheckSuccess(check) {
				stats.Passed++
			} else {
				stats.Failed++
			}
		} else {
			stats.Pending++
		}
	}
	return stats
}

func analyzeChecks(checks []CheckStatus) (CheckResult, string) {
	if len(checks) == 0 {
		return CheckResultPending, "No checks found yet..."
	}

	stats := countCheckStatuses(checks)
	return stats.Result(), stats.String()
}

// getCheckMarker returns the display marker for a check
func getCheckMarker(check CheckStatus) string {
	if !isCheckComplete(check) {
		return MarkerPending
	}

	switch check.Conclusion {
	case CheckConclusionSuccess:
		return MarkerSuccess
	case CheckConclusionFailure:
		return MarkerFailure
	default:
		return MarkerPending
	}
}

// getCheckStatusDisplay returns the status string to display for a check
func getCheckStatusDisplay(check CheckStatus) string {
	if isCheckComplete(check) {
		return check.Conclusion
	}
	return check.Status
}

func printCheckDetails(checks []CheckStatus) {
	fmt.Println("\nCheck details:")
	for _, check := range checks {
		marker := getCheckMarker(check)
		status := getCheckStatusDisplay(check)
		fmt.Printf("  [%s] %s: %s\n", marker, check.Name, status)
	}
}
