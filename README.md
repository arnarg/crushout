# crushout

A [Crush](https://github.com/charmbracelet/crush) hook that auto-approves read-only bash commands.

When an agent runs `git status`, `ls -la`, or `cat file.txt`, crushout skips the permission prompt. When it runs `rm -rf /`, `git push`, or anything it can't confidently classify, crushout stays silent and the normal permission flow takes over.

It never denies, it only fast-tracks safe commands or gets out of the way.

## How it works

crushout parses bash commands with [tree-sitter](https://tree-sitter.github.io/tree-sitter/) and walks the AST to extract every individual command. It checks each one against a recursive rule system:

- **Non-mutable** - always allowed (`ls`, `cat`, `grep`, `git status`, `go test`)
- **Mutable** - always requires confirmation (`rm`, `sudo`, `git push`, `make`)
- **Flag-dependent** - allowed unless a dangerous flag is present (`sed` is fine, `sed -i` is not; `find` is fine, `find -exec` is not)
- **Subcommand-dependent** - resolved recursively (`git remote -v` ✓, `git remote add` ✗; `go mod graph` ✓, `go mod tidy` ✗)
- **Unknown** - requires confirmation (deny by default)

If **any** command in a chain (`&&`, `||`, `|`, `;`) is mutable, the whole thing goes through the normal permission prompt.

### Path tracking

crushout tracks working directory across `cd` calls. If a `cd` would move outside the initial project root, the command is not auto-approved.

`git -C`, `git --work-tree`, and `git --git-dir` are no auto-approved as they bypass the tracked working directory.

### Safety guards

Commands are rejected from auto-approval (not denied, just not fast-tracked) if they contain:

- Command substitutions: `$(...)`
- Process substitutions: `<(...)`
- Subshells: `(...)`
- Arithmetic expansion: `$((...))`
- Parse errors
- Dynamic command names like `$CMD`

## Install

```bash
go build -o crushout ./cmd/crushout
```

Move the binary somewhere on your `$PATH`, or use a full path in the config.

## Configure

Add to your project's `crush.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "^bash$",
        "command": "crushout",
        "timeout": 5
      }
    ]
  }
}
```

Or use a full path:

```json
"command": "/usr/local/bin/crushout"
```

## What gets approved

| Command | Result | Why |
|---|---|---|
| `ls -la` | ✅ auto | always non-mutable |
| `cat file \| grep foo \| sort` | ✅ auto | all non-mutable |
| `git status` | ✅ auto | read-only subcommand |
| `git diff HEAD~3` | ✅ auto | read-only subcommand |
| `git log --oneline -10` | ✅ auto | read-only subcommand |
| `git remote -v` | ✅ auto | nested: `remote` → `-v` |
| `git remote show origin` | ✅ auto | nested: `remote` → `show` |
| `git branch -l` | ✅ auto | nested: `branch` → `-l` |
| `git branch --show-current` | ✅ auto | nested: `branch` → `--show-current` |
| `git tag -l` | ✅ auto | nested: `tag` → `-l` |
| `git stash list` | ✅ auto | nested: `stash` → `list` |
| `git config --get user.name` | ✅ auto | `config` without `--global`/`--system` |
| `find . -name '*.go'` | ✅ auto | `find` without `-exec`/`-delete` |
| `sed 's/old/new/g' file` | ✅ auto | `sed` without `-i` |
| `go test ./...` | ✅ auto | read-only subcommand |
| `go mod graph` | ✅ auto | nested: `mod` → `graph` |
| `cd src && ls -la && cd ..` | ✅ auto | `cd` stays within root |
| `pwd && whoami && uname -a` | ✅ auto | all non-mutable |
| `jq '.name' data.json` | ✅ auto | always non-mutable |
| `sha256sum file` | ✅ auto | always non-mutable |
| `xxd file.bin` | ✅ auto | always non-mutable |
| `yq '.key' file.yaml` | ✅ auto | `yq` without `-i` |
| `cargo test` | ✅ auto | read-only subcommand |
| `cargo tree` | ✅ auto | read-only subcommand |
| `gh pr list` | ✅ auto | nested: `pr` → `list` |
| `gh repo view owner/repo` | ✅ auto | nested: `repo` → `view` |
| `kubectl get pods` | ✅ auto | read-only subcommand |
| `kubectl logs my-pod` | ✅ auto | read-only subcommand |
| `docker ps` | ✅ auto | read-only subcommand |
| `docker images` | ✅ auto | read-only subcommand |
| `docker volume ls` | ✅ auto | nested: `volume` → `ls` |
| | | |
| `rm -rf /` | 🔒 prompt | mutable command |
| `sudo rm -rf /` | 🔒 prompt | mutable command |
| `git push` | 🔒 prompt | mutable subcommand |
| `git commit -m 'fix'` | 🔒 prompt | mutable subcommand |
| `git branch new-feature` | 🔒 prompt | `branch` without list flag |
| `git branch -D old` | 🔒 prompt | `-D` in deny flags |
| `git tag v1.0.0` | 🔒 prompt | `tag` without `-l` |
| `git stash` | 🔒 prompt | bare `stash` = push |
| `git -C /tmp status` | 🔒 prompt | `-C` is denied |
| `git config --global user.name X` | 🔒 prompt | `--global` is denied |
| `find . -exec rm {} \;` | 🔒 prompt | `-exec` is denied |
| `sed -i 's/old/new/g' file` | 🔒 prompt | `-i` is denied |
| `go build ./...` | 🔒 prompt | mutable subcommand |
| `go mod tidy` | 🔒 prompt | nested: `mod` → `tidy` |
| `cd /tmp && ls` | 🔒 prompt | `cd` escapes root |
| `cd .. && ls` | 🔒 prompt | `cd` escapes root |
| `echo $(whoami)` | 🔒 prompt | command substitution |
| `make build` | 🔒 prompt | mutable command |
| `docker run nginx` | 🔒 prompt | mutable subcommand |
| `npm install` | 🔒 prompt | mutable command |
| `unknown_cmd --foo` | 🔒 prompt | not in rules |

## Project structure

```
crushout/
├── cmd/crushout/main.go          # entry point, crush protocol I/O
├── internal/
│   ├── hook/protocol.go          # HookInput, HookOutput types
│   ├── bash/parse.go             # tree-sitter parsing and AST traversal
│   ├── rules/rule.go             # recursive Rule type and resolution
│   ├── rules/defaults.go         # built-in rule definitions
│   └── checker/checker.go        # orchestrator, path tracking
├── go.mod
└── go.sum
```

## Customizing rules

Edit `internal/rules/defaults.go`. The rule type is recursive:

```go
var Default = map[string]*Rule{
    "my-tool": {
        Default:   true,                          // allow unknown subcommands
        DenyFlags: []string{"--dangerous"},       // block these flags
        Subcommands: map[string]*Rule{
            "read":  {Default: true},
            "write": {Default: false},
            "db": {
                Default: false,
                Subcommands: map[string]*Rule{
                    "migrate": {Default: false},
                    "seed":    {Default: false},
                },
            },
        },
    },
}
```

Resolution walks arguments left-to-right:

1. `DenyFlags` are checked at each level against all remaining args
2. If an arg matches a `Subcommands` key, descend into that rule
3. When no deeper match is found, use `Default`

## Hook protocol

crushout implements the [Crush PreToolUse hook protocol](https://github.com/charmbracelet/crush/blob/main/docs/hooks/README.md). It reads JSON on stdin and writes JSON on stdout.

Input:

```json
{
  "event": "PreToolUse",
  "session_id": "abc123",
  "cwd": "/home/user/project",
  "tool_name": "bash",
  "tool_input": { "command": "git status" }
}
```

Output (auto-approve):

```json
{"version": 1, "decision": "allow"}
```

Output (no opinion, let the normal permission prompt handle it):

```json
{}
```

## Why not deny?

A deny from a hook is final and the model sees the error and tries something else, which costs tokens and time. The permission prompt already exists as the human-in-the-loop gate. crushout only short-circuits when it's **confident** the command is safe, and stays silent for everything else.

This also means false negatives (a safe command that isn't auto-approved) are just a minor inconvenience. False positives (a dangerous command that gets auto-approved) are the real threat, and the conservative default-unknown-to-prompt posture avoids them.

## License

MIT
