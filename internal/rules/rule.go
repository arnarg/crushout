package rules

import (
	"fmt"
	"strings"
)

type Decision int

const (
	NoOpinion Decision = iota
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
		return NoOpinion, nil
	default:
		return NoOpinion, fmt.Errorf("invalid decision %q (use allow, deny, or prompt)", s)
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
	DenyFlags       []string         `yaml:"deny_flags,omitempty"`
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
			for _, flag := range current.DenyFlags {
				if arg == flag || strings.HasPrefix(arg, flag+"=") {
					return NoOpinion, "denied flag"
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
