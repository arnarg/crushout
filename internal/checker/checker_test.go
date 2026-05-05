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

// ── Success cases: IsReadOnly returns true ─────────────────

func TestIsReadOnly_SimpleLs(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("ls -la")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitStatus(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git status")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitDiff(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git diff HEAD~3")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitLog(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git log --oneline -10")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitRemoteV(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git remote -v")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitRemoteShow(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git remote show origin")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitBranchList(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git branch -l")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitBranchShowCurrent(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git branch --show-current")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitTagList(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git tag -l")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitStashList(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git stash list")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitSubmoduleStatus(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git submodule status")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_Pipeline(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("cat README.md | grep -i todo | sort | uniq")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_CompoundList(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git status && git diff && ls -la")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_Find(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("find . -name '*.go'")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_Sed(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("sed 's/old/new/g' file.txt")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GoTest(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go test ./...")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GoModGraph(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go mod graph")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GoEnv(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go env GOPATH")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GoToolPprof(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go tool pprof cpu.prof")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_PwdWhoamiUname(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("pwd && whoami && uname -a")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_Echo(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("echo hello")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_CDWithinRoot(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("cd src && ls -la && cd ..")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

func TestIsReadOnly_GitConfigGet(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("git config get user.name")
	assertNoError(t, err)
	if !ok {
		t.Error("expected read-only")
	}
}

// ── Failure cases: IsReadOnly returns false ────────────────

func TestIsReadOnly_Rm(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("rm -rf /")
	if ok {
		t.Error("rm should not be read-only")
	}
}

func TestIsReadOnly_SudoRm(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("sudo rm -rf /")
	if ok {
		t.Error("sudo should not be read-only")
	}
}

func TestIsReadOnly_GitPush(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git push")
	if ok {
		t.Error("git push should not be read-only")
	}
}

func TestIsReadOnly_GitPushOrigin(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git push origin main")
	if ok {
		t.Error("git push origin main should not be read-only")
	}
}

func TestIsReadOnly_GitCommit(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git commit -m 'fix'")
	if ok {
		t.Error("git commit should not be read-only")
	}
}

func TestIsReadOnly_GitCheckout(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git checkout -b feature")
	if ok {
		t.Error("git checkout should not be read-only")
	}
}

func TestIsReadOnly_GitRemoteAdd(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git remote add origin git@github.com:user/repo.git")
	if ok {
		t.Error("git remote add should not be read-only")
	}
}

func TestIsReadOnly_GitBranchDelete(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git branch -D old-branch")
	if ok {
		t.Error("git branch -D should not be read-only")
	}
}

func TestIsReadOnly_GitBranchCreate(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git branch new-feature")
	if ok {
		t.Error("git branch new-feature should not be read-only")
	}
}

func TestIsReadOnly_GitTagCreate(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git tag v1.0.0")
	if ok {
		t.Error("git tag v1.0.0 should not be read-only")
	}
}

func TestIsReadOnly_GitStashBare(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git stash")
	if ok {
		t.Error("git stash (bare) should not be read-only")
	}
}

func TestIsReadOnly_GitStashPop(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git stash pop")
	if ok {
		t.Error("git stash pop should not be read-only")
	}
}

func TestIsReadOnly_GitC(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git -C /tmp status")
	if ok {
		t.Error("git -C should not be read-only")
	}
}

func TestIsReadOnly_GitConfigGlobal(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git config --global user.name New")
	if ok {
		t.Error("git config --global should not be read-only")
	}
}

func TestIsReadOnly_FindExec(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("find . -name '*.go' -exec rm {} \\;")
	if ok {
		t.Error("find -exec should not be read-only")
	}
}

func TestIsReadOnly_SedInPlace(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("sed -i 's/old/new/g' file.txt")
	if ok {
		t.Error("sed -i should not be read-only")
	}
}

func TestIsReadOnly_Make(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("make build")
	if ok {
		t.Error("make should not be read-only")
	}
}

func TestIsReadOnly_Docker(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("docker ps")
	assertNoError(t, err)
	if !ok {
		t.Error("docker ps should be read-only")
	}
}

func TestIsReadOnly_DockerRun(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("docker run nginx")
	if ok {
		t.Error("docker run should not be read-only")
	}
}

func TestIsReadOnly_Npm(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("npm install")
	if ok {
		t.Error("npm should not be read-only")
	}
}

func TestIsReadOnly_Python(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("python -c 'import os; os.remove(\"x\")'")
	if ok {
		t.Error("python should not be read-only")
	}
}

func TestIsReadOnly_Curl(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("curl https://example.com")
	if ok {
		t.Error("curl should not be read-only")
	}
}

func TestIsReadOnly_GoBuild(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("go build ./...")
	if !ok {
		t.Error("go build ./... should be read-only")
	}
}

func TestIsReadOnly_GoModTidy(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("go mod tidy")
	if ok {
		t.Error("go mod tidy should not be read-only")
	}
}

func TestIsReadOnly_GoTestC(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("go test -c ./...")
	if ok {
		t.Error("go test -c should not be read-only")
	}
}

func TestIsReadOnly_UnknownCommand(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("some_unknown_command --foo")
	if ok {
		t.Error("unknown command should not be read-only")
	}
}

func TestIsReadOnly_CommandSubstitution(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("echo $(whoami)")
	if ok {
		t.Error("command substitution should not be read-only")
	}
}

func TestIsReadOnly_Subshell(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("(cd /tmp && ls)")
	if ok {
		t.Error("subshell should not be read-only")
	}
}

// ── CD path tracking ───────────────────────────────────────

func TestIsReadOnly_CDEscapesRoot(t *testing.T) {
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
			ok, _ := c.IsReadOnly(input)
			if ok {
				t.Error("cd escaping root should not be read-only")
			}
		})
	}
}

