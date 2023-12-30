package main

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/prabhatexit0/vis-call/astutils"
)


func main() {
	if len(os.Args) != 2 {
		fmt.Println("Error: invalid arguments passed")
		return
	}

	filePath := os.Args[1]
	jsAstString, err := astutils.GenerateJsAst(filePath)
	if err != nil {
		fmt.Println("Error: Could not parse the provided ast file")
		return
	}

	fmt.Println("======== JavaScript AST ========")
	var jsAstMap map[string]interface{}
	err = json.Unmarshal([]byte(jsAstString), &jsAstMap)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(jsAstMap)
}
