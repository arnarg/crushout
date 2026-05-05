package rules

import (
	"strings"
	"testing"
)

func TestRule_SimpleAllow(t *testing.T) {
	r := &Rule{Default: Allow}
	if !r.Allow(nil) {
		t.Error("expected allow")
	}
	if !r.Allow([]string{"-la", "/tmp"}) {
		t.Error("expected allow with args")
	}
}

func TestRule_SimpleDeny(t *testing.T) {
	r := &Rule{}
	if r.Allow(nil) {
		t.Error("expected deny")
	}
	if r.Allow([]string{"-rf", "/"}) {
		t.Error("expected deny with args")
	}
}

func TestRule_SubcommandAllow(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"status": {Default: Allow},
			"diff":   {Default: Allow},
		},
	}
	if !r.Allow([]string{"status"}) {
		t.Error("git status should be allowed")
	}
	if !r.Allow([]string{"diff", "HEAD~3"}) {
		t.Error("git diff HEAD~3 should be allowed")
	}
}

func TestRule_SubcommandDeny(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"push":   {},
			"commit": {},
		},
	}
	if r.Allow([]string{"push"}) {
		t.Error("git push should be denied")
	}
	if r.Allow([]string{"commit", "-m", "fix"}) {
		t.Error("git commit should be denied")
	}
}

func TestRule_SubcommandFallbackToDefault(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"status": {Default: Allow},
		},
	}
	// "clone" is not in subcommands, falls to Default (NoOpinion)
	if r.Allow([]string{"clone"}) {
		t.Error("unknown git subcommand should fall to default (deny)")
	}
}

func TestRule_DenyFlags(t *testing.T) {
	r := &Rule{
		Default:   Allow,
		DenyFlags: []string{"-i", "--in-place"},
	}
	if !r.Allow([]string{"s/old/new/g", "file.txt"}) {
		t.Error("sed without -i should be allowed")
	}
	if r.Allow([]string{"-i", "s/old/new/g", "file.txt"}) {
		t.Error("sed -i should be denied")
	}
	if r.Allow([]string{"--in-place", "s/old/new/g"}) {
		t.Error("sed --in-place should be denied")
	}
}

func TestRule_DenyFlagsAtNestedLevel(t *testing.T) {
	r := &Rule{
		DenyFlags: []string{"-C", "--work-tree"},
		Subcommands: map[string]*Rule{
			"branch": {
				DenyFlags: []string{"-d", "-D", "--delete", "-m", "-M", "--move"},
				Subcommands: map[string]*Rule{
					"-l":     {Default: Allow},
					"--list": {Default: Allow},
				},
			},
		},
	}
	if !r.Allow([]string{"branch", "-l"}) {
		t.Error("git branch -l should be allowed")
	}
	if r.Allow([]string{"branch", "-D", "foo"}) {
		t.Error("git branch -D should be denied by nested DenyFlags")
	}
	if r.Allow([]string{"-C", "/tmp", "status"}) {
		t.Error("git -C should be denied by top-level DenyFlags")
	}
}

func TestRule_DeepNesting(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"remote": {
				Subcommands: map[string]*Rule{
					"-v":      {Default: Allow},
					"show":    {Default: Allow},
					"get-url": {Default: Allow},
					"add":     {},
					"remove":  {},
				},
			},
		},
	}
	if !r.Allow([]string{"remote", "-v"}) {
		t.Error("git remote -v should be allowed")
	}
	if !r.Allow([]string{"remote", "show", "origin"}) {
		t.Error("git remote show origin should be allowed")
	}
	if !r.Allow([]string{"remote", "get-url", "origin"}) {
		t.Error("git remote get-url origin should be allowed")
	}
	if r.Allow([]string{"remote", "add", "origin", "url"}) {
		t.Error("git remote add should be denied")
	}
	if r.Allow([]string{"remote", "remove", "origin"}) {
		t.Error("git remote remove should be denied")
	}
}

func TestRule_ThreeLevels(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"mod": {
				Subcommands: map[string]*Rule{
					"graph":    {Default: Allow},
					"download": {Default: Allow},
					"tidy":     {},
				},
			},
		},
	}
	if !r.Allow([]string{"mod", "graph"}) {
		t.Error("go mod graph should be allowed")
	}
	if !r.Allow([]string{"mod", "download"}) {
		t.Error("go mod download should be allowed")
	}
	if r.Allow([]string{"mod", "tidy"}) {
		t.Error("go mod tidy should be denied")
	}
}

