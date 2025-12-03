package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/theaino/gotk/lib"
)


func main() {
	cwd, _ := os.Getwd()
	root := flag.String("root", cwd, "Specify the project root dir")

	flag.Parse()

	path := flag.Arg(0)

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	args := lib.TkArgs{
		Root: *root,
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
