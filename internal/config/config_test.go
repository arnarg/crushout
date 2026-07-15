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

func decisionPtr(d rules.Decision) *Decision {
	conv := Decision(d)
	return &conv
}

func newRuleWithDecision(d rules.Decision) *rules.Rule {
	return &rules.Rule{Default: d, DefaultExplicit: true}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadFirst_notFound(t *testing.T) {
	cf, err := loadFirst("/nonexistent/path", repoConfigNames)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cf != nil {
		t.Fatalf("expected nil ConfigFile, got %+v", cf)
	}
}

func TestLoadFirst_yaml(t *testing.T) {
	content := `overwrite_defaults: true
rules:
  nix:
    decision: prompt
    subcommands:
      build:
        decision: allow
`

	dir := t.TempDir()
	writeFile(t, dir, ".crushout.yml", content)

	cf, err := loadFirst(dir, repoConfigNames)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cf == nil {
		t.Fatal("expected ConfigFile, got nil")
	}
	if cf.OverwriteDefaults == nil || !*cf.OverwriteDefaults {
		t.Error("expected OverwriteDefaults to be true")
	}
	nix, ok := cf.Rules["nix"]
	if !ok {
		t.Fatal("expected nix rule")
	}
	if nix.Decision == nil || *nix.Decision != Decision(rules.Prompt) {
		t.Error("expected nix.decision to be prompt (Prompt)")
	}
	build, ok := nix.Subcommands["build"]
	if !ok {
		t.Fatal("expected build subcommand")
	}
	if build.Decision == nil || *build.Decision != Decision(rules.Allow) {
		t.Fatal("expected build.decision to be allow")
	}
}

func TestLoadFirst_yamlShorthand(t *testing.T) {
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
	writeFile(t, dir, ".crushout.yml", content)

	cf, err := loadFirst(dir, repoConfigNames)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cf == nil {
		t.Fatal("expected ConfigFile, got nil")
	}

	ls := cf.Rules["ls"]
	if ls == nil || ls.Decision == nil || *ls.Decision != Decision(rules.Allow) {
		t.Error("expected ls to be allow via shorthand")
	}

	rm := cf.Rules["rm"]
	if rm == nil || rm.Decision == nil || *rm.Decision != Decision(rules.Deny) {
		t.Error("expected rm to be deny via shorthand")
	}

	kubectl := cf.Rules["kubectl"]
	if kubectl == nil || kubectl.Decision == nil || *kubectl.Decision != Decision(rules.Prompt) {
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

func TestLoadFirst_ymlPriority(t *testing.T) {
	yml := `rules: {}`
	yaml := `rules:
  foo: allow`

	dir := t.TempDir()
	writeFile(t, dir, ".crushout.yml", yml)
	writeFile(t, dir, ".crushout.yaml", yaml)

	cf, err := loadFirst(dir, repoConfigNames)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cf == nil {
		t.Fatal("expected ConfigFile, got nil")
	}
	// .crushout.yml takes precedence
	if _, ok := cf.Rules["foo"]; ok {
		t.Error(".crushout.yml should take precedence over .crushout.yaml")
	}
}

func TestLoadFirst_globalYmlPriority(t *testing.T) {
	yml := `rules: {}`
	yaml := `rules:
  foo: allow`

	dir := t.TempDir()
	writeFile(t, dir, "crushout.yml", yml)
	writeFile(t, dir, "crushout.yaml", yaml)

	cf, err := loadFirst(dir, globalConfigNames)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cf == nil {
		t.Fatal("expected ConfigFile, got nil")
	}
	// crushout.yml takes precedence over crushout.yaml
	if _, ok := cf.Rules["foo"]; ok {
		t.Error("crushout.yml should take precedence over crushout.yaml")
	}
}

func TestLoadFirst_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".crushout.yml", "invalid: [yaml")

	_, err := loadFirst(dir, repoConfigNames)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadFirst_invalidDecision(t *testing.T) {
	dir := t.TempDir()
	content := `rules:
  ls: maybe`
	writeFile(t, dir, ".crushout.yml", content)

	_, err := loadFirst(dir, repoConfigNames)
	if err == nil {
		t.Error("expected error for invalid decision value")
	}
}

func TestResolveRtkRewrite_defaults(t *testing.T) {
	if got := resolveRtkRewrite(nil, nil); got != true {
		t.Errorf("expected true, got %v", got)
	}
}

func TestResolveRtkRewrite_globalOnly(t *testing.T) {
	if got := resolveRtkRewrite(&ConfigFile{RtkRewrite: boolPtr(false)}, nil); got != false {
		t.Errorf("expected false, got %v", got)
	}
}

func TestResolveRtkRewrite_repoWins(t *testing.T) {
	got := resolveRtkRewrite(&ConfigFile{RtkRewrite: boolPtr(false)}, &ConfigFile{RtkRewrite: boolPtr(true)})
	if got != true {
		t.Errorf("expected repo to win (true), got %v", got)
	}
}

func TestResolveRtkRewrite_repoUnsetFallsBackToGlobal(t *testing.T) {
	got := resolveRtkRewrite(&ConfigFile{RtkRewrite: boolPtr(false)}, &ConfigFile{})
	if got != false {
		t.Errorf("expected global (false) when repo unset, got %v", got)
	}
}

func TestResolveRtkRewrite_bothUnset(t *testing.T) {
	if got := resolveRtkRewrite(&ConfigFile{}, &ConfigFile{}); got != true {
		t.Errorf("expected true when both unset, got %v", got)
	}
}

func TestApplyLayer_mergeMode(t *testing.T) {
	cf := &ConfigFile{
		OverwriteDefaults: boolPtr(false),
		Rules: map[string]*RuleConfig{
			"nix": {Decision: decisionPtr(rules.Prompt)},
		},
	}

	result := applyLayer(rules.Default, cf)

	if _, ok := result["ls"]; !ok {
		t.Error("ls should be preserved from defaults")
	}
	nix, ok := result["nix"]
	if !ok {
		t.Fatal("nix should be present")
	}
	if nix.Default != rules.Prompt {
		t.Error("nix.default should be Prompt")
	}
}

func TestApplyLayer_overwriteMode(t *testing.T) {
	cf := &ConfigFile{
		OverwriteDefaults: boolPtr(true),
		Rules: map[string]*RuleConfig{
			"nix": {Decision: decisionPtr(rules.Prompt)},
		},
	}

	result := applyLayer(rules.Default, cf)

	if _, ok := result["ls"]; ok {
		t.Error("overwrite=true should drop default rules like ls")
	}
	nix, ok := result["nix"]
	if !ok {
		t.Fatal("nix should be present")
	}
	if nix.Default != rules.Prompt {
		t.Error("nix.default should be Prompt")
	}
}

func TestApplyLayer_unspecifiedDefaultsToMerge(t *testing.T) {
	cf := &ConfigFile{
		Rules: map[string]*RuleConfig{
			"nix": {Decision: decisionPtr(rules.Prompt)},
		},
	}

	result := applyLayer(rules.Default, cf)

	if _, ok := result["ls"]; !ok {
		t.Error("ls should be preserved when overwrite_defaults is unset")
	}
}

func TestBuildRules_noConfig(t *testing.T) {
	result := buildRules(nil, nil)
	if result == nil {
		t.Fatal("expected non-nil result for nil config")
	}
	if len(result) == 0 {
		t.Error("nil config should return non-empty rules")
	}
}

// TestBuildRules_layeredOverwrite validates the overwrite_defaults matrix
// across the global and repo layers:
//
//	global   repo    app defaults survive?   global survive?
//	no       no      yes                      yes
//	yes      no      no                       yes
//	no       yes     no                       no
//	yes      yes     no                       no
func TestBuildRules_layeredOverwrite(t *testing.T) {
	t.Run("global merge, repo merge", func(t *testing.T) {
		global := &ConfigFile{
			Rules: map[string]*RuleConfig{"gnix": {Decision: decisionPtr(rules.Allow)}},
		}
		repo := &ConfigFile{
			Rules: map[string]*RuleConfig{"rnix": {Decision: decisionPtr(rules.Allow)}},
		}

		result := buildRules(global, repo)

		// app defaults survive
		if _, ok := result["ls"]; !ok {
			t.Error("ls should survive")
		}
		// global survives
		if _, ok := result["gnix"]; !ok {
			t.Error("gnix should survive")
		}
		// repo present
		if _, ok := result["rnix"]; !ok {
			t.Error("rnix should be present")
		}
	})

	t.Run("global overwrite, repo merge", func(t *testing.T) {
		global := &ConfigFile{
			OverwriteDefaults: boolPtr(true),
			Rules:             map[string]*RuleConfig{"gnix": {Decision: decisionPtr(rules.Allow)}},
		}
		repo := &ConfigFile{
			Rules: map[string]*RuleConfig{"rnix": {Decision: decisionPtr(rules.Allow)}},
		}

		result := buildRules(global, repo)

		// app defaults dropped by global overwrite
		if _, ok := result["ls"]; ok {
			t.Error("ls should be dropped (global overwrote defaults)")
		}
		// global survives
		if _, ok := result["gnix"]; !ok {
			t.Error("gnix should survive")
		}
		// repo present (merged over the effective base, which is global-only)
		if _, ok := result["rnix"]; !ok {
			t.Error("rnix should be present")
		}
	})

	t.Run("global merge, repo overwrite", func(t *testing.T) {
		global := &ConfigFile{
			Rules: map[string]*RuleConfig{"gnix": {Decision: decisionPtr(rules.Allow)}},
		}
		repo := &ConfigFile{
			OverwriteDefaults: boolPtr(true),
			Rules:             map[string]*RuleConfig{"rnix": {Decision: decisionPtr(rules.Allow)}},
		}

		result := buildRules(global, repo)

		// app defaults dropped by repo overwrite
		if _, ok := result["ls"]; ok {
			t.Error("ls should be dropped (repo overwrote effective base)")
		}
		// global dropped by repo overwrite
		if _, ok := result["gnix"]; ok {
			t.Error("gnix should be dropped (repo overwrote effective base)")
		}
		// repo present
		if _, ok := result["rnix"]; !ok {
			t.Error("rnix should be present")
		}
	})

	t.Run("global overwrite, repo overwrite", func(t *testing.T) {
		global := &ConfigFile{
			OverwriteDefaults: boolPtr(true),
			Rules:             map[string]*RuleConfig{"gnix": {Decision: decisionPtr(rules.Allow)}},
		}
		repo := &ConfigFile{
			OverwriteDefaults: boolPtr(true),
			Rules:             map[string]*RuleConfig{"rnix": {Decision: decisionPtr(rules.Allow)}},
		}

		result := buildRules(global, repo)

		// app defaults dropped
		if _, ok := result["ls"]; ok {
			t.Error("ls should be dropped")
		}
		// global dropped by repo overwrite
		if _, ok := result["gnix"]; ok {
			t.Error("gnix should be dropped")
		}
		// repo present
		if _, ok := result["rnix"]; !ok {
			t.Error("rnix should be present")
		}
	})
}

func TestLoad_noConfig(t *testing.T) {
	cfg, err := load("", t.TempDir())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if !cfg.RtkRewrite {
		t.Error("expected RtkRewrite to default to true")
	}
	if _, ok := cfg.Rules["ls"]; !ok {
		t.Error("ls should be present from defaults")
	}
}

func TestLoad_globalOnly(t *testing.T) {
	globalDir := t.TempDir()
	writeFile(t, globalDir, "crushout.yml", `rtk_rewrite: false
rules:
  gnix: allow
`)
	repoDir := t.TempDir()

	cfg, err := load(globalDir, repoDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.RtkRewrite {
		t.Error("expected RtkRewrite to be false from global")
	}
	if _, ok := cfg.Rules["gnix"]; !ok {
		t.Error("gnix should be present")
	}
	if _, ok := cfg.Rules["ls"]; !ok {
		t.Error("ls should be preserved from defaults")
	}
}

func TestLoad_repoOnly(t *testing.T) {
	repoDir := t.TempDir()
	writeFile(t, repoDir, ".crushout.yml", `rtk_rewrite: false
rules:
  rnix: allow
`)

	cfg, err := load("", repoDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.RtkRewrite {
		t.Error("expected RtkRewrite to be false from repo")
	}
	if _, ok := cfg.Rules["rnix"]; !ok {
		t.Error("rnix should be present")
	}
	if _, ok := cfg.Rules["ls"]; !ok {
		t.Error("ls should be preserved from defaults")
	}
}

func TestLoad_bothMerge(t *testing.T) {
	globalDir := t.TempDir()
	writeFile(t, globalDir, "crushout.yml", `rules:
  gnix: allow
`)
	repoDir := t.TempDir()
	writeFile(t, repoDir, ".crushout.yml", `rules:
  rnix: allow
`)

	cfg, err := load(globalDir, repoDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, ok := cfg.Rules["gnix"]; !ok {
		t.Error("gnix should be present from global")
	}
	if _, ok := cfg.Rules["rnix"]; !ok {
		t.Error("rnix should be present from repo")
	}
	if _, ok := cfg.Rules["ls"]; !ok {
		t.Error("ls should be preserved from defaults")
	}
}

func TestLoad_repoRtkRewriteOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	writeFile(t, globalDir, "crushout.yml", `rtk_rewrite: false
`)
	repoDir := t.TempDir()
	writeFile(t, repoDir, ".crushout.yml", `rtk_rewrite: true
`)

	cfg, err := load(globalDir, repoDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !cfg.RtkRewrite {
		t.Error("expected RtkRewrite to be true (repo wins)")
	}
}

func TestLoad_globalUnsetRepoUnsetDefaultsTrue(t *testing.T) {
	globalDir := t.TempDir()
	writeFile(t, globalDir, "crushout.yml", `rules: {}
`)
	repoDir := t.TempDir()
	writeFile(t, repoDir, ".crushout.yml", `rules: {}
`)

	cfg, err := load(globalDir, repoDir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !cfg.RtkRewrite {
		t.Error("expected RtkRewrite to default to true when both unset")
	}
}

func TestLoad_malformedGlobalErrors(t *testing.T) {
	globalDir := t.TempDir()
	writeFile(t, globalDir, "crushout.yml", "invalid: [yaml")
	repoDir := t.TempDir()

	_, err := load(globalDir, repoDir)
	if err == nil {
		t.Error("expected error for malformed global config")
	}
}

func TestLoad_malformedRepoErrors(t *testing.T) {
	globalDir := t.TempDir()
	repoDir := t.TempDir()
	writeFile(t, repoDir, ".crushout.yml", "invalid: [yaml")

	_, err := load(globalDir, repoDir)
	if err == nil {
		t.Error("expected error for malformed repo config")
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
			Decision: decisionPtr(rules.Prompt),
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
	if nix.Default != rules.Prompt {
		t.Error("expected nix.default to be Prompt")
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
		"ls": {},
	}

	result := ToRules(cfg)

	ls, ok := result["ls"]
	if !ok {
		t.Fatal("expected ls rule")
	}
	if ls.Default != rules.Prompt {
		t.Error("expected ls.default to be Prompt for nil decision")
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
	if nix.Default != rules.Prompt {
		t.Error("nix.default should be Prompt")
	}
}

func TestMerge_overrideDefault(t *testing.T) {
	base := map[string]*rules.Rule{
		"ls": {Default: rules.Allow},
	}
	user := map[string]*rules.Rule{
		"ls": {Default: rules.Prompt, DefaultExplicit: true},
	}

	result := Merge(base, user)

	if result["ls"].Default != rules.Prompt {
		t.Error("user should win: ls.default should be Prompt")
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
				"status": newRuleWithDecision(rules.Prompt),
				"fetch":  {},
			},
		},
	}

	result := Merge(base, user)

	git := result["git"]
	if git.Default != rules.Prompt {
		t.Error("git.default should remain Prompt")
	}
	if _, ok := git.Subcommands["push"]; !ok {
		t.Error("git.push should be preserved from base")
	}
	if git.Subcommands["status"].Default != rules.Prompt {
		t.Error("git.status should be overridden to Prompt")
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
						"show": newRuleWithDecision(rules.Prompt),
					},
				},
			},
		},
	}

	result := Merge(base, user)

	remote := result["git"].Subcommands["remote"]
	if remote.Subcommands["show"].Default != rules.Prompt {
		t.Error("git.remote.show should be overridden to Prompt")
	}
	if remote.Subcommands["add"].Default != rules.Prompt {
		t.Error("git.remote.add should be preserved as Prompt")
	}
}

func TestMerge_promptFlagsOverride(t *testing.T) {
	base := map[string]*rules.Rule{
		"find": {
			Default:     rules.Allow,
			PromptFlags: []string{"-exec", "-delete"},
		},
	}
	user := map[string]*rules.Rule{
		"find": {
			PromptFlags: []string{"-exec", "-fprint"},
		},
	}

	result := Merge(base, user)

	find := result["find"]
	if len(find.PromptFlags) != 2 {
		t.Fatalf("expected 2 prompt flags, got %d", len(find.PromptFlags))
	}
	if find.PromptFlags[0] != "-exec" || find.PromptFlags[1] != "-fprint" {
		t.Errorf("expected [-exec -fprint], got %v", find.PromptFlags)
	}
}

func TestMerge_preservePromptFlagsWhenUserHasNone(t *testing.T) {
	base := map[string]*rules.Rule{
		"sed": {
			Default:     rules.Allow,
			PromptFlags: []string{"-i"},
		},
	}
	user := map[string]*rules.Rule{
		"sed": {},
	}

	result := Merge(base, user)

	sed := result["sed"]
	if len(sed.PromptFlags) != 1 || sed.PromptFlags[0] != "-i" {
		t.Errorf("expected [-i], got %v", sed.PromptFlags)
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

func TestDeepCopyRule(t *testing.T) {
	orig := &rules.Rule{
		Default:     rules.Allow,
		PromptFlags: []string{"-i"},
		Message:     "test message",
		Subcommands: map[string]*rules.Rule{
			"sub": {},
		},
	}

	copy_ := deepCopyRule(orig)

	if copy_ == orig {
		t.Error("deepCopyRule should return a new instance")
	}
	if copy_.Subcommands["sub"] == orig.Subcommands["sub"] {
		t.Error("Subcommands map should be deep copied")
	}
	if len(copy_.PromptFlags) != len(orig.PromptFlags) || copy_.PromptFlags[0] != "-i" {
		t.Error("PromptFlags slice should be deep copied with correct values")
	}

	if copy_.Default != rules.Allow || copy_.PromptFlags[0] != "-i" || copy_.Message != "test message" || copy_.Subcommands["sub"].Default != rules.Prompt {
		t.Error("copied values should match original")
	}
}

func TestDeepCopyRules_nil(t *testing.T) {
	result := deepCopyRules(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}
}

func TestMerge_allowFlagsOverride(t *testing.T) {
	base := map[string]*rules.Rule{
		"gofmt": {
			Default:    rules.Allow,
			AllowFlags: []string{"-l", "-d"},
		},
	}
	user := map[string]*rules.Rule{
		"gofmt": {
			AllowFlags: []string{"-l", "-d", "-s"},
		},
	}

	result := Merge(base, user)

	gofmt := result["gofmt"]
	if len(gofmt.AllowFlags) != 3 {
		t.Fatalf("expected 3 allow flags, got %d", len(gofmt.AllowFlags))
	}
}

func TestMerge_preserveAllowFlagsWhenUserHasNone(t *testing.T) {
	base := map[string]*rules.Rule{
		"gofmt": {
			Default:    rules.Allow,
			AllowFlags: []string{"-l", "-d"},
		},
	}
	user := map[string]*rules.Rule{
		"gofmt": {},
	}

	result := Merge(base, user)

	gofmt := result["gofmt"]
	if len(gofmt.AllowFlags) != 2 {
		t.Errorf("expected allow flags preserved, got %v", gofmt.AllowFlags)
	}
}

func TestDeepCopyRule_allowFlags(t *testing.T) {
	orig := &rules.Rule{
		AllowFlags: []string{"-l", "-d"},
	}
	copy_ := deepCopyRule(orig)
	copy_.AllowFlags[0] = "-w"
	if orig.AllowFlags[0] == "-w" {
		t.Error("AllowFlags should be deep copied")
	}
}
