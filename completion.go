package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// listWorktreesFn is replaceable for testing
var listWorktreesFn = defaultListWorktrees

func defaultListWorktrees() ([]string, error) {
	root, err := gitMainRoot()
	if err != nil {
		return nil, err
	}

	worktreesDir := filepath.Join(root, ".worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var worktrees []string
	for _, entry := range entries {
		if entry.IsDir() {
			worktrees = append(worktrees, entry.Name())
		}
	}
	return worktrees, nil
}

func listWorktrees() ([]string, error) {
	return listWorktreesFn()
}

// completion generates shell completion scripts
func completion(shell string, w io.Writer) error {
	switch shell {
	case "bash":
		return bashCompletion(w)
	case "zsh":
		return zshCompletion(w)
	case "fish":
		return fishCompletion(w)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}

func bashCompletion(w io.Writer) error {
	script := `_wt_completions() {
    local cur prev words cword
    _init_completion || return

    local commands="jump create remove list gha completion"

    case "${prev}" in
        wt)
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
            return
            ;;
        jump)
            local worktrees
            worktrees=$(wt __complete jump 2>/dev/null)
            COMPREPLY=($(compgen -W "${worktrees}" -- "${cur}"))
            return
            ;;
        remove)
            local worktrees
            worktrees=$(wt __complete remove 2>/dev/null)
            COMPREPLY=($(compgen -W "${worktrees}" -- "${cur}"))
            return
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            return
            ;;
        --hook)
            _filedir
            return
            ;;
    esac

    case "${cur}" in
        -*)
            COMPREPLY=($(compgen -W "--hook -h --help" -- "${cur}"))
            return
            ;;
    esac

    # Default to commands if nothing matched
    if [[ ${cword} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
    fi
}

complete -F _wt_completions wt
`
	_, err := fmt.Fprint(w, script)
	return err
}

func zshCompletion(w io.Writer) error {
	script := `#compdef wt

_wt_worktrees() {
    local worktrees
    worktrees=(${(f)"$(wt __complete jump 2>/dev/null)"})
    _describe -t worktrees 'worktrees' worktrees
}

_wt() {
    local -a commands
    commands=(
        'jump:Jump to a worktree or repo root'
        'create:Create a new worktree with branch'
        'remove:Remove a worktree and its branch'
        'list:List all worktrees'
        'gha:Monitor GitHub Actions status for current branch PR'
        'completion:Generate shell completion script'
    )

    local -a shells
    shells=(bash zsh fish)

    _arguments -C \
        '(-h --help)'{-h,--help}'[Show help message]' \
        '--hook[Custom hook script to run after create]:hook file:_files' \
        '1: :->command' \
        '*: :->args'

    case $state in
        command)
            _describe -t commands 'wt commands' commands
            ;;
        args)
            case $words[2] in
                jump)
                    _wt_worktrees
                    ;;
                remove)
                    _wt_worktrees
                    ;;
                completion)
                    _describe -t shells 'shells' shells
                    ;;
                create)
                    # No completion for branch names (user provides new name)
                    ;;
            esac
            ;;
    esac
}

_wt "$@"
`
	_, err := fmt.Fprint(w, script)
	return err
}

func fishCompletion(w io.Writer) error {
	script := `# Fish completion for wt

function __wt_worktrees
    wt __complete jump 2>/dev/null
end

# Disable file completion by default
complete -c wt -f

# Commands
complete -c wt -n "__fish_use_subcommand" -a "jump" -d "Jump to a worktree or repo root"
complete -c wt -n "__fish_use_subcommand" -a "create" -d "Create a new worktree with branch"
complete -c wt -n "__fish_use_subcommand" -a "remove" -d "Remove a worktree and its branch"
complete -c wt -n "__fish_use_subcommand" -a "list" -d "List all worktrees"
complete -c wt -n "__fish_use_subcommand" -a "gha" -d "Monitor GitHub Actions status"
complete -c wt -n "__fish_use_subcommand" -a "completion" -d "Generate shell completion script"

# Options
complete -c wt -s h -l help -d "Show help message"
complete -c wt -l hook -r -d "Custom hook script to run after create"

# Worktree completion for jump
complete -c wt -n "__fish_seen_subcommand_from jump" -a "(__wt_worktrees)"

# Worktree completion for remove
complete -c wt -n "__fish_seen_subcommand_from remove" -a "(__wt_worktrees)"

# Shell completion for completion command
complete -c wt -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
`
	_, err := fmt.Fprint(w, script)
	return err
}

// completeWorktrees outputs worktree names for shell completion
func completeWorktrees(w io.Writer) error {
	worktrees, err := listWorktrees()
	if err != nil {
		return err
	}
	for _, wt := range worktrees {
		fmt.Fprintln(w, wt)
	}
	return nil
}
