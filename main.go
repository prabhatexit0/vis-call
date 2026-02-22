package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

// CallExpr represents a function call found inside a function body.
type CallExpr struct {
	Name     string
	Line     uint32
	Probable bool // true if inside a conditional branch
}

// FuncDecl represents a parsed function declaration.
type FuncDecl struct {
	Name  string
	Line  uint32
	Doc   string
	Calls []CallExpr
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: viscall <file>\n")
		os.Exit(1)
	}

	filePath := os.Args[1]
	ext := filepath.Ext(filePath)

	languages := map[string]*sitter.Language{
		".js": javascript.GetLanguage(),
	}

	lang, ok := languages[ext]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unsupported file type %q\n", ext)
		os.Exit(1)
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing: %v\n", err)
		os.Exit(1)
	}

	root := tree.RootNode()
	funcs := extractFunctions(root, source)
	funcNames := buildFuncSet(funcs)
	printCallGraph(funcs, funcNames)
}

// extractFunctions walks the top-level AST and returns all function declarations.
func extractFunctions(root *sitter.Node, source []byte) []FuncDecl {
	var funcs []FuncDecl

	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)

		switch child.Type() {
		case "function_declaration":
			fd := parseFuncDecl(child, root, source)
			funcs = append(funcs, fd)

		case "lexical_declaration", "variable_declaration":
			// Handle: const foo = function() {} or const foo = () => {}
			for j := 0; j < int(child.NamedChildCount()); j++ {
				declarator := child.NamedChild(j)
				if declarator.Type() != "variable_declarator" {
					continue
				}
				nameNode := declarator.ChildByFieldName("name")
				valueNode := declarator.ChildByFieldName("value")
				if nameNode == nil || valueNode == nil {
					continue
				}
				if valueNode.Type() == "arrow_function" || valueNode.Type() == "function" {
					fd := parseFuncExpr(nameNode, valueNode, root, source)
					funcs = append(funcs, fd)
				}
			}
		}
	}

	return funcs
}

// parseFuncDecl extracts a FuncDecl from a function_declaration node.
func parseFuncDecl(node *sitter.Node, root *sitter.Node, source []byte) FuncDecl {
	nameNode := node.ChildByFieldName("name")
	bodyNode := node.ChildByFieldName("body")

	name := ""
	if nameNode != nil {
		name = nameNode.Content(source)
	}

	fd := FuncDecl{
		Name: name,
		Line: node.StartPoint().Row + 1,
		Doc:  extractDoc(node, source),
	}

	if bodyNode != nil {
		fd.Calls = extractCalls(bodyNode, bodyNode, source)
	}

	return fd
}

// parseFuncExpr extracts a FuncDecl from an arrow function or function expression.
func parseFuncExpr(nameNode, valueNode *sitter.Node, root *sitter.Node, source []byte) FuncDecl {
	name := nameNode.Content(source)
	bodyNode := valueNode.ChildByFieldName("body")

	fd := FuncDecl{
		Name: name,
		Line: nameNode.StartPoint().Row + 1,
		Doc:  extractDoc(nameNode.Parent().Parent(), source), // variable_declarator -> declaration
	}

	if bodyNode != nil {
		fd.Calls = extractCalls(bodyNode, bodyNode, source)
	}

	return fd
}

// extractDoc looks for a comment node immediately preceding the given node.
func extractDoc(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}

	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}

	text := prev.Content(source)
	// Only treat block comments starting with /** as doc comments
	if !strings.HasPrefix(text, "/**") {
		return ""
	}

	// Clean up the doc comment
	text = strings.TrimPrefix(text, "/**")
	text = strings.TrimSuffix(text, "*/")

	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, " ")
}

// conditionalTypes are AST node types that represent conditional control flow.
var conditionalTypes = map[string]bool{
	"if_statement":          true,
	"switch_statement":      true,
	"ternary_expression":    true,
	"try_statement":         true,
	"for_statement":         true,
	"for_in_statement":      true,
	"while_statement":       true,
	"do_statement":          true,
	"catch_clause":          true,
	"binary_expression":     true, // short-circuit &&, ||
	"logical_expression":    true, // some grammars use this
	"conditional_expression": true,
}

// extractCalls recursively walks a function body and collects call expressions.
func extractCalls(node *sitter.Node, boundary *sitter.Node, source []byte) []CallExpr {
	var calls []CallExpr

	var walk func(n *sitter.Node, inConditional bool)
	walk = func(n *sitter.Node, inConditional bool) {
		if n.Type() == "call_expression" {
			name := extractCallName(n, source)
			if name != "" {
				calls = append(calls, CallExpr{
					Name:     name,
					Line:     n.StartPoint().Row + 1,
					Probable: inConditional,
				})
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			childConditional := inConditional || conditionalTypes[n.Type()]
			walk(child, childConditional)
		}
	}

	walk(node, false)
	return calls
}

// extractCallName gets the function name from a call_expression node.
func extractCallName(node *sitter.Node, source []byte) string {
	fn := node.ChildByFieldName("function")
	if fn == nil {
		return ""
	}

	switch fn.Type() {
	case "identifier":
		return fn.Content(source)
	case "member_expression":
		return fn.Content(source)
	default:
		return fn.Content(source)
	}
}

// buildFuncSet returns a set of all declared function names.
func buildFuncSet(funcs []FuncDecl) map[string]bool {
	set := make(map[string]bool)
	for _, f := range funcs {
		set[f.Name] = true
	}
	return set
}

// printCallGraph prints each function and its call tree.
func printCallGraph(funcs []FuncDecl, funcNames map[string]bool) {
	if len(funcs) == 0 {
		fmt.Println("No functions found.")
		return
	}

	for i, f := range funcs {
		// Header
		header := fmt.Sprintf("%s", f.Name)
		if f.Doc != "" {
			header += fmt.Sprintf("  [doc: %s]", f.Doc)
		}
		header += fmt.Sprintf("  (line %d)", f.Line)
		fmt.Println(header)

		if len(f.Calls) == 0 {
			fmt.Println("  (no calls)")
		} else {
			// Deduplicate calls, keeping track of all lines
			type callInfo struct {
				name     string
				probable bool
				lines    []uint32
				external bool
			}
			seen := make(map[string]*callInfo)
			var order []string

			for _, c := range f.Calls {
				key := c.Name
				if info, ok := seen[key]; ok {
					info.lines = append(info.lines, c.Line)
					// If any call is definite, mark as definite
					if !c.Probable {
						info.probable = false
					}
				} else {
					seen[key] = &callInfo{
						name:     c.Name,
						probable: c.Probable,
						lines:    []uint32{c.Line},
						external: !funcNames[c.Name],
					}
					order = append(order, key)
				}
			}

			for j, key := range order {
				info := seen[key]
				isLast := j == len(order)-1

				prefix := "├── "
				if isLast {
					prefix = "└── "
				}

				label := info.name + "()"
				if info.probable {
					label += " [probable]"
				} else {
					label += " [definite]"
				}
				if info.external {
					label += " [external]"
				}

				// Show line numbers
				lineStrs := make([]string, len(info.lines))
				for k, l := range info.lines {
					lineStrs[k] = fmt.Sprintf("%d", l)
				}
				if len(info.lines) == 1 {
					label += fmt.Sprintf(" (line %s)", lineStrs[0])
				} else {
					label += fmt.Sprintf(" (lines %s)", strings.Join(lineStrs, ", "))
				}

				fmt.Println(prefix + label)
			}
		}

		if i < len(funcs)-1 {
			fmt.Println()
		}
	}
}
