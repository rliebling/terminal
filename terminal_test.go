// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// +build !plan9,!windows

package terminal

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

func TestRawMode(t *testing.T) {
	term, err := New(syscall.Stderr)
	if err != nil {
		t.Fatal(err)
	}

	if err = term.MakeRaw(); err != nil {
		t.Error("expected set raw mode:", err)
	}

	if err = term.SetEcho(false); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}
	if err = term.SetEcho(true); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}

	if err = term.Restore(); err != nil {
		t.Error("expected to restore:", err)
	}

	// == Restoring from a saved state.
	term, _ = New(syscall.Stderr)
	state := term.OriginalState()

	if err = term.SetEcho(false); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}
	if err = Restore(term.fd, state); err != nil {
		t.Error("expected to restore from saved state:", err)
	}
}

func TestInformation(t *testing.T) {
	term, _ := New(syscall.Stderr)
	defer term.Restore()

	if !CheckANSI() {
		t.Error("expected to support this terminal")
	}

	if !IsTerminal(term.fd) {
		t.Error("expected to be a terminal")
	}

	if _, err := TTYName(term.fd); err != nil {
		t.Error("expected to get the terminal name", err)
	}
}

func TestSize(t *testing.T) {
	term, _ := New(syscall.Stderr)
	defer term.Restore()

	row, col, err := term.GetSize()
	if err != nil {
		t.Fatal(err)
	}
	if row == 0 || col == 0 {
		t.Error("expected to get size of rows and columns")
	}

	rowE, colE := term.GetSizeFromEnv()
	if rowE == 0 || colE == 0 {
		t.Error("expected to get size from environment")
	}

	// == Detect window size

	TrapSize()
	fmt.Println("[Change the size of the terminal]")

	go func() { // I want to finish the test
		time.Sleep(5 * time.Second)
		WinSizeChan <- 0
	}()

	<-WinSizeChan

	row2, col2, err := term.GetSize()
	if err != nil {
		t.Fatal(err)
	}
	if row == row2 || col == col2 {
		t.Error("the terminal size got the same value")
	}
}
