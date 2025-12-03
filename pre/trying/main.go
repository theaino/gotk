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

// Insert an error checking structure at each expr
func (t *trying) modifyAst() {
	for _, block := range t.tryingBlocks {
		newStmts := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			hasTryingExpr := false
			stmt := astutil.Apply(stmt, nil, func(c *astutil.Cursor) bool {
				expr, ok := c.Node().(ast.Expr)
				if !ok {
					return true
				}
				if finished, _ := t.finishedExprs[expr]; finished {
					return true
				}
				idx := slices.Index(t.tryingExprs, expr)
				if idx == -1 {
					return true
				}
				resultTypes := make([]string, 0)
				switch exprType := t.info.TypeOf(expr).(type) {
				case *types.Tuple:
					for typeVar := range exprType.Variables() {
						resultTypes = append(resultTypes, typeVar.Type().String())
					}
				default:
					resultTypes = append(resultTypes, exprType.String())
				}
				resultFields := make([]*ast.Field, 0, len(resultTypes) - 1)
				for _, resultType := range resultTypes {
					if resultType == "error" {
						continue
					}
					resultFields = append(resultFields, &ast.Field{
						Type: &ast.Ident{Name: resultType},
					})
				}
				varExprs := make([]ast.Expr, len(resultTypes))
				varExprsWithoutErr := make([]ast.Expr, 0, len(resultFields))
				for idx, resultType := range resultTypes {
					name := "err"
					if resultType != "error" {
						name = fmt.Sprintf("tryingR%d", idx)
						varExprsWithoutErr = append(varExprsWithoutErr, &ast.Ident{Name: name})
					}
					varExprs[idx] = &ast.Ident{Name: name}
				}
				varDecls := make([]ast.Stmt, 0, len(resultFields))
				for idx, varExpr := range varExprs {
					if varExpr.(*ast.Ident).Name == "err" {
						continue
					}
					varDecls = append(varDecls, &ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{varExpr.(*ast.Ident)},
									Type: &ast.Ident{Name: resultTypes[idx]},
								},
							},
						},
					})
				}
				hasTryingExpr = true
				c.Replace(&ast.CallExpr{
					Fun: &ast.FuncLit{
						Type: &ast.FuncType{
							Results: &ast.FieldList{List: resultFields},
						},
						Body: &ast.BlockStmt{
							List: append(
								varDecls,
								&ast.AssignStmt{
									Tok: token.ASSIGN,
									Lhs: varExprs,
									Rhs: []ast.Expr{expr},
								},
								&ast.ReturnStmt{Results: varExprsWithoutErr},
							),
						},
					},
					Args: []ast.Expr{},
				})
				t.finishedExprs[expr] = true
				return true
			}).(ast.Stmt)
			newStmts = append(newStmts, stmt)
			if hasTryingExpr {
				newStmts = append(newStmts, &ast.IfStmt{
					Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{},
						},
					},
				})
			}
		}
		block.List = newStmts
	}
}


// BFS to find the deepest BlockStmt which includes a specific Expr
func (t *trying) findSurroundingBlocks() {
	t.tryingBlocks = make([]*ast.BlockStmt, len(t.tryingExprs))
	nodeQueue := []ast.Node{t.file}
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
				for idx, expr := range t.tryingExprs {
					if nodeExpr == expr {
						t.tryingBlocks[idx] = block
					}
				}
				return true
			})
		}
		nodeQueue = append(nodeQueue, slices.Collect(ast.Preorder(node))[1:]...)
	}
}

// Load the ast types
func (t *trying) loadTypes(pkg string) (err error) {
	t.info = &types.Info{
    Types: make(map[ast.Expr]types.TypeAndValue),
    Defs:  make(map[*ast.Ident]types.Object),
    Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{
		FakeImportC: true,
		Importer: importer.ForCompiler(t.fset, "source", nil),
		Error: func(err error) {
			fmt.Printf("%v\n", err)
		},
	}
	_, err = conf.Check(pkg, t.fset, []*ast.File{t.file}, t.info)
	return
}

// Find each Expr which is is directly before a position
func (t *trying) findWrappedExprs() (err error) {
	t.tryingExprs = make([]ast.Expr, len(t.tryingPositions))
	ast.Inspect(t.file, func(n ast.Node) bool {
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		for idx, position := range t.tryingPositions {
			if int(expr.End()) > position + 1 {
				continue
			}
			wrappedExpr := t.tryingExprs[idx]
			if wrappedExpr == nil {
				t.tryingExprs[idx] = expr
				continue
			}
			if expr.End() > wrappedExpr.End() {
				t.tryingExprs[idx] = expr
			} else if expr.End() == wrappedExpr.End() && expr.Pos() < wrappedExpr.End() {
				t.tryingExprs[idx] = expr
			}
		}
		return true
	})
	return
}

// Parse a golang source
func (t *trying) parseSource() (err error) {
	t.fset = token.NewFileSet()
	t.file, err = parser.ParseFile(t.fset, "", t.source, parser.AllErrors)
	return err
}

// Replace a position array in the source with space characters
// This makes the resulting source be parsable
func (t *trying) replacePositions() {
	var builder strings.Builder
	var lastIdx int
	for _, pos := range t.tryingPositions {
		builder.WriteString(t.source[lastIdx:pos])
		builder.WriteString(" ")
		lastIdx = pos + 1
	}
	if lastIdx < len(t.source) {
		builder.WriteString(t.source[lastIdx:])
	}
	t.source = builder.String()
}

// Get the '?' chars in the source
// Using the parsing errors which are at an '?' char
func (t *trying) getCharPositions() {
	t.tryingPositions = make([]int, 0)
	fset := token.NewFileSet()
	_, errs := parser.ParseFile(fset, "", t.source, parser.AllErrors)
	if errs == nil {
		return
	}
	for _, err := range errs.(scanner.ErrorList) {
		char := t.source[err.Pos.Offset]
		if char != '?' {
			continue
		}
		if slices.Contains(t.tryingPositions, err.Pos.Offset) {
			continue
		}
		t.tryingPositions = append(t.tryingPositions, err.Pos.Offset)
	}
	slices.Sort(t.tryingPositions)
	return
}

