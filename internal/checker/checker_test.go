package checker

import (
	"testing"

	"github.com/arnarg/crushout/internal/rules"
)

func newTestChecker() *Checker {
	return &Checker{
		RootDir: "/home/user/project",
		HomeDir: "/home/user",
		Rules:   rules.Default,
	}
}

// ── Success cases: Check returns Allow ────────────────────

func TestCheck_SimpleLs(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("ls -la")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitStatus(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git status")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitDiff(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git diff HEAD~3")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitLog(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git log --oneline -10")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitRemoteV(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git remote -v")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitRemoteShow(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git remote show origin")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitBranchList(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git branch -l")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitBranchShowCurrent(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git branch --show-current")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitTagList(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git tag -l")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitStashList(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git stash list")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitSubmoduleStatus(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git submodule status")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_Pipeline(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("cat README.md | grep -i todo | sort | uniq")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_CompoundList(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git status && git diff && ls -la")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_Find(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("find . -name '*.go'")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_Sed(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("sed 's/old/new/g' file.txt")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GoTest(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go test ./...")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GoModGraph(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go mod graph")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GoEnv(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go env GOPATH")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GoToolPprof(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go tool pprof cpu.prof")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_PwdWhoamiUname(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("pwd && whoami && uname -a")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_Echo(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("echo hello")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_CDWithinRoot(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("cd src && ls -la && cd ..")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

func TestCheck_GitConfigGet(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("git config get user.name")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("expected allow")
	}
}

// ── Failure cases: Check returns NoOpinion ────────────────

func TestCheck_Rm(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("rm -rf /")
	if d != rules.NoOpinion {
		t.Error("rm should be no-opinion")
	}
}

func TestCheck_SudoRm(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("sudo rm -rf /")
	if d != rules.NoOpinion {
		t.Error("sudo should be no-opinion")
	}
}

func TestCheck_GitPush(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git push")
	if d != rules.NoOpinion {
		t.Error("git push should be no-opinion")
	}
}

func TestCheck_GitPushOrigin(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git push origin main")
	if d != rules.NoOpinion {
		t.Error("git push origin main should be no-opinion")
	}
}

func TestCheck_GitCommit(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git commit -m 'fix'")
	if d != rules.NoOpinion {
		t.Error("git commit should be no-opinion")
	}
}

func TestCheck_GitCheckout(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git checkout -b feature")
	if d != rules.NoOpinion {
		t.Error("git checkout should be no-opinion")
	}
}

func TestCheck_GitRemoteAdd(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git remote add origin git@github.com:user/repo.git")
	if d != rules.NoOpinion {
		t.Error("git remote add should be no-opinion")
	}
}

func TestCheck_GitBranchDelete(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git branch -D old-branch")
	if d != rules.NoOpinion {
		t.Error("git branch -D should be no-opinion")
	}
}

func TestCheck_GitBranchCreate(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git branch new-feature")
	if d != rules.NoOpinion {
		t.Error("git branch new-feature should be no-opinion")
	}
}

func TestCheck_GitTagCreate(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git tag v1.0.0")
	if d != rules.NoOpinion {
		t.Error("git tag v1.0.0 should be no-opinion")
	}
}

func TestCheck_GitStashBare(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git stash")
	if d != rules.NoOpinion {
		t.Error("git stash (bare) should be no-opinion")
	}
}

func TestCheck_GitStashPop(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git stash pop")
	if d != rules.NoOpinion {
		t.Error("git stash pop should be no-opinion")
	}
}

func TestCheck_GitC(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git -C /tmp status")
	if d != rules.NoOpinion {
		t.Error("git -C should be no-opinion")
	}
}

func TestCheck_GitConfigGlobal(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git config --global user.name New")
	if d != rules.NoOpinion {
		t.Error("git config --global should be no-opinion")
	}
}

func TestCheck_FindExec(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("find . -name '*.go' -exec rm {} \\;")
	if d != rules.NoOpinion {
		t.Error("find -exec should be no-opinion")
	}
}

func TestCheck_SedInPlace(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("sed -i 's/old/new/g' file.txt")
	if d != rules.NoOpinion {
		t.Error("sed -i should be no-opinion")
	}
}

func TestCheck_Make(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("make build")
	if d != rules.NoOpinion {
		t.Error("make should be no-opinion")
	}
}

func TestCheck_Docker(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("docker ps")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("docker ps should be allow")
	}
}

func TestCheck_DockerRun(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("docker run nginx")
	if d != rules.NoOpinion {
		t.Error("docker run should be no-opinion")
	}
}

func TestCheck_Npm(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("npm install")
	if d != rules.NoOpinion {
		t.Error("npm should be no-opinion")
	}
}

func TestCheck_Python(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("python -c 'import os; os.remove(\"x\")'")
	if d != rules.NoOpinion {
		t.Error("python should be no-opinion")
	}
}

func TestCheck_Curl(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("curl https://example.com")
	if d != rules.NoOpinion {
		t.Error("curl should be no-opinion")
	}
}

func TestCheck_GoBuild(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("go build ./...")
	if d != rules.Allow {
		t.Error("go build ./... should be allow")
	}
}

func TestCheck_GoModTidy(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("go mod tidy")
	if d != rules.NoOpinion {
		t.Error("go mod tidy should be no-opinion")
	}
}

func TestCheck_GoTestC(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("go test -c ./...")
	if d != rules.NoOpinion {
		t.Error("go test -c should be no-opinion")
	}
}

func TestCheck_UnknownCommand(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("some_unknown_command --foo")
	if d != rules.NoOpinion {
		t.Error("unknown command should be no-opinion")
	}
}

func TestCheck_CommandSubstitution(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("echo $(whoami)")
	if d != rules.NoOpinion {
		t.Error("command substitution should be no-opinion")
	}
}

func TestCheck_Subshell(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("(cd /tmp && ls)")
	if d != rules.NoOpinion {
		t.Error("subshell should be no-opinion")
	}
}

// ── CD path tracking ───────────────────────────────────────

func TestCheck_CDEscapesRoot(t *testing.T) {
	c := newTestChecker()

	tests := []string{
		"cd /tmp && ls",
		"cd .. && ls",
		"cd ../.. && ls",
		"cd / && ls",
		"cd /home/user && ls", // parent of project
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			d, _, _ := c.Check(input)
			if d != rules.NoOpinion {
				t.Error("cd escaping root should be no-opinion")
			}
		})
	}
}

