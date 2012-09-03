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

//cgo type struct_termios

// A Terminal represents a general terminal interface.
type Terminal struct {
	fd int // file descriptor

	width, height int

	// To checking if restore is needed
	isNewState bool
	IsRawMode  bool

	// Contains the state of a terminal
	oldState *termios // in order to restore the original settings
	newState *termios
}

// New creates a new terminal interface in the file descriptor.
// Note that 0 is the file descriptor for the standard input.
func New(fd int) (*Terminal, error) {
	t := new(Terminal)

	// Get the actual state
	t.newState = new(termios)
	if err := tcgetattr(fd, t.newState); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	t.oldState = new(termios)
	*t.oldState = *t.newState

	t.fd = fd
	return t, nil
}

// == Modes
//

//cgo const (BRKINT, IGNBRK, ICRNL, INLCR, IGNCR, ISTRIP, IXON, PARMRK)
//cgo const (OPOST, ECHO, ECHONL, ICANON, IEXTEN, ISIG)
//cgo const (CSIZE, PARENB, CS8, VMIN, VTIME)

// MakeRaw sets the terminal to something like the "raw" mode. Input is available
// character by character, echoing is disabled, and all special processing of
// terminal input and output characters is disabled.
//
// NOTE: in tty "raw mode", CR+LF is used for output and CR is used for input.
func (t *Terminal) MakeRaw() error {
	if t.IsRawMode {
		return nil
	}

	// Input modes - no break, no CR to NL, no NL to CR, no carriage return,
	// no strip char, no start/stop output control, no parity check.
	t.newState.Iflag &^= (_BRKINT | _IGNBRK | _ICRNL | _INLCR | _IGNCR |
		_ISTRIP | _IXON | _PARMRK)

	// Output modes - disable post processing.
	t.newState.Oflag &^= (_OPOST)

	// Local modes - echoing off, canonical off, no extended functions,
	// no signal chars (^Z,^C).
	t.newState.Lflag &^= (_ECHO | _ECHONL | _ICANON | _IEXTEN | _ISIG)

	// Control modes - set 8 bit chars.
	t.newState.Cflag &^= (_CSIZE | _PARENB)
	t.newState.Cflag |= (_CS8)

	// Control chars - set return condition: min number of bytes and timer.
	// We want read to return every single byte, without timeout.
	t.newState.Cc[_VMIN] = 1 // Read returns when one char is available.
	t.newState.Cc[_VTIME] = 0

	// Put the terminal in raw mode after flushing
	if err := tcsetattr(t.fd, _TCSAFLUSH, t.newState); err != nil {
		return fmt.Errorf("terminal: could not set raw mode: %s", err)
	}
	t.IsRawMode = true
	return nil
}

// SetEcho turns the echo mode.
func (t *Terminal) SetEcho(echo bool) error {
	if !echo {
		t.newState.Lflag &^= (_ECHO)
	} else {
		t.newState.Lflag |= (_ECHO)
	}

	if err := tcsetattr(t.fd, _TCSANOW, t.newState); err != nil {
		return fmt.Errorf("terminal: could not turn echo mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// SetCharMode sets the terminal to single-character mode.
func (t *Terminal) SetCharMode() error {
	// Disable canonical mode, and set buffer size to 1 byte.
	t.newState.Lflag &^= (_ICANON)
	t.newState.Cc[_VTIME] = 0
	t.newState.Cc[_VMIN] = 1

	if err := tcsetattr(t.fd, _TCSANOW, t.newState); err != nil {
		return fmt.Errorf("terminal: could not set single-character mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// == Restore
//

type State struct {
	wrap *termios
}

// OriginalState returns the terminal's original state.
func (t *Terminal) OriginalState() State {
	return State{t.oldState}
}

// Restore restores the original settings for the terminal.
func (t *Terminal) Restore() error {
	if t.IsRawMode || t.isNewState {
		*t.newState = *t.oldState

		if err := tcsetattr(t.fd, _TCSANOW, t.newState); err != nil {
			return fmt.Errorf("terminal: could not restore: %s", err)
		}
		t.IsRawMode = false
		t.isNewState = false
	}
	return nil
}

// Restore restores the settings from State.
func Restore(fd int, st State) error {
	if err := tcsetattr(fd, _TCSANOW, st.wrap); err != nil {
		return fmt.Errorf("terminal: could not restore: %s", err)
	}
	return nil
}

// == Information
//

var unsupportedTerm = []string{"dumb", "cons25"}

// CheckANSI checks if the terminal supports ANSI escape controls.
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
func (t *Terminal) GetSize() (row, column int, err error) {
	ws := new(winsize)
	if e := getWinsize(t.fd, ws); e != nil {
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
func (t *Terminal) GetSizeFromEnv() (row, column int) {
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
