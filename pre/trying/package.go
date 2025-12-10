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
		Mode: packages.LoadSyntax,
	}, root)
	if err != nil {
		return nil, err
	}
	pkg.Package = packages[0]
	return pkg, nil
}

func processPackageFiles(pkg *packages.Package) (err error) {
	for idx, file := range pkg.Syntax {
		path := pkg.CompiledGoFiles[idx]
		if err = processFile(file, path); err != nil {
			return
		}
	}
	return
}