func TestRule_FlagDenyBeforeSubcommand(t *testing.T) {
	// git -C /tmp status → -C should be caught at top level
	r := &Rule{
		DenyFlags: []string{"-C"},
		Subcommands: map[string]*Rule{
			"status": {Default: Allow},
		},
	}
	if r.Allow([]string{"-C", "/tmp", "status"}) {
		t.Error("git -C /tmp status should be denied")
	}
}

func TestRule_FindDenyFlags(t *testing.T) {
	r := &Rule{
		Default:   Allow,
		DenyFlags: []string{"-exec", "-execdir", "-ok", "-okdir", "-delete"},
	}
	if !r.Allow([]string{".", "-name", "*.go"}) {
		t.Error("find without -exec should be allowed")
	}
	if r.Allow([]string{".", "-name", "*.go", "-exec", "rm", "{}", ";"}) {
		t.Error("find -exec should be denied")
	}
	if r.Allow([]string{".", "-delete"}) {
		t.Error("find -delete should be denied")
	}
}

func TestRule_StashNested(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"stash": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"show": {Default: Allow},
					"push": {},
					"pop":  {},
				},
			},
		},
	}
	if !r.Allow([]string{"stash", "list"}) {
		t.Error("git stash list should be allowed")
	}
	if !r.Allow([]string{"stash", "show"}) {
		t.Error("git stash show should be allowed")
	}
	if r.Allow([]string{"stash"}) {
		t.Error("git stash (bare) should be denied")
	}
	if r.Allow([]string{"stash", "pop"}) {
		t.Error("git stash pop should be denied")
	}
	if r.Allow([]string{"stash", "push"}) {
		t.Error("git stash push should be denied")
	}
}

func TestRule_DenyFlags_ExactMatch(t *testing.T) {
	r := &Rule{
		Default:   Allow,
		DenyFlags: []string{"-c", "--coverprofile"},
	}
	if !r.Allow([]string{"-count=1", "./..."}) {
		t.Error("go test -count=1 should be allowed (-c should not match -count)")
	}
	if !r.Allow([]string{"-cover", "./..."}) {
		t.Error("go test -cover should be allowed (-c should not match -cover)")
	}
	if r.Allow([]string{"-c", "./..."}) {
		t.Error("go test -c should be denied")
	}
	if r.Allow([]string{"-c=output", "./..."}) {
		t.Error("go test -c=output should be denied")
	}
}

func TestRule_DenyFlags_EqualsSuffix(t *testing.T) {
	r := &Rule{
		Default:   Allow,
		DenyFlags: []string{"--global"},
	}
	if r.Allow([]string{"--global=value"}) {
		t.Error("--global=value should be denied")
	}
	if !r.Allow([]string{"--globalish", "value"}) {
		t.Error("--globalish should be allowed")
	}
}

func TestRule_GoTestDenyFlags(t *testing.T) {
	r := &Rule{
		Subcommands: map[string]*Rule{
			"test": {
				Default:   Allow,
				DenyFlags: []string{"-c", "--coverprofile"},
			},
		},
	}
	if !r.Allow([]string{"test", "./..."}) {
		t.Error("go test ./... should be allowed")
	}
	if r.Allow([]string{"test", "-c", "./..."}) {
		t.Error("go test -c should be denied")
	}
}

func TestRule_ConfigDenyFlags(t *testing.T) {
	r := &Rule{
		Default:   Allow,
		DenyFlags: []string{"--global", "--system", "--file"},
	}
	if !r.Allow([]string{"--get", "user.name"}) {
		t.Error("git config --get should be allowed")
	}
	if r.Allow([]string{"--global", "user.name", "New"}) {
		t.Error("git config --global should be denied")
	}
}

func TestRule_Resolve_DenyDecision(t *testing.T) {
	r := &Rule{
		Default: Deny,
		Message: "this command is blocked",
	}
	d, msg := r.Resolve(nil)
	if d != Deny {
		t.Error("expected deny decision")
	}
	if msg == "" {
		t.Error("expected deny message")
	}
}

