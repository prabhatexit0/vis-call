package	astutils

import (
	"os"
	"os/exec"
	"bytes"
)

func GenerateJsAst(filepath string) (string, error) {
	// run node.js with the @babel/parser package to generate ast
	cmd := exec.Command("node", "scripts/js-ast.js", filepath)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
