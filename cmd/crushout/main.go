package main

import (
	"fmt"
	"os"

	"github.com/arnarg/crushout/internal/checker"
	"github.com/arnarg/crushout/internal/config"
	"github.com/arnarg/crushout/internal/hook"
	"github.com/arnarg/crushout/internal/rewrite"
	"github.com/arnarg/crushout/internal/rules"
)

func main() {
	input, err := hook.Decode(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "crushout: %v\n", err)
		os.Exit(1)
	}

	// If tool is not bash we just skip it
	if !hook.IsBashTool(input) {
		out, err := input.FormatDecision(rules.NoOpinion, "", "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not serialize output: %v\n", err)
			os.Exit(1)
		}

		os.Stdout.Write(out)
		return
	}

	// We want to extract the CWD so we can only auto-allow
	// whitelisted commands that also stay within the CWD (track cd)
	rootDir := input.CWD()
	if rootDir == "" {
		rootDir, _ = os.Getwd()
	}
	homeDir, _ := os.UserHomeDir()

	// Load custom config if present
	ruleSet := rules.Default
	cfg, err := config.Load(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse user config: %v\n", err)
	} else if cfg != nil {
		ruleSet = config.ToRulesWithDefaults(cfg)
	}

	c := &checker.Checker{
		RootDir: rootDir,
		HomeDir: homeDir,
		Rules:   ruleSet,
	}

	d, reason, err := c.Check(input.Command())
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not check command: %v\n", err)
		os.Exit(1)
	}

	var rewritten string
	rtkEnabled := cfg == nil || cfg.RtkRewrite
	if d != rules.Deny && rtkEnabled {
		if rw, ok := rewrite.TryRtkRewrite(input.Command()); ok {
			rewritten = rw
		}
	}

	out, err := input.FormatDecision(d, reason, rewritten)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not serialize output: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(out)
}
