package main

import (
	"fmt"
	"os"
	"strings"
)

func fail(format string, a ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(2)
}
