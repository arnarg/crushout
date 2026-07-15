package rules

import (
	"fmt"
	"strings"
)

type Decision int

const (
	Prompt Decision = iota
	Allow
	Deny
)

func ParseDecision(s string) (Decision, error) {
	switch s {
	case "allow":
		return Allow, nil
	case "deny":
		return Deny, nil
	case "prompt":
		return Prompt, nil
	default:
		return Prompt, fmt.Errorf("invalid decision %q (use allow, deny, or prompt)", s)
	}
}

func (d Decision) String() string {
	switch d {
	case Allow:
		return "allow"
	case Deny:
		return "deny"
	default:
		return "prompt"
	}
}

type Rule struct {
	Subcommands     map[string]*Rule `yaml:"subcommands,omitempty"`
	PromptFlags     []string         `yaml:"prompt_flags,omitempty"`
	AllowFlags      []string         `yaml:"allow_flags,omitempty"`
	Default         Decision         `yaml:"default"`
	DefaultExplicit bool             `yaml:"-"`
	Message         string           `yaml:"message,omitempty"`
}

func (r *Rule) Allow(args []string) bool {
	d, _ := r.Resolve(args)
	return d == Allow
}

func (r *Rule) Resolve(args []string) (Decision, string) {
	current := r

	remaining := args
	for len(remaining) > 0 {
		for _, arg := range remaining {
			for _, flag := range current.PromptFlags {
				if arg == flag || strings.HasPrefix(arg, flag+"=") {
					return Prompt, "denied flag"
				}
			}
			if len(current.AllowFlags) > 0 && strings.HasPrefix(arg, "-") {
				allowed := false
				for _, af := range current.AllowFlags {
					if arg == af || strings.HasPrefix(arg, af+"=") {
						allowed = true
						break
					}
				}
				if !allowed {
					return Prompt, "unrecognized flag"
				}
			}
		}

		if current.Subcommands == nil {
			break
		}
		next, found := current.Subcommands[remaining[0]]
		if !found {
			break
		}
		current = next
		remaining = remaining[1:]
	}

	msg := ""
	if current.Default == Deny {
		msg = current.denyMessage()
	}
	return current.Default, msg
}

func (r *Rule) denyMessage() string {
	if r.Message != "" {
		return r.Message
	}
	return "denied by rule"
}
