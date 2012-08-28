// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// +build ignore

package console

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

func TestSize(t *testing.T) {
	con, _ := New(syscall.Stderr)
	defer con.Restore()

	row, col, err := con.GetSize()
	if err != nil {
		t.Fatal(err)
	}
	if row == 0 || col == 0 {
		t.Error("expected to get size of rows and columns")
	}

	/*row, col, _ := con.GetSize()
	if row == 0 || col == 0 {
		t.Error("expected to get size in characters of rows and columns")
	}*/

	// == Detect window size

	TrapSize()
	fmt.Println("[Change the size of the console]")

	go func() { // I want to finish the test
		time.Sleep(5 * time.Second)
		WinSizeChan <- 0
	}()

	<-WinSizeChan

	row2, col2, err := con.GetSize()
	if err != nil {
		t.Fatal(err)
	}
	if row == row2 || col == col2 {
		t.Error("the console size got the same value")
	}
}
