package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestUsageText(t *testing.T) {
	text := usageText()
	if len(text) == 0 {
		t.Error("usageText() returned empty string")
	}
	// Verify key content is present
	if !bytes.Contains([]byte(text), []byte("wt [options] <name>")) {
		t.Error("usageText() missing usage line")
	}
	if !bytes.Contains([]byte(text), []byte("--hook")) {
		t.Error("usageText() missing --hook option")
	}
}

func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	if buf.Len() == 0 {
		t.Error("printUsage() wrote nothing")
	}
	if buf.String() != usageText() {
		t.Error("printUsage() output doesn't match usageText()")
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    string
		wantName   string
		wantHook   string
		wantErr    error
		wantErrMsg string
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: errShowHelpFail,
		},
		{
			name:    "help flag -h",
			args:    []string{"-h"},
			wantErr: errShowHelp,
		},
		{
			name:    "help flag --help",
			args:    []string{"--help"},
			wantErr: errShowHelp,
		},
		{
			name:    "help flag in middle",
			args:    []string{"create", "--help", "foo"},
			wantErr: errShowHelp,
		},
		{
			name:     "simple name (default create)",
			args:     []string{"my-feature"},
			wantCmd:  "create",
			wantName: "my-feature",
			wantHook: defaultHook,
		},
		{
			name:     "explicit create",
			args:     []string{"create", "my-feature"},
			wantCmd:  "create",
			wantName: "my-feature",
			wantHook: defaultHook,
		},
		{
			name:     "remove command",
			args:     []string{"remove", "my-feature"},
			wantCmd:  "remove",
			wantName: "my-feature",
			wantHook: defaultHook,
		},
		{
			name:     "create with hook",
			args:     []string{"--hook", "setup.sh", "my-feature"},
			wantCmd:  "create",
			wantName: "my-feature",
			wantHook: "setup.sh",
		},
		{
			name:     "create explicit with hook",
			args:     []string{"create", "--hook", "setup.sh", "my-feature"},
			wantCmd:  "create",
			wantName: "my-feature",
			wantHook: "setup.sh",
		},
		{
			name:       "hook without path",
			args:       []string{"--hook"},
			wantErrMsg: "--hook requires a path argument",
		},
		{
			name:       "unknown flag",
			args:       []string{"--unknown", "foo"},
			wantErrMsg: "unknown flag --unknown",
		},
		{
			name:       "unknown short flag",
			args:       []string{"-x", "foo"},
			wantErrMsg: "unknown flag -x",
		},
		{
			name:       "create without name",
			args:       []string{"create"},
			wantErrMsg: "branch name required",
		},
		{
			name:       "remove without name",
			args:       []string{"remove"},
			wantErrMsg: "branch name required",
		},
		{
			name:       "extra argument",
			args:       []string{"create", "foo", "bar"},
			wantErrMsg: "unexpected argument: bar",
		},
		{
			name:       "hook at end without value",
			args:       []string{"create", "--hook"},
			wantErrMsg: "--hook requires a path argument",
		},
		{
			name:     "gha command no args",
			args:     []string{"gha"},
			wantCmd:  "gha",
			wantName: "",
			wantHook: defaultHook,
		},
		{
			name:       "gha command with extra arg",
			args:       []string{"gha", "extra"},
			wantErrMsg: "unexpected argument: extra",
		},
		{
			name:     "completion command bash",
			args:     []string{"completion", "bash"},
			wantCmd:  "completion",
			wantName: "bash",
			wantHook: defaultHook,
		},
		{
			name:     "completion command zsh",
			args:     []string{"completion", "zsh"},
			wantCmd:  "completion",
			wantName: "zsh",
			wantHook: defaultHook,
		},
		{
			name:     "completion command fish",
			args:     []string{"completion", "fish"},
			wantCmd:  "completion",
			wantName: "fish",
			wantHook: defaultHook,
		},
		{
			name:       "completion without shell",
			args:       []string{"completion"},
			wantErrMsg: "shell name required (bash, zsh, fish)",
		},
		{
			name:       "completion with extra arg",
			args:       []string{"completion", "bash", "extra"},
			wantErrMsg: "unexpected argument: extra",
		},
		{
			name:     "__complete remove",
			args:     []string{"__complete", "remove"},
			wantCmd:  "__complete",
			wantName: "remove",
			wantHook: defaultHook,
		},
		{
			name:       "__complete without subcommand",
			args:       []string{"__complete"},
			wantErrMsg: "subcommand required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, name, hook, err := parseArgs(tt.args)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("parseArgs() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if tt.wantErrMsg != "" {
				if err == nil {
					t.Errorf("parseArgs() error = nil, want error containing %q", tt.wantErrMsg)
					return
				}
				if err.Error() != tt.wantErrMsg {
					t.Errorf("parseArgs() error = %q, want %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseArgs() unexpected error: %v", err)
				return
			}

			if cmd != tt.wantCmd {
				t.Errorf("parseArgs() cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if name != tt.wantName {
				t.Errorf("parseArgs() name = %q, want %q", name, tt.wantName)
			}
			if hook != tt.wantHook {
				t.Errorf("parseArgs() hook = %q, want %q", hook, tt.wantHook)
			}
		})
	}
}

