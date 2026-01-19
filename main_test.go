package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUsageText(t *testing.T) {
	text := usageText()
	if len(text) == 0 {
		t.Error("usageText() returned empty string")
	}
	// Verify key content is present
	if !bytes.Contains([]byte(text), []byte("wt <command>")) {
		t.Error("usageText() missing usage line")
	}
	if !bytes.Contains([]byte(text), []byte("--hook")) {
		t.Error("usageText() missing --hook option")
	}
	// Verify auto-detect is documented
	if !bytes.Contains([]byte(text), []byte("auto-detects")) {
		t.Error("usageText() missing auto-detect documentation")
	}
	// Verify jump command is documented
	if !bytes.Contains([]byte(text), []byte("jump")) {
		t.Error("usageText() missing 'jump' command")
	}
	// Verify jump navigation example
	if !bytes.Contains([]byte(text), []byte("Jump to a worktree")) {
		t.Error("usageText() missing 'Jump to a worktree' text")
	}
	// Verify version command is documented
	if !bytes.Contains([]byte(text), []byte("version")) {
		t.Error("usageText() missing 'version' command")
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

func TestIsValidCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"create", "create", true},
		{"remove", "remove", true},
		{"jump", "jump", true},
		{"list", "list", true},
		{"completion", "completion", true},
		{"version", "version", true},
		{"__complete", "__complete", true},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidCommand(tt.cmd); got != tt.want {
				t.Errorf("isValidCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestIsHelpRequested(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", []string{}, false},
		{"-h", []string{"-h"}, true},
		{"--help", []string{"--help"}, true},
		{"help in middle", []string{"create", "--help", "foo"}, true},
		{"no help", []string{"create", "foo"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHelpRequested(tt.args); got != tt.want {
				t.Errorf("isHelpRequested(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    string
		wantIdx    int
		wantErrMsg string
	}{
		{"empty", []string{}, "", 0, ""},
		{"create", []string{"create", "foo"}, "create", 1, ""},
		{"remove", []string{"remove", "foo"}, "remove", 1, ""},
		{"jump", []string{"jump"}, "jump", 1, ""},
		{"unknown command", []string{"my-branch"}, "", 0, "unknown command: my-branch"},
		{"unknown with args", []string{"foo", "bar"}, "", 0, "unknown command: foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, idx, err := parseCommand(tt.args)

			if tt.wantErrMsg != "" {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("parseCommand() error = %v, want %q", err, tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseCommand() unexpected error: %v", err)
				return
			}

			if cmd != tt.wantCmd {
				t.Errorf("parseCommand() cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if idx != tt.wantIdx {
				t.Errorf("parseCommand() idx = %d, want %d", idx, tt.wantIdx)
			}
		})
	}
}

func TestParseHookFlag(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		idx        int
		wantIdx    int
		wantHook   string
		wantErrMsg string
	}{
		{"no hook", []string{"foo"}, 0, 0, DefaultHook, ""},
		{"with hook", []string{"--hook", "setup.sh", "foo"}, 0, 2, "setup.sh", ""},
		{"hook missing value", []string{"--hook"}, 0, 0, "", "--hook requires a path argument"},
		{"unknown flag", []string{"-x", "foo"}, 0, 0, "", "unknown flag -x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, hook, err := parseHookFlag(tt.args, tt.idx, DefaultHook)

			if tt.wantErrMsg != "" {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("parseHookFlag() error = %v, want %q", err, tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseHookFlag() unexpected error: %v", err)
				return
			}

			if idx != tt.wantIdx {
				t.Errorf("parseHookFlag() idx = %d, want %d", idx, tt.wantIdx)
			}
			if hook != tt.wantHook {
				t.Errorf("parseHookFlag() hook = %q, want %q", hook, tt.wantHook)
			}
		})
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
			name:    "no args shows help",
			args:    []string{},
			wantErr: errShowHelp,
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
			name:       "unknown command (no implicit create)",
			args:       []string{"my-feature"},
			wantErrMsg: "unknown command: my-feature",
		},
		{
			name:     "explicit create",
			args:     []string{"create", "my-feature"},
			wantCmd:  "create",
			wantName: "my-feature",
			wantHook: DefaultHook,
		},
		{
			name:     "remove command",
			args:     []string{"remove", "my-feature"},
			wantCmd:  "remove",
			wantName: "my-feature",
			wantHook: DefaultHook,
		},
		{
			name:       "hook without command is unknown flag",
			args:       []string{"--hook", "setup.sh", "my-feature"},
			wantErrMsg: "unknown command: --hook",
		},
		{
			name:     "create explicit with hook",
			args:     []string{"create", "--hook", "setup.sh", "my-feature"},
			wantCmd:  "create",
			wantName: "my-feature",
			wantHook: "setup.sh",
		},
		{
			name:       "hook without path requires command",
			args:       []string{"--hook"},
			wantErrMsg: "unknown command: --hook",
		},
		{
			name:       "create hook without path",
			args:       []string{"create", "--hook"},
			wantErrMsg: "--hook requires a path argument",
		},
		{
			name:       "unknown flag requires command",
			args:       []string{"--unknown", "foo"},
			wantErrMsg: "unknown command: --unknown",
		},
		{
			name:       "create unknown flag",
			args:       []string{"create", "--unknown", "foo"},
			wantErrMsg: "unknown flag --unknown",
		},
		{
			name:       "unknown short flag requires command",
			args:       []string{"-x", "foo"},
			wantErrMsg: "unknown command: -x",
		},
		{
			name:       "create unknown short flag",
			args:       []string{"create", "-x", "foo"},
			wantErrMsg: "unknown flag -x",
		},
		{
			name:       "create without name",
			args:       []string{"create"},
			wantErrMsg: "branch name required",
		},
		{
			name:     "remove without name (auto-detect)",
			args:     []string{"remove"},
			wantCmd:  "remove",
			wantName: "",
			wantHook: DefaultHook,
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
			name:     "jump command no args",
			args:     []string{"jump"},
			wantCmd:  "jump",
			wantName: "",
			wantHook: DefaultHook,
		},
		{
			name:     "jump command with name",
			args:     []string{"jump", "my-feature"},
			wantCmd:  "jump",
			wantName: "my-feature",
			wantHook: DefaultHook,
		},
		{
			name:       "jump command with extra arg",
			args:       []string{"jump", "my-feature", "extra"},
			wantErrMsg: "unexpected argument: extra",
		},
		{
			name:     "list command no args",
			args:     []string{"list"},
			wantCmd:  "list",
			wantName: "",
			wantHook: DefaultHook,
		},
		{
			name:       "list command with extra arg",
			args:       []string{"list", "extra"},
			wantErrMsg: "unexpected argument: extra",
		},
		{
			name:     "completion command bash",
			args:     []string{"completion", "bash"},
			wantCmd:  "completion",
			wantName: "bash",
			wantHook: DefaultHook,
		},
		{
			name:     "completion command zsh",
			args:     []string{"completion", "zsh"},
			wantCmd:  "completion",
			wantName: "zsh",
			wantHook: DefaultHook,
		},
		{
			name:     "completion command fish",
			args:     []string{"completion", "fish"},
			wantCmd:  "completion",
			wantName: "fish",
			wantHook: DefaultHook,
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
			wantHook: DefaultHook,
		},
		{
			name:       "__complete without subcommand",
			args:       []string{"__complete"},
			wantErrMsg: "subcommand required",
		},
		{
			name:     "version command",
			args:     []string{"version"},
			wantCmd:  "version",
			wantName: "",
			wantHook: DefaultHook,
		},
		{
			name:       "version command with extra arg",
			args:       []string{"version", "extra"},
			wantErrMsg: "unexpected argument: extra",
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
	origGitRoot := gitMainRootFn
	origGitCmd := gitCmdFn
	defer func() {
		gitMainRootFn = origGitRoot
		gitCmdFn = origGitCmd
	}()

	t.Run("no args shows help", func(t *testing.T) {
		err := run([]string{})
		if !errors.Is(err, errShowHelp) {
			t.Errorf("run() error = %v, want %v", err, errShowHelp)
		}
	})

	t.Run("jump command runs jump", func(t *testing.T) {
		tmpDir := t.TempDir()
		origGetwd := getwdFn
		defer func() { getwdFn = origGetwd }()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		getwdFn = func() (string, error) {
			return "/some/other/dir", nil
		}

		err := run([]string{"jump"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})

	t.Run("jump command with name", func(t *testing.T) {
		tmpDir := t.TempDir()
		worktreesDir := filepath.Join(tmpDir, WorktreesDir)
		os.MkdirAll(filepath.Join(worktreesDir, "my-feature"), 0755)

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}

		err := run([]string{"jump", "my-feature"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})

	t.Run("help error propagates", func(t *testing.T) {
		err := run([]string{"--help"})
		if !errors.Is(err, errShowHelp) {
			t.Errorf("run() error = %v, want %v", err, errShowHelp)
		}
	})

	t.Run("create command calls create", func(t *testing.T) {
		gitMainRootFn = func() (string, error) {
			return "", errors.New("mock: not in git repo")
		}

		err := run([]string{"create", "my-feature"})
		if err == nil || err.Error() != "mock: not in git repo" {
			t.Errorf("run() error = %v, want 'mock: not in git repo'", err)
		}
	})

	t.Run("unknown command returns error", func(t *testing.T) {
		err := run([]string{"my-feature"})
		if err == nil || err.Error() != "unknown command: my-feature" {
			t.Errorf("run() error = %v, want 'unknown command: my-feature'", err)
		}
	})

	t.Run("remove command calls remove", func(t *testing.T) {
		gitMainRootFn = func() (string, error) {
			return "", errors.New("mock: not in git repo for remove")
		}

		err := run([]string{"remove", "my-feature"})
		if err == nil || err.Error() != "mock: not in git repo for remove" {
			t.Errorf("run() error = %v, want 'mock: not in git repo for remove'", err)
		}
	})

	t.Run("remove without name detects current worktree", func(t *testing.T) {
		origGetwd := getwdFn
		defer func() { getwdFn = origGetwd }()

		tmpDir := t.TempDir()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		gitCmdFn = func(dir string, args ...string) error {
			return nil
		}
		// Simulate being inside a worktree
		getwdFn = func() (string, error) {
			return tmpDir + "/" + WorktreesDir + "/auto-detected", nil
		}

		err := run([]string{"remove"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})

	t.Run("remove without name not inside worktree", func(t *testing.T) {
		origGetwd := getwdFn
		defer func() { getwdFn = origGetwd }()

		tmpDir := t.TempDir()

		gitMainRootFn = func() (string, error) {
			return tmpDir, nil
		}
		// Simulate being outside worktree
		getwdFn = func() (string, error) {
			return "/some/other/dir", nil
		}

		err := run([]string{"remove"})
		if err == nil || err.Error() != "not inside a worktree (specify branch name)" {
			t.Errorf("run() error = %v, want 'not inside a worktree (specify branch name)'", err)
		}
	})

	t.Run("remove without name git root error", func(t *testing.T) {
		gitMainRootFn = func() (string, error) {
			return "", errors.New("mock: not in git repo")
		}

		err := run([]string{"remove"})
		if err == nil || err.Error() != "mock: not in git repo" {
			t.Errorf("run() error = %v, want 'mock: not in git repo'", err)
		}
	})

	t.Run("list command calls list", func(t *testing.T) {
		origListWorktrees := listWorktreesFn
		defer func() { listWorktreesFn = origListWorktrees }()

		listWorktreesFn = func() ([]string, error) {
			return []string{"feature-a", "feature-b"}, nil
		}

		err := run([]string{"list"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})

	t.Run("list command with error", func(t *testing.T) {
		origListWorktrees := listWorktreesFn
		defer func() { listWorktreesFn = origListWorktrees }()

		listWorktreesFn = func() ([]string, error) {
			return nil, errors.New("mock: not in git repo")
		}

		err := run([]string{"list"})
		if err == nil || err.Error() != "mock: not in git repo" {
			t.Errorf("run() error = %v, want 'mock: not in git repo'", err)
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

	t.Run("__complete jump calls completeWorktrees", func(t *testing.T) {
		origListWorktrees := listWorktreesFn
		defer func() { listWorktreesFn = origListWorktrees }()

		listWorktreesFn = func() ([]string, error) {
			return []string{"test-worktree"}, nil
		}

		err := run([]string{"__complete", "jump"})
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

	t.Run("version command", func(t *testing.T) {
		err := run([]string{"version"})
		if err != nil {
			t.Errorf("run() unexpected error: %v", err)
		}
	})
}

func TestVersionFunc(t *testing.T) {
	var buf bytes.Buffer
	err := version(&buf)
	if err != nil {
		t.Errorf("version() returned error: %v", err)
	}
	output := buf.String()
	if output != Version+"\n" {
		t.Errorf("version() = %q, want %q", output, Version+"\n")
	}
}

// TestMainFunc tests the main() function by mocking exitFn and os.Args
func TestMainFunc(t *testing.T) {
	// Save and restore original values
	origArgs := os.Args
	origExit := exitFn
	origGitRoot := gitMainRootFn
	defer func() {
		os.Args = origArgs
		exitFn = origExit
		gitMainRootFn = origGitRoot
	}()

	tests := []struct {
		name     string
		args     []string
		wantExit int
		mockRoot bool
	}{
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
			name:     "error from run (unknown command)",
			args:     []string{"wt", "test-branch"},
			wantExit: 1,
		},
		{
			name:     "error from run (create fails)",
			args:     []string{"wt", "create", "test-branch"},
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
			gitMainRootFn = func() (string, error) {
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
	origGitRoot := gitMainRootFn
	origGitCmd := gitCmdFn
	origStdout := os.Stdout
	defer func() {
		os.Args = origArgs
		exitFn = origExit
		gitMainRootFn = origGitRoot
		gitCmdFn = origGitCmd
		os.Stdout = origStdout
	}()

	// Create temp dir with .worktrees
	tmpDir := t.TempDir()
	os.MkdirAll(tmpDir+"/"+WorktreesDir, 0755)

	exitCalled := false
	exitFn = func(code int) {
		exitCalled = true
	}

	gitMainRootFn = func() (string, error) {
		return tmpDir, nil
	}
	gitCmdFn = func(dir string, args ...string) error {
		return nil
	}

	// Capture stdout to prevent output during test
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"wt", "create", "test-branch"}
	main()

	w.Close()
	r.Close()

	if exitCalled {
		t.Error("main() should not call exit on success")
	}
}
