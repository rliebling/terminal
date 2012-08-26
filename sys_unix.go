// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// Reference: man termios

// +build darwin freebsd linux netbsd openbsd

package console

// #include <termios.h>
// #include <unistd.h>
import "C"

import "fmt"

//C	int tcgetattr(int fd, struct termios *termios_p)
//C	int tcsetattr(int fd, int optional_actions, const struct termios *termios_p)

// tcgetattr gets the console state.
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
}

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
