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

Commands:
  create    Create a new worktree with branch (default if no command given)
  remove    Remove a worktree and its branch

Options:
  --hook <path>    Custom hook script to run after create (default: .worktree-hook)
  -h, --help       Show this help message

Examples:
  wt my-feature              Create worktree for 'my-feature' branch
  wt create my-feature       Same as above
  wt --hook setup.sh feat    Create worktree, run setup.sh as hook
  wt remove my-feature       Remove worktree and branch
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

	// parseArgs guarantees cmd is "create" or "remove"
	if cmd == "remove" {
		return remove(name)
	}
	return create(name, hookPath)
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
