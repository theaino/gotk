package main

import (
	"errors"
	"fmt"
	"strconv"
	"gopkg.in/yaml.v3"
)

func main() {
	err := run()
	panic(err)
}

var testData = `
a: foo
b:
	c: bar
`

func run() (err error) {
	var res any
	yaml.Unmarshal([]byte(testData), &res)?
	fmt.Println(foo(res)?)

	return
}

func foo(n any) (string, error) {
	return strconv.Itoa(1337 / n), nil
}
