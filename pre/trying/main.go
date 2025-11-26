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
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/theaino/gotk/lib"
)


func main() {
	args, err := lib.GetArgs()
	if err != nil {
		panic(err)
	}
	
	positions := getTryingCharPositions(args.Source)
	source := replacePositionsWithSpaces(args.Source, positions)
	fset, f, err := parseSource(source)
	if err != nil {
		panic(err)
	}
	exprs, err := findWrappedExprs(f, positions)
	if err != nil {
		panic(err)
	}
	exprTypes, err := getExprTypes(fset, f, slices.Collect(maps.Values(exprs)), args.Root)
	if err != nil {
		panic(err)
	}
	modifyAst(fset, f, slices.Collect(maps.Values(exprs)), exprTypes)

	fset = token.NewFileSet()
	err = printer.Fprint(os.Stdout, fset, f)
	if err != nil {
		fmt.Print(err)
	}
}

// Insert an error checking structure at each expr
func modifyAst(_ *token.FileSet, f *ast.File, exprs []ast.Expr, exprTypes map[ast.Expr]types.Type) {
	blocks := deepestBlocksWithExprs(f, exprs)
	for expr, block := range blocks {
		newStmts := make([]ast.Stmt, 0)
		for _, stmt := range block.List {
			newStmts = append(newStmts, stmt)
			ast.Inspect(stmt, func(n ast.Node) bool {
				if n != expr {
					return true
				}
				ifStmt := &ast.IfStmt{
					Cond: &ast.BinaryExpr{X: &ast.BasicLit{Value: "err"}, Op: token.NEQ, Y: &ast.BasicLit{Value: "nil"}},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{},
						},
					},
				}
				newStmts = append(newStmts, ifStmt)
				return true
			})
		}
		block.List = newStmts
	}
}

// BFS to find the deepest BlockStmt which includes a specific Expr
func deepestBlocksWithExprs(n ast.Node, exprs []ast.Expr) (blocks map[ast.Expr]*ast.BlockStmt) {
	blocks = make(map[ast.Expr]*ast.BlockStmt)
	nodeQueue := []ast.Node{n}
	for len(nodeQueue) > 0 {
		node := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		block, ok := node.(*ast.BlockStmt)
		if ok {
			ast.Inspect(block, func(n ast.Node) bool {
				nodeExpr, ok := n.(ast.Expr)
				if !ok {
					return true
				}
				for _, expr := range exprs {
					if nodeExpr == expr {
						blocks[expr] = block
					}
				}
				return true
			})
		}
		nodeQueue = append(nodeQueue, slices.Collect(ast.Preorder(node))[1:]...)
	}
	return
}

// Get the types for each expr
func getExprTypes(fset *token.FileSet, f *ast.File, exprs []ast.Expr, pkg string) (exprTypes map[ast.Expr]types.Type, err error) {
	info := &types.Info{
    Types: make(map[ast.Expr]types.TypeAndValue),
    Defs:  make(map[*ast.Ident]types.Object),
    Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{
		FakeImportC: true,
		Importer: importer.ForCompiler(fset, "source", nil),
		Error: func(err error) {
			fmt.Printf("%v\n", err)
		},
	}
	_, err = conf.Check(pkg, fset, []*ast.File{f}, info)
	if err != nil {
		return
	}
	exprTypes = make(map[ast.Expr]types.Type)
	for _, expr := range exprs {
		exprTypes[expr] = info.TypeOf(expr)
	}
	return
}

// Find each Expr which is is directly before a position
func findWrappedExprs(f *ast.File, positions []int) (exprs map[int]ast.Expr, err error) {
	exprs = make(map[int]ast.Expr)
	ast.Inspect(f, func(n ast.Node) bool {
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		for _, position := range positions {
			if int(expr.End()) > position + 1 {
				continue
			}
			wrappedExpr, ok := exprs[position]
			if !ok {
				exprs[position] = expr
				continue
			}
			if expr.End() > wrappedExpr.End() {
				exprs[position] = expr
			} else if expr.End() == wrappedExpr.End() && expr.Pos() < wrappedExpr.End() {
				exprs[position] = expr
			}
		}
		return true
	})
	return
}

// Parse a golang source
func parseSource(source string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", source, parser.AllErrors)
	return fset, f, err
}

// Replace a position array in the source with space characters
// This makes the resulting source be parsable
func replacePositionsWithSpaces(source string, positions []int) string {
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
func getTryingCharPositions(source string) (positions []int) {
	positions = make([]int, 0)
	fset := token.NewFileSet()
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

