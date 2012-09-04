// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package terminal

import (
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
