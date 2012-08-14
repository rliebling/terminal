// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/* Reference: man termios */

// +build darwin freebsd linux netbsd openbsd

package console

// #include <termios.h>
// #include <unistd.h>
import "C"

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"unsafe"
)

// A Console represents a general console interface.
type Console struct {
	Fd int // file descriptor

	// To checking if restore is needed
	isNewState bool
	IsRawMode  bool

	// Contains the state of a console
	oldState *C.struct_termios // in order to restore the original settings
	newState *C.struct_termios
}

// New creates a new console interface in the file descriptor.
// Note that 0 is the file descriptor for the standard input.
func New(fd int) (*Console, error) {
	t := new(Console)

	// Get the actual state
	t.newState = new(C.struct_termios)
	if err := tcgetattr(fd, t.newState); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	t.oldState = new(C.struct_termios)
	*t.oldState = *t.newState

	t.Fd = fd
	return t, nil
}

// == Modes
//

// MakeRaw sets the console to something like the "raw" mode. Input is available
// character by character, echoing is disabled, and all special processing of
// console input and output characters is disabled.
//
// NOTE: in tty "raw mode", CR+LF is used for output and CR is used for input.
func (t *Console) MakeRaw() error {
	if t.IsRawMode {
		return nil
	}

	// Based in the system call: void cfmakeraw(struct termios *termios_p)

	// Input modes - no break, no CR to NL, no NL to CR, no carriage return,
	// no strip char, no start/stop output control, no parity check.
	t.newState.c_iflag &^= (C.BRKINT | C.IGNBRK | C.ICRNL | C.INLCR | C.IGNCR |
		C.ISTRIP | C.IXON | C.PARMRK)

	// Output modes - disable post processing.
	t.newState.c_oflag &^= (C.OPOST)

	// Local modes - echoing off, canonical off, no extended functions,
	// no signal chars (^Z,^C).
	t.newState.c_lflag &^= (C.ECHO | C.ECHONL | C.ICANON | C.IEXTEN | C.ISIG)

	// Control modes - set 8 bit chars.
	t.newState.c_cflag &^= (C.CSIZE | C.PARENB)
	t.newState.c_cflag |= (C.CS8)

	// Control chars - set return condition: min number of bytes and timer.
	// We want read to return every single byte, without timeout.
	t.newState.c_cc[C.VMIN] = 1 // Read returns when one char is available.
	t.newState.c_cc[C.VTIME] = 0

	// Put the console in raw mode after flushing
	if err := t.tcsetattr(C.TCSAFLUSH); err != nil {
		return fmt.Errorf("console: could not set raw mode: %s", err)
	}
	t.IsRawMode = true
	return nil
}

