// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// +build !plan9,!windows

package terminal

import (
	"bufio"
	"fmt"
	"testing"
	"time"
)

func TestRawMode(t *testing.T) {
	term, err := New(INPUT_FD)
	if err != nil {
		t.Fatal(err)
	}

	oldState := term.oldState

	if err = term.MakeRaw(); err != nil {
		t.Error("expected set raw mode:", err)
	}
	if err = term.Restore(); err != nil {
		t.Error("expected to restore:", err)
	}

	lastState := term.State

	if oldState.Iflag != lastState.Iflag ||
		oldState.Oflag != lastState.Oflag ||
		oldState.Cflag != lastState.Cflag ||
		oldState.Lflag != lastState.Lflag {

		t.Error("expected to restore all settings")
	}

	// Restore from a saved state
	term, _ = New(INPUT_FD)
	state := term.OriginalState()

	if err = Restore(term.fd, state); err != nil {
		t.Error("expected to restore from saved state:", err)
	}
}

func TestModes(t *testing.T) {
	term, _ := New(INPUT_FD)

	// Single-character

	err := term.SetSingleChar()
	if err != nil {
		t.Error("expected to set mode:", err)
	} else {
		buf := bufio.NewReaderSize(INPUT, 4)
		exit := false

		fmt.Print("\n + Mode to single character\n")

		if !*fInteractive {
			reply := []string{
				"a",
				"€",
				"~",
			}

			go func() {
				for _, r := range reply {
					time.Sleep(time.Duration(*fTime) * time.Second)
					fmt.Fprint(OUTPUT, r)
				}
				exit = true
			}()
		} else {
			exit = true
		}

		for {
			fmt.Print(" Press key: ")
			rune, _, err := buf.ReadRune()
			if err != nil {
				term.Restore()
				t.Fatal(err)
			}
			fmt.Printf("\n pressed: %q\n", string(rune))

			if exit {
				break
			}
		}
	}
	term.Restore()

	// Echo

	if term.SetEcho(false); err != nil {
		t.Error("expected to set mode:", err)
	} else {
		buf := bufio.NewReader(INPUT)

		fmt.Print("\n + Mode to echo off\n")

		if !*fInteractive {
			go func() {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprint(OUTPUT, "Karma\n")
			}()
		}
		fmt.Print(" Write (enter to finish): ")
		line, err := buf.ReadString('\n')
		if err != nil {
			term.Restore()
			t.Fatal(err)
		}
		fmt.Printf("\n entered: %q\n", line)

		term.SetEcho(true)
		fmt.Print("\n + Mode to echo on\n")

		if !*fInteractive {
			go func() {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprint(OUTPUT, "hotel\n")
			}()
		}
		fmt.Print(" Write (enter to finish): ")
		line, _ = buf.ReadString('\n')
		if !*fInteractive {
			fmt.Println()
		}
		fmt.Printf(" entered: %q\n", line)
	}

	term.Restore()
	fmt.Println()
}

func TestPassword(t *testing.T) {
	fmt.Print("\n Password: ")
	pass := make([]byte, 8)

	/*if !*fInteractive {
		go func() {
			time.Sleep(time.Duration(*fTime) * time.Second)
			fmt.Fprint(OUTPUT, "Parallel universe\n\n")
		}()
	}*/

	n, err := ReadPassword(INPUT_FD, pass)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("\n entered: %q\n number: %d\n\n", pass, n)
}

func TestInformation(t *testing.T) {
	term, _ := New(INPUT_FD)
	defer term.Restore()

	if !CheckANSI() {
		t.Error("expected to support this terminal")
	}

	if !IsTerminal(term.fd) {
		t.Error("expected to be a terminal")
	}

	/*if _, err := TTYName(term.fd); err != nil {
		t.Error("expected to get the terminal name", err)
	}*/
}

func TestSize(t *testing.T) {
	term, _ := New(INPUT_FD)
	defer term.Restore()

	row, col, err := term.GetSize()
	if err != nil {
		term.Restore()
		t.Fatal(err)
	}
	if row == 0 || col == 0 {
		t.Error("expected to get size")
	}

	rowE, colE := GetSizeFromEnv()
	if rowE == 0 || colE == 0 {
		t.Error("expected to get size from environment")
	}

	// Detect window size

	TrapSize()
	fmt.Println("[Change the size of the terminal]")

	go func() { // I want to finish the test
		time.Sleep(5 * time.Second)
		WinSizeChan <- 0
	}()

	<-WinSizeChan

	row2, col2, err := term.GetSize()
	if err != nil {
		term.Restore()
		t.Fatal(err)
	}
	if row == row2 || col == col2 {
		t.Error("the terminal size got the same value")
	}
}
