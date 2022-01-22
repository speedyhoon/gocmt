package main

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/speedyhoon/gocmt/process"
)

// parseFile parses and modifies the input file if necessary. Returns AST represents of (new) source, a boolean
// to report whether the source file was modified, and any error if occurred.
func parseFile(fSet *token.FileSet, filePath, template string) (af *ast.File, modified bool, err error) {
	af, err = parser.ParseFile(fSet, filePath, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return
	}

	return process.File(af, fSet, commentBase, commentIndentedBase, template, *parenComment)
}
