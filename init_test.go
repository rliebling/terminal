// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

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