func TestRun(t *testing.T) {
	// Save original functions and restore after test
	origGitRoot := gitRootFn
	origGitCmd := gitCmdFn
	origGhPRView := ghPRViewFn
	defer func() {
		gitRootFn = origGitRoot
		gitCmdFn = origGitCmd
		ghPRViewFn = origGhPRView
	}()

	t.Run("parse error propagates", func(t *testing.T) {
		err := run([]string{})
		if !errors.Is(err, errShowHelpFail) {
			t.Errorf("run() error = %v, want %v", err, errShowHelpFail)
		}
	})

	t.Run("help error propagates", func(t *testing.T) {
		err := run([]string{"--help"})
		if !errors.Is(err, errShowHelp) {
			t.Errorf("run() error = %v, want %v", err, errShowHelp)
		}
	})

	t.Run("create command calls create", func(t *testing.T) {
		gitRootFn = func() (string, error) {
			return "", errors.New("mock: not in git repo")
		}

		err := run([]string{"my-feature"})
		if err == nil || err.Error() != "mock: not in git repo" {
			t.Errorf("run() error = %v, want 'mock: not in git repo'", err)
		}
	})

	t.Run("remove command calls remove", func(t *testing.T) {
		gitRootFn = func() (string, error) {
			return "", errors.New("mock: not in git repo for remove")
		}

		err := run([]string{"remove", "my-feature"})
		if err == nil || err.Error() != "mock: not in git repo for remove" {
			t.Errorf("run() error = %v, want 'mock: not in git repo for remove'", err)
		}
	})

	t.Run("gha command calls gha", func(t *testing.T) {
		ghPRViewFn = func() (*PRStatus, error) {
			return nil, errors.New("mock: no PR found for current branch")
		}

		err := run([]string{"gha"})
		if err == nil || err.Error() != "mock: no PR found for current branch" {
			t.Errorf("run() error = %v, want 'mock: no PR found for current branch'", err)
		}
	})

	t.Run("completion command calls completion", func(t *testing.T) {
		err := run([]string{"completion", "bash"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})

	t.Run("completion command with invalid shell", func(t *testing.T) {
		err := run([]string{"completion", "invalid"})
		if err == nil || !strings.Contains(err.Error(), "unsupported shell") {
			t.Errorf("run() error = %v, want error containing 'unsupported shell'", err)
		}
	})

	t.Run("__complete remove calls completeWorktrees", func(t *testing.T) {
		origListWorktrees := listWorktreesFn
		defer func() { listWorktreesFn = origListWorktrees }()

		listWorktreesFn = func() ([]string, error) {
			return []string{"test-worktree"}, nil
		}

		err := run([]string{"__complete", "remove"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})

	t.Run("__complete with other subcommand", func(t *testing.T) {
		err := run([]string{"__complete", "create"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})
}

// TestMainFunc tests the main() function by mocking exitFn and os.Args
func TestMainFunc(t *testing.T) {
	// Save and restore original values
	origArgs := os.Args
	origExit := exitFn
	origGitRoot := gitRootFn
	defer func() {
		os.Args = origArgs
		exitFn = origExit
		gitRootFn = origGitRoot
	}()

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "no args",
			args:     []string{"wt"},
			wantExit: 1,
		},
		{
			name:     "help flag",
			args:     []string{"wt", "--help"},
			wantExit: 0,
		},
		{
			name:     "help flag -h",
			args:     []string{"wt", "-h"},
			wantExit: 0,
		},
		{
			name:     "error from run",
			args:     []string{"wt", "test-branch"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var exitCode int
			exitFn = func(code int) {
				exitCode = code
			}

			// Mock gitRoot to return an error (not in git repo)
			gitRootFn = func() (string, error) {
				return "", errors.New("not in a git repository")
			}

			os.Args = tt.args
			main()

			if exitCode != tt.wantExit {
				t.Errorf("main() exit code = %d, want %d", exitCode, tt.wantExit)
			}
		})
	}
}

// TestMainSuccess tests main() when run() succeeds
func TestMainSuccess(t *testing.T) {
	origArgs := os.Args
	origExit := exitFn
	origGitRoot := gitRootFn
	origGitCmd := gitCmdFn
	defer func() {
		os.Args = origArgs
		exitFn = origExit
		gitRootFn = origGitRoot
		gitCmdFn = origGitCmd
	}()

	// Create temp dir with .worktrees
	tmpDir := t.TempDir()
	os.MkdirAll(tmpDir+"/.worktrees", 0755)

	exitCalled := false
	exitFn = func(code int) {
		exitCalled = true
	}

	gitRootFn = func() (string, error) {
		return tmpDir, nil
	}
	gitCmdFn = func(dir string, args ...string) error {
		return nil
	}

	os.Args = []string{"wt", "test-branch"}
	main()

	if exitCalled {
		t.Error("main() should not call exit on success")
	}
}
