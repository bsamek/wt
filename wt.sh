# wt - Git worktree manager shell wrapper
# Source this file in your .bashrc or .zshrc to enable directory changing
#
# Usage:
#   source /path/to/wt.sh
#
# Or add to your shell rc file:
#   source /path/to/wt.sh

wt() {
    # Find the real wt binary (use type -P to bypass the function)
    local wt_bin
    wt_bin=$(type -P wt 2>/dev/null)
    if [[ -z "$wt_bin" ]]; then
        echo "error: wt binary not found in PATH" >&2
        return 1
    fi

    # Pass through commands that produce non-directory output
    case "$1" in
        completion|__complete|"")
            "$wt_bin" "$@"
            return $?
            ;;
    esac

    # Run wt and capture stdout (the directory path)
    local dir
    dir=$("$wt_bin" "$@")
    local exit_code=$?

    # If successful and we got a directory path, cd into it
    if [[ $exit_code -eq 0 && -n "$dir" && -d "$dir" ]]; then
        cd "$dir" || return 1
    fi

    return $exit_code
}
