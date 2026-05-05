# AGENTS.md

## What This Project Is

**crushout** is a [Crush](https://crush.dev) hook binary that auto-allows read-only bash commands. It reads a JSON payload from stdin (Crush's hook protocol), parses the bash command with tree-sitter, checks every command against a rule set, and outputs a JSON allow/deny decision on stdout.

## Commands

```bash
go test ./...          # run all tests
go build ./cmd/crushout # build the binary
```

No Makefile, no CI config, no linter config, no other tooling.

## Architecture & Data Flow

```
stdin (Crush hook JSON)
  → main.go: decode Input, skip non-bash/empty commands
  → bash.Parse(): tree-sitter parses bash into []Command
  → checker.IsReadOnly(): iterate commands, track cd, apply rules
  → stdout: JSON Output with decision="allow" or empty {}
```

### Packages

| Package | Purpose |
|---|---|
| `cmd/crushout` | Entrypoint. Reads stdin JSON, wires checker, writes stdout JSON. |
| `internal/hook` | `Input` and `Output` structs for the Crush hook JSON protocol. |
| `internal/bash` | Tree-sitter–based bash parser. Extracts `Command` structs from a bash AST. Detects "complex" constructs (command substitution `$()`, subshells `()`, process substitution `<()`, arithmetic `$(())`) and parse errors, both will not auto-allow. |
| `internal/checker` | Orchestrates parse + rule check. Tracks `cd` to ensure the working directory stays within `RootDir`. Resolves `~/...` paths against `HomeDir`. Any `cd` that would escape the root causes a deny. |
| `internal/rules` | Recursive rule engine (`Rule` struct with `Subcommands`, `DenyFlags`, `Default`). `defaults.go` contains the full built-in ruleset mapping command names to their allow/deny rules. |

## Key Design Decisions

- **Fail-closed**: anything ambiguous or unknown is denied. Parse errors, complex bash constructs, unknown commands, commands with `$` in the name, and missing rules all return `{}` (no decision).
- **"Complex" bash is rejected outright**: command substitution, process substitution, subshells, and arithmetic expansion set `IsComplex=true` and skip command extraction entirely.
- **`cd` tracking**: the checker tracks the current working directory across `cd` commands in a chain. `cd` with `$VAR`, `~` (if home is outside root), `-`, or any path resolving outside `RootDir` denies the whole command.
- **Output convention**: returning `{}` (empty JSON object) means "no opinion". Only `{"version":1,"decision":"allow"}` auto-allows.

## Testing Patterns

- Standard `testing` package only, no assertion libraries. Tests define local helpers (`assertNoError`, `assertCommand`).
- `checker_test.go` uses `newTestChecker()` with `RootDir: "/home/user/project"` and `HomeDir: "/home/user"`.
- Rule tests in `rule_test.go` construct minimal `Rule` trees to test subcommand resolution, deny flags, and nesting, then also test the full `Default` ruleset.
- The `cmd/crushout` and `internal/hook` packages have no tests.

## Gotchas

- The tree-sitter bash grammar treats `git -C /tmp status` with `-C` as an anonymous (un-named) child node. This means `-C` does **not** appear in `cmd.Args`. The deny works via `DenyFlags` on the rule which checks the raw args, but the tree-sitter parse won't include it in the structured args. This is why `checker.isReadOnly` uses `filepath.Base` on names with `/` and rejects names containing `$` or backticks.
- `Rule.resolve` walks args as a subcommand chain. The first arg matching a subcommand key descends into that sub-rule, consuming the arg. Remaining args are then checked against `DenyFlags` at the new level. This means flag position matters: `git -C /tmp status` has `-C` checked at the top-level git rule, but `git branch -l` descends into the `branch` sub-rule.
- `cd` with no arguments (bare `cd`) resolves to `$HOME`. It's only allowed if `HomeDir` is within `RootDir`.
