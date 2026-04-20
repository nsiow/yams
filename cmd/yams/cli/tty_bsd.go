//go:build darwin || freebsd || netbsd || openbsd

package cli

import "golang.org/x/sys/unix"

const ttyGetTermios = unix.TIOCGETA
