package bash

import (
	"strings"

	gotreesitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// Command is a single command extracted from a bash AST.
type Command struct {
	Name string   // binary name (e.g. "git")
	Args []string // arguments (field name="argument" children)
	Raw  string   // full raw text of the AST node
}

// Result holds extracted commands and parse metadata.
type Result struct {
	Commands    []Command
	HasError    bool // tree has parse errors
	IsComplex   bool // contains $(), <(), (), $(())
	HasRedirect bool // contains output redirections (>, >>, &>, etc.)
}

// Parse extracts commands from a bash string.
func Parse(source string) (*Result, error) {
	lang := grammars.BashLanguage()
	parser := gotreesitter.NewParser(lang)
	tree, err := parser.Parse([]byte(source))
	if err != nil {
		return nil, err
	}
	defer tree.Release()

	root := tree.RootNode()
	src := []byte(source)

	result := &Result{
		HasError:  root.HasError(),
		IsComplex: isComplex(root, lang),
	}

	if !result.HasError && !result.IsComplex {
		walkNode(root, lang, src, result)
	}

	return result, nil
}

// ── AST traversal ──────────────────────────────────────────

var complexTypes = map[string]bool{
	"command_substitution": true,
	"process_substitution": true,
	"subshell":             true,
	"arithmetic_expansion": true,
}

func isComplex(node *gotreesitter.Node, lang *gotreesitter.Language) bool {
	if node == nil {
		return false
	}
	if complexTypes[node.Type(lang)] {
		return true
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		if isComplex(node.Child(i), lang) {
			return true
		}
	}
	return false
}

func walkNode(node *gotreesitter.Node, lang *gotreesitter.Language, src []byte, result *Result) {
	if node == nil {
		return
	}

	switch node.Type(lang) {
	case "command":
		result.Commands = append(result.Commands, extractCommand(node, lang, src))

	case "declaration_command", "unset_command":
		raw := node.Text(src)
		keyword := strings.Fields(raw)[0]
		result.Commands = append(result.Commands, Command{Name: keyword, Raw: raw})

	case "test_command":
		result.Commands = append(result.Commands, Command{Name: "test", Raw: node.Text(src)})

	case "program", "list", "pipeline",
		"subshell", "compound_statement",
		"if_statement", "elif_clause", "else_clause",
		"for_statement", "c_style_for_statement",
		"while_statement", "case_statement", "case_item",
		"do_group", "function_definition",
		"negated_command":
		recurseNamed(node, lang, src, result)

	case "redirected_statement":
		if hasOutputRedirect(node, lang, src) {
			result.HasRedirect = true
			return
		}
		body := node.ChildByFieldName("body", lang)
		if body != nil {
			walkNode(body, lang, src, result)
		} else {
			recurseNamed(node, lang, src, result)
		}
	}
}

func recurseNamed(node *gotreesitter.Node, lang *gotreesitter.Language, src []byte, result *Result) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		walkNode(node.NamedChild(i), lang, src, result)
	}
}

var outputRedirectOps = map[string]bool{
	">":  true,
	">>": true,
	"&>": true,
	">&": true,
}

func hasOutputRedirect(node *gotreesitter.Node, lang *gotreesitter.Language, src []byte) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type(lang) != "file_redirect" {
			continue
		}
		op, target := redirectParts(child, lang, src)
		if !outputRedirectOps[op] {
			continue
		}
		if op == ">&" && target.t == "number" {
			continue
		}
		return true
	}
	return false
}

type redirectPart struct {
	text string
	t    string
}

func redirectParts(redirect *gotreesitter.Node, lang *gotreesitter.Language, src []byte) (string, redirectPart) {
	var op string
	var target redirectPart
	for i := 0; i < int(redirect.ChildCount()); i++ {
		c := redirect.Child(i)
		if c.IsNamed() {
			if c.Type(lang) == "file_descriptor" {
				continue
			}
			target = redirectPart{text: c.Text(src), t: c.Type(lang)}
		} else {
			op = c.Text(src)
		}
	}
	return op, target
}

func extractCommand(node *gotreesitter.Node, lang *gotreesitter.Language, src []byte) Command {
	nameNode := node.ChildByFieldName("name", lang)
	if nameNode == nil {
		return Command{Raw: node.Text(src)}
	}
	name := nameNode.Text(src)

	var args []string
	for i := 0; i < int(node.ChildCount()); i++ {
		if arg, ok := extractChild(node, i, lang, src); ok {
			args = append(args, arg)
		}
	}

	return Command{Name: name, Args: args, Raw: node.Text(src)}
}

func extractChild(node *gotreesitter.Node, index int, lang *gotreesitter.Language, src []byte) (string, bool) {
	child := node.Child(index)

	if child.IsNamed() && node.FieldNameForChild(index, lang) == "argument" {
		return child.Text(src), true
	}

	return "", false
}
