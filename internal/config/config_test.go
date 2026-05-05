package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arnarg/crushout/internal/rules"
)

func boolPtr(b bool) *bool {
	return &b
}

func newRuleWithDefault(val bool) *rules.Rule {
	return &rules.Rule{Default: val, DefaultExplicit: true}
}

func TestLoad_notFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config, got %+v", cfg)
	}
}

func TestLoad_yaml(t *testing.T) {
	content := `overwrite_defaults: true
rules:
  nix:
    default: false
    subcommands:
      build:
        default: true
`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".crushout.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if !cfg.OverwriteDefaults {
		t.Error("expected OverwriteDefaults to be true")
	}
	nix, ok := cfg.Rules["nix"]
	if !ok {
		t.Fatal("expected nix rule")
	}
	if nix.Default == nil || *nix.Default {
		t.Error("expected nix.default to be false")
	}
	build, ok := nix.Subcommands["build"]
	if !ok {
		t.Fatal("expected build subcommand")
	}
	if build.Default == nil || !*build.Default {
		t.Fatal("expected build.default to be true")
	}
}

func TestLoad_yamlPriority(t *testing.T) {
	yml := `rules: {}`
	yaml := `rules:
  foo: {default: true}`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".crushout.yml"), []byte(yml), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".crushout.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	// .crushout.yml takes precedence
	if _, ok := cfg.Rules["foo"]; ok {
		t.Error(".crushout.yml should take precedence over .crushout.yaml")
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".crushout.yml"), []byte("invalid: [yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestToRules_empty(t *testing.T) {
	result := ToRules(nil)
	if result == nil {
		t.Fatal("expected non-nil map for nil input")
	}
	result = ToRules(map[string]*RuleConfig{})
	if result == nil {
		t.Fatal("expected non-nil map for empty input")
	}
}

func TestToRules_basic(t *testing.T) {
	cfg := map[string]*RuleConfig{
		"nix": {
			Default: boolPtr(false),
			Subcommands: map[string]*RuleConfig{
				"build": {Default: boolPtr(true)},
			},
		},
	}

	rules := ToRules(cfg)

	nix, ok := rules["nix"]
	if !ok {
		t.Fatal("expected nix rule")
	}
	if nix.Default {
		t.Error("expected nix.default to be false")
	}
	build, ok := nix.Subcommands["build"]
	if !ok {
		t.Fatal("expected build subcommand")
	}
	if !build.Default {
		t.Error("expected build.default to be true")
	}
}

func TestToRules_nilDefault(t *testing.T) {
	cfg := map[string]*RuleConfig{
		"ls": {}, // no default, no subcommands
	}

	result := ToRules(cfg)

	ls, ok := result["ls"]
	if !ok {
		t.Fatal("expected ls rule")
	}
	// Default should be false (zero value of bool)
	if ls.Default {
		t.Error("expected ls.default to be false for nil *bool")
	}
}

func TestMerge_addNewRule(t *testing.T) {
	base := map[string]*rules.Rule{
		"ls": {Default: true},
	}
	user := map[string]*rules.Rule{
		"nix": {Default: false},
	}

	result := Merge(base, user)

	if _, ok := result["ls"]; !ok {
		t.Error("ls should be preserved from base")
	}
	nix, ok := result["nix"]
	if !ok {
		t.Fatal("nix should be added from user")
	}
	if nix.Default {
		t.Error("nix.default should be false")
	}
}

func TestMerge_overrideDefault(t *testing.T) {
	base := map[string]*rules.Rule{
		"ls": {Default: true},
	}
	user := map[string]*rules.Rule{
		"ls": {Default: false},
	}

	result := Merge(base, user)

	if !result["ls"].Default {
		t.Error("user should win: ls.default should be false")
	}
}

func TestMerge_deepMergeSubcommands(t *testing.T) {
	base := map[string]*rules.Rule{
		"git": {
			Default: false,
			Subcommands: map[string]*rules.Rule{
				"status": {Default: true},
				"push":   {Default: false},
			},
		},
	}
	user := map[string]*rules.Rule{
		"git": {
			Subcommands: map[string]*rules.Rule{
				"status": newRuleWithDefault(false),
				"fetch":  {Default: false},
			},
		},
	}

	result := Merge(base, user)

	git := result["git"]
	if git.Default {
		t.Error("git.default should remain false")
	}
	if _, ok := git.Subcommands["push"]; !ok {
		t.Error("git.push should be preserved from base")
	}
	if git.Subcommands["status"].Default {
		t.Error("git.status should be overridden to false")
	}
	if _, ok := git.Subcommands["fetch"]; !ok {
		t.Error("git.fetch should be added from user")
	}
}

