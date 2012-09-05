// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package terminal

import (
	"fmt"
)

//cgo type struct_termios

// A Terminal represents a general terminal interface.
type Terminal struct {
	// To checking if restore is needed
	isNewState bool
	IsRawMode  bool

	fd int // file descriptor

	// Contains the state of a terminal
	oldState *termios // in order to restore the original settings
	State    *termios
}

// New creates a new terminal interface in the file descriptor.
// Note that an input file descriptor should be used.
func New(fd int) (*Terminal, error) {
	t := new(Terminal)

	// Get the actual state
	t.State = new(termios)
	if err := tcgetattr(fd, t.State); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	t.oldState = new(termios)
	*t.oldState = *t.State

	t.fd = fd
	return t, nil
}

// Fd returns the Unix file descriptor referencing the terminal.
func (t *Terminal) Fd() int {
	return t.fd
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
		if err := tcsetattr(t.fd, _TCSANOW, t.oldState); err != nil {
			return fmt.Errorf("terminal: could not restore: %s", err)
		}
		*t.State = *t.oldState
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

// == Modes
//

//cgo const (BRKINT, IGNBRK, ICRNL, INLCR, IGNCR, ISTRIP, IXON, PARMRK)
//cgo const (OPOST, ECHO, ECHONL, ICANON, IEXTEN, ISIG)
//cgo const (CSIZE, PARENB, CS8, VMIN, VTIME)
//cgo const (ECHO, ECHOE, ECHOK, ECHONL)

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
	t.State.Iflag &^= (_BRKINT | _IGNBRK | _ICRNL | _INLCR | _IGNCR |
		_ISTRIP | _IXON | _PARMRK)

	// Output modes - disable post processing.
	t.State.Oflag &^= _OPOST

	// Local modes - echoing off, canonical off, no extended functions,
	// no signal chars (^Z,^C).
	t.State.Lflag &^= (_ECHO | _ECHONL | _ICANON | _IEXTEN | _ISIG)

	// Control modes - set 8 bit chars.
	t.State.Cflag &^= (_CSIZE | _PARENB)
	t.State.Cflag |= _CS8

	// Control chars - set return condition: min number of bytes and timer.
	// We want read to return every single byte, without timeout.
	t.State.Cc[_VMIN] = 1 // Read returns when one char is available.
	t.State.Cc[_VTIME] = 0

	// Put the terminal in raw mode after flushing
	if err := tcsetattr(t.fd, _TCSAFLUSH, t.State); err != nil {
		return fmt.Errorf("terminal: could not set raw mode: %s", err)
	}
	t.IsRawMode = true
	return nil
}

// SetEcho turns the echo mode.
func (t *Terminal) SetEcho(echo bool) error {
	if !echo {
		t.State.Lflag &^= _ECHO
	} else {
		t.State.Lflag |= _ECHO
	}

	if err := tcsetattr(t.fd, _TCSANOW, t.State); err != nil {
		return fmt.Errorf("terminal: could not turn echo mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// SetSingleChar sets the terminal to single-character mode.
func (t *Terminal) SetSingleChar() error {
	// Disable canonical mode, and set buffer size to 1 byte.
	t.State.Lflag &^= _ICANON
	t.State.Cc[_VTIME] = 0
	t.State.Cc[_VMIN] = 1

	if err := tcsetattr(t.fd, _TCSANOW, t.State); err != nil {
		return fmt.Errorf("terminal: could not set single-character mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// == Utility
//

// GetSize gets the number of rows and columns from the kernel.
func (t *Terminal) GetSize() (row, column int, err error) {
	ws := new(winsize)
	if e := getWinsize(t.fd, ws); e != nil {
		err = e
		return
	}
	return int(ws.Row), int(ws.Col), nil
}
