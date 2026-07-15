package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arnarg/crushout/internal/rules"
	"gopkg.in/yaml.v3"
)

// Config is the fully resolved configuration returned to callers.
// All scalar fields have been resolved across layers and Rules has been
// merged with the built-in defaults.
type Config struct {
	RtkRewrite bool
	Rules      map[string]*rules.Rule
}

// Default returns a config with default values.
func Default() *Config {
	return &Config{
		RtkRewrite: true,
		Rules:      deepCopyRules(rules.Default),
	}
}

// repoConfigNames are the filenames searched for in a project root.
var repoConfigNames = []string{".crushout.yml", ".crushout.yaml"}

// globalConfigNames are the filenames searched for in the user config dir.
var globalConfigNames = []string{"crushout.yml", "crushout.yaml"}

// ConfigFile is the raw YAML representation of a single config file.
// It is used while merging layers: scalar fields are *bool so that an
// unset value can be distinguished from an explicit one.
type ConfigFile struct {
	RtkRewrite        *bool                  `yaml:"rtk_rewrite"`
	OverwriteDefaults *bool                  `yaml:"overwrite_defaults"`
	Rules             map[string]*RuleConfig `yaml:"rules"`
}

// RuleConfig is a rule as specified in YAML.
type RuleConfig struct {
	Decision    *Decision              `yaml:"decision"`
	PromptFlags []string               `yaml:"prompt_flags"`
	AllowFlags  []string               `yaml:"allow_flags"`
	Message     string                 `yaml:"message"`
	Subcommands map[string]*RuleConfig `yaml:"subcommands"`
}

// Decision is a YAML-friendly wrapper around rules.Decision.
type Decision rules.Decision

func (d *Decision) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := rules.ParseDecision(value.Value)
	if err != nil {
		return err
	}
	*d = Decision(parsed)
	return nil
}

// UnmarshalYAML implements custom YAML unmarshalling for RuleConfig.
// Supports both shorthand (string like "allow") and full mapping form.
func (rc *RuleConfig) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		d, err := rules.ParseDecision(value.Value)
		if err != nil {
			return fmt.Errorf("invalid rule: %w", err)
		}
		conv := Decision(d)
		rc.Decision = &conv
		return nil

	case yaml.MappingNode:
		type plain RuleConfig
		return value.Decode((*plain)(rc))

	default:
		return fmt.Errorf("rule must be a string or mapping, got %s", value.Tag)
	}
}

// loadFirst reads the first existing config file from dir, trying names in
// order. Returns (nil, nil) if no file exists.
// Returns an error if a file exists but is malformed.
func loadFirst(dir string, names []string) (*ConfigFile, error) {
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var cf ConfigFile
		if err := yaml.Unmarshal(data, &cf); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		return &cf, nil
	}

	return nil, nil
}

// load reads global and repo config files, merges them, and returns the
// resolved Config. Either directory may be empty (skipped).
func load(globalDir, repoDir string) (*Config, error) {
	var (
		global, repo *ConfigFile
		err          error
	)

	if globalDir != "" {
		global, err = loadFirst(globalDir, globalConfigNames)
		if err != nil {
			return nil, err
		}
	}

	repo, err = loadFirst(repoDir, repoConfigNames)
	if err != nil {
		return nil, err
	}

	return &Config{
		RtkRewrite: resolveRtkRewrite(global, repo),
		Rules:      buildRules(global, repo),
	}, nil
}

// Load resolves configuration from two layers, merged with the repo config
// taking precedence over the user (global) config:
//
//   - global: $XDG_CONFIG_HOME/crushout/crushout.{yml,yaml}
//   - repo:   <rootDir>/.crushout.{yml,yaml}
//
// It returns a fully resolved *Config (always non-nil on success) with Rules
// already merged with the built-in defaults. A malformed file in either layer
// causes an error.
func Load(rootDir string) (*Config, error) {
	return load(userConfigDir(), rootDir)
}

// userConfigDir returns the global config directory following the XDG base
// directory specification via os.UserConfigDir (which falls back to
// $HOME/.config). Returns "" if it cannot be determined.
func userConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "crushout")
}

// resolveRtkRewrite resolves rtk_rewrite across layers: repo wins, then global,
// then defaults to true.
func resolveRtkRewrite(global, repo *ConfigFile) bool {
	if repo != nil && repo.RtkRewrite != nil {
		return *repo.RtkRewrite
	}
	if global != nil && global.RtkRewrite != nil {
		return *global.RtkRewrite
	}
	return true
}

// buildRules builds the final rule set by applying each layer in sequence over
// the built-in defaults. Within a layer, overwrite_defaults: true replaces the
// effective base entirely; otherwise the layer is deep-merged over it.
//
// Layer order (each builds on the previous): defaults -> global -> repo.
func buildRules(global, repo *ConfigFile) map[string]*rules.Rule {
	base := deepCopyRules(rules.Default)
	if global != nil {
		base = applyLayer(base, global)
	}
	if repo != nil {
		base = applyLayer(base, repo)
	}
	return base
}

// applyLayer applies a single config layer over base.
func applyLayer(base map[string]*rules.Rule, cf *ConfigFile) map[string]*rules.Rule {
	if cf.OverwriteDefaults != nil && *cf.OverwriteDefaults {
		return ToRules(cf.Rules)
	}
	return Merge(base, ToRules(cf.Rules))
}

// ToRules converts a map of RuleConfig to the runtime rules map.
func ToRules(cfg map[string]*RuleConfig) map[string]*rules.Rule {
	result := make(map[string]*rules.Rule, len(cfg))
	for name, rc := range cfg {
		result[name] = rc.toRule()
	}
	return result
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
	if rc.Decision != nil {
		r.Default = rules.Decision(*rc.Decision)
		r.DefaultExplicit = true
	}
	r.PromptFlags = rc.PromptFlags
	r.AllowFlags = rc.AllowFlags
	r.Message = rc.Message
	if len(rc.Subcommands) > 0 {
		r.Subcommands = make(map[string]*rules.Rule, len(rc.Subcommands))
		for name, sub := range rc.Subcommands {
			r.Subcommands[name] = sub.toRule()
		}
	}
	return r
}

// mergeRule merges userRule into baseRule.
// User values win: Default (if explicitly set), PromptFlags, AllowFlags, Message,
// and Subcommands are merged recursively.
func mergeRule(base, user *rules.Rule) *rules.Rule {
	result := &rules.Rule{
		Default:     base.Default,
		PromptFlags: base.PromptFlags,
		AllowFlags:  base.AllowFlags,
		Message:     base.Message,
	}
	if base.Subcommands != nil {
		result.Subcommands = deepCopyRules(base.Subcommands)
	}

	if user.DefaultExplicit {
		result.Default = user.Default
	}
	if len(user.PromptFlags) > 0 {
		result.PromptFlags = user.PromptFlags
	}
	if len(user.AllowFlags) > 0 {
		result.AllowFlags = user.AllowFlags
	}
	if user.Message != "" {
		result.Message = user.Message
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
		PromptFlags:     append([]string(nil), src.PromptFlags...),
		AllowFlags:      append([]string(nil), src.AllowFlags...),
		Message:         src.Message,
	}
	if len(src.Subcommands) > 0 {
		dst.Subcommands = deepCopyRules(src.Subcommands)
	}
	return dst
}
