package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arnarg/crushout/internal/rules"
	"gopkg.in/yaml.v3"
)

// Config represents the .crushout.yml configuration.
type Config struct {
	OverwriteDefaults bool                   `yaml:"overwrite_defaults"`
	Rules             map[string]*RuleConfig `yaml:"rules"`
}

// RuleConfig is a rule as specified in YAML.
type RuleConfig struct {
	Default     *bool                  `yaml:"default"`
	DenyFlags   []string               `yaml:"deny_flags"`
	Subcommands map[string]*RuleConfig `yaml:"subcommands"`
}

// Load reads .crushout.yml or .crushout.yaml from dir.
// Returns (nil, nil) if no config file exists.
// Returns an error if the file exists but is malformed.
func Load(dir string) (*Config, error) {
	for _, name := range []string{".crushout.yml", ".crushout.yaml"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		return &cfg, nil
	}

	return nil, nil
}

// ToRules converts a map of RuleConfig to the runtime rules map.
func ToRules(cfg map[string]*RuleConfig) map[string]*rules.Rule {
	result := make(map[string]*rules.Rule, len(cfg))
	for name, rc := range cfg {
		result[name] = rc.toRule()
	}
	return result
}

// ToRulesWithDefaults merges cfg with rules.Default.
// If cfg.OverwriteDefaults is true, cfg.Rules replaces the defaults entirely.
// Otherwise, cfg.Rules is deep-merged over rules.Default with user values winning.
func ToRulesWithDefaults(cfg *Config) map[string]*rules.Rule {
	if cfg == nil {
		return rules.Default
	}
	if cfg.OverwriteDefaults {
		return ToRules(cfg.Rules)
	}
	userRules := ToRules(cfg.Rules)
	return Merge(rules.Default, userRules)
}

// Merge deep-merges userRules over baseRules.
// User-provided rules win at every level.
func Merge(baseRules, userRules map[string]*rules.Rule) map[string]*rules.Rule {
	result := deepCopyRules(baseRules)
	for name, userRule := range userRules {
		if baseRule, ok := result[name]; ok {
			result[name] = mergeRule(baseRule, userRule)
		} else {
			result[name] = userRule
		}
	}
	return result
}

// toRule converts a RuleConfig to the runtime Rule type.
func (rc *RuleConfig) toRule() *rules.Rule {
	r := &rules.Rule{}
	if rc.Default != nil {
		r.Default = *rc.Default
		r.DefaultExplicit = true
	}
	r.DenyFlags = rc.DenyFlags
	if len(rc.Subcommands) > 0 {
		r.Subcommands = make(map[string]*rules.Rule, len(rc.Subcommands))
		for name, sub := range rc.Subcommands {
			r.Subcommands[name] = sub.toRule()
		}
	}
	return r
}

// mergeRule merges userRule into baseRule.
// User values win: Default (if explicitly set), DenyFlags, and Subcommands are merged recursively.
func mergeRule(base, user *rules.Rule) *rules.Rule {
	result := &rules.Rule{
		Default:   base.Default,
		DenyFlags: base.DenyFlags,
	}
	if base.Subcommands != nil {
		result.Subcommands = deepCopyRules(base.Subcommands)
	}

	// Only override Default if user explicitly set it (DefaultExplicit=true)
	if user.DefaultExplicit {
		result.Default = user.Default
	}
	if len(user.DenyFlags) > 0 {
		result.DenyFlags = user.DenyFlags
	}

	if len(user.Subcommands) > 0 {
		if result.Subcommands == nil {
			result.Subcommands = make(map[string]*rules.Rule)
		}
		for name, userSub := range user.Subcommands {
			if baseSub, ok := result.Subcommands[name]; ok {
				result.Subcommands[name] = mergeRule(baseSub, userSub)
			} else {
				result.Subcommands[name] = userSub
			}
		}
	}

	return result
}

// deepCopyRules returns a deep copy of a rule map.
func deepCopyRules(src map[string]*rules.Rule) map[string]*rules.Rule {
	if src == nil {
		return nil
	}
	dst := make(map[string]*rules.Rule, len(src))
	for k, v := range src {
		dst[k] = deepCopyRule(v)
	}
	return dst
}

// deepCopyRule returns a deep copy of a single rule.
func deepCopyRule(src *rules.Rule) *rules.Rule {
	if src == nil {
		return nil
	}
	dst := &rules.Rule{
		Default:         src.Default,
		DefaultExplicit: src.DefaultExplicit,
		DenyFlags:       append([]string(nil), src.DenyFlags...),
	}
	if len(src.Subcommands) > 0 {
		dst.Subcommands = deepCopyRules(src.Subcommands)
	}
	return dst
}
