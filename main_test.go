package main

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

func parseJS(t *testing.T, source string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return tree.RootNode()
}

func TestExtractFunctions_Declaration(t *testing.T) {
	source := `function foo() { bar(); }`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	if funcs[0].Name != "foo" {
		t.Errorf("expected name 'foo', got %q", funcs[0].Name)
	}
	if len(funcs[0].Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(funcs[0].Calls))
	}
	if funcs[0].Calls[0].Name != "bar" {
		t.Errorf("expected call to 'bar', got %q", funcs[0].Calls[0].Name)
	}
}

func TestExtractFunctions_ArrowFunction(t *testing.T) {
	source := `const greet = () => { hello(); };`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	if funcs[0].Name != "greet" {
		t.Errorf("expected name 'greet', got %q", funcs[0].Name)
	}
	if len(funcs[0].Calls) != 1 || funcs[0].Calls[0].Name != "hello" {
		t.Errorf("expected call to 'hello', got %v", funcs[0].Calls)
	}
}

func TestExtractFunctions_FunctionExpression(t *testing.T) {
	source := `const greet = function() { hello(); };`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	if funcs[0].Name != "greet" {
		t.Errorf("expected name 'greet', got %q", funcs[0].Name)
	}
}

func TestExtractFunctions_Multiple(t *testing.T) {
	source := `
function a() {}
function b() { a(); }
function c() { a(); b(); }
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(funcs))
	}
	if len(funcs[0].Calls) != 0 {
		t.Errorf("expected 0 calls in a(), got %d", len(funcs[0].Calls))
	}
	if len(funcs[1].Calls) != 1 {
		t.Errorf("expected 1 call in b(), got %d", len(funcs[1].Calls))
	}
	if len(funcs[2].Calls) != 2 {
		t.Errorf("expected 2 calls in c(), got %d", len(funcs[2].Calls))
	}
}

func TestProbableCallDetection(t *testing.T) {
	source := `
function doStuff(flag) {
    always();
    if (flag) {
        maybe();
    }
}
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}

	calls := funcs[0].Calls
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}

	// always() should be definite
	if calls[0].Name != "always" || calls[0].Probable {
		t.Errorf("expected 'always' to be definite, got name=%q probable=%v", calls[0].Name, calls[0].Probable)
	}
	// maybe() should be probable
	if calls[1].Name != "maybe" || !calls[1].Probable {
		t.Errorf("expected 'maybe' to be probable, got name=%q probable=%v", calls[1].Name, calls[1].Probable)
	}
}

func TestProbableCall_SwitchAndLoop(t *testing.T) {
	source := `
function process(x) {
    switch(x) {
        case 1: handleOne(); break;
    }
    for (let i = 0; i < 10; i++) {
        tick();
    }
    while (true) {
        spin();
    }
}
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	calls := funcs[0].Calls
	for _, c := range calls {
		if !c.Probable {
			t.Errorf("expected %q to be probable (inside control flow), got definite", c.Name)
		}
	}
}

func TestDocExtraction(t *testing.T) {
	source := `
/** Adds two numbers. */
function add(a, b) { return a + b; }
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}
	if funcs[0].Doc != "Adds two numbers." {
		t.Errorf("expected doc 'Adds two numbers.', got %q", funcs[0].Doc)
	}
}

func TestDocExtraction_NonJSDoc(t *testing.T) {
	source := `
// regular comment
function noop() {}
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if funcs[0].Doc != "" {
		t.Errorf("expected no doc for non-JSDoc comment, got %q", funcs[0].Doc)
	}
}

func TestMemberExpressionCalls(t *testing.T) {
	source := `
function log() {
    console.log("hi");
    obj.method();
}
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	calls := funcs[0].Calls
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Name != "console.log" {
		t.Errorf("expected 'console.log', got %q", calls[0].Name)
	}
	if calls[1].Name != "obj.method" {
		t.Errorf("expected 'obj.method', got %q", calls[1].Name)
	}
}

func TestBuildFuncSet(t *testing.T) {
	funcs := []FuncDecl{
		{Name: "alpha"},
		{Name: "beta"},
	}
	set := buildFuncSet(funcs)

	if !set["alpha"] || !set["beta"] {
		t.Errorf("expected both functions in set")
	}
	if set["gamma"] {
		t.Errorf("unexpected function in set")
	}
}

func TestNoFunctions(t *testing.T) {
	source := `const x = 42;`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	if len(funcs) != 0 {
		t.Errorf("expected 0 functions, got %d", len(funcs))
	}
}

func TestNestedCalls(t *testing.T) {
	source := `function wrap() { outer(inner()); }`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))

	calls := funcs[0].Calls
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls (outer + inner), got %d", len(calls))
	}

	names := map[string]bool{}
	for _, c := range calls {
		names[c.Name] = true
	}
	if !names["outer"] || !names["inner"] {
		t.Errorf("expected both 'outer' and 'inner', got %v", names)
	}
}

func TestTryCatch(t *testing.T) {
	source := `
function risky() {
    safe();
    try {
        dangerous();
    } catch(e) {
        recover(e);
    }
}
`
	root := parseJS(t, source)
	funcs := extractFunctions(root, []byte(source))
	calls := funcs[0].Calls

	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
	if calls[0].Name != "safe" || calls[0].Probable {
		t.Errorf("expected 'safe' definite, got name=%q probable=%v", calls[0].Name, calls[0].Probable)
	}
	if calls[1].Name != "dangerous" || !calls[1].Probable {
		t.Errorf("expected 'dangerous' probable, got name=%q probable=%v", calls[1].Name, calls[1].Probable)
	}
	if calls[2].Name != "recover" || !calls[2].Probable {
		t.Errorf("expected 'recover' probable, got name=%q probable=%v", calls[2].Name, calls[2].Probable)
	}
}
