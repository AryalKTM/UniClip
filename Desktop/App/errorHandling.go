package main

import(
	"fmt"
	"os"
	"io"
)

func handleError(err error) {
	if err == io.EOF {
		fmt.Println("Disconnected")
	} else {
		fmt.Fprintln(os.Stderr, "error: ["+err.Error()+"]")
	}
}

func debug(a ...interface{}) {
	if printDebugInfo {
		fmt.Println("verbose:", a)
	}
}