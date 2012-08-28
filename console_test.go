// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

package console

import (
	"syscall"
	"testing"
)

func TestRawMode(t *testing.T) {
	con, err := New(syscall.Stderr)
	if err != nil {
		t.Fatal(err)
	}

	if err = con.MakeRaw(); err != nil {
		t.Error("expected set raw mode:", err)
	}

/*	if err = con.SetEcho(false); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}
	if err = con.SetEcho(true); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}
*/
	if err = con.Restore(); err != nil {
		t.Error("expected to restore:", err)
	}
/*
	// == Restoring from a saved state.
	con, _ = New(syscall.Stderr)
	state := con.OriginalState()

	if err = con.SetEcho(false); err != nil {
		t.Error("expected to turn the echo mode:", err)
	}
	if err = Restore(con.fd, state); err != nil {
		t.Error("expected to restore from saved state:", err)
	}*/
}
/*
func TestInformation(t *testing.T) {
	con, _ := New(syscall.Stderr)
	defer con.Restore()

	if !CheckANSI() {
		t.Error("expected to support this terminal")
	}

	_, err := IsTTY(con.fd)
	if err != nil {
		t.Error("expected to be a terminal")
	}

	if _, err = TTYName(con.fd); err != nil {
		t.Error("expected to get the terminal name", err)
	}
}*/
