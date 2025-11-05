package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/theaino/gotk/lib"
)


func main() {
	flag.Parse()

	path := flag.Arg(0)

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	args := lib.TkArgs{
		Root: cwd,
		Path: path,
		Source: string(data),
	}

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
