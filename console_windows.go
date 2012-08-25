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

package console

import (
  "fmt"
	"syscall"
)

type Console struct {
	fd syscall.Handle

	// To checking if restore is needed
	isNewState bool
	IsRawMode  bool

	// Contains the state of a console
	oldState uint32
	newState uint32
}

// New creates a new console interface in the file descriptor.
func New(fd syscall.Handle) (*Console, error) {
	co := new(Console)

	// Get the actual state
	if err := getConsoleMode(fd, &co.newState); err != nil {
		return nil, err
	}

	// The actual state is copied to another one
	co.oldState = co.newState

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

	co.newState = 0

  raw &^= _ENABLE_LINE_INPUT | _ENABLE_ECHO_INPUT | _ENABLE_PROCESSED_INPUT | _ENABLE_WINDOW_INPUT

	// Put the console in raw mode
  if err := setConsoleMode(co.fd, co.newState); err != nil {
    return fmt.Errorf("console: could not set raw mode: %s", err)
  }

	co.IsRawMode = true
	return nil
}

// SetEcho turns the echo mode.
func (co *Console) SetEcho(echo bool) error {
  //if !echo {
    //co.newState &^= (_ENABLE_ECHO_INPUT)
  //} else {
    co.newState = _ENABLE_ECHO_INPUT|_ENABLE_LINE_INPUT
  //}

  if err := setConsoleMode(co.fd, co.newState); err != nil {
    return fmt.Errorf("console: could not turn echo mode: %s", err)
  }
  co.isNewState = true
  return nil
}


/*
var (
  modkernel32 = syscall.MustLoadDLL("kernel32.dll")

  procGetConsoleMode = modkernel32.MustFindProc("GetConsoleMode")
)
*/
