# AGENTS.md

## What This Project Is

**crushout** is a [Crush](https://crush.dev) / [Claude Code](https://code.claude.com/docs/en/hooks) hook binary that auto-allows read-only bash commands. It reads a JSON payload from stdin (hook protocol), detects whether it's Crush or Claude Code format, parses the bash command with tree-sitter, checks every command against a rule set, and outputs a JSON allow/deny/prompt decision on stdout.

## Commands

```bash
go test ./...          # run all tests
go build ./cmd/crushout # build the binary
```

No Makefile, no CI config, no linter config, no other tooling.

## Architecture & Data Flow

```
stdin (hook JSON, Crush or Claude Code format)
  → main.go: hook.Decode() detects protocol, skip non-bash/empty commands
  → config.Load(): read .crushout.yml if present, merge with defaults
  → bash.Parse(): tree-sitter parses bash into []Command
  → checker.Check(): iterate commands, track cd, apply rules
  → stdout: JSON Output (protocol-specific allow/deny/prompt)
```

### Packages

| Package | Purpose |
|---|---|
| `cmd/crushout` | Entrypoint. Reads stdin JSON, loads config, wires checker, writes stdout JSON. |
| `internal/hook` | `Hook` interface, `CrushInput`/`ClaudeInput` structs, `Decode()` auto-detects protocol from field names. |
| `internal/bash` | Tree-sitter–based bash parser. Extracts `Command` structs from a bash AST. Detects "complex" constructs (command substitution `$()`, subshells `()`, process substitution `<()`, arithmetic `$(())`) and parse errors, both will not auto-allow. |
| `internal/checker` | Orchestrates parse + rule check. Tracks `cd` to ensure the working directory stays within `RootDir`. Resolves `~/...` paths against `HomeDir`. Any `cd` that would escape the root causes a deny. |
| `internal/config` | Loads `.crushout.yml`/`.crushout.yaml`, deep-merges user rules over built-in defaults. Supports `overwrite_defaults: true` to replace defaults entirely. |
| `internal/rules` | Recursive rule engine (`Rule` struct with `Subcommands`, `DenyFlags`, `Default`). `defaults.go` contains the full built-in ruleset mapping command names to their allow/deny rules. |

## Key Design Decisions

- **Fail-closed**: anything ambiguous or unknown falls through to the normal permission prompt. Parse errors, complex bash constructs, unknown commands, commands with `$` in the name, and missing rules all produce a "no opinion" / "ask" result.
- **Three outcomes**: `allow` (auto-approve), `deny` (hard-block with reason), and the default "no opinion" (fall through to normal permission flow). The built-in rules never deny; deny rules come from `.crushout.yml` config only.
- **"Complex" bash is rejected outright**: command substitution, process substitution, subshells, and arithmetic expansion set `IsComplex=true` and skip command extraction entirely.
- **`cd` tracking**: the checker tracks the current working directory across `cd` commands in a chain. `cd` with `$VAR`, `~` (if home is outside root), `-`, or any path resolving outside `RootDir` denies the whole command.
- **Output convention**: returning a protocol-specific "no opinion" payload (`{}` for Crush, `ask` for Claude Code) means fall through to normal prompt. `allow` auto-approves. `deny` hard-blocks with a reason string.

## Testing Patterns

- Standard `testing` package only, no assertion libraries. Tests define local helpers (`assertNoError`, `assertCommand`).
- `checker_test.go` uses `newTestChecker()` with `RootDir: "/home/user/project"` and `HomeDir: "/home/user"`.
- Config tests in `config_test.go` cover YAML parsing, shorthand syntax, and merge behavior.
- Rule tests in `rule_test.go` construct minimal `Rule` trees to test subcommand resolution, deny flags, and nesting, then also test the full `Default` ruleset.
- E2E tests in `tests/e2e/` run the full binary against JSONL test cases (`cases_crush.jsonl`, `cases_claude.jsonl`) via `run.sh`.
- The `cmd/crushout` and `internal/hook` packages have no unit tests.

## Gotchas

- The tree-sitter bash grammar treats `git -C /tmp status` with `-C` as an anonymous (un-named) child node. This means `-C` does **not** appear in `cmd.Args`. The deny works via `DenyFlags` on the rule which checks the raw args, but the tree-sitter parse won't include it in the structured args. This is why `checker.isReadOnly` uses `filepath.Base` on names with `/` and rejects names containing `$` or backticks.
- `Rule.resolve` walks args as a subcommand chain. The first arg matching a subcommand key descends into that sub-rule, consuming the arg. Remaining args are then checked against `DenyFlags` at the new level. This means flag position matters: `git -C /tmp status` has `-C` checked at the top-level git rule, but `git branch -l` descends into the `branch` sub-rule.
- `cd` with no arguments (bare `cd`) resolves to `$HOME`. It's only allowed if `HomeDir` is within `RootDir`.
