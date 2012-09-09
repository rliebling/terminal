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

// A Terminal represents a general terminal interface.
type Terminal struct {
	fd int // File descriptor
	mod mode

	// Size
	row, column int

	// Contain the state of a terminal, allowing to restore the original settings
	oldState, lastState termios
}

// New creates a new terminal interface in the file descriptor.
// Note that an input file descriptor should be used.
func New(fd int) (*Terminal, error) {
	var t Terminal

	// Get the actual state
	if err := tcgetattr(fd, &t.lastState); err != nil {
		return nil, err
	}

	t.oldState = t.lastState // the actual state is copied to another one
	t.fd = fd
	return &t, nil
}

// == Restore
//

type State struct {
	wrap termios
}

// OriginalState returns the terminal's original state.
func (t *Terminal) OriginalState() State {
	return State{t.oldState}
}

// Restore restores the original settings for the terminal.
func (t *Terminal) Restore() error {
	if t.mod != 0 {
		if err := tcsetattr(t.fd, _TCSANOW, &t.oldState); err != nil {
			return fmt.Errorf("terminal: could not restore: %s", err)
		}
		t.lastState = t.oldState
		t.mod = 0
	}
	return nil
}

// Restore restores the settings from State.
func Restore(fd int, st State) error {
	if err := tcsetattr(fd, _TCSANOW, &st.wrap); err != nil {
		return fmt.Errorf("terminal: could not restore: %s", err)
	}
	return nil
}

// == Modes
//

// RawMode sets the terminal to something like the "raw" mode. Input is available
// character by character, echoing is disabled, and all special processing of
// terminal input and output characters is disabled.
//
// NOTE: in tty "raw mode", CR+LF is used for output and CR is used for input.
func (t *Terminal) RawMode() error {
	if t.mod&rawMode != 0 {
		return nil
	}

	// Input modes - no break, no CR to NL, no NL to CR, no carriage return,
	// no strip char, no start/stop output control, no parity check.
	t.lastState.Iflag &^= (BRKINT | IGNBRK | ICRNL | INLCR | IGNCR | ISTRIP | IXON | PARMRK)

	// Output modes - disable post processing.
	t.lastState.Oflag &^= OPOST

	// Local modes - echoing off, canonical off, no extended functions,
	// no signal chars (^Z,^C).
	t.lastState.Lflag &^= (ECHO | ECHONL | ICANON | IEXTEN | ISIG)

	// Control modes - set 8 bit chars.
	t.lastState.Cflag &^= (CSIZE | PARENB)
	t.lastState.Cflag |= CS8

	// Control chars - set return condition: min number of bytes and timer.
	// We want read to return every single byte, without timeout.
	t.lastState.Cc[VMIN] = 1 // Read returns when one char is available.
	t.lastState.Cc[VTIME] = 0

	// Put the terminal in raw mode after flushing
	if err := tcsetattr(t.fd, _TCSAFLUSH, &t.lastState); err != nil {
		return fmt.Errorf("terminal: could not set raw mode: %s", err)
	}
	t.mod |= rawMode
	return nil
}

// EchoMode turns the echo mode.
func (t *Terminal) EchoMode(echo bool) error {
	if echo {
		t.lastState.Lflag |= ECHO
	} else {
		t.lastState.Lflag &^= ECHO
	}

	if err := tcsetattr(t.fd, _TCSANOW, &t.lastState); err != nil {
		return fmt.Errorf("terminal: could not turn echo mode: %s", err)
	}

	if echo {
		t.mod |= echoMode
	} else {
		t.mod &^= echoMode
	}
	return nil
}

// CharMode sets the terminal to single-character mode.
func (t *Terminal) CharMode() error {
	// Disable canonical mode, and set buffer size to 1 byte.
	t.lastState.Lflag &^= ICANON
	t.lastState.Cc[VTIME] = 0
	t.lastState.Cc[VMIN] = 1

	if err := tcsetattr(t.fd, _TCSANOW, &t.lastState); err != nil {
		return fmt.Errorf("terminal: could not set single-character mode: %s", err)
	}
	t.mod |= charMode
	return nil
}

// SetMode sets the terminal attributes given by state.
// Warning: The use of this function could do your code not cross-system.
func (t *Terminal) SetMode(state termios) error {
	if err := tcsetattr(t.fd, _TCSANOW, &state); err != nil {
		return fmt.Errorf("terminal: could not set new mode: %s", err)
	}

	t.lastState = state
	t.mod |= otherMode
	return nil
}

// == Utility
//

// Fd returns the Unix file descriptor referencing the terminal.
func (t *Terminal) Fd() int {
	return t.fd
}

// GetSize returns the size of the terminal.
func (t *Terminal) GetSize() (row, column int, err error) {
/*fmt.Println(t.row)
	if t.row != 0 {
println("OPS!")
		return t.row, t.column, nil
	}
*/
	ws := new(winsize)

	if e := getWinsize(t.fd, ws); e != nil {
		err = e
		return
	}
	t.row, t.column = int(ws.Row), int(ws.Col)
	return t.row, t.column, nil
}
