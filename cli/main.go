package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/theaino/gotk/lib"
)


func main() {
	cwd, _ := os.Getwd()

	flag.Parse()

	modPath := flag.Arg(0)

	mod, err := lib.CloneModule(modPath, path.Join(cwd, ".tk"))
	if err != nil {
		panic(err)
	}

	args := lib.TkArgs{Mod: mod}

	// Temp. testing preprocessors
	cmd, err := args.Call("go", "run", "./pre/trying")
	if err != nil {
		panic(err)
	}
	cmd.Stderr = os.Stderr
	source, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Print(string(source))
}
