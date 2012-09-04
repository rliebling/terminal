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
	INPUT_FD = syscall.Stderr

	P_RD *io.PipeReader
	P_WR *io.PipeWriter
)

func init() {
	flag.Parse()

	if *fInteractive {
		INPUT = os.Stderr
	} else {
		P_RD, P_WR = io.Pipe()
		INPUT = P_RD
	}
}
