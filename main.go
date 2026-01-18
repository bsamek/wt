package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const defaultHook = ".worktree-hook"

// Sentinel errors for testing
var (
	errShowHelp     = errors.New("show help")
	errShowHelpFail = errors.New("show help (error)")
)

// exitFn is the exit function, replaceable for testing
var exitFn = os.Exit

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

// parseArgs parses command line arguments and returns (command, name, hookPath, error)
func parseArgs(args []string) (cmd string, name string, hookPath string, err error) {
	if len(args) == 0 {
		return "", "", "", errShowHelpFail
	}

	// Check for help flag anywhere
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return "", "", "", errShowHelp
		}
	}

	// Default values
	cmd = "create"
	hookPath = defaultHook

	i := 0

	// Check if first arg is a command
	if len(args) > 0 {
		switch args[0] {
		case "create":
			cmd = "create"
			i++
		case "remove":
			cmd = "remove"
			i++
		case "gha":
			cmd = "gha"
			i++
		case "completion":
			cmd = "completion"
			i++
		case "__complete":
			cmd = "__complete"
			i++
		}
	}

	// Parse flags
	for i < len(args) {
		if args[i] == "--hook" {
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--hook requires a path argument")
			}
			hookPath = args[i+1]
			i += 2
		} else if len(args[i]) > 0 && args[i][0] == '-' {
			return "", "", "", fmt.Errorf("unknown flag %s", args[i])
		} else {
			break
		}
	}

	// gha command takes no additional arguments
	if cmd == "gha" {
		if i < len(args) {
			return "", "", "", fmt.Errorf("unexpected argument: %s", args[i])
		}
		return cmd, "", hookPath, nil
	}

	// completion command takes a shell name
	if cmd == "completion" {
		if i >= len(args) {
			return "", "", "", fmt.Errorf("shell name required (bash, zsh, fish)")
		}
		name = args[i]
		if i+1 < len(args) {
			return "", "", "", fmt.Errorf("unexpected argument: %s", args[i+1])
		}
		return cmd, name, hookPath, nil
	}

	// __complete command takes a subcommand name
	if cmd == "__complete" {
		if i >= len(args) {
			return "", "", "", fmt.Errorf("subcommand required")
		}
		name = args[i]
		return cmd, name, hookPath, nil
	}

	// Remaining arg should be the name
	if i >= len(args) {
		return "", "", "", fmt.Errorf("branch name required")
	}

	name = args[i]

	// Validate no extra args
	if i+1 < len(args) {
		return "", "", "", fmt.Errorf("unexpected argument: %s", args[i+1])
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
