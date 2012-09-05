// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package editline provides simple functions for editing lines.
//
// Features:
//
//   Unicode support
//   History
//   Multi-line editing
//
// List of key sequences enabled (just like in GNU Readline):
//
//   Backspace / Ctrl+h
//
//   Delete
//   Home / Ctrl+a
//   End  / Ctrl+e
//
//   Left arrow  / Ctrl+b
//   Right arrow / Ctrl+f
//   Up arrow    / Ctrl+p
//   Down arrow  / Ctrl+n
//   Ctrl+left arrow
//   Ctrl+right arrow
//
//   Ctrl+t : swap actual character by the previous one
//   Ctrl+k : delete from current to end of line
//   Ctrl+u : delete the whole line
//   Ctrl+l : clear screen
//
//   Ctrl+c
//   Ctrl+d : exit
//
// Note that There are several default values:
//
// + For the buffer: BufferCap, BufferLen.
//
// + For the history file: HistoryCap, HistoryPerm.
//
// Important: the TTY is set in "raw mode" so there is to use CR+LF ("\r\n") for
// writing a new line.
package editline

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/kless/terminal"
)

// Default values for prompts.
const (
	_PS1 = "$ "
	_PS2 = "> "
)

// Input / Output
var (
	InputFd int       = syscall.Stdin
	Input   io.Reader = os.Stdin
	Output  io.Writer = os.Stdout
)

// To detect if has been pressed Ctrl+C
var ChanCtrlC = make(chan byte)

func init() {
	if !terminal.CheckANSI() {
		panic("Your terminal does not support ANSI")
	}
}

// A Line represents a line.
type Line struct {
	useHistory bool
	lenPS1     int      // Primary prompt size
	ps1        string   // Primary prompt
	ps2        string   // Command continuations
	buf        *buffer  // Text buffer
	hist       *history // History file
	term       *terminal.Terminal
}

// NewLine returns a line using both prompts ps1 and ps2, and setting the TTY to
// raw mode.
// lenAnsi is the length of ANSI codes that the prompt ps1 could have.
// If the history is nil then it is not used.
func NewLine(ps1, ps2 string, lenAnsi int, hist *history) (*Line, error) {
	term, err := terminal.New(InputFd)
	if err != nil {
		return nil, err
	}
	if err = term.MakeRaw(); err != nil {
		return nil, err
	}

	lenPS1 := len(ps1) - lenAnsi
	_, col, err := term.GetSize()
	if err != nil {
		return nil, err
	}

	buf := newBuffer(lenPS1, col)
	buf.insertRunes([]rune(ps1))

	return &Line{
		hasHistory(hist),
		lenPS1,
		ps1,
		ps2,
		buf,
		hist,
		term,
	}, nil
}

// NewDefaultLine returns a line type using the prompt by default, and setting
// the TTY to raw mode.
// If the history is nil then it is not used.
func NewDefaultLine(hist *history) (*Line, error) {
	term, err := terminal.New(InputFd)
	if err != nil {
		return nil, err
	}
	if err = term.MakeRaw(); err != nil {
		return nil, err
	}

	_, col, err := term.GetSize()
	if err != nil {
		return nil, err
	}

	buf := newBuffer(len(_PS1), col)
	buf.insertRunes([]rune(_PS1))

	return &Line{
		hasHistory(hist),
		len(_PS1),
		_PS1,
		_PS2,
		buf,
		hist,
		term,
	}, nil
}

// Restore restores the terminal settings, so it is disabled the raw mode.
func (ln *Line) Restore() {
	ln.term.Restore()
}

