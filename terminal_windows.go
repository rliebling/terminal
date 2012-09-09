// Copyright 2012 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// System calls
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms682073%28v=vs.85%29.aspx
//
//sys getConsoleMode(handle syscall.Handle, mode *uint32) (err error) = GetConsoleMode
//sys setConsoleMode(handle syscall.Handle, mode uint32) (err error) = SetConsoleMode
//sys getConsoleScreenBufferInfo(handle syscall.Handle, info *_CONSOLE_SCREEN_BUFFER_INFO) (err error) = GetConsoleScreenBufferInfo
//sys readConsoleInput(handleIn syscall.Handle, buf *_INPUT_RECORD, length uint32, numEvents *uint32) (err error) = ReadConsoleInputW

package terminal

import (
	"fmt"
	"os"
	"syscall"
)

type Terminal struct {
	// To checking if restore is needed
	isNewState, isRawMode bool

	// Handler
	handle syscall.Handle

	// Size
	row, column int

	// Contain the state of a terminal, allowing to restore the original settings
	oldState, lastState uint32
}

// New creates a new terminal interface in the file descriptor.
func New(handle syscall.Handle) (*Terminal, error) {
	var t Terminal

	// Get the actual state
	if err := getConsoleMode(handle, &t.lastState); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	t.oldState = t.lastState

	t.handle = handle
	return &t, nil
}

// == Restore
//

type State struct {
	wrap uint32
}

// OriginalState returns the terminal's original state.
func (t *Terminal) OriginalState() State {
	return State{t.oldState}
}

// Restore restores the original settings for the terminal.
func (t *Terminal) Restore() error {
	if t.isRawMode || t.isNewState {
		if err := setConsoleMode(t.handle, t.oldState); err != nil {
			return fmt.Errorf("terminal: could not restore: %s", err)
		}
		t.lastState = t.oldState
		t.isRawMode = false
		t.isNewState = false
	}
	return nil
}

// Restore restores the settings from State.
func Restore(handle syscall.Handle, st State) error {
	if err := setConsoleMode(handle, st.wrap); err != nil {
		return fmt.Errorf("terminal: could not restore: %s", err)
	}
	return nil
}

// == Modes
//

// MakeRaw sets the terminal to something like the "raw" mode. Input is available
// character by character, echoing is disabled, and all special processing of
// terminal input and output characters is disabled.
//
// NOTE: in tty "raw mode", CR+LF is used for output and CR is used for input.
func (t *Terminal) MakeRaw() error {
	if t.isRawMode {
		return nil
	}

	t.lastState = 0
	t.lastState &^= (ENABLE_LINE_INPUT | ENABLE_PROCESSED_INPUT | ENABLE_ECHO_INPUT |
		ENABLE_WINDOW_INPUT)

// in Stdout
//	t.lastState &^= (ENABLE_PROCESSED_OUTPUT | ENABLE_WRAP_AT_EOL_OUTPUT)

	// Put the terminal in raw mode
	if err := setConsoleMode(t.handle, t.lastState); err != nil {
		return fmt.Errorf("terminal: could not set raw mode: %s", err)
	}
	t.isRawMode = true
	return nil
}

// SetEcho turns the echo mode.
func (t *Terminal) SetEcho(echo bool) error {
	if !echo {
		t.lastState &^= ENABLE_ECHO_INPUT
	} else {
		t.lastState |= ENABLE_ECHO_INPUT
	}

	if err := setConsoleMode(t.handle, t.lastState); err != nil {
		return fmt.Errorf("terminal: could not turn echo mode: %s", err)
	}
	t.isNewState = true
	return nil
}

// SetSingleChar sets the terminal to single-character mode.
func (t *Terminal) SetSingleChar() (err error) {
	t.lastState |= ENABLE_WINDOW_INPUT // | ENABLE_MOUSE_INPUT

//	t.lastState &^= ENABLE_PROCESSED_OUTPUT

	if err = setConsoleMode(t.handle, t.lastState); err != nil {
		return fmt.Errorf("terminal: could not set single-character mode: %s", err)
	}
	t.isNewState = true

	var input _INPUT_RECORD
	var numEvents uint32 = 1

	go func() {
		for {
			err = readConsoleInput(t.handle, &input, 1, &numEvents)
			if err != nil {
fmt.Println("ERR:", err)
//				return(err)
break
			}

			/*if input.EventType == _KEY_EVENT {
			
			}*/
		}
	}()
	return nil
}

// SetMode sets the terminal attributes given by state.
// Warning: The use of this function could do your code not cross-system.
func (t *Terminal) SetMode(state uint32) error {
	if err := setConsoleMode(t.handle, state); err != nil {
		return fmt.Errorf("terminal: could not set new mode: %s", err)
	}
	t.lastState = state
	t.isNewState = true
	return nil
}

// == Utility
//
/*
type WinSize struct {
	Row    int16
	Col    int16
	Xpixel int16
	Ypixel int16
}
*/

// Fd returns the handle referencing the terminal.
func (t *Terminal) Fd() int {
	return int(t.handle)
}

/*func (t *Terminal) GetName() (name string, err error) {
	var title string
	if _, e := getConsoleTitle(&title, 128); e != nil {
		return "", os.NewSyscallError("getConsoleTitle", e)
	}
	return title, nil
}*/

// GetSize returns the size of the terminal.
func (t *Terminal) GetSize() (row, column int, err error) {
/*	if t.row != 0 {
		return t.row, t.column, nil
	}
*/
	info := new(_CONSOLE_SCREEN_BUFFER_INFO)
	if e := getConsoleScreenBufferInfo(t.handle, info); e != nil {
		err = os.NewSyscallError("getConsoleScreenBufferInfo", e)
		return
	}
	t.row, t.column = int(info.dwSize.x), int(info.dwSize.y)
	return t.row, t.column, nil
}