func TestIsReadOnly_CDIntoSubdir(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("cd src && ls")
	assertNoError(t, err)
	if !ok {
		t.Error("cd into subdir within root should be read-only")
	}
}

func TestIsReadOnly_CDRoundTrip(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("cd src && ls && cd ..")
	assertNoError(t, err)
	if !ok {
		t.Error("cd round trip within root should be read-only")
	}
}

func TestIsReadOnly_CDTilde(t *testing.T) {
	c := newTestChecker()
	// ~ resolves to /home/user which is NOT /home/user/project
	ok, _ := c.IsReadOnly("cd ~ && ls")
	if ok {
		t.Error("cd ~ should escape root")
	}
}

func TestIsReadOnly_CDTildeSubdir(t *testing.T) {
	c := newTestChecker()
	// ~/project resolves to /home/user/project which IS the root
	ok, err := c.IsReadOnly("cd ~/project && ls")
	assertNoError(t, err)
	if !ok {
		t.Error("cd ~/project should stay within root")
	}
}

func TestIsReadOnly_CDDynamicPath(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("cd $DIR && ls")
	if ok {
		t.Error("cd $DIR should not be read-only (dynamic)")
	}
}

func TestIsReadOnly_CDBack(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("cd - && ls")
	if ok {
		t.Error("cd - should not be read-only (dynamic)")
	}
}

func TestIsReadOnly_CDBare_HomeWithinRoot(t *testing.T) {
	c := &Checker{
		RootDir: "/home/user",
		HomeDir: "/home/user",
		Rules:   rules.Default,
	}
	ok, err := c.IsReadOnly("cd && ls")
	assertNoError(t, err)
	if !ok {
		t.Error("cd (bare) should be read-only when home is within root")
	}
}

func TestIsReadOnly_CDBare_HomeOutsideRoot(t *testing.T) {
	c := newTestChecker() // root=/home/user/project, home=/home/user
	ok, _ := c.IsReadOnly("cd && ls")
	if ok {
		t.Error("cd (bare) should not be read-only when home escapes root")
	}
}

// ── Mixed: one bad command spoils the whole chain ──────────

func TestIsReadOnly_GoodThenBad(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("git status && rm -rf /")
	if ok {
		t.Error("chain with one mutable command should not be read-only")
	}
}

func TestIsReadOnly_LsThenGitPush(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("ls && git push")
	if ok {
		t.Error("chain with git push should not be read-only")
	}
}

func TestIsReadOnly_OutputRedirect(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("echo hello > file.txt")
	if ok {
		t.Error("output redirect should not be read-only")
	}
}

func TestIsReadOnly_AppendRedirect(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("echo hello >> file.txt")
	if ok {
		t.Error("append redirect should not be read-only")
	}
}

func TestIsReadOnly_DupToFile(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("echo hello >& /tmp/file")
	if ok {
		t.Error("redirect to file should not be read-only")
	}
}

func TestIsReadOnly_DupFd(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go test ./... 2>&1")
	assertNoError(t, err)
	if !ok {
		t.Error("2>&1 fd duplication should be read-only")
	}
}

func TestIsReadOnly_InputRedirect(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("grep pattern < input.txt")
	assertNoError(t, err)
	if !ok {
		t.Error("input redirect should be read-only")
	}
}

func TestIsReadOnly_EnvCommand(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("env FOO=bar rm -rf /")
	if ok {
		t.Error("env with command should not be read-only")
	}
}

func TestIsReadOnly_Printenv(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("printenv HOME")
	assertNoError(t, err)
	if !ok {
		t.Error("printenv should be read-only")
	}
}

func TestIsReadOnly_GoTestCount(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go test -count=1 ./...")
	assertNoError(t, err)
	if !ok {
		t.Error("go test -count=1 should be read-only")
	}
}

func TestIsReadOnly_GoTestCover(t *testing.T) {
	c := newTestChecker()
	ok, err := c.IsReadOnly("go test -cover ./...")
	assertNoError(t, err)
	if !ok {
		t.Error("go test -cover should be read-only")
	}
}

// ── Edge cases ─────────────────────────────────────────────

func TestIsReadOnly_Empty(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("")
	if ok {
		t.Error("empty input should not be read-only")
	}
}

func TestIsReadOnly_Whitespace(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("   ")
	if ok {
		t.Error("whitespace input should not be read-only")
	}
}

func TestIsReadOnly_InvalidBash(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("if if if")
	if ok {
		t.Error("invalid bash should not be read-only")
	}
}

func TestIsReadOnly_DynamicCommandName(t *testing.T) {
	c := newTestChecker()
	ok, _ := c.IsReadOnly("$CMD args")
	// This may or may not parse cleanly depending on tree-sitter,
	// but if it does parse, $CMD should be denied
	if ok {
		t.Error("dynamic command name should not be read-only")
	}
}

// ── helpers ────────────────────────────────────────────────

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