func TestCheck_CDIntoSubdir(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("cd src && ls")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("cd into subdir within root should be allow")
	}
}

func TestCheck_CDRoundTrip(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("cd src && ls && cd ..")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("cd round trip within root should be allow")
	}
}

func TestCheck_CDTilde(t *testing.T) {
	c := newTestChecker()
	// ~ resolves to /home/user which is NOT /home/user/project
	d, _, _ := c.Check("cd ~ && ls")
	if d != rules.NoOpinion {
		t.Error("cd ~ should escape root")
	}
}

func TestCheck_CDTildeSubdir(t *testing.T) {
	c := newTestChecker()
	// ~/project resolves to /home/user/project which IS the root
	d, _, err := c.Check("cd ~/project && ls")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("cd ~/project should stay within root")
	}
}

func TestCheck_CDDynamicPath(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("cd $DIR && ls")
	if d != rules.NoOpinion {
		t.Error("cd $DIR should be no-opinion (dynamic)")
	}
}

func TestCheck_CDBack(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("cd - && ls")
	if d != rules.NoOpinion {
		t.Error("cd - should be no-opinion (dynamic)")
	}
}

func TestCheck_CDBare_HomeWithinRoot(t *testing.T) {
	c := &Checker{
		RootDir: "/home/user",
		HomeDir: "/home/user",
		Rules:   rules.Default,
	}
	d, _, err := c.Check("cd && ls")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("cd (bare) should be allow when home is within root")
	}
}

func TestCheck_CDBare_HomeOutsideRoot(t *testing.T) {
	c := newTestChecker() // root=/home/user/project, home=/home/user
	d, _, _ := c.Check("cd && ls")
	if d != rules.NoOpinion {
		t.Error("cd (bare) should be no-opinion when home escapes root")
	}
}

// ── Mixed: one bad command spoils the whole chain ──────────

func TestCheck_GoodThenBad(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("git status && rm -rf /")
	if d != rules.NoOpinion {
		t.Error("chain with one mutable command should be no-opinion")
	}
}

func TestCheck_LsThenGitPush(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("ls && git push")
	if d != rules.NoOpinion {
		t.Error("chain with git push should be no-opinion")
	}
}

