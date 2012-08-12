// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// +build !lookup
package editline

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"
	"testing"
	"time"
)

var (
	fInteractive = flag.Bool("int", false, "mode interactive")
	fTime        = flag.Uint("t", 2, "time in seconds to wait to write answer")

	pr *io.PipeReader
	pw *io.PipeWriter
)

func init() {
	flag.Parse()

	if *fInteractive {
		Input = os.Stderr
	} else {
		pr, pw = io.Pipe()
		Input = pr
	}
	InputFd = syscall.Stderr
}

func TestReadLine(t *testing.T) {
	tempHistory := path.Join(os.TempDir(), "test_editline")

	hist, err := NewHistory(tempHistory)
	if err != nil {
		t.Fatal(err)
	}
	hist.Load()

	fmt.Println("\n== Reading line")
	fmt.Printf("Press ^D to exit\n\n")

	ln, err := NewDefaultLine(hist)
	if err != nil {
		goto _end
	}
	defer ln.Restore()

	if !*fInteractive {
		reply := []string{
			"I have heard that the night is all magic",
			"and that a goblin invites you to dream",
		}

		go func() {
			for _, r := range reply {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprintf(pw, "%s\r\n", r)
			}
			time.Sleep(time.Duration(*fTime) * time.Second)
			pw.Write([]byte{4}) // Ctrl+D
		}()
	}

	for {
		if _, err = ln.Read(); err != nil {
			if err == ErrCtrlD {
				hist.Save()
				err = nil
			} else {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
			break
		}
	}

_end:
	os.Remove(tempHistory)
	if err != nil {
		t.Fatal(err)
	}
}
