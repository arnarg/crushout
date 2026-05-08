# crushout

A PreToolUse hook for the Bash tool that auto-approves read-only commands and can hard-block dangerous ones via config. Works with [Crush](https://github.com/charmbracelet/crush) and [Claude Code](https://code.claude.com/docs/en/hooks).

> **Note:** I use crushout with Crush. The Claude Code protocol is implemented based on their docs but I haven't personally verified it end-to-end. Bug reports welcome.

When an agent runs `git status`, `ls -la`, or `cat file.txt`, crushout skips the permission prompt. When it runs `rm -rf /`, `git push`, or anything it can't confidently classify, crushout stays silent and the normal permission flow takes over. With a `.crushout.yml` config, you can also hard-deny specific commands so they're blocked outright.

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

### rtk rewrite

If [rtk](https://github.com/rtk-ai/rtk) is installed in `$PATH`, crushout passes non-denied commands through `rtk rewrite` for Crush or Claude Code to run instead.

This is enabled by default. Disable it in `.crushout.yml`:

```yaml
rtk_rewrite: false
```

The rewrite is invisible to the checker as it runs after the allow/deny decision. If `rtk` is not installed, the command passes through unchanged.

## Install

```bash
go build -o crushout ./cmd/crushout
```

Move the binary somewhere on your `$PATH`, or use a full path in the config.

## Configure

### Crush

Add to your project's `.crush.json`:

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

### Claude Code

Add to your project's `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "crushout"
          }
        ]
      }
    ]
  }
}
```

Or use a full path in either:

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
| `git config get user.name` | ✅ auto | `config` with `list`/`get` |
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
| `go mod tidy` | 🔒 prompt | nested: `mod` → `tidy` |
| `cd /tmp && ls` | 🔒 prompt | `cd` escapes root |
| `cd .. && ls` | 🔒 prompt | `cd` escapes root |
| `echo $(whoami)` | 🔒 prompt | command substitution |
| `make build` | 🔒 prompt | mutable command |
| `docker run nginx` | 🔒 prompt | mutable subcommand |
| `npm install` | 🔒 prompt | mutable command |
| `unknown_cmd --foo` | 🔒 prompt | not in rules |
| | | |
| `curl https://example.com` *(with `rules: {curl: deny}`)* | 🚫 deny | hard-denied by config |
| `kubectl exec -it pod -- bash` *(with deny rule)* | 🚫 deny | hard-denied by config |

## Project structure

```
crushout/
├── cmd/crushout/main.go          # entry point, crush protocol I/O
├── internal/
│   ├── hook/protocol.go          # Hook interface, Crush/Claude Code protocol types
│   ├── bash/parse.go             # tree-sitter parsing and AST traversal
│   ├── rewrite/rewrite.go        # rtk rewrite integration
│   ├── config/config.go          # .crushout.yml loading and merging
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
        Default:   rules.Allow,                    // allow unknown subcommands
        DenyFlags: []string{"--dangerous"},        // prompt on these flags
        Subcommands: map[string]*Rule{
            "read":  {Default: rules.Allow},
            "write": {Default: rules.NoOpinion},   // prompt
            "db": {
                Subcommands: map[string]*Rule{
                    "migrate": {Default: rules.NoOpinion},
                    "seed":    {Default: rules.NoOpinion},
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

## Custom config file

Instead of editing the source, you can drop `.crushout.yml` or `.crushout.yaml` in your project root. crushout looks for it in the `cwd` passed by the hook (typically the repo root).

### YAML format

```yaml
overwrite_defaults: false
rules:
  nix:
    decision: prompt
    subcommands:
      build:
        decision: allow
  git:
    subcommands:
      status:
        decision: prompt  # require confirmation even for status
```

#### Shorthand syntax

Rules can be written as a simple string instead of a full mapping:

```yaml
rules:
  ls: allow
  rm: deny
  kubectl:
    decision: prompt    # default if no subcommand matches
    subcommands:
      get: allow        # allow `kubectl get *`
      exec:             # deny `kubectl exec *`
        decision: deny
        message: "no remote execution"
