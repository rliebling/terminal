// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

/* Reference: man termios ; man tty_ioctl */

// +build darwin freebsd linux netbsd openbsd

package console

// #include <unistd.h>
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

//cgo TCGETS, TCSETS, TCSETSW, TCSETSF, TCSANOW, TCSADRAIN, TCSAFLUSH

//C	int tcgetattr(int fd, struct termios *termios_p)
//C	int tcsetattr(int fd, int optional_actions, const struct termios *termios_p)

func tcgetattr(fd int, state *termios) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(TCGETS), uintptr(unsafe.Pointer(state)))
	if e1 != 0 {
		err = e1
	}
	return
}

func tcsetattr(fd int, action int, state *termios) (err error) {
	switch action {
	case TCSANOW:
		action = TCSETS
	case TCSADRAIN:
		action = TCSETSW
	case TCSAFLUSH:
		action = TCSETSF
	}

	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(action), uintptr(unsafe.Pointer(state)))
	if e1 != 0 {
		err = e1
	}
	return
}

/*// tcgetattr gets the console state.
func tcgetattr(fd int, state *C.struct_termios) error {
	exitCode, errno := C.tcgetattr(C.int(fd), state)
	if exitCode == 0 {
		return nil
	}
	return fmt.Errorf("console.tcgetattr: %s", errno)
}

// tcsetattr sets the console state.
func tcsetattr(fd, actions int, state *C.struct_termios) error {
	exitCode, errno := C.tcsetattr(C.int(fd), C.int(actions), state)
	if exitCode == 0 {
		return nil
	}
	return fmt.Errorf("console.tcsetattr: %s", errno)
}*/

//C	int isatty(int fd)
//C	char *ttyname(int fd)

// IsTTY determines if the device is a console.
func IsTTY(fd int) (bool, error) {
	exitCode, errno := C.isatty(C.int(fd))
	if exitCode == 1 {
		return true, nil
	}
	return false, fmt.Errorf("console.IsTTY: %s", errno)
}

// TTYName gets the name of a console.
func TTYName(fd int) (string, error) {
	name, errno := C.ttyname(C.int(fd))
	if errno != nil {
		return "", fmt.Errorf("console.TTYName: %s", errno)
	}
	return C.GoString(name), nil
}

// * * *

//cgo TIOCGWINSZ

// getWinsize returns the winsize struct with the console size set by the kernel.
// It is used the TIOCGWINSZ ioctl() call on the tty device.
func getWinsize(fd int) (ws *winsize, err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if e1 != 0 {
		err = e1
	}
	return
}
