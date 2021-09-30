package myerr

import "fmt"

// Assert
func Assert(cond bool, msg string) {
	if !cond {
		fmt.Println(msg)
	}
}