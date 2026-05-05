package bash

import (
	"testing"
)

func TestParse_SimpleCommand(t *testing.T) {
	result, err := Parse("ls -la /tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasError {
		t.Fatal("unexpected parse error")
	}
	if result.IsComplex {
		t.Fatal("should not be complex")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	cmd := result.Commands[0]
	assertCommand(t, cmd, "ls", []string{"-la", "/tmp"})
}

func TestParse_Pipeline(t *testing.T) {
	result, err := Parse("cat file.txt | grep foo | sort | uniq")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 4 {
		t.Fatalf("expected 4 commands, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "cat", []string{"file.txt"})
	assertCommand(t, result.Commands[1], "grep", []string{"foo"})
	assertCommand(t, result.Commands[2], "sort", nil)
	assertCommand(t, result.Commands[3], "uniq", nil)
}

func TestParse_List(t *testing.T) {
	result, err := Parse("git status && git diff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "git", []string{"status"})
	assertCommand(t, result.Commands[1], "git", []string{"diff"})
}

func TestParse_ListOrPipeline(t *testing.T) {
	result, err := Parse("git status && git diff || cat file.txt | grep foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 4 {
		t.Fatalf("expected 4 commands, got %d: %+v", len(result.Commands), result.Commands)
	}
	assertCommand(t, result.Commands[0], "git", []string{"status"})
	assertCommand(t, result.Commands[1], "git", []string{"diff"})
	assertCommand(t, result.Commands[2], "cat", []string{"file.txt"})
	assertCommand(t, result.Commands[3], "grep", []string{"foo"})
}

func TestParse_OutputRedirect(t *testing.T) {
	result, err := Parse("cat file.txt > output.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasRedirect {
		t.Fatal("expected HasRedirect for output redirection")
	}
	if len(result.Commands) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(result.Commands))
	}
}

func TestParse_AppendRedirect(t *testing.T) {
	result, err := Parse("echo hello >> output.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasRedirect {
		t.Fatal("expected HasRedirect for append redirection")
	}
}

func TestParse_DupToFile(t *testing.T) {
	result, err := Parse("echo hello >& /tmp/file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasRedirect {
		t.Fatal(">& to file should set HasRedirect")
	}
}

func TestParse_DupFd(t *testing.T) {
	result, err := Parse("echo hello 2>&1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasRedirect {
		t.Fatal("2>&1 should not set HasRedirect")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "echo", []string{"hello"})
}

func TestParse_CombinedRedirect(t *testing.T) {
	result, err := Parse("go test ./... 2>&1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasRedirect {
		t.Fatal("2>&1 should not set HasRedirect")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "go", []string{"test", "./..."})
}

func TestParse_InputRedirect(t *testing.T) {
	result, err := Parse("grep pattern < input.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasRedirect {
		t.Fatal("input redirect should not set HasRedirect")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "grep", []string{"pattern"})
}

func TestParse_NegatedCommand(t *testing.T) {
	result, err := Parse("! grep -q pattern file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "grep", []string{"-q", "pattern", "file.txt"})
}

func TestParse_DeclarationCommand(t *testing.T) {
	result, err := Parse("export FOO=bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	if result.Commands[0].Name != "export" {
		t.Errorf("expected name 'export', got %q", result.Commands[0].Name)
	}
}

func TestParse_TestCommand(t *testing.T) {
	result, err := Parse("[ -f file.txt ]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	if result.Commands[0].Name != "test" {
		t.Errorf("expected name 'test', got %q", result.Commands[0].Name)
	}
}

func TestParse_GitSubcommands(t *testing.T) {
	tests := []struct {
		input string
		name  string
		args  []string
	}{
		{"git status", "git", []string{"status"}},
		{"git diff HEAD~3", "git", []string{"diff", "HEAD~3"}},
		{"git log --oneline -10", "git", []string{"log", "--oneline", "-10"}},
		{"git remote -v", "git", []string{"remote", "-v"}},
		{"git remote show origin", "git", []string{"remote", "show", "origin"}},
		{"git branch -l", "git", []string{"branch", "-l"}},
		{"git stash list", "git", []string{"stash", "list"}},
		{"git push origin main", "git", []string{"push", "origin", "main"}},
		{"git commit -m 'fix'", "git", []string{"commit", "-m", "'fix'"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Commands) != 1 {
				t.Fatalf("expected 1 command, got %d", len(result.Commands))
			}
			assertCommand(t, result.Commands[0], tt.name, tt.args)
		})
	}
}

func TestParse_CommandSubstitution_IsComplex(t *testing.T) {
	result, err := Parse("echo $(whoami)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsComplex {
		t.Error("expected IsComplex for command substitution")
	}
	if len(result.Commands) != 0 {
		t.Error("expected no commands extracted from complex input")
	}
}

func TestParse_Subshell_IsComplex(t *testing.T) {
	result, err := Parse("(cd /tmp && ls)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsComplex {
		t.Error("expected IsComplex for subshell")
	}
}

func TestParse_ProcessSubstitution_IsComplex(t *testing.T) {
	result, err := Parse("diff <(sort a.txt) <(sort b.txt)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsComplex {
		t.Error("expected IsComplex for process substitution")
	}
}

func TestParse_ArithmeticExpansion_IsComplex(t *testing.T) {
	result, err := Parse("echo $((1 + 2))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsComplex {
		t.Error("expected IsComplex for arithmetic expansion")
	}
}

func TestParse_InvalidInput_HasError(t *testing.T) {
	// Deeply mismatched constructs can produce parse errors
	result, err := Parse("if if if")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasError {
		t.Error("expected HasError for malformed input")
	}
}

func TestParse_EmptyInput(t *testing.T) {
	result, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(result.Commands))
	}
}

func TestParse_WhitespaceOnly(t *testing.T) {
	result, err := Parse("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(result.Commands))
	}
}

func TestParse_CompoundCommand(t *testing.T) {
	result, err := Parse("cd src && ls -la && cd ..")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "cd", []string{"src"})
	assertCommand(t, result.Commands[1], "ls", []string{"-la"})
	assertCommand(t, result.Commands[2], "cd", []string{".."})
}

func TestParse_GitWithFlagsBeforeSubcommand(t *testing.T) {
	result, err := Parse("git -C /tmp status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	// -C is an anonymous token before command_name; args should contain status
	cmd := result.Commands[0]
	if cmd.Name != "git" {
		t.Errorf("expected name 'git', got %q", cmd.Name)
	}
}

func TestParse_EchoWithFlags(t *testing.T) {
	result, err := Parse("echo -n 'hello world'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "echo", []string{"-n", "'hello world'"})
}

func TestParse_MultipleSemicolons(t *testing.T) {
	result, err := Parse("pwd; whoami; uname -a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(result.Commands))
	}
	assertCommand(t, result.Commands[0], "pwd", nil)
	assertCommand(t, result.Commands[1], "whoami", nil)
	assertCommand(t, result.Commands[2], "uname", []string{"-a"})
}

// ── helpers ────────────────────────────────────────────────

func assertCommand(t *testing.T, cmd Command, expectedName string, expectedArgs []string) {
	t.Helper()
	if cmd.Name != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, cmd.Name)
	}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("expected args %v, got %v", expectedArgs, cmd.Args)
		return
	}
	for i, arg := range expectedArgs {
		if cmd.Args[i] != arg {
			t.Errorf("args[%d]: expected %q, got %q", i, arg, cmd.Args[i])
		}
	}
}
