// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

package console

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

func TestModeRaw(t *testing.T) {
	term, err := New(syscall.Stderr)
	if err != nil {
		t.Fatal(err)
	}

	if err = term.MakeRaw(); err != nil {
		t.Error("expected set raw mode:", err)
	}
	if err = term.Restore(); err != nil {
		t.Error("expected to restore:", err)
	}

	// == Restoring from a saved state.
	term, _ = New(syscall.Stderr)
	state := term.OriginalState()

	if err = term.Echo(false); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}
	if err = Restore(term.Fd, state); err != nil {
		t.Error("expected to restore from saved state:", err)
	}
}

func TestInformation(t *testing.T) {
	term, _ := New(syscall.Stderr)
	defer term.Restore()

	if !CheckANSI() {
		t.Error("expected to support this terminal")
	}

	_, err := IsTTY(term.Fd)
	if err != nil {
		t.Error("expected to be a terminal")
	}

	if _, err = TTYName(term.Fd); err != nil {
		t.Error("expected to get the terminal name", err)
	}
}

func TestSize(t *testing.T) {
	term, _ := New(syscall.Stderr)
	defer term.Restore()

	ws, err := term.WinSize()
	if err != nil {
		t.Fatal(err)
	}
	if ws.Row == 0 || ws.Col == 0 {
		t.Error("expected to get size of rows and columns")
	}

	row, col := term.GetSize()
	if row == 0 || col == 0 {
		t.Error("expected to get size in characters of rows and columns")
	}

	// == Detect window size

	TrapSize()
	fmt.Println("[Change the size of the window]")

	go func() { // I want to finish the test
		time.Sleep(5 * time.Second)
		WinSizeChan <- 0
	}()

	<-WinSizeChan

	ws2, err := term.WinSize()
	if err != nil {
		t.Fatal(err)
	}

	if ws.Row == ws2.Row || ws.Col == ws2.Col {
		t.Error("the window size got the same value")
	}
}
