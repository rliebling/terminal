// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package editline

import (
	"bufio"
	"fmt"
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
