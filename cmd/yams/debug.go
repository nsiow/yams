package main

import (
	"fmt"
	"strings"
)

var IS_DEBUG_ENABLED bool

const DEBUG_PREAMBLE = "[DEBUG] "

func debug(msg string, args ...any) {
	if IS_DEBUG_ENABLED {
		if !strings.HasPrefix(msg, DEBUG_PREAMBLE) {
			msg = DEBUG_PREAMBLE + msg
		}
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Printf(msg, args...)
	}
}
