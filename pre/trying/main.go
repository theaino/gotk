package main

import (
	"github.com/theaino/gotk/lib"
)

func main() {
	args, err := lib.GetArgs()
	if err != nil {
		panic(err)
	}

	p, err := loadPackage(args.Mod.Root)
	if err != nil {
		panic(err)
	}

	err = p.process()
	if err != nil {
		panic(err)
	}
}
