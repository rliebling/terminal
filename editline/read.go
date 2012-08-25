// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package editline

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// ReadKey reads the key pressed. The argument prompt is written to the standard
// output, if any.
func ReadKey(prompt string) (rune rune, err error) {
	in := bufio.NewReaderSize(Input, 4)

	if prompt != "" {
		fmt.Print(prompt)
	}

	rune, _, err = in.ReadRune()
	if err != nil {
		return 0, err
	}

	return rune, nil
}

// ReadBytes reads a line from input until Return is pressed (stripping a trailing
// newline), and returning it in bytes.
// The argument prompt is written to standard output, if any.
func ReadBytes(prompt string) (line []byte, err error) {
	in := bufio.NewReader(Input)

	if prompt != "" {
		fmt.Print(prompt)
	}

	line, err = in.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	return bytes.TrimRight(line, "\n"), nil
}

// ReadString reads a line from input until Return is pressed (stripping a trailing
// newline), returning it into a string.
// The argument prompt is written to standard output, if any.
func ReadString(prompt string) (line string, err error) {
	in := bufio.NewReader(Input)

	if prompt != "" {
		fmt.Print(prompt)
	}

	line, err = in.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(line, "\n"), nil
}
