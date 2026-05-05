package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arnarg/crushout/internal/checker"
	"github.com/arnarg/crushout/internal/config"
	"github.com/arnarg/crushout/internal/hook"
	"github.com/arnarg/crushout/internal/rules"
)

func main() {
	// Read hook input from stdin
	var input hook.Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "crushout: %v\n", err)
		os.Exit(1)
	}

	// If tool is not bash or command is empty we just skip it
	if input.ToolName != "bash" || input.ToolInput.Command == "" {
		fmt.Println("{}")
		return
	}

	// We want to extract the CWD so we can only auto-allow
	// whitelisted commands that also stay within the CWD (track cd)
	rootDir := input.CWD
	if rootDir == "" {
		rootDir, _ = os.Getwd()
	}
	homeDir, _ := os.UserHomeDir()

	// Load custom config if present
	ruleSet := rules.Default
	cfg, err := config.Load(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "crushout: %v\n", err)
	} else if cfg != nil {
		ruleSet = config.ToRulesWithDefaults(cfg)
	}

	c := &checker.Checker{
		RootDir: rootDir,
		HomeDir: homeDir,
		Rules:   ruleSet,
	}

	switch d, reason, _ := c.Check(input.ToolInput.Command); d {
	case rules.Allow:
		json.NewEncoder(os.Stdout).Encode(hook.Output{
			Version:  1,
			Decision: "allow",
		})
	case rules.Deny:
		json.NewEncoder(os.Stdout).Encode(hook.Output{
			Version:  1,
			Decision: "deny",
			Reason:   reason,
		})
	default:
		fmt.Println("{}")
	}
}
