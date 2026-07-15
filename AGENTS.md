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
  → config.Load(): load global + repo config, merge with defaults
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
| `internal/config` | Layered config loading. Reads global (`$XDG_CONFIG_HOME/crushout/crushout.{yml,yaml}`) and repo (`<rootDir>/.crushout.{yml,yaml}`) config, deep-merges both over built-in defaults. Repo config wins over global. Supports `overwrite_defaults: true` per-layer to replace the accumulated base entirely. |
| `internal/rules` | Recursive rule engine (`Rule` struct with `Subcommands`, `PromptFlags`, `AllowFlags`, `Default`). `defaults.go` contains the full built-in ruleset mapping command names to their allow/deny rules. |

## Key Design Decisions

- **Fail-closed**: anything ambiguous or unknown falls through to the normal permission prompt. Parse errors, complex bash constructs, unknown commands, commands with `$` in the name, and missing rules all produce a `prompt` result.
- **Three outcomes**: `allow` (auto-approve), `deny` (hard-block with reason), and `prompt` (fall through to normal permission flow). The built-in rules never deny; deny rules come from `.crushout.yml` config only.
- **"Complex" bash is rejected outright**: command substitution, process substitution, subshells, and arithmetic expansion set `IsComplex=true` and skip command extraction entirely.
- **Layered config**: crushout reads two config layers merged in order — global (`$XDG_CONFIG_HOME/crushout/crushout.{yml,yaml}`) then repo (`<rootDir>/.crushout.{yml,yaml}`). Each layer is deep-merged over the accumulated base, with later layers winning. `overwrite_defaults: true` in a layer replaces the accumulated base entirely (so a global `overwrite_defaults: true` drops built-ins that no later layer can recover). Scalar fields like `rtk_rewrite` use `*bool` internally so an unset value is distinguishable from an explicit `false`.
- **`cd` tracking**: the checker tracks the current working directory across `cd` commands in a chain. `cd` with `$VAR`, `~` (if home is outside root), `-`, or any path resolving outside `RootDir` denies the whole command.
- **Output convention**: returning a protocol-specific `prompt` payload (`{}` for Crush, `ask` for Claude Code) means fall through to normal prompt. `allow` auto-approves. `deny` hard-blocks with a reason string.

## Testing Patterns

- Standard `testing` package only, no assertion libraries. Tests define local helpers (`assertNoError`, `assertCommand`).
- `checker_test.go` uses `newTestChecker()` with `RootDir: "/home/user/project"` and `HomeDir: "/home/user"`.
- Config tests in `config_test.go` cover YAML parsing, shorthand syntax, single-layer loading (`loadFirst`), layered merging (`load`), `resolveRtkRewrite`, and `applyLayer`/`buildRules` overwrite matrix.
- Rule tests in `rule_test.go` construct minimal `Rule` trees to test subcommand resolution, `PromptFlags`, `AllowFlags`, and nesting, then also test the full `Default` ruleset.
- E2E tests in `tests/e2e/` run the full binary against JSONL test cases (`cases_crush.jsonl`, `cases_claude.jsonl`) via `run.sh`.
- The `cmd/crushout` and `internal/hook` packages have no unit tests.

## Gotchas

- The tree-sitter bash grammar treats `git -C /tmp status` with `-C` as an anonymous (un-named) child node. This means `-C` does **not** appear in `cmd.Args`. The deny works via `PromptFlags` on the rule which checks the raw args, but the tree-sitter parse won't include it in the structured args. This is why `checker.isReadOnly` uses `filepath.Base` on names with `/` and rejects names containing `$` or backticks.
- `Rule.resolve` walks args as a subcommand chain. The first arg matching a subcommand key descends into that sub-rule, consuming the arg. Remaining args are then checked against `PromptFlags` and `AllowFlags` at the new level. This means flag position matters: `git -C /tmp status` has `-C` checked at the top-level git rule, but `git branch -l` descends into the `branch` sub-rule.
- `cd` with no arguments (bare `cd`) resolves to `$HOME`. It's only allowed if `HomeDir` is within `RootDir`.