func TestMerge_deepMergeNestedSubcommands(t *testing.T) {
	base := map[string]*rules.Rule{
		"git": {
			Subcommands: map[string]*rules.Rule{
				"remote": {
					Subcommands: map[string]*rules.Rule{
						"show": {Default: true},
						"add":  {Default: false},
					},
				},
			},
		},
	}
	user := map[string]*rules.Rule{
		"git": {
			Subcommands: map[string]*rules.Rule{
				"remote": {
					Subcommands: map[string]*rules.Rule{
						"show": newRuleWithDefault(false),
					},
				},
			},
		},
	}

	result := Merge(base, user)

	remote := result["git"].Subcommands["remote"]
	if remote.Subcommands["show"].Default {
		t.Error("git.remote.show should be overridden to false")
	}
	if remote.Subcommands["add"].Default {
		t.Error("git.remote.add should be preserved from base")
	}
}

func TestMerge_denyFlagsOverride(t *testing.T) {
	base := map[string]*rules.Rule{
		"find": {
			Default:   true,
			DenyFlags: []string{"-exec", "-delete"},
		},
	}
	user := map[string]*rules.Rule{
		"find": {
			DenyFlags: []string{"-exec", "-fprint"},
		},
	}

	result := Merge(base, user)

	find := result["find"]
	if len(find.DenyFlags) != 2 {
		t.Fatalf("expected 2 deny flags, got %d", len(find.DenyFlags))
	}
	if find.DenyFlags[0] != "-exec" || find.DenyFlags[1] != "-fprint" {
		t.Errorf("expected [-exec -fprint], got %v", find.DenyFlags)
	}
}

func TestMerge_preserveDenyFlagsWhenUserHasNone(t *testing.T) {
	base := map[string]*rules.Rule{
		"sed": {
			Default:   true,
			DenyFlags: []string{"-i"},
		},
	}
	user := map[string]*rules.Rule{
		"sed": {
			Default: true,
			// DenyFlags not specified, should preserve base
		},
	}

	result := Merge(base, user)

	sed := result["sed"]
	if len(sed.DenyFlags) != 1 || sed.DenyFlags[0] != "-i" {
		t.Errorf("expected [-i], got %v", sed.DenyFlags)
	}
}

func TestToRulesWithDefaults_noConfig(t *testing.T) {
	result := ToRulesWithDefaults(nil)
	if result == nil {
		t.Fatal("expected non-nil result for nil config")
	}
	if len(result) == 0 {
		t.Error("nil config should return non-empty rules")
	}
}

func TestToRulesWithDefaults_overwriteTrue(t *testing.T) {
	cfg := &Config{
		OverwriteDefaults: true,
		Rules: map[string]*RuleConfig{
			"nix": {Default: boolPtr(false)},
		},
	}

	result := ToRulesWithDefaults(cfg)

	if _, ok := result["ls"]; ok {
		t.Error("overwrite=true should drop default rules like ls")
	}
	nix, ok := result["nix"]
	if !ok {
		t.Fatal("nix should be present")
	}
	if nix.Default {
		t.Error("nix.default should be false")
	}
}

func TestToRulesWithDefaults_merge(t *testing.T) {
	cfg := &Config{
		OverwriteDefaults: false,
		Rules: map[string]*RuleConfig{
			"nix": {Default: boolPtr(false)},
		},
	}

	result := ToRulesWithDefaults(cfg)

	if _, ok := result["ls"]; !ok {
		t.Error("ls should be preserved from defaults")
	}
	nix, ok := result["nix"]
	if !ok {
		t.Fatal("nix should be present")
	}
	if nix.Default {
		t.Error("nix.default should be false")
	}
}

func TestDeepCopyRule(t *testing.T) {
	orig := &rules.Rule{
		Default:   true,
		DenyFlags: []string{"-i"},
		Subcommands: map[string]*rules.Rule{
			"sub": {Default: false},
		},
	}

	copy_ := deepCopyRule(orig)

	// Verify it's a separate copy
	if copy_ == orig {
		t.Error("deepCopyRule should return a new instance")
	}
	if copy_.Subcommands["sub"] == orig.Subcommands["sub"] {
		t.Error("Subcommands map should be deep copied")
	}
	if len(copy_.DenyFlags) != len(orig.DenyFlags) || copy_.DenyFlags[0] != orig.DenyFlags[0] {
		t.Error("DenyFlags slice should be deep copied with correct values")
	}

	// Verify values are correct
	if !copy_.Default || copy_.DenyFlags[0] != "-i" || copy_.Subcommands["sub"].Default {
		t.Error("copied values should match original")
	}
}

func TestDeepCopyRules_nil(t *testing.T) {
	result := deepCopyRules(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}
}
