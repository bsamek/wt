# wt

A lightweight CLI for managing Git worktrees. Work on multiple branches simultaneously without switching contexts.

## Installation

**Option 1: Go install (recommended)**

If you have a working Go installation and `$GOPATH/bin` (or `$GOBIN`) in your PATH:

```bash
go install .
```

**Option 2: Build from source**

```bash
go build -o wt .
```

Move the binary to a directory in your PATH.

**Option 3: Download from GitHub releases**

Download the executable for your platform from the [GitHub releases](https://github.com/bsamek/wt/releases) page and place it in a directory in your PATH.

## Usage

```
wt [options] <name>
wt create [options] <name>
wt remove <name>
```

### Commands

| Command | Description |
|---------|-------------|
| `create` | Create a new worktree with branch (default if no command given) |
| `remove` | Remove a worktree and its branch |

### Options

| Option | Description |
|--------|-------------|
| `--hook <path>` | Custom hook script to run after create (default: `.worktree-hook`) |
| `-h, --help` | Show help message |

### Examples

```bash
wt my-feature              # Create worktree for 'my-feature' branch
wt create my-feature       # Same as above
wt --hook setup.sh feat    # Create worktree, run setup.sh as hook
wt remove my-feature       # Remove worktree and branch
```

## How It Works

Worktrees are created in a `.worktrees/` directory at the repository root:

```
my-repo/
├── .worktrees/
│   ├── my-feature/      # Working directory for my-feature branch
│   └── bugfix/          # Working directory for bugfix branch
└── ...
```

Each worktree has its own working directory, so you can have different branches checked out simultaneously.

### Claude Code Support

If your repository has a `.claude/` directory (used by [Claude Code](https://claude.ai/code) for settings and context), `wt` automatically creates a symlink to it in each new worktree. This keeps your Claude configuration in sync across all worktrees without needing to copy or merge changes.

## Development

Run tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

The CI pipeline enforces 100% test coverage.
