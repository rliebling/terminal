// Copyright 2012 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build windows

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/kless/terminal"
)

var (
	fInteractive = flag.Bool("int", false, "mode interactive")
	fTime        = flag.Uint("t", 2, "time in seconds to wait to write")
)

var (
	INPUT    io.Reader
	OUTPUT   io.Writer
	INPUT_FD = syscall.Stdin
)

func main() {
	flag.Parse()

	if *fInteractive {
		INPUT = os.Stdin
		OUTPUT = os.Stdout
	} else {
		INPUT, OUTPUT = io.Pipe()
	}

	var t *testing.T
	TestRawMode(t)
	TestModes(t)
	TestInformation(t)
}

func TestRawMode(t *testing.T) {
	term, err := terminal.New(INPUT_FD)
	if err != nil {
		t.Fatal(err)
	}

//	oldState := term.oldState

	if err = term.MakeRaw(); err != nil {
		t.Error("expected set raw mode:", err)
	}

/*name, err := term.GetName()
if err != nil {
	log.Print(err)
}
println("Console", name)*/

	if err = term.Restore(); err != nil {
		t.Error("expected to restore:", err)
	}

/*	lastState := term.State
	if oldState != lastState {
		t.Error("expected to restore settings")
	}
*/
	// Restore from a saved state
	term, _ = terminal.New(INPUT_FD)
	state := term.OriginalState()

	if err = terminal.Restore(INPUT_FD, state); err != nil {
		t.Error("expected to restore from saved state:", err)
	}
}

func TestModes(t *testing.T) {
	term, _ := terminal.New(INPUT_FD)

	// Single-character

	err := term.SetSingleChar()
	if err != nil {
		log.Print("expected to set mode:", err)
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
				log.Fatal(err)
			}
			fmt.Printf("\n pressed: %q\n", string(rune))

			if exit {
				break
			}
		}
	}
	term.Restore()

	// Echo

	if err := term.SetEcho(false); err != nil {
		log.Print("expected to set mode:", err)
	} else {
		buf := bufio.NewReader(INPUT)

		fmt.Print("\n + Mode to echo off\n")

		if !*fInteractive {
			go func() {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprint(OUTPUT, "Karma\r\n")
			}()
		}
		fmt.Print(" Write (enter to finish): ")
		line, err := buf.ReadString('\n')
		if err != nil {
			term.Restore()
			log.Fatal(err)
		}
		fmt.Printf("\n entered: %q\n", line)

		term.SetEcho(true)
		fmt.Print("\n + Mode to echo on\n")

		if !*fInteractive {
			go func() {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprint(OUTPUT, "hotel\r\n")
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

	// Password

	/*if !*fInteractive {
		go func() {
			time.Sleep(time.Duration(*fTime) * time.Second)
			fmt.Fprint(OUTPUT, "Parallel universe\n\n")
		}()
	}*/
/*
	if *fInteractive {
		fmt.Print("\n Password: ")
		pass := make([]byte, 8)

		n, err := ReadPassword(INPUT_FD, pass)
		if err != nil {
			log.Error(err)
		}
		fmt.Printf("\n entered: %q\n number: %d\n", pass, n)
	}

	fmt.Println()*/
}


func TestInformation(t *testing.T) {
	term, _ := terminal.New(INPUT_FD)
	defer term.Restore()

/*	if !CheckANSI() {
		t.Error("expected to support this terminal")
	}
*/
/*	if !terminal.IsTerminal(syscall.Handle(3)) {
		t.Error("expected to be a terminal")
	}*/

	/*if _, err := TTYName(term.fd); err != nil {
		t.Error("expected to get the terminal name", err)
	}*/
}
/*
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
}*/
