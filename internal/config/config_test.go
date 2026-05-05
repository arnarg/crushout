package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arnarg/crushout/internal/rules"
)

func decisionPtr(d rules.Decision) *Decision {
	conv := Decision(d)
	return &conv
}

func newRuleWithDecision(d rules.Decision) *rules.Rule {
	return &rules.Rule{Default: d, DefaultExplicit: true}
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
    decision: prompt
    subcommands:
      build:
        decision: allow
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
	if nix.Decision == nil || *nix.Decision != Decision(rules.NoOpinion) {
		t.Error("expected nix.decision to be prompt (NoOpinion)")
	}
	build, ok := nix.Subcommands["build"]
	if !ok {
		t.Fatal("expected build subcommand")
	}
	if build.Decision == nil || *build.Decision != Decision(rules.Allow) {
		t.Fatal("expected build.decision to be allow")
	}
}

func TestLoad_yamlShorthand(t *testing.T) {
	content := `rules:
  ls: allow
  rm: deny
  kubectl:
    decision: prompt
    subcommands:
      get: allow
      exec: deny
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

	ls := cfg.Rules["ls"]
	if ls == nil || ls.Decision == nil || *ls.Decision != Decision(rules.Allow) {
		t.Error("expected ls to be allow via shorthand")
	}

	rm := cfg.Rules["rm"]
	if rm == nil || rm.Decision == nil || *rm.Decision != Decision(rules.Deny) {
		t.Error("expected rm to be deny via shorthand")
	}

	kubectl := cfg.Rules["kubectl"]
	if kubectl == nil || kubectl.Decision == nil || *kubectl.Decision != Decision(rules.NoOpinion) {
		t.Error("expected kubectl to be prompt")
	}

	get := kubectl.Subcommands["get"]
	if get == nil || get.Decision == nil || *get.Decision != Decision(rules.Allow) {
		t.Error("expected kubectl.get to be allow via shorthand")
	}

	exec := kubectl.Subcommands["exec"]
	if exec == nil || exec.Decision == nil || *exec.Decision != Decision(rules.Deny) {
		t.Error("expected kubectl.exec to be deny via shorthand")
	}
}

func TestLoad_yamlPriority(t *testing.T) {
	yml := `rules: {}`
	yaml := `rules:
  foo: allow`

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

func TestLoad_invalidDecision(t *testing.T) {
	dir := t.TempDir()
	content := `rules:
  ls: maybe`
	if err := os.WriteFile(filepath.Join(dir, ".crushout.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for invalid decision value")
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
			Decision: decisionPtr(rules.NoOpinion),
			Subcommands: map[string]*RuleConfig{
				"build": {Decision: decisionPtr(rules.Allow)},
			},
		},
	}

	rs := ToRules(cfg)

	nix, ok := rs["nix"]
	if !ok {
		t.Fatal("expected nix rule")
	}
	if nix.Default != rules.NoOpinion {
		t.Error("expected nix.default to be NoOpinion")
	}
	build, ok := nix.Subcommands["build"]
	if !ok {
		t.Fatal("expected build subcommand")
	}
	if build.Default != rules.Allow {
		t.Error("expected build.default to be Allow")
	}
}

func TestToRules_nilDecision(t *testing.T) {
	cfg := map[string]*RuleConfig{
		"ls": {}, // no decision, no subcommands
	}

	result := ToRules(cfg)

	ls, ok := result["ls"]
	if !ok {
		t.Fatal("expected ls rule")
	}
	// Default should be NoOpinion (zero value)
	if ls.Default != rules.NoOpinion {
		t.Error("expected ls.default to be NoOpinion for nil decision")
	}
}

func TestToRules_denyWithMessage(t *testing.T) {
	cfg := map[string]*RuleConfig{
		"curl": {
			Decision: decisionPtr(rules.Deny),
			Message:  "curl is not allowed",
		},
	}

	rs := ToRules(cfg)

	curl := rs["curl"]
	if curl.Default != rules.Deny {
		t.Error("expected curl to be denied")
	}
	if curl.Message != "curl is not allowed" {
		t.Errorf("expected message, got %q", curl.Message)
	}
}

func TestMerge_addNewRule(t *testing.T) {
	base := map[string]*rules.Rule{
		"ls": {Default: rules.Allow},
	}
	user := map[string]*rules.Rule{
		"nix": {},
	}

	result := Merge(base, user)

	if _, ok := result["ls"]; !ok {
		t.Error("ls should be preserved from base")
	}
	nix, ok := result["nix"]
	if !ok {
		t.Fatal("nix should be added from user")
	}
	if nix.Default != rules.NoOpinion {
		t.Error("nix.default should be NoOpinion")
	}
}

func TestMerge_overrideDefault(t *testing.T) {
	base := map[string]*rules.Rule{
		"ls": {Default: rules.Allow},
	}
	user := map[string]*rules.Rule{
		"ls": {Default: rules.NoOpinion, DefaultExplicit: true},
	}

	result := Merge(base, user)

	if result["ls"].Default != rules.NoOpinion {
		t.Error("user should win: ls.default should be NoOpinion")
	}
}

func TestMerge_deepMergeSubcommands(t *testing.T) {
	base := map[string]*rules.Rule{
		"git": {
			Subcommands: map[string]*rules.Rule{
				"status": {Default: rules.Allow},
				"push":   {},
			},
		},
	}
	user := map[string]*rules.Rule{
		"git": {
			Subcommands: map[string]*rules.Rule{
				"status": newRuleWithDecision(rules.NoOpinion),
				"fetch":  {},
			},
		},
	}

	result := Merge(base, user)

	git := result["git"]
	if git.Default != rules.NoOpinion {
		t.Error("git.default should remain NoOpinion")
	}
	if _, ok := git.Subcommands["push"]; !ok {
		t.Error("git.push should be preserved from base")
	}
	if git.Subcommands["status"].Default != rules.NoOpinion {
		t.Error("git.status should be overridden to NoOpinion")
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
						"show": {Default: rules.Allow},
						"add":  {},
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
						"show": newRuleWithDecision(rules.NoOpinion),
					},
				},
			},
		},
	}

	result := Merge(base, user)

	remote := result["git"].Subcommands["remote"]
	if remote.Subcommands["show"].Default != rules.NoOpinion {
		t.Error("git.remote.show should be overridden to NoOpinion")
	}
	if remote.Subcommands["add"].Default != rules.NoOpinion {
		t.Error("git.remote.add should be preserved from base")
	}
}

func TestMerge_denyFlagsOverride(t *testing.T) {
	base := map[string]*rules.Rule{
		"find": {
			Default:   rules.Allow,
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
			Default:   rules.Allow,
			DenyFlags: []string{"-i"},
		},
	}
	user := map[string]*rules.Rule{
		"sed": {
			Default: rules.Allow,
			// DenyFlags not specified, should preserve base
		},
	}

	result := Merge(base, user)

	sed := result["sed"]
	if len(sed.DenyFlags) != 1 || sed.DenyFlags[0] != "-i" {
		t.Errorf("expected [-i], got %v", sed.DenyFlags)
	}
}

func TestMerge_messageOverride(t *testing.T) {
	base := map[string]*rules.Rule{
		"curl": {
			Default: rules.Deny,
			Message: "base message",
		},
	}
	user := map[string]*rules.Rule{
		"curl": {
			Message: "user message",
		},
	}

	result := Merge(base, user)

	curl := result["curl"]
	if curl.Message != "user message" {
		t.Errorf("expected user message, got %q", curl.Message)
	}
}

func TestMerge_preserveMessageWhenUserHasNone(t *testing.T) {
	base := map[string]*rules.Rule{
		"curl": {
			Default: rules.Deny,
			Message: "base message",
		},
	}
	user := map[string]*rules.Rule{
		"curl": {
			DefaultExplicit: true,
		},
	}

	result := Merge(base, user)

	curl := result["curl"]
	if curl.Message != "base message" {
		t.Errorf("expected base message preserved, got %q", curl.Message)
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
			"nix": {Decision: decisionPtr(rules.NoOpinion)},
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
	if nix.Default != rules.NoOpinion {
		t.Error("nix.default should be NoOpinion")
	}
}

func TestToRulesWithDefaults_merge(t *testing.T) {
	cfg := &Config{
		OverwriteDefaults: false,
		Rules: map[string]*RuleConfig{
			"nix": {Decision: decisionPtr(rules.NoOpinion)},
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
	if nix.Default != rules.NoOpinion {
		t.Error("nix.default should be NoOpinion")
	}
}

func TestDeepCopyRule(t *testing.T) {
	orig := &rules.Rule{
		Default:   rules.Allow,
		DenyFlags: []string{"-i"},
		Message:   "test message",
		Subcommands: map[string]*rules.Rule{
			"sub": {},
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
	if copy_.Default != rules.Allow || copy_.DenyFlags[0] != "-i" || copy_.Message != "test message" || copy_.Subcommands["sub"].Default != rules.NoOpinion {
		t.Error("copied values should match original")
	}
}

func TestDeepCopyRules_nil(t *testing.T) {
	result := deepCopyRules(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}
}
