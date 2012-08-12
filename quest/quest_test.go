// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package quest

import (
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/kless/console/editline"
)

var (
	fInteractive = flag.Bool("int", false, "mode interactive")
	fTime        = flag.Uint("t", 2, "time in seconds to wait to write answer")

	// Mode interactive:
	// It is needed a fifo for os.Stdin (os.Stderr is used by 'go test') since
	// it is connected to the TTY. So, you write go to the terminal.
	pr *io.PipeReader
	pw *io.PipeWriter
)

func init() {
	flag.Parse()

	if *fInteractive {
		editline.Input = os.Stderr
	} else {
		pr, pw = io.Pipe()
		editline.Input = pr
	}
	editline.InputFd = syscall.Stderr
}

func TestQuest(t *testing.T) {
	var ans interface{}
	var err error

	fmt.Print("\n== Questions\n\n")

	q := NewDefault()
	defer q.Restore()

	if *fInteractive {
		q.ExitAtCtrlC(1)
	} else {
		Int := "-11"
		Float := "1.23"
		String := "foo"
		False := "false"
		True := "true"

		auto := map[int][]string{
			1:  []string{Float, True, "", "R. C."},
			2:  []string{""},
			3:  []string{String, Int},
			4:  []string{True, Int, ""},
			5:  []string{String, Float},
			6:  []string{String, ""},
			7:  []string{False},
			8:  []string{String, ""},
			9:  []string{String, "5", "2"},
			10: []string{"2", "1"},
			11: []string{String, ""},
			12: []string{"photo", "cryp", ""},
		}

		go func() {
			for i := 1; i <= 12; i++ {
				for _, v := range auto[i] {
					time.Sleep(time.Duration(*fTime) * time.Second)
					fmt.Fprintf(pw, "%s%s", v, editline.CRLF) // remember that terminal is in raw mode
				}
			}
		}()
	}

	q.NewPrompt("1. What is your name?").Mod(REQUIRED)
	ans, err = q.ReadString()
	print("\""+ans.(string)+"\"", err)

	q.NewPrompt("2. What color is your hair?").Default("brown")
	ans, err = q.ReadString()
	print(ans, err)

	q.NewPrompt("3. What temperature is there?").Default(-2)
	ans, err = q.ReadInt()
	print(ans, err)

	q.NewPrompt("4. How old are you?").Default(uint(16))
	ans, err = q.ReadUint()
	print(ans, err)

	q.NewPrompt("5. How tall are you?").Mod(REQUIRED)
	ans, err = q.ReadFloat()
	print(ans, err)

	q.NewPrompt("6. Do you watch television?").Default(true)
	ans, err = q.ReadBool()
	print(ans, err)

	q.NewPrompt("7. Do you read books?").Default(false)
	ans, err = q.ReadBool()
	print(ans, err)

	color := []string{"red", "blue", "black"}
	q.NewPrompt("8. What is your favourite color?").Default("blue")
	ans, err = q.ChoiceString(color)
	print(ans, err)

	q.NewPrompt("9. Another favourite color?")
	ans, err = q.ChoiceString(color)
	print(ans, err)

	q.NewPrompt("10. Choose number").Default(uint(3))
	ans, err = q.ChoiceUint([]uint{1, 3, 5})
	print(ans, err)

	q.NewPrompt("11. Email").Default("ja@contac.me")
	ans, err = q.ReadEmail()
	print(ans, err)

	q.NewPrompt("12. Hobby")
	ans, err = q.ReadMultipleString()
	print(ans, err)
}

func TestQuestExtraBoolean(t *testing.T) {
	fmt.Println("\n===\n  *NOTE:* It has been added the boolean strings 'oui', 'non'")

	q := New(" > ", "  ERR:", "oui", "non")
	defer q.Restore()

	if *fInteractive {
		q.ExitAtCtrlC(1)
	} else {
		go func() {
			for _, v := range []string{"ja", "oui"} {
				time.Sleep(time.Duration(*fTime) * time.Second)
				fmt.Fprintf(pw, "%s%s", v, editline.CRLF)
			}
		}()
	}

	q.NewPrompt("13. Are you french?").Mod(REQUIRED)
	ans, err := q.ReadBool()
	print(ans, err)
}

// * * *

// Prints values returned.
func print(a interface{}, err error) {
	if err == nil {
		fmt.Printf("  answer: %v\r\n", a)
	} else if err != editline.ErrCtrlD {
		fmt.Printf("%s\r\n", err)
	}
}
