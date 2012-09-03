// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// +build !lookup

package editline

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/kless/terminal"
)

func TestCharMode(t *testing.T) {
	term, _ := terminal.New(syscall.Stderr)
	fmt.Println("\n== Terminal")

	// == Read single key
	if !*fInteractive {
		reply := []string{
			"H",
			"The starts light up\n",
			"you\n",
		}

		go func() {
			for _, r := range reply {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprintf(pw, "%s", r)
			}
		}()
	}

	term.SetCharMode()
	rune, _ := ReadKey("\n + Mode on single character: ")
	fmt.Printf("\n  pressed: %q\n", string(rune))
	term.Restore()

	// == Echo
	//term.Echo(false)
	// TODO: add password

	//term.Echo(true)
}
