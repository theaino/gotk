package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/ast/astutil"
)


// Insert an error checking structure at each expr
func (p *tryingPackage) modifyAst(path string, blocks []*ast.BlockStmt) {
	finished := make(map[ast.Expr]bool)
	for _, block := range blocks {
		newStmts := make([]ast.Stmt, 0, len(block.List))
		for _, stmt := range block.List {
			hasTryingExpr := false
			stmt := astutil.Apply(stmt, nil, func(c *astutil.Cursor) bool {
				expr, ok := c.Node().(ast.Expr)
				if !ok {
					return true
				}
				if finished[expr] {
					return true
				}
				idx := slices.Index(p.Exprs[path], expr)
				if idx == -1 {
					return true
				}
				resultTypes := make([]string, 0)
				switch exprType := p.Package.TypesInfo.TypeOf(expr).(type) {
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
				finished[expr] = true
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
