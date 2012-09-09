// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build !plan9,!windows

package terminal

import (
	"flag"
	"io"
	"os"
	"syscall"
)

var (
	fInteractive = flag.Bool("int", false, "mode interactive")
	fTime        = flag.Uint("t", 2, "time in seconds to wait to write")
)

var (
	INPUT    io.Reader
	OUTPUT   io.Writer
	INPUT_FD = syscall.Stderr
)

func init() {
	flag.Parse()

	if *fInteractive {
		INPUT = os.Stderr
		OUTPUT = os.Stdout
	} else {
		INPUT, OUTPUT = io.Pipe()
	}
}