```

The string form (`allow`, `deny`, or `prompt`) is equivalent to `{decision: <value>}`.

### Merge behavior

When `overwrite_defaults: false` (the default), user rules are **deep-merged** with the built-in rules. Your values win where they differ, but anything you omit is inherited from the defaults.

When `overwrite_defaults: true`, only the rules you specify are active; the built-ins are ignored entirely.

### Fields

| Field | Type | Description |
|---|---|---|
| `rtk_rewrite` | bool | If `true`, pass commands through `rtk rewrite`. Default is `true`. |
| `overwrite_defaults` | bool | If `true`, ignore built-in rules. Default is `false`. |
| `rules` | map | Map of command name → rule. |
| `rules.*` | string or map | Shorthand (`allow`, `deny`, `prompt`) or full rule mapping. |
| `rules.*.decision` | string | Decision for unknown subcommands: `allow`, `deny`, or `prompt`. Defaults to `prompt` if not set. |
| `rules.*.deny_flags` | []string | Flags that always require confirmation. |
| `rules.*.message` | string | Custom message shown when denied. Only used with `decision: deny`. |
| `rules.*.subcommands` | map | Recursive map of subcommand name → rule. |

### Examples

**Allow `nix build` but prompt for everything else under `nix`:**

```yaml
rules:
  nix:
    decision: prompt
    subcommands:
      build: allow
```

**Prompt on `git status` (normally allowed):**

```yaml
rules:
  git:
    subcommands:
      status:
        decision: prompt
```

**Hard-deny `curl` and `kubectl exec`:**

```yaml
rules:
  curl: deny
  kubectl:
    decision: prompt
    subcommands:
      get: allow
      exec:
        decision: deny
        message: "no remote execution in this project"
```

**Start fresh with only `ls` allowed (everything else is prompted):**

```yaml
overwrite_defaults: true
rules:
  ls: allow
```

## Hook protocol

crushout implements the [Crush PreToolUse hook protocol](https://github.com/charmbracelet/crush/blob/main/docs/hooks/README.md) and the [Claude Code PreToolUse hook protocol](https://code.claude.com/docs/en/hooks). It reads JSON on stdin, detects which format is being used based on the field names, and writes the correct output format on stdout.

### Crush

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

Output (auto-approve with rtk rewrite):

```json
{"version": 1, "decision": "allow", "updated_input": {"command": "rewritten command"}}
```

Output (no opinion, let the normal permission prompt handle it):

```json
{}
```

Output (no opinion with rtk rewrite):

```json
{"updated_input": {"command": "rewritten command"}}
```

Output (hard-deny):

```json
{"version": 1, "decision": "deny", "reason": "crushout: no curl allowed (rule for 'curl')"}
```

### Claude Code

Input:

```json
{
  "session_id": "abc123",
  "transcript_path": "/home/user/.claude/projects/.../transcript.jsonl",
  "cwd": "/home/user/project",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "git status" }
}
```

Output (auto-approve):

```json
{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow"}}
```

Output (auto-approve with rtk rewrite):

```json
{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "allow", "updatedInput": {"command": "rewritten command"}}}
```

Output (ask user, let the normal permission prompt handle it):

```json
{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "ask"}}
```

Output (ask user with rtk rewrite):

```json
{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "ask", "updatedInput": {"command": "rewritten command"}}}
```

Output (hard-deny):

```json
{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "crushout: no curl allowed (rule for 'curl')"}}
```

## About deny

By default, crushout only has two outcomes: **allow** (auto-approve) and **no opinion** (fall through to the normal permission prompt). The built-in rules never deny, they either fast-track safe commands or get out of the way.

Through `.crushout.yml`, you can add explicit **deny** rules. A deny is final, the model sees the error and tries something else. This is useful when you want to block a command the defaults would otherwise allow (or just prompt for):

- Deny `kubectl exec` outright instead of letting the user approve it each time
- Deny `curl` entirely in a sensitive project
- Deny `git push --force` specifically while still prompting for plain `git push`

Use `deny` sparingly. The permission prompt already exists as the human-in-the-loop gate. False negatives (a safe command that isn't auto-approved) are just a minor inconvenience. False positives (a dangerous command that gets auto-approved) are the real threat, and the conservative default-unknown-to-prompt posture avoids them.

## License

MIT
