package rules

import "strings"

// Rule is a recursive allow/deny rule.
type Rule struct {
	Subcommands map[string]*Rule `json:"subcommands,omitempty"`
	DenyFlags   []string         `json:"deny_flags,omitempty"`
	Default     bool             `json:"default"`
}

// Allow walks args as a subcommand chain, checking DenyFlags at each level.
func (r *Rule) Allow(args []string) bool {
	ok, _ := r.resolve(args)
	return ok
}

func (r *Rule) resolve(args []string) (bool, string) {
	current := r

	remaining := args
	for len(remaining) > 0 {
		for _, arg := range remaining {
			for _, flag := range current.DenyFlags {
				if arg == flag || strings.HasPrefix(arg, flag+"=") {
					return false, "denied flag"
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

	return current.Default, ""
}

