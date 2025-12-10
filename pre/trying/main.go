package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"go/types"
	"os"
	"slices"
	"strings"

	"github.com/theaino/gotk/lib"
	"golang.org/x/tools/go/ast/astutil"
)

type trying struct {
	source string
	fset *token.FileSet
	file *ast.File
	info *types.Info
	tryingPositions []int
	tryingExprs []ast.Expr
	tryingBlocks []*ast.BlockStmt
	finishedExprs map[ast.Expr]bool
}

func main() {
	args, err := lib.GetArgs()
	if err != nil {
		panic(err)
	}

	t := trying{
		source: args.Source,
		finishedExprs: make(map[ast.Expr]bool),
	}

	t.getCharPositions()
	t.replacePositions()
	err = t.parseSource()
	if err != nil {
		panic(err)
	}
	err = t.findWrappedExprs()
	if err != nil {
		panic(err)
	}
	err = t.loadTypes(args.Root)
	if err != nil {
		panic(err)
	}
	t.findSurroundingBlocks()
	t.modifyAst()

	err = printer.Fprint(os.Stdout, t.fset, t.file)
	if err != nil {
		fmt.Print(err)
	}
}

