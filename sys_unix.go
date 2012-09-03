// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/* Reference: man termios ; man tty_ioctl */

// +build !plan9,!windows

package terminal

// #include <unistd.h>
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

// == System calls
//

//cgo const (TCSANOW, TCSADRAIN, TCSAFLUSH)
//cgo const TIOCGWINSZ
//cgo type struct_winsize

//sys	int tcgetattr(int fd, struct termios *termios_p)

func tcgetattr(fd int, state *termios) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(_TCGETS), uintptr(unsafe.Pointer(state)))
	if e1 != 0 {
		err = e1
	}
	return
}

//sys	int tcsetattr(int fd, int optional_actions, const struct termios *termios_p)

func tcsetattr(fd int, action uint, state *termios) (err error) {
	switch action {
	case _TCSANOW:
		action = _TCSETS
	case _TCSADRAIN:
		action = _TCSETSW
	case _TCSAFLUSH:
		action = _TCSETSF
	}

	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(action), uintptr(unsafe.Pointer(state)))
	if e1 != 0 {
		err = e1
	}
	return
}

// getWinsize gets the winsize struct with the terminal size set by the kernel.
func getWinsize(fd int, ws *winsize) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(_TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if e1 != 0 {
		err = e1
	}
	return
}

// == C library functions
//

//C	int isatty(int fd)
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/posix/isatty.c

// IsTerminal returns true if the file descriptor is a terminal.
func IsTerminal(fd int) bool {
	return tcgetattr(fd, &termios{}) == nil
}

//C	char *ttyname(int fd)

// TTYName gets the name of a terminal.
func TTYName(fd int) (string, error) {
	name, errno := C.ttyname(C.int(fd))
	if errno != nil {
		return "", fmt.Errorf("terminal.TTYName: %s", errno)
	}
	return C.GoString(name), nil
}
