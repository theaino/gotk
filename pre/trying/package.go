package main

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

type tryingPackage struct {
	Package *packages.Package
	Positions map[string][]int
	Exprs map[string][]ast.Expr
}

func loadPackage(root string) (*tryingPackage, error) {
	pkg := &tryingPackage{
		Positions: make(map[string][]int),
		Exprs: make(map[string][]ast.Expr),
	}
	packages, err := packages.Load(&packages.Config{
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parseTryingFile(pkg, fset, filename, string(src))
		},
		Dir: root,
		Mode: packages.LoadFiles | packages.LoadImports | packages.LoadTypes | packages.LoadSyntax,
	}, root)
	if err != nil {
		return nil, err
	}
	pkg.Package = packages[0]
	return pkg, nil
}

func (p *tryingPackage) process() (err error) {
	for idx, file := range p.Package.Syntax {
		path := p.Package.CompiledGoFiles[idx]
		if err = p.processFile(file, path); err != nil {
			return
		}
		if err = p.writeFile(file, path); err != nil {
			return
		}
	}
	return
}
