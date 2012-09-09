// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var shellsNotANSI = []string{"cmd.exe", "command.com"}

// SupportANSI checks if the terminal supports ANSI escape sequences.
func SupportANSI() bool {
	term := os.Getenv("ComSpec") // full path to the command processor
	if term == "" {
		return false
	}
	term = filepath.Base(term)

	for _, v := range shellsNotANSI {
		if v == term {
			return false
		}
	}
	return true
}

// Default values
const (
	d_ROW    = 24
	d_COLUMN = 80
)

/*// #include <unistd.h>
import "C"*/

/*
//C	char *ttyname(int fd)
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/unix/sysv/linux/ttyname.c;hb=HEAD
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/posix/ttyname.c;hb=HEAD
It uses readlink and the /proc filesystem to query the kernel, and not a
dedicated syscall. Armed with that information, you can probably find the
relevant kernel code which outputs the tty information in a format compatible
with readlink.

// GetName gets the name of a terminal.
func GetName(fd int) (string, error) {
	name, errno := C.ttyname(C.int(fd))
	if errno != nil {
		return "", fmt.Errorf("terminal.TTYName: %s", errno)
	}
	return C.GoString(name), nil
}

*/
//C	int isatty(int fd)
// http://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/posix/isatty.c;hb=HEAD

// IsTerminal returns true if the handler is a terminal.
func IsTerminal(handle syscall.Handle) bool {
	var st uint32
	return getConsoleMode(handle, &st) == nil
}
/*
// ReadPassword reads the input until '\n' without echo.
// Returns the number of bytes read.
func ReadPassword(fd int, pass []byte) (n int, err error) {
	var oldState, newState termios

	if err = tcgetattr(fd, &oldState); err != nil {
		return 0, err
	}

	// Turn off echo
	newState = oldState
	newState.Lflag &^= (ECHO | ECHOE | ECHOK | ECHONL)

	if err = tcsetattr(fd, _TCSANOW, &newState); err != nil {
		return 0, fmt.Errorf("terminal: could not turn off echo: %s", err)
	}

	// Block SIGINT & SIGTSTP (CTRL-C, CTRL-Z)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTSTP)
	go func() {
		for {
			select {
			case <-sig:
				// ignore
			}
		}
	}()

	tmpPass := make([]byte, len(pass))

	for i, exit := 0, false; ; i++ { // to store all data read until '\n'
		n, err = syscall.Read(fd, tmpPass)
		if err != nil {
			tcsetattr(fd, _TCSANOW, &oldState)
			return 0, err
		}

		if tmpPass[n-1] == '\n' {
			n--
			exit = true
		}
		if i == 0 {
			copy(pass, tmpPass[:n])
		}
		if exit {
			if i != 0 {
				n = len(pass)
			}
			break
		}
	}

	tmpPass = tmpPass[:0] // reset
	tcsetattr(fd, _TCSANOW, &oldState)
	return
}

// WinSizeChan allocates a channel to know when the window size has changed
// through TrapSize.
var WinSizeChan = make(chan byte, 1)

// TrapSize caughts a signal named SIGWINCH whenever the window size changes.
func TrapSize() {
	change := make(chan os.Signal, 1)
	signal.Notify(change, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-change:
				WinSizeChan <- 1 // Send a signal
			}
		}
	}()
}*/