// Read reads charactes from input to write them to output, enabling line editing.
// The errors that could return are to indicate if Ctrl+D was pressed, and for
// both input/output errors.
func (ln *Line) Read() (line string, err error) {
	var anotherLine []rune // For lines got from history.
	var isHistoryUsed bool // If the history has been accessed.

	in := bufio.NewReader(Input) // Read input.
	esc := make([]byte, 2)       // For escape sequences.
	extEsc := make([]byte, 3)    // Extended escape sequences.

	// Print the primary prompt.
	if err = ln.Prompt(); err != nil {
		return "", err
	}

	// == Detect change of window size.
	terminal.TrapSize()

	go func() {
		for {
			select {
			case <-terminal.WinSizeChan: // Wait for.
				_, col, err := ln.term.GetSize()
				if err != nil {
					ln.buf.columns = col
					ln.buf.refresh()
				}
			}
		}
	}()

	for {
		rune, _, err := in.ReadRune()
		if err != nil {
			return "", inputError(err.Error())
		}

		switch rune {
		default:
			if err = ln.buf.insertRune(rune); err != nil {
				return "", err
			}
			continue

		case 13: // enter
			line = ln.buf.toString()

			if ln.useHistory {
				ln.hist.Add(line)
			}
			if _, err = Output.Write(CRLF); err != nil {
				return "", outputError(err.Error())
			}

			return strings.TrimSpace(line), nil

		case 127, 8: // backspace, Ctrl+h
			if err = ln.buf.deleteCharPrev(); err != nil {
				return "", err
			}
			continue

		case 9: // horizontal tab
			// TODO: disabled by now
			continue

		case 3: // Ctrl+c
			if err = ln.buf.insertRunes(ctrlC); err != nil {
				return "", err
			}
			if _, err = Output.Write(CRLF); err != nil {
				return "", outputError(err.Error())
			}

			ChanCtrlC <- 1

			if err = ln.Prompt(); err != nil {
				return "", err
			}

			continue

		case 4: // Ctrl+d
			if err = ln.buf.insertRunes(ctrlD); err != nil {
				return "", err
			}
			if _, err = Output.Write(CRLF); err != nil {
				return "", outputError(err.Error())
			}

			ln.Restore()
			return "", ErrCtrlD

		// Escape sequence
		case 27: // Escape: Ctrl+[ ("\x1b" in hexadecimal, "033" in octal)
			if _, err = in.Read(esc); err != nil {
				return "", inputError(err.Error())
			}

			if esc[0] == 79 { // 'O'
				switch esc[1] {
				case 72: // Home: "\x1b O H"
					goto _start
				case 70: // End: "\x1b O F"
					goto _end
				}
			}

			if esc[0] == 91 { // '['
				switch esc[1] {
				case 68: // "\x1b [ D"
					goto _leftArrow
				case 67: // "\x1b [ C"
					goto _rightArrow
				case 65, 66: // Up: "\x1b [ A"; Down: "\x1b [ B"
					goto _upDownArrow
				}

				// Extended escape.
				if esc[1] > 48 && esc[1] < 55 {
					if _, err = in.Read(extEsc); err != nil {
						return "", inputError(err.Error())
					}

					if extEsc[0] == 126 { // '~'
						switch esc[1] {
						//case 50: // Insert: "\x1b [ 2 ~"

						case 51: // Delete: "\x1b [ 3 ~"
							if err = ln.buf.deleteChar(); err != nil {
								return "", err
							}
							continue
							//case 53: // RePag: "\x1b [ 5 ~"

							//case 54: // AvPag: "\x1b [ 6 ~"

						}
					}
					if esc[1] == 49 && extEsc[0] == 59 && extEsc[1] == 53 { // "1;5"
						switch extEsc[2] {
						case 68: // Ctrl+left arrow: "\x1b [ 1 ; 5 D"
							// move to last word
							if err = ln.buf.wordBackward(); err != nil {
								return "", err
							}
							continue
						case 67: // Ctrl+right arrow: "\x1b [ 1 ; 5 C"
							// move to next word
							if err = ln.buf.wordForward(); err != nil {
								return "", err
							}
							continue
						}
					}
				}
			}
			continue

		case 20: // Ctrl+t, swap actual character by the previous one.
			if err = ln.buf.swap(); err != nil {
				return "", err
			}
			continue

		case 21: // Ctrl+u, delete the whole line.
			if err = ln.buf.deleteLine(); err != nil {
				return "", err
			}
			if err = ln.Prompt(); err != nil {
				return "", err
			}
			continue

		case 12: // Ctrl+l, clear screen.
			if _, err = Output.Write(delScreenToUpper); err != nil {
				return "", err
			}
			if err = ln.Prompt(); err != nil {
				return "", err
			}
			continue

		case 11: // Ctrl+k, delete from current to end of line.
			if err = ln.buf.deleteToRight(); err != nil {
				return "", err
			}
			continue

		case 1: // Ctrl+a, go to the start of the line.
			goto _start

		case 5: // Ctrl+e, go to the end of the line.
			goto _end

		case 2: // Ctrl+b
			goto _leftArrow

		case 6: // Ctrl+f
			goto _rightArrow

		case 16: // Ctrl+p
			esc[1] = 65
			goto _upDownArrow

		case 14: // Ctrl+n
			esc[1] = 66
			goto _upDownArrow
		}

	_upDownArrow: // Up and down arrow: history
		if !ln.useHistory {
			continue
		}

		// Up
		if esc[1] == 65 {
			anotherLine, err = ln.hist.Prev()
			// Down
		} else {
			anotherLine, err = ln.hist.Next()
		}
		if err != nil {
			continue
		}

		// Update the current history entry before to overwrite it with
		// the next one.
		// TODO: it has to be removed before of to be saved the history
		if !isHistoryUsed {
			ln.hist.Add(ln.buf.toString())
		}
		isHistoryUsed = true

		ln.buf.grow(len(anotherLine))
		ln.buf.size = len(anotherLine) + ln.buf.promptLen
		copy(ln.buf.data[ln.lenPS1:], anotherLine)

		if err = ln.buf.refresh(); err != nil {
			return "", err
		}
		continue

	_leftArrow:
		if _, err = ln.buf.backward(); err != nil {
			return "", err
		}
		continue

	_rightArrow:
		if _, err = ln.buf.forward(); err != nil {
			return "", err
		}
		continue

	_start:
		if err = ln.buf.start(); err != nil {
			return "", err
		}
		continue

	_end:
		if _, err = ln.buf.end(); err != nil {
			return "", err
		}
		continue
	}
	return
}

// Prompt prints the primary prompt.
func (ln *Line) Prompt() (err error) {
	if _, err = Output.Write(DelLine_CR); err != nil {
		return outputError(err.Error())
	}
	if _, err = fmt.Fprint(Output, ln.ps1); err != nil {
		return outputError(err.Error())
	}

	ln.buf.pos, ln.buf.size = ln.lenPS1, ln.lenPS1
	return
}

// == Utility

// hasHistory checks whether has an history file.
func hasHistory(h *history) bool {
	if h == nil {
		return false
	}
	return true
}
