package main

import (
	"go/ast"
	"slices"
)

// BFS to find the deepest BlockStmt which includes a specific Expr
func (p *tryingPackage) findSurroundingBlocks(path string, file *ast.File) (blocks []*ast.BlockStmt) {
	blocks = make([]*ast.BlockStmt, len(p.Exprs[path]))
	nodeQueue := []ast.Node{file}
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
				for idx, expr := range p.Exprs[path] {
					if nodeExpr == expr {
						blocks[idx] = block
					}
				}
				return true
			})
		}
		nodeQueue = append(nodeQueue, slices.Collect(ast.Preorder(node))[1:]...)
	}
	return
}

// Find each Expr which is is directly before a position
func (p *tryingPackage) findWrappedExprs(path string, file *ast.File) (err error) {
	p.Exprs[path] = make([]ast.Expr, len(p.Positions[path]))
	ast.Inspect(file, func(n ast.Node) bool {
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		for idx, position := range p.Positions[path] {
			if int(expr.End()) > position + 1 {
				continue
			}
			wrappedExpr := p.Exprs[path][idx]
			if wrappedExpr == nil {
				p.Exprs[path][idx] = expr
				continue
			}
			if expr.End() > wrappedExpr.End() {
				p.Exprs[path][idx] = expr
			} else if expr.End() == wrappedExpr.End() && expr.Pos() < wrappedExpr.End() {
				p.Exprs[path][idx] = expr
			}
		}
		return true
	})
	return
}
