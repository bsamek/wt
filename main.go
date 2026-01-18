package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Sentinel errors for testing
var (
	errShowHelp     = errors.New("show help")
	errShowHelpFail = errors.New("show help (error)")
)

// exitFn is the exit function, replaceable for testing
var exitFn = os.Exit

// validCommands lists all valid command names
var validCommands = []string{"create", "remove", "gha", "completion", "__complete"}

func usageText() string {
	return `Usage: wt [options] <name>
       wt create [options] <name>
       wt remove <name>
       wt gha
       wt completion <shell>

Commands:
  create      Create a new worktree with branch (default if no command given)
  remove      Remove a worktree and its branch
  gha         Monitor GitHub Actions status for current branch's PR
  completion  Generate shell completion script (bash, zsh, fish)

Options:
  --hook <path>    Custom hook script to run after create (default: .worktree-hook)
  -h, --help       Show this help message

Examples:
  wt my-feature              Create worktree for 'my-feature' branch
  wt create my-feature       Same as above
  wt --hook setup.sh feat    Create worktree, run setup.sh as hook
  wt remove my-feature       Remove worktree and branch
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

// parseCommand extracts the command from arguments, defaulting to "create"
func parseCommand(args []string) (cmd string, startIdx int) {
	if len(args) == 0 {
		return "create", 0
	}
	if isValidCommand(args[0]) {
		return args[0], 1
	}
	return "create", 0
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
		return "", "", "", errShowHelpFail
	}

	if isHelpRequested(args) {
		return "", "", "", errShowHelp
	}

	cmd, idx := parseCommand(args)

	// Parse hook flag
	idx, hookPath, err = parseHookFlag(args, idx, DefaultHook)
	if err != nil {
		return "", "", "", err
	}

	// gha command takes no additional arguments
	if cmd == "gha" {
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

// run executes the CLI with the given arguments
func run(args []string) error {
	cmd, name, hookPath, err := parseArgs(args)
	if err != nil {
		return err
	}

	switch cmd {
	case "remove":
		return remove(name)
	case "gha":
		return gha()
	case "completion":
		return completion(name, os.Stdout)
	case "__complete":
		if name == "remove" {
			return completeWorktrees(os.Stdout)
		}
		return nil
	default:
		return create(name, hookPath)
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
		if errors.Is(err, errShowHelpFail) {
			printUsage(os.Stderr)
			exitFn(1)
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitFn(1)
		return
	}
}
