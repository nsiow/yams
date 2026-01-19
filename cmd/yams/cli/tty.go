package cli

import (
	"os"

	"golang.org/x/sys/unix"
)

// IsTTY returns true if the given file descriptor is a terminal
func IsTTY(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	return err == nil
}

// StdoutIsTTY returns true if stdout is a terminal
func StdoutIsTTY() bool {
	return IsTTY(int(os.Stdout.Fd()))
}

// StderrIsTTY returns true if stderr is a terminal
func StderrIsTTY() bool {
	return IsTTY(int(os.Stderr.Fd()))
}

// StdinIsTTY returns true if stdin is a terminal
func StdinIsTTY() bool {
	return IsTTY(int(os.Stdin.Fd()))
}
