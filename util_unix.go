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

// Default values
const (
	d_ROW    = 24
	d_COLUMN = 80
)

// GetSizeFromEnv gets the number of rows and columns through environment
// variables, else returns default values.
func GetSizeFromEnv() (row, column int) {
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

// ReadPassword reads the input until '\n' without echo.
// Returns the number of bytes read.
func ReadPassword(fd int, pass []byte) (n int, err error) {
	var oldState, newState termios

	if err = tcgetattr(fd, &oldState); err != nil {
		return 0, err
	}

	// Turn off echo
	newState = oldState
	newState.Lflag &^= (ECHO | ECHOE | ECHOK | ECHONL)

	if err = tcsetattr(fd, _TCSANOW, &newState); err != nil {
		return 0, fmt.Errorf("terminal: could not turn off echo: %s", err)
	}

	// Block SIGINT & SIGTSTP (CTRL-C, CTRL-Z)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTSTP)
	go func() {
		for {
			select {
			case <-sig:
				// ignore
			}
		}
	}()

	tmpPass := make([]byte, len(pass))

	for i, exit := 0, false; ; i++ { // to store all data read until '\n'
		n, err = syscall.Read(fd, tmpPass)
		if err != nil {
			tcsetattr(fd, _TCSANOW, &oldState)
			return 0, err
		}

		if tmpPass[n-1] == '\n' {
			n--
			exit = true
		}
		if i == 0 {
			copy(pass, tmpPass[:n])
		}
		if exit {
			if i != 0 {
				n = len(pass)
			}
			break
		}
	}

	tmpPass = tmpPass[:0] // reset
	tcsetattr(fd, _TCSANOW, &oldState)
	return
}

// WinSizeChan allocates a channel to know when the window size has changed
// through TrapSize.
var WinSizeChan = make(chan byte, 1)

// TrapSize caughts a signal named SIGWINCH whenever the window size changes.
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
