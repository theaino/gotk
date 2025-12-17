package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"os"
	"slices"
	"strings"
)

type tryingFile struct {
	path string
	source string
	tryingPositions []int
}

// Parse function for go/packages.Load(...)
func parseTryingFile(pkg *tryingPackage, fset *token.FileSet, path string, source string) (*ast.File, error) {
	positions := charPositions(fset, source)
	source = replacePositions(source, positions)
	pkg.Positions[path] = positions
	return parser.ParseFile(fset, path, source, parser.AllErrors)
}

// Replace a position array in the source with space characters
// This makes the resulting source be parsable
func replacePositions(source string, positions []int) string {
	var builder strings.Builder
	var lastIdx int
	for _, pos := range positions {
		builder.WriteString(source[lastIdx:pos])
		builder.WriteString(" ")
		lastIdx = pos + 1
	}
	if lastIdx < len(source) {
		builder.WriteString(source[lastIdx:])
	}
	return builder.String()
}

// Get the '?' chars in the source
// Using the parsing errors which are at an '?' char
func charPositions(fset *token.FileSet, source string) (positions []int) {
	positions = make([]int, 0)
	_, errs := parser.ParseFile(fset, "", source, parser.AllErrors)
	if errs == nil {
		return
	}
	for _, err := range errs.(scanner.ErrorList) {
		char := source[err.Pos.Offset]
		if char != '?' {
			continue
		}
		if slices.Contains(positions, err.Pos.Offset) {
			continue
		}
		positions = append(positions, err.Pos.Offset)
	}
	slices.Sort(positions)
	return
}

func (p *tryingPackage) processFile(file *ast.File, path string) error {
	err := p.findWrappedExprs(path, file)
	if err != nil {
		return err
	}
	blocks := p.findSurroundingBlocks(path, file)
	p.modifyAst(path, blocks)
	return nil
}

func (p *tryingPackage) writeFile(file *ast.File, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return printer.Fprint(f, p.Package.Fset, file)
}
