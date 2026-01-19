package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Sentinel errors for testing
var errShowHelp = errors.New("show help")

// exitFn is the exit function, replaceable for testing
var exitFn = os.Exit

// validCommands lists all valid command names
var validCommands = []string{"create", "remove", "jump", "list", "gha", "completion", "__complete"}

func usageText() string {
	return `Usage: wt <command> [options] [args]

Commands:
  jump          Jump to a worktree or repository root
  create        Create a new worktree with branch
  remove        Remove a worktree and its branch (auto-detects if inside worktree)
  list          List all worktrees
  gha           Monitor GitHub Actions status for current branch's PR
  completion    Generate shell completion script (bash, zsh, fish)

Options:
  --hook <path>    Custom hook script to run after create (default: .worktree-hook)
  -h, --help       Show this help message

Examples:
  wt jump                    Navigate to repository root (from worktree)
  wt jump my-feature         Jump to 'my-feature' worktree
  wt create my-feature       Create worktree for 'my-feature' branch
  wt create --hook setup.sh feat    Create worktree, run setup.sh as hook
  wt remove my-feature       Remove worktree and branch
  wt remove                  Remove current worktree (when inside one)
  wt list                    List all worktrees
  wt gha                     Wait for GHA checks on current branch's PR
  wt completion bash         Generate bash completion script
`
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, usageText())
}

// isValidCommand checks if a string is a valid command name
func isValidCommand(s string) bool {
	for _, cmd := range validCommands {
		if s == cmd {
			return true
		}
	}
	return false
}

// isHelpRequested checks if any argument is a help flag
func isHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// parseCommand extracts the command from arguments, returning an error for unknown commands
func parseCommand(args []string) (cmd string, startIdx int, err error) {
	if len(args) == 0 {
		return "", 0, nil
	}
	if isValidCommand(args[0]) {
		return args[0], 1, nil
	}
	return "", 0, fmt.Errorf("unknown command: %s", args[0])
}

// parseHookFlag parses the --hook flag from arguments starting at idx
// Returns the new index, hook path, and any error
func parseHookFlag(args []string, idx int, defaultHook string) (int, string, error) {
	hookPath := defaultHook

	for idx < len(args) {
		if args[idx] == "--hook" {
			if idx+1 >= len(args) {
				return 0, "", fmt.Errorf("--hook requires a path argument")
			}
			hookPath = args[idx+1]
			idx += 2
		} else if len(args[idx]) > 0 && args[idx][0] == '-' {
			return 0, "", fmt.Errorf("unknown flag %s", args[idx])
		} else {
			break
		}
	}

	return idx, hookPath, nil
}

// parseArgs parses command line arguments and returns (command, name, hookPath, error)
func parseArgs(args []string) (cmd string, name string, hookPath string, err error) {
	if len(args) == 0 {
		return "", "", "", errShowHelp
	}

	if isHelpRequested(args) {
		return "", "", "", errShowHelp
	}

	cmd, idx, err := parseCommand(args)
	if err != nil {
		return "", "", "", err
	}

	// Parse hook flag
	idx, hookPath, err = parseHookFlag(args, idx, DefaultHook)
	if err != nil {
		return "", "", "", err
	}

	// jump command takes an optional worktree name
	if cmd == "jump" {
		if idx < len(args) {
			name = args[idx]
			if idx+1 < len(args) {
				return "", "", "", fmt.Errorf("unexpected argument: %s", args[idx+1])
			}
		}
		return cmd, name, hookPath, nil
	}

	// gha command takes no additional arguments
	if cmd == "gha" {
		if idx < len(args) {
			return "", "", "", fmt.Errorf("unexpected argument: %s", args[idx])
		}
		return cmd, "", hookPath, nil
	}

	// list command takes no additional arguments
	if cmd == "list" {
		if idx < len(args) {
			return "", "", "", fmt.Errorf("unexpected argument: %s", args[idx])
		}
		return cmd, "", hookPath, nil
	}

	// completion command takes a shell name
	if cmd == "completion" {
		if idx >= len(args) {
			return "", "", "", fmt.Errorf("shell name required (bash, zsh, fish)")
		}
		name = args[idx]
		if idx+1 < len(args) {
			return "", "", "", fmt.Errorf("unexpected argument: %s", args[idx+1])
		}
		return cmd, name, hookPath, nil
	}

	// __complete command takes a subcommand name
	if cmd == "__complete" {
		if idx >= len(args) {
			return "", "", "", fmt.Errorf("subcommand required")
		}
		name = args[idx]
		return cmd, name, hookPath, nil
	}

	// remove command: name is optional (can detect from current worktree)
	if cmd == "remove" && idx >= len(args) {
		return cmd, "", hookPath, nil
	}

	// Remaining arg should be the name
	if idx >= len(args) {
		return "", "", "", fmt.Errorf("branch name required")
	}

	name = args[idx]

	// Validate no extra args
	if idx+1 < len(args) {
		return "", "", "", fmt.Errorf("unexpected argument: %s", args[idx+1])
	}

	return cmd, name, hookPath, nil
}

// runRemove executes the remove command, detecting current worktree if name is empty
func runRemove(name string) error {
	if name == "" {
		wm, err := NewWorktreeManager()
		if err != nil {
			return err
		}
		// CurrentWorktreeName returns empty string if not in worktree (never errors)
		name, _ = wm.CurrentWorktreeName()
		if name == "" {
			return fmt.Errorf("not inside a worktree (specify branch name)")
		}
	}
	return remove(name)
}

// run executes the CLI with the given arguments
func run(args []string) error {
	cmd, name, hookPath, err := parseArgs(args)
	if err != nil {
		return err
	}

	switch cmd {
	case "jump":
		return jump(name)
	case "create":
		return create(name, hookPath)
	case "remove":
		return runRemove(name)
	case "list":
		return list(os.Stdout)
	case "gha":
		return gha()
	case "completion":
		return completion(name, os.Stdout)
	default: // __complete
		if name == "remove" || name == "jump" {
			return completeWorktrees(os.Stdout)
		}
		return nil
	}
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		if errors.Is(err, errShowHelp) {
			printUsage(os.Stdout)
			exitFn(0)
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitFn(1)
		return
	}
}
