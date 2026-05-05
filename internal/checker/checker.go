package checker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arnarg/crushout/internal/bash"
	"github.com/arnarg/crushout/internal/rules"
)

type Checker struct {
	RootDir string
	HomeDir string
	Rules   map[string]*rules.Rule
}

// Check evaluates the input command string and returns a Decision.
// Allow means auto-approve, Deny means hard block, NoOpinion means
// let the normal permission prompt handle it.
func (c *Checker) Check(input string) (rules.Decision, string, error) {
	result, err := bash.Parse(input)
	if err != nil {
		return rules.NoOpinion, "", nil
	}

	if result.HasError || result.IsComplex || result.HasRedirect || len(result.Commands) == 0 {
		return rules.NoOpinion, "", nil
	}

	final := rules.Allow
	cwd := c.RootDir
	for _, cmd := range result.Commands {
		if cmd.Name == "cd" {
			if !c.isSafeCD(cmd, &cwd) {
				return rules.NoOpinion, "", nil
			}
			continue
		}

		d, msg := c.checkCommand(cmd)
		switch d {
		case rules.Deny:
			return rules.Deny, msg, nil
		case rules.NoOpinion:
			final = rules.NoOpinion
		}
	}

	return final, "", nil
}

func (c *Checker) checkCommand(cmd bash.Command) (rules.Decision, string) {
	name := cmd.Name
	if strings.Contains(name, "/") {
		name = filepath.Base(name)
	}
	if strings.Contains(name, "$") || strings.Contains(name, "`") {
		return rules.NoOpinion, ""
	}

	rule, exists := c.Rules[name]
	if !exists {
		return rules.NoOpinion, ""
	}

	d, msg := rule.Resolve(cmd.Args)
	if d == rules.Deny && msg != "" {
		return d, fmt.Sprintf("crushout: %s (rule for '%s')", msg, name)
	}
	return d, msg
}

func (c *Checker) isSafeCD(cmd bash.Command, cwd *string) bool {
	dirArg := firstNonFlagArg(cmd.Args)
	if dirArg == "" {
		if c.HomeDir != "" && isWithinRoot(c.RootDir, c.HomeDir) {
			*cwd = c.HomeDir
			return true
		}
		return false
	}

	resolved := c.resolvePath(dirArg, *cwd)
	if resolved == "" {
		return false
	}

	if !isWithinRoot(c.RootDir, resolved) {
		return false
	}

	*cwd = resolved
	return true
}

// ── Path helpers ───────────────────────────────────────────

func firstNonFlagArg(args []string) string {
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

func isWithinRoot(root, target string) bool {
	cleanRoot := filepath.Clean(root)
	cleanTarget := filepath.Clean(target)
	rel, err := filepath.Rel(cleanRoot, cleanTarget)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

func (c *Checker) resolvePath(arg, cwd string) string {
	switch {
	case strings.Contains(arg, "$"):
		return ""
	case arg == "-":
		return ""
	case arg == "~" || strings.HasPrefix(arg, "~/"):
		if c.HomeDir == "" {
			return ""
		}
		if arg == "~" {
			return c.HomeDir
		}
		return filepath.Join(c.HomeDir, strings.TrimPrefix(arg, "~/"))
	case filepath.IsAbs(arg):
		return filepath.Clean(arg)
	default:
		return filepath.Clean(filepath.Join(cwd, arg))
	}
}
