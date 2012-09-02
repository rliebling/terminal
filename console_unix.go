// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build darwin freebsd linux netbsd openbsd

package console

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

//cgo type struct_termios

// A Console represents a general console interface.
type Console struct {
	fd int // file descriptor

	width, height int

	// To checking if restore is needed
	isNewState bool
	IsRawMode  bool

	// Contains the state of a console
	oldState *termios // in order to restore the original settings
	newState *termios
}

// New creates a new console interface in the file descriptor.
// Note that 0 is the file descriptor for the standard input.
func New(fd int) (*Console, error) {
	co := new(Console)

	// Get the actual state
	co.newState = new(termios)
	if err := tcgetattr(fd, co.newState); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	co.oldState = new(termios)
	*co.oldState = *co.newState

	co.fd = fd
	return co, nil
}

// == Modes
//

//cgo const (BRKINT, IGNBRK, ICRNL, INLCR, IGNCR, ISTRIP, IXON, PARMRK)
//cgo const (OPOST, ECHO, ECHONL, ICANON, IEXTEN, ISIG)
//cgo const (CSIZE, PARENB, CS8, VMIN, VTIME)

// MakeRaw sets the console to something like the "raw" mode. Input is available
// character by character, echoing is disabled, and all special processing of
// console input and output characters is disabled.
//
// NOTE: in tty "raw mode", CR+LF is used for output and CR is used for input.
func (co *Console) MakeRaw() error {
	if co.IsRawMode {
		return nil
	}

	// Input modes - no break, no CR to NL, no NL to CR, no carriage return,
	// no strip char, no start/stop output control, no parity check.
	co.newState.Iflag &^= (_BRKINT | _IGNBRK | _ICRNL | _INLCR | _IGNCR |
		_ISTRIP | _IXON | _PARMRK)

	// Output modes - disable post processing.
	co.newState.Oflag &^= (_OPOST)

	// Local modes - echoing off, canonical off, no extended functions,
	// no signal chars (^Z,^C).
	co.newState.Lflag &^= (_ECHO | _ECHONL | _ICANON | _IEXTEN | _ISIG)

	// Control modes - set 8 bit chars.
	co.newState.Cflag &^= (_CSIZE | _PARENB)
	co.newState.Cflag |= (_CS8)

	// Control chars - set return condition: min number of bytes and timer.
	// We want read to return every single byte, without timeout.
	co.newState.Cc[_VMIN] = 1 // Read returns when one char is available.
	co.newState.Cc[_VTIME] = 0

	// Put the console in raw mode after flushing
	if err := tcsetattr(co.fd, _TCSAFLUSH, co.newState); err != nil {
		return fmt.Errorf("console: could not set raw mode: %s", err)
	}
	co.IsRawMode = true
	return nil
}

// SetEcho turns the echo mode.
func (co *Console) SetEcho(echo bool) error {
	if !echo {
		co.newState.Lflag &^= (_ECHO)
	} else {
		co.newState.Lflag |= (_ECHO)
	}

	if err := tcsetattr(co.fd, _TCSANOW, co.newState); err != nil {
		return fmt.Errorf("console: could not turn echo mode: %s", err)
	}
	co.isNewState = true
	return nil
}

// SetCharMode sets the console to single-character mode.
func (co *Console) SetCharMode() error {
	// Disable canonical mode, and set buffer size to 1 byte.
	co.newState.Lflag &^= (_ICANON)
	co.newState.Cc[_VTIME] = 0
	co.newState.Cc[_VMIN] = 1

	if err := tcsetattr(co.fd, _TCSANOW, co.newState); err != nil {
		return fmt.Errorf("console: could not set single-character mode: %s", err)
	}
	co.isNewState = true
	return nil
}

// == Restore
//

type State struct {
	wrap *termios
}

// OriginalState returns the console's original state.
func (co *Console) OriginalState() State {
	return State{co.oldState}
}

// Restore restores the original settings for the console.
func (co *Console) Restore() error {
	if co.IsRawMode || co.isNewState {
		*co.newState = *co.oldState

		if err := tcsetattr(co.fd, _TCSANOW, co.newState); err != nil {
			return fmt.Errorf("console: could not restore: %s", err)
		}
		co.IsRawMode = false
		co.isNewState = false
	}
	return nil
}

// Restore restores the settings from State.
func Restore(fd int, st State) error {
	if err := tcsetattr(fd, _TCSANOW, st.wrap); err != nil {
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

// == Size

// GetSize gets the number of rows and columns from the kernel.
func (co *Console) GetSize() (row, column int, err error) {
	ws, e := getWinsize(co.fd)
	if e != nil {
		err = e
		return
	}
	return int(ws.Row), int(ws.Col), nil
}

// Default values
const (
	d_ROW    = 24
	d_COLUMN = 80
)

// GetSizeFromEnv gets the number of rows and columns through environment
// variables, else returns default values.
func (co *Console) GetSizeFromEnv() (row, column int) {
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
