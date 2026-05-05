package checker

import (
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

// IsReadOnly returns true only if every command is confidently non-mutable.
// Returns false for anything ambiguous, unknown, or mutable.
func (c *Checker) IsReadOnly(input string) (bool, error) {
	result, err := bash.Parse(input)
	if err != nil {
		return false, nil
	}

	if result.HasError || result.IsComplex || result.HasRedirect || len(result.Commands) == 0 {
		return false, nil
	}

	cwd := c.RootDir
	for _, cmd := range result.Commands {
		if cmd.Name == "cd" {
			if !c.isSafeCD(cmd, &cwd) {
				return false, nil
			}
			continue
		}
		if !c.isReadOnly(cmd) {
			return false, nil
		}
	}

	return true, nil
}

func (c *Checker) isReadOnly(cmd bash.Command) bool {
	name := cmd.Name
	if strings.Contains(name, "/") {
		name = filepath.Base(name)
	}
	if strings.Contains(name, "$") || strings.Contains(name, "`") {
		return false
	}

	rule, exists := c.Rules[name]
	if !exists {
		return false
	}
	return rule.Allow(cmd.Args)
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
