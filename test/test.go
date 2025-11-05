package main

import (
	"errors"
	"fmt"
	"strconv"
)

func main() {
	run()
}

func run() (err error) {
	fmt.Print(foo(1)? )
	fmt.Print(foo(0)? )

	//===

	fmt.Print(func()string{
		r1, r2 := foo(1)
		err = r2
		return r1
	})
	if err != nil {
		return
	}

	return nil
}

func foo(n int) (string, error) {
	if n == 0 {
		return "", errors.New("division by (?) zero")
	}
	return strconv.Itoa(1337 / n), nil
}
