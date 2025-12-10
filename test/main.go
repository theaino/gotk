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

func run() (err error) {
	fmt.Println(foo(3)?)

	return
}

func foo(n int) (string, error) {
	if n == 0 {
		return "", errors.New("division by (?) zero")
	}
	return strconv.Itoa(1337 / n), nil
}
