package main

import (
	"context"
	"fmt"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"os"
	"path/filepath"
)

func serverDotSh() {}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Error: Invalid number of arguments passed")
		os.Exit(1)
	}

	filePath := os.Args[1]
	ext := filepath.Ext(filePath)
	ctx := context.Background()

	language := map[string]*sitter.Language{
		".js": javascript.GetLanguage(),
	}

	parser := sitter.NewParser()
	if lang, ok := language[ext]; ok {
		parser.SetLanguage(lang)
	} else {
		fmt.Println("Error: Can not detect the language or the language is not supported")
		os.Exit(1)
	}

	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	ast, err := parser.ParseCtx(ctx, nil, sourceCode)
	if err != nil {
		fmt.Printf("Error: Can not parse the source code into AST - %v", err)
	}

	fmt.Println("----- AST -----")
	fmt.Println(ast.RootNode().Content(sourceCode))

	fmt.Println("----- idk why but keeping it here -----")
	fmt.Println(ast.RootNode().Child(0).Child(0).Content(sourceCode))
	fmt.Println(ast.RootNode().Child(0).Child(1).Type())
	fmt.Println(ast.RootNode().Child(0).Child(1).Content(sourceCode))
}
