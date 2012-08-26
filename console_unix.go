// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build darwin freebsd linux netbsd openbsd

package console

// #include <termios.h>
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
	fd int // file descriptor

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
	co := new(Console)

	// Get the actual state
	co.newState = new(C.struct_termios)
	if err := tcgetattr(fd, co.newState); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	co.oldState = new(C.struct_termios)
	*co.oldState = *co.newState

	co.fd = fd
	return co, nil
}

// == Modes
//

// MakeRaw sets the console to something like the "raw" mode. Input is available
// character by character, echoing is disabled, and all special processing of
// console input and output characters is disabled.
//
// NOTE: in tty "raw mode", CR+LF is used for output and CR is used for input.
func (co *Console) MakeRaw() error {
	if co.IsRawMode {
		return nil
	}

	C.cfmakeraw(co.newState)

	// Put the console in raw mode after flushing
	if err := co.tcsetattr(C.TCSAFLUSH); err != nil {
		return fmt.Errorf("console: could not set raw mode: %s", err)
	}
	co.IsRawMode = true
	return nil
}

// SetEcho turns the echo mode.
func (co *Console) SetEcho(echo bool) error {
	if !echo {
		co.newState.c_lflag &^= C.ECHO
	} else {
		co.newState.c_lflag |= C.ECHO
	}

	if err := co.tcsetattr(C.TCSANOW); err != nil {
		return fmt.Errorf("console: could not turn echo mode: %s", err)
	}
	co.isNewState = true
	return nil
}

// SetCharMode sets the console to single-character mode.
func (co *Console) SetCharMode() error {
	// Disable canonical mode, and set buffer size to 1 byte.
	co.newState.c_lflag &^= (C.ICANON)
	co.newState.c_cc[C.VTIME] = 0
	co.newState.c_cc[C.VMIN] = 1

	if err := co.tcsetattr(C.TCSANOW); err != nil {
		return fmt.Errorf("console: could not set single-character mode: %s", err)
	}
	co.isNewState = true
	return nil
}

// tcsetattr sets the console state; use arguments from Console.
func (co *Console) tcsetattr(actions int) error {
	return tcsetattr(co.fd, actions, co.newState)
}

// == Restore
//

type State struct {
	wrap *C.struct_termios
}

// OriginalState returns the console's original state.
func (co *Console) OriginalState() State {
	return State{co.oldState}
}

// Restore restores the original settings for the console.
func (co *Console) Restore() error {
	if co.IsRawMode || co.isNewState {
		*co.newState = *co.oldState

		if err := co.tcsetattr(C.TCSANOW); err != nil {
			return fmt.Errorf("console: could not restore: %s", err)
		}
		co.IsRawMode = false
		co.isNewState = false
	}
	return nil
}

// Restore restores the settings from State.
func Restore(fd int, st State) error {
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

// Default values
const (
	d_ROW    = 24
	d_COLUMN = 80
)

// GetSize gets the number of rows and columns of the window or console.
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

// GetSize gets the number of rows and columns of the actual window or console.
func (co *Console) GetSize() (row, column int) {
	return GetSize(co.fd)
}

// WinSize gets the window size of the actual window.
func (co *Console) WinSize() (*Winsize, error) {
	return WinSize(co.fd)
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
