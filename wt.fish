# wt - Git worktree manager shell wrapper for Fish
# Source this file or add to ~/.config/fish/functions/wt.fish
#
# Usage:
#   source /path/to/wt.fish
#
# Or copy to ~/.config/fish/functions/wt.fish for auto-loading

function wt --description "Git worktree manager with auto-cd"
    # Find the real wt binary
    set -l wt_bin (command -v wt 2>/dev/null)
    if test -z "$wt_bin"
        echo "error: wt binary not found in PATH" >&2
        return 1
    end

    # Pass through commands that produce non-directory output (including no args for help)
    if test (count $argv) -eq 0
        $wt_bin
        return $status
    end

    switch $argv[1]
        case completion __complete
            $wt_bin $argv
            return $status
    end

    # Run wt and capture stdout (the directory path)
    set -l dir ($wt_bin $argv)
    set -l exit_code $status

    # If successful and we got a directory path, cd into it
    if test $exit_code -eq 0 -a -n "$dir" -a -d "$dir"
        cd $dir
    end

    return $exit_code
end