// Echo turns the echo mode.
func (t *Console) Echo(echo bool) error {
	if !echo {
		t.newState.c_lflag &^= (C.ECHO)
	} else {
		t.newState.c_lflag |= (C.ECHO)
	}

	if err := t.tcsetattr(C.TCSANOW); err != nil {
		return fmt.Errorf("console: could not turn echo mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// ModeChar sets the console to single-character mode.
func (t *Console) ModeChar() error {
	// Disable canonical mode, and set buffer size to 1 byte.
	t.newState.c_lflag &^= (C.ICANON)
	t.newState.c_cc[C.VTIME] = 0
	t.newState.c_cc[C.VMIN] = 1

	if err := t.tcsetattr(C.TCSANOW); err != nil {
		return fmt.Errorf("console: could not set single-character mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// * * *

// tcgetattr gets the console state.
func tcgetattr(fd int, state *C.struct_termios) error {
	// int tcgetattr(int fd, struct termios *termios_p);
	exitCode, errno := C.tcgetattr(C.int(fd), state)
	if exitCode == 0 {
		return nil
	}
	return fmt.Errorf("console.tcgetattr: %s", errno)
}

// tcsetattr sets the console state.
func tcsetattr(fd, optional_actions int, state *C.struct_termios) error {
	// int tcsetattr(int fd, int optional_actions, const struct termios *termios_p);
	exitCode, errno := C.tcsetattr(C.int(fd), C.int(optional_actions), state)
	if exitCode == 0 {
		return nil
	}
	return fmt.Errorf("console.tcsetattr: %s", errno)
}

// tcsetattr sets the console state; use arguments from Console.
func (t *Console) tcsetattr(optional_actions int) error {
	return tcsetattr(t.Fd, optional_actions, t.newState)
}

// == Restore
//

type State struct {
	wrap *C.struct_termios
}

// OriginalState returns the console's original state.
func (t *Console) OriginalState() *State {
	return &State{t.oldState}
}

// Restore restores the original settings for the console.
func (t *Console) Restore() error {
	if t.IsRawMode || t.isNewState {
		*t.newState = *t.oldState

		if err := t.tcsetattr(C.TCSANOW); err != nil {
			return fmt.Errorf("console: could not restore: %s", err)
		}
		t.IsRawMode = false
		t.isNewState = false
	}
	return nil
}

// Restore restores the settings from State.
func Restore(fd int, st *State) error {
	if err := tcsetattr(fd, C.TCSANOW, st.wrap); err != nil {
		return fmt.Errorf("console: could not restore: %s", err)
	}
	return nil
}

// == Information
//

var unsupportedTerm = []string{"dumb", "cons25"}

// CheckANSI checks if the console supports ANSI escape controls.
func CheckANSI() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	for _, v := range unsupportedTerm {
		if v == term {
			return false
		}
	}
	return true
}

// IsTTY determines if the device is a console.
func IsTTY(fd int) (bool, error) {
	// int isatty(int fd);
	exitCode, errno := C.isatty(C.int(fd))
	if exitCode == 1 {
		return true, nil
	}
	return false, fmt.Errorf("console.IsTTY: %s", errno)
}

// TTYName gets the name of a console.
func TTYName(fd int) (string, error) {
	// char *ttyname(int fd);
	name, errno := C.ttyname(C.int(fd))
	if errno != nil {
		return "", fmt.Errorf("console.TTYName: %s", errno)
	}
	return C.GoString(name), nil
}

// * * *

// Default values
const (
	d_ROW    = 24
	d_COLUMN = 80
)

// GetSize gets the number of rows and columns of the window or terminal.
// In the first is used WinSize(), else it is tried using the environment
// variables, and lastly returns values by default.
func GetSize(fd int) (row, column int) {
	ws, err := WinSize(fd)

	// If there is any error, then to try get the values through environment.
	// Else, it is used values by default.
	if err != nil {
		sRow := os.Getenv("LINES")
		sCol := os.Getenv("COLUMNS")

		if sRow == "" {
			row = d_ROW
		} else {
			iRow, err := strconv.Atoi(sRow)
			if err == nil {
				row = iRow
			} else {
				row = d_ROW
			}
		}

		if sCol == "" {
			column = d_COLUMN
		} else {
			iCol, err := strconv.Atoi(sCol)
			if err == nil {
				column = iCol
			} else {
				column = d_COLUMN
			}
		}
		return
	}

	return int(ws.Row), int(ws.Col)
}

// The winsize structure is defined in Linux in "termios.h".
// But it is the same in system Darwin.
type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// WinSize gets the window size.
// It is used the TIOCGWINSZ ioctl() call on the tty device.
func WinSize(fd int) (*Winsize, error) {
	ws := new(Winsize)

	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(_TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)

	if int(r1) == -1 {
		return nil, os.NewSyscallError("WinSize", errno)
	}
	return ws, nil
}

// GetSize gets the number of rows and columns of the actual window or terminal.
func (t *Console) GetSize() (row, column int) {
	return GetSize(t.Fd)
}

// WinSize gets the window size of the actual window.
func (t *Console) WinSize() (*Winsize, error) {
	return WinSize(t.Fd)
}

// == Utility
//

var WinSizeChan = make(chan byte, 1) // Allocate a channel for TrapSize()

// TrapSize caughts a signal named SIGWINCH whenever the screen size changes.
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
}