func TestRule_Resolve_DenyWithSubcommand(t *testing.T) {
	r := &Rule{
		Default: Deny,
		Message: "kubectl blocked",
		Subcommands: map[string]*Rule{
			"get": {Default: Allow},
		},
	}
	// "get" matches subcommand, should be allowed
	d, _ := r.Resolve([]string{"get", "pods"})
	if d != Allow {
		t.Error("kubectl get should be allowed via subcommand")
	}
	// no subcommand match, falls to default deny
	d, msg := r.Resolve([]string{"exec", "-it", "pod", "--", "bash"})
	if d != Deny {
		t.Error("kubectl exec should be denied")
	}
	if msg == "" {
		t.Error("expected deny message")
	}
}

func TestRule_Resolve_DenyMessageFallback(t *testing.T) {
	r := &Rule{Default: Deny}
	_, msg := r.Resolve(nil)
	if msg != "denied by rule" {
		t.Errorf("expected fallback message, got %q", msg)
	}
}

func TestRule_Resolve_DenyMessageCustom(t *testing.T) {
	r := &Rule{
		Default: Deny,
		Message: "no curl allowed",
	}
	_, msg := r.Resolve(nil)
	if msg != "no curl allowed" {
		t.Errorf("expected custom message, got %q", msg)
	}
}

// ── Test the full Default ruleset ─────────────────────────

func TestDefaultRules_NonMutableCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"ls", []string{"-la"}},
		{"cat", []string{"file.txt"}},
		{"grep", []string{"pattern", "file"}},
		{"sort", nil},
		{"uniq", nil},
		{"pwd", nil},
		{"whoami", nil},
		{"echo", []string{"hello"}},
		{"find", []string{".", "-name", "*.go"}},
		{"sed", []string{"s/old/new/g", "file.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, exists := Default[tt.name]
			if !exists {
				t.Fatalf("no rule for %q", tt.name)
			}
			if !rule.Allow(tt.args) {
				t.Errorf("%q %v should be allowed", tt.name, tt.args)
			}
		})
	}
}

func TestDefaultRules_MutableCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"rm", []string{"-rf", "/"}},
		{"mv", []string{"a", "b"}},
		{"cp", []string{"a", "b"}},
		{"sudo", []string{"rm", "-rf", "/"}},
		{"curl", []string{"https://example.com"}},
		{"make", []string{"build"}},
		{"npm", []string{"install"}},
		{"python", []string{"-c", "print(1)"}},
		{"bash", []string{"-c", "echo hi"}},
		{"xargs", []string{"rm"}},
		{"env", []string{"FOO=bar", "rm", "-rf", "/"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, exists := Default[tt.name]
			if !exists {
				t.Fatalf("no rule for %q", tt.name)
			}
			if rule.Allow(tt.args) {
				t.Errorf("%q %v should be denied", tt.name, tt.args)
			}
		})
	}
}

