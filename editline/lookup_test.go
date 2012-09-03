// Copyright 2012 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build lookup

package editline

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/kless/terminal"
)

func init() {
	Input = os.Stderr
	InputFd = syscall.Stderr
}

// TestLookup prints the decimal code at pressing a key.
func TestLookup(t *testing.T) {
	term, err := terminal.New(InputFd)
	if err != nil {
		t.Fatal(err)
	}
	defer term.Restore()

	if err = term.MakeRaw(); err != nil {
		t.Error(err)
	} else {
		buf := bufio.NewReader(Input)
		runes := make([]int32, 0)
		chars := make([]string, 0)

		fmt.Print("[Press Enter to exit]\r\n")
		fmt.Print("> ")

	L:
		for {
			rune, _, err := buf.ReadRune()
			if err != nil {
				t.Error(err)
				continue
			}

			switch rune {
			default:
				fmt.Print(rune)
				runes = append(runes, rune)
				char := strconv.QuoteRune(rune)
				chars = append(chars, char[1:len(char)-1])
				continue

			case 13:
				fmt.Printf("\r\n\r\n%v\r\n\"%s\"\r\n\r\n", runes, strings.Join(chars, " "))
				break L
			}
		}
	}
}