func TestCheck_OutputRedirect(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("echo hello > file.txt")
	if d != rules.NoOpinion {
		t.Error("output redirect should be no-opinion")
	}
}

func TestCheck_AppendRedirect(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("echo hello >> file.txt")
	if d != rules.NoOpinion {
		t.Error("append redirect should be no-opinion")
	}
}

func TestCheck_DupToFile(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("echo hello >& /tmp/file")
	if d != rules.NoOpinion {
		t.Error("redirect to file should be no-opinion")
	}
}

func TestCheck_DupFd(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go test ./... 2>&1")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("2>&1 fd duplication should be allow")
	}
}

func TestCheck_InputRedirect(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("grep pattern < input.txt")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("input redirect should be allow")
	}
}

func TestCheck_EnvCommand(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("env FOO=bar rm -rf /")
	if d != rules.NoOpinion {
		t.Error("env with command should be no-opinion")
	}
}

func TestCheck_Printenv(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("printenv HOME")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("printenv should be allow")
	}
}

func TestCheck_GoTestCount(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go test -count=1 ./...")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("go test -count=1 should be allow")
	}
}

func TestCheck_GoTestCover(t *testing.T) {
	c := newTestChecker()
	d, _, err := c.Check("go test -cover ./...")
	assertNoError(t, err)
	if d != rules.Allow {
		t.Error("go test -cover should be allow")
	}
}

// ── Deny from config ───────────────────────────────────────

func TestCheck_DenyFromRule(t *testing.T) {
	c := &Checker{
		RootDir: "/home/user/project",
		HomeDir: "/home/user",
		Rules: map[string]*rules.Rule{
			"curl": {
				Default: rules.Deny,
				Message: "curl is not allowed in this project",
			},
		},
	}
	d, reason, _ := c.Check("curl https://example.com")
	if d != rules.Deny {
		t.Error("curl should be denied")
	}
	if reason == "" {
		t.Error("expected deny reason")
	}
}

func TestCheck_DenyShortCircuits(t *testing.T) {
	c := &Checker{
		RootDir: "/home/user/project",
		HomeDir: "/home/user",
		Rules: map[string]*rules.Rule{
			"ls": {Default: rules.Allow},
			"curl": {
				Default: rules.Deny,
				Message: "curl is not allowed in this project",
			},
		},
	}
	d, reason, _ := c.Check("ls && curl https://example.com && ls")
	if d != rules.Deny {
		t.Error("should be denied because of curl")
	}
	if reason == "" {
		t.Error("expected deny reason")
	}
}

func TestCheck_DenyWithSubcommandOverride(t *testing.T) {
	c := &Checker{
		RootDir: "/home/user/project",
		HomeDir: "/home/user",
		Rules: map[string]*rules.Rule{
			"kubectl": {
				Default: rules.Deny,
				Message: "kubectl is blocked",
				Subcommands: map[string]*rules.Rule{
					"get": {Default: rules.Allow},
				},
			},
		},
	}

	// get matches subcommand, should be allowed
	d, _, _ := c.Check("kubectl get pods")
	if d != rules.Allow {
		t.Error("kubectl get should be allowed via subcommand")
	}

	// exec doesn't match subcommand, falls to deny
	d, reason, _ := c.Check("kubectl exec -it my-pod -- bash")
	if d != rules.Deny {
		t.Error("kubectl exec should be denied")
	}
	if reason == "" {
		t.Error("expected deny reason")
	}
}

// ── Edge cases ─────────────────────────────────────────────

func TestCheck_Empty(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("")
	if d != rules.NoOpinion {
		t.Error("empty input should be no-opinion")
	}
}

func TestCheck_Whitespace(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("   ")
	if d != rules.NoOpinion {
		t.Error("whitespace input should be no-opinion")
	}
}

func TestCheck_InvalidBash(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("if if if")
	if d != rules.NoOpinion {
		t.Error("invalid bash should be no-opinion")
	}
}

func TestCheck_DynamicCommandName(t *testing.T) {
	c := newTestChecker()
	d, _, _ := c.Check("$CMD args")
	// This may or may not parse cleanly depending on tree-sitter,
	// but if it does parse, $CMD should be denied
	if d == rules.Allow {
		t.Error("dynamic command name should not be allow")
	}
}

// ── helpers ────────────────────────────────────────────────

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