func TestDefaultRules_FlagDependent(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		allowed bool
	}{
		{"sed", []string{"s/old/new/g", "file.txt"}, true},
		{"sed", []string{"-i", "s/old/new/g", "file.txt"}, false},
		{"sed", []string{"--in-place", "s/old/new/g"}, false},
		{"find", []string{".", "-name", "*.go"}, true},
		{"find", []string{".", "-exec", "rm", "{}", ";"}, false},
		{"find", []string{".", "-delete"}, false},
	}

	for _, tt := range tests {
		name := tt.name + " " + tt.args[0]
		t.Run(name, func(t *testing.T) {
			rule, exists := Default[tt.name]
			if !exists {
				t.Fatalf("no rule for %q", tt.name)
			}
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_GitSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		allowed bool
	}{
		// Read-only
		{[]string{"status"}, true},
		{[]string{"diff"}, true},
		{[]string{"diff", "HEAD~3"}, true},
		{[]string{"log", "--oneline", "-10"}, true},
		{[]string{"show"}, true},
		{[]string{"blame", "file.go"}, true},
		{[]string{"reflog"}, true},
		{[]string{"ls-files"}, true},
		{[]string{"ls-tree", "HEAD"}, true},
		{[]string{"ls-remote"}, true},
		{[]string{"describe"}, true},
		{[]string{"rev-parse", "HEAD"}, true},
		{[]string{"remote", "-v"}, true},
		{[]string{"remote", "show", "origin"}, true},
		{[]string{"remote", "get-url", "origin"}, true},
		{[]string{"branch", "-l"}, true},
		{[]string{"branch", "--list"}, true},
		{[]string{"branch", "--show-current"}, true},
		{[]string{"tag", "-l"}, true},
		{[]string{"stash", "list"}, true},
		{[]string{"stash", "show"}, true},
		{[]string{"submodule", "status"}, true},
		{[]string{"config", "get", "user.name"}, true},

		// Mutable
		{[]string{"commit", "-m", "fix"}, false},
		{[]string{"push"}, false},
		{[]string{"push", "origin", "main"}, false},
		{[]string{"pull"}, false},
		{[]string{"merge"}, false},
		{[]string{"rebase"}, false},
		{[]string{"reset"}, false},
		{[]string{"checkout", "main"}, false},
		{[]string{"switch", "main"}, false},
		{[]string{"clone", "url"}, false},
		{[]string{"init"}, false},
		{[]string{"add", "."}, false},
		{[]string{"clean", "-fd"}, false},
		{[]string{"stash"}, false},
		{[]string{"stash", "pop"}, false},
		{[]string{"stash", "push"}, false},
		{[]string{"branch", "new-feature"}, false},
		{[]string{"branch", "-D", "old"}, false},
		{[]string{"tag", "v1.0.0"}, false},
		{[]string{"tag", "-d", "v1.0.0"}, false},
		{[]string{"remote", "add", "origin", "url"}, false},
		{[]string{"remote", "remove", "origin"}, false},
		{[]string{"submodule", "update"}, false},

		// DenyFlags
		{[]string{"-C", "/tmp", "status"}, false},
		{[]string{"--work-tree", "/tmp", "status"}, false},
		{[]string{"config", "--global", "user.name", "X"}, false},
	}

	for _, tt := range tests {
		name := "git " + tt.args[0]
		if len(tt.args) > 1 {
			name += " " + tt.args[1]
		}
		t.Run(name, func(t *testing.T) {
			rule := Default["git"]
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("git Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_GoSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		allowed bool
	}{
		{[]string{"version"}, true},
		{[]string{"env", "GOPATH"}, true},
		{[]string{"list", "./..."}, true},
		{[]string{"doc", "fmt.Println"}, true},
		{[]string{"vet", "./..."}, true},
		{[]string{"test", "./..."}, true},
		{[]string{"mod", "graph"}, true},
		{[]string{"mod", "download"}, true},
		{[]string{"mod", "verify"}, true},
		{[]string{"tool", "pprof", "cpu.prof"}, true},
		{[]string{"build", "./..."}, true},
		{[]string{"build", "./cmd/crushout"}, false},
		{[]string{"run", "main.go"}, false},
		{[]string{"install", "./..."}, false},
		{[]string{"get", "github.com/foo/bar"}, false},
		{[]string{"generate", "./..."}, false},
		{[]string{"mod", "tidy"}, false},
		{[]string{"mod", "init"}, false},
		{[]string{"test", "-c", "./..."}, false},
	}

	for _, tt := range tests {
		name := "go " + tt.args[0]
		if len(tt.args) > 1 {
			name += " " + tt.args[1]
		}
		t.Run(name, func(t *testing.T) {
			rule := Default["go"]
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("go Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_CargoSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		allowed bool
	}{
		{[]string{"version"}, true},
		{[]string{"check"}, true},
		{[]string{"test"}, true},
		{[]string{"doc"}, true},
		{[]string{"tree"}, true},
		{[]string{"metadata"}, true},
		{[]string{"search", "serde"}, true},
		{[]string{"build"}, false},
		{[]string{"run"}, false},
		{[]string{"install"}, false},
		{[]string{"add", "serde"}, false},
		{[]string{"remove", "serde"}, false},
		{[]string{"publish"}, false},
		{[]string{"init"}, false},
		{[]string{"new", "myproject"}, false},
		{[]string{"clean"}, false},
		{[]string{"update"}, false},
	}

	for _, tt := range tests {
		name := "cargo " + tt.args[0]
		t.Run(name, func(t *testing.T) {
			rule := Default["cargo"]
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("cargo Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_GhSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		allowed bool
	}{
		{[]string{"version"}, true},
		{[]string{"api", "/repos/owner/repo"}, true},
		{[]string{"pr", "list"}, true},
		{[]string{"pr", "view", "123"}, true},
		{[]string{"pr", "status"}, true},
		{[]string{"pr", "diff", "123"}, true},
		{[]string{"pr", "create"}, false},
		{[]string{"pr", "merge", "123"}, false},
		{[]string{"pr", "close", "123"}, false},
		{[]string{"issue", "list"}, true},
		{[]string{"issue", "view", "456"}, true},
		{[]string{"issue", "create"}, false},
		{[]string{"issue", "close", "456"}, false},
		{[]string{"repo", "list"}, true},
		{[]string{"repo", "view"}, true},
		{[]string{"repo", "clone", "owner/repo"}, true},
		{[]string{"repo", "create"}, false},
		{[]string{"repo", "delete"}, false},
		{[]string{"release", "list"}, true},
		{[]string{"release", "view", "v1.0"}, true},
		{[]string{"release", "create", "v1.0"}, false},
		{[]string{"release", "delete", "v1.0"}, false},
		{[]string{"workflow", "list"}, true},
		{[]string{"workflow", "view"}, true},
		{[]string{"workflow", "run"}, false},
		{[]string{"run", "list"}, true},
		{[]string{"run", "view"}, true},
		{[]string{"run", "watch"}, true},
		{[]string{"run", "rerun"}, false},
		{[]string{"auth", "status"}, true},
		{[]string{"auth", "login"}, false},
		{[]string{"gpg-key", "list"}, true},
		{[]string{"ssh-key", "list"}, true},
	}

	for _, tt := range tests {
		name := "gh " + strings.Join(tt.args, " ")
		t.Run(name, func(t *testing.T) {
			rule := Default["gh"]
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("gh Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_KubectlSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		allowed bool
	}{
		{[]string{"get", "pods"}, true},
		{[]string{"describe", "pod", "my-pod"}, true},
		{[]string{"logs", "my-pod"}, true},
		{[]string{"top", "pods"}, true},
		{[]string{"api-resources"}, true},
		{[]string{"api-versions"}, true},
		{[]string{"version"}, true},
		{[]string{"cluster-info"}, true},
		{[]string{"auth", "can-i", "get", "pods"}, true},
		{[]string{"apply", "-f", "deploy.yaml"}, false},
		{[]string{"create", "-f", "deploy.yaml"}, false},
		{[]string{"delete", "pod", "my-pod"}, false},
		{[]string{"edit", "deploy/my-deploy"}, false},
		{[]string{"exec", "-it", "my-pod", "--", "bash"}, false},
		{[]string{"scale", "--replicas=3", "deploy/my-deploy"}, false},
		{[]string{"rollout", "status", "deploy/my-deploy"}, false},
	}

	for _, tt := range tests {
		name := "kubectl " + strings.Join(tt.args, " ")
		t.Run(name, func(t *testing.T) {
			rule := Default["kubectl"]
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("kubectl Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_DockerSubcommands(t *testing.T) {
	tests := []struct {
		args    []string
		allowed bool
	}{
		{[]string{"version"}, true},
		{[]string{"info"}, true},
		{[]string{"images"}, true},
		{[]string{"ps"}, true},
		{[]string{"inspect", "my-container"}, true},
		{[]string{"logs", "my-container"}, true},
		{[]string{"stats"}, true},
		{[]string{"top", "my-container"}, true},
		{[]string{"history", "my-image"}, true},
		{[]string{"search", "nginx"}, true},
		{[]string{"volume", "ls"}, true},
		{[]string{"volume", "inspect", "my-vol"}, true},
		{[]string{"network", "ls"}, true},
		{[]string{"network", "inspect", "my-net"}, true},
		{[]string{"build", "-t", "img", "."}, false},
		{[]string{"run", "nginx"}, false},
		{[]string{"pull", "nginx"}, false},
		{[]string{"push", "my-image"}, false},
		{[]string{"stop", "my-container"}, false},
		{[]string{"kill", "my-container"}, false},
		{[]string{"exec", "-it", "my-container", "bash"}, false},
		{[]string{"rm", "my-container"}, false},
		{[]string{"rmi", "my-image"}, false},
		{[]string{"volume", "create"}, false},
		{[]string{"volume", "rm", "my-vol"}, false},
		{[]string{"network", "create", "my-net"}, false},
		{[]string{"network", "rm", "my-net"}, false},
		{[]string{"compose", "up"}, false},
	}

	for _, tt := range tests {
		name := "docker " + strings.Join(tt.args, " ")
		t.Run(name, func(t *testing.T) {
			rule := Default["docker"]
			if rule.Allow(tt.args) != tt.allowed {
				t.Errorf("docker Allow(%v) = %v, want %v", tt.args, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestDefaultRules_UnknownCommand(t *testing.T) {
	_, exists := Default["foobarbaz"]
	if exists {
		t.Error("unknown command should not exist in rules")
	}
}
