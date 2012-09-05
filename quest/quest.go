// Copyright 2010 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package quest provides functions for printing questions and validating answers.
//
// The test enables to run in mode interactive (flag -int), or automatically (by
// default) where the answers are written directly every n seconds (flag -t).
package quest

import (
	"fmt"
	"os"
	"strings"

	"github.com/kless/terminal"
	"github.com/kless/terminal/editline"
	"github.com/kless/validate"
)

// == Copied from package validate.

// Represents the modifications to do to the text.
const (
	NONE       validate.Modifier = 1 << iota
	REQUIRED                     // a required field
	TRIM_SPACE                   // trim spaces
	LOWERCASE                    // convert to lower case
	UPPERCASE                    // convert to upper case
)

// * * *

// A Question represents a question.
type Question struct {
	isMultiple bool              // accept multiple answers
	mod        validate.Modifier // modifiers used at getting the value

	prefix    string // string placed before of questions
	errPrefix string // string placed before of error messages
	prompt    string // the question

	defString string      // default value in string, to be showed in the question
	defValue  interface{} // default value to return if the input is empty

	trueString  string // strings that represent booleans
	falseString string
	extraBool   map[string]bool // to pass it to validate.Atob

	// To restore the terminal original settings
	termFD    int
	termState terminal.State
}

// New returns a Question with the given arguments.
//
// prefix is the text placed before of the prompt, errPrefix is placed before of
// show any error.
//
// trueString and falseString are the strings to be showed when the question
// needs a boolean like answer and it is being used a default value.
// It is already handled the next strings like boolean values (from validate.Atob):
//   1, t, T, TRUE, true, True, y, Y, yes, YES, Yes
//   0, f, F, FALSE, false, False, n, N, no, NO, No
func New(prefix, errPrefix, trueString, falseString string) *Question {
	term, err := terminal.New(editline.InputFd)
	if err != nil {
		panic(err)
	}

	extraBool := make(map[string]bool)
	val := validate.New(validate.Bool, validate.NONE)

	// Add strings of boolean values if there are not validated like boolean.
	if _, err = val.Atob(trueString); err != nil {
		extraBool = map[string]bool{
			strings.ToLower(trueString): true,
			strings.ToUpper(trueString): true,
			strings.Title(trueString):   true,
		}
	}
	if _, err = val.Atob(falseString); err != nil {
		extraBool[strings.ToLower(falseString)] = false
		extraBool[strings.ToUpper(falseString)] = false
		extraBool[strings.Title(falseString)] = false
	}

	return &Question{
		false,
		NONE,

		prefix,
		errPrefix,
		"",

		"",
		nil,

		trueString,
		falseString,
		extraBool,

		term.Fd(),
		term.OriginalState(),
	}
}

// Values by default to use in NewDefault.
const (
	q_PREFIX          = " + "
	q_MULTIPLE_PREFIX = "   * "
	q_ERR_PREFIX      = "  ERROR:"
	q_TRUE_STRING     = "y"
	q_FALSE_STRING    = "n"
)

// NewDefault returns a Question using default values.
func NewDefault() *Question {
	return New(q_PREFIX, q_ERR_PREFIX, q_TRUE_STRING, q_FALSE_STRING)
}

// Restore restores terminal settings.
func (q *Question) Restore() error {
	if err := terminal.Restore(q.termFD, q.termState); err != nil {
		return err
	}
	return nil
}

// ExitAtCtrlC exits when it is pressed Ctrl-C, with value n.
func (q *Question) ExitAtCtrlC(n int) {
	go func() {
		select {
		case <-editline.ChanCtrlC:
			q.Restore()
			editline.Output.Write(editline.DelLine_CR)
			os.Exit(n)
		}
	}()
}

// * * *

// NewPrompt sets a new prompt.
func (q *Question) NewPrompt(str string) *Question {
	q.prompt = str
	return q
}

// Mod sets the modifier to apply to the input.
func (q *Question) Mod(m validate.Modifier) *Question {
	q.mod = m
	return q
}

// Default sets a value by default.
func (q *Question) Default(def interface{}) *Question {
	// Check if it is a nil string.
	if str, ok := def.(string); ok && str == "" {
		return q
	}
	q.defValue = def
	return q
}

// clean removes data related to the prompt.
func (q *Question) clean() {
	q.prompt = ""
	q.defString = ""
	q.defValue = nil
	q.mod = NONE
}

// == Generic to read

// read is the base to read.
func (q *Question) read(line *editline.Line, valida *validate.Validate) (interface{}, error) {
	var hadError bool

	for {
		input, err := line.Read()
		if err != nil {
			return "", err
		}

		// == Validation
		val, err := valida.Get(input)

		if err == validate.ErrRequired {
			os.Stderr.Write(editline.DelLine_CR)
			fmt.Fprintf(os.Stderr, "%s it %s", q.errPrefix, err)
			editline.Output.Write(editline.CursorUp)
			hadError = true
			continue
		}

		// Error of type.
		if err != nil {
			os.Stderr.Write(editline.DelLine_CR)
			fmt.Fprintf(os.Stderr, "%s %q %s", q.errPrefix, input, err)
			editline.Output.Write(editline.CursorUp)
			hadError = true
			continue
		}

		if hadError {
			os.Stderr.Write(editline.DelLine_CR)
		}
		return val, nil
	}
	return nil, nil
}

// == Basic types

// readType is the base to read the basic types.
func (q *Question) readType(kind validate.Kind) (value interface{}, err error) {
	valida := validate.New(kind, q.mod)

	if kind == validate.Bool {
		valida.SetBoolString(q.extraBool)
	}

	// Default value
	if q.defValue != nil {
		valida.SetDefault(q.defValue)

		if kind != validate.Bool {
			q.defString = defaultToPrint(q.defValue)
		} else {
			q.defString = q.defaultBoolToPrint(q.defValue.(bool))
		}
	}

	value, err = q.read(q.newLine(), valida)
	q.clean()
	return
}

// ReadBool prints the prompt waiting to get a string that represents a boolean.
func (q *Question) ReadBool() (bool, error) {
	value, err := q.readType(validate.Bool)
	return value.(bool), err
}

// ReadInt prints the prompt waiting to get an integer number.
func (q *Question) ReadInt() (int, error) {
	value, err := q.readType(validate.Int)
	return value.(int), err
}

// ReadUint prints the prompt waiting to get an unsigned integer number.
func (q *Question) ReadUint() (uint, error) {
	value, err := q.readType(validate.Uint)
	return value.(uint), err
}

// ReadFloat prints the prompt waiting to get a float number.
func (q *Question) ReadFloat() (float32, error) {
	value, err := q.readType(validate.Float32)
	return value.(float32), err
}

// ReadString prints the prompt waiting to get a string.
func (q *Question) ReadString() (string, error) {
	value, err := q.readType(validate.String)
	return value.(string), err
}

// ReadMultipleString is like ReadString but it can read multiple strings.
// To do not read any more, press Enter to finish.
func (q *Question) ReadMultipleString() ([]string, error) {
	res := make([]string, 0)

	if err := q.newLine().Prompt(); err != nil {
		return nil, err
	}
	editline.Output.Write(editline.CRLF)
	q.isMultiple = true

	for {
		v, err := q.ReadString()
		if err != nil {
			return nil, err
		}

		if v == "" {
			break
		}
		res = append(res, v)
	}
	return res, nil
}

// == Choices

// readChoice is the base to read choices.
func (q *Question) readChoice(choices interface{}) (value interface{}, err error) {
	valida := validate.NewSlice(choices, q.mod)

	// Default value
	if q.defValue != nil {
		valida.SetDefault(q.defValue)
		q.defString = defaultToPrint(q.defValue)
	}

	fmt.Fprintf(editline.Output, "  >>> %s\r\n", valida.JoinChoices())

	value, err = q.read(q.newLine(), valida)
	q.clean()
	return
}

// ChoiceInt prints the prompt waiting to get an integer number that is in the slice.
func (q *Question) ChoiceInt(choices []int) (int, error) {
	value, err := q.readChoice(choices)
	return value.(int), err
}

// ChoiceUint prints the prompt waiting to get an unsigned number that is in the slice.
func (q *Question) ChoiceUint(choices []uint) (uint, error) {
	value, err := q.readChoice(choices)
	return value.(uint), err
}

// ChoiceFloat prints the prompt waiting to get a float number that is in the slice.
func (q *Question) ChoiceFloat(choices []float32) (float32, error) {
	value, err := q.readChoice(choices)
	return value.(float32), err
}

// ChoiceString prints the prompt waiting to get a string that is in the slice.
func (q *Question) ChoiceString(choices []string) (string, error) {
	value, err := q.readChoice(choices)
	return value.(string), err
}

// == Regexp

// ReadRegexp prints the prompt waiting to get a string that matches with a regular expression.
func (q *Question) ReadRegexp(options, pattern string) (string, error) {
	valida := validate.NewRegexp(q.mod, options, pattern)
	if q.defValue != nil {
		valida.SetDefault(q.defValue)
		q.defString = defaultToPrint(q.defValue)
	}

	value, err := q.read(q.newLine(), valida)
	q.clean()
	return value.(string), err
}

// ReadEmail prints the prompt waiting to get a string that be valid like email.
func (q *Question) ReadEmail() (string, error) {
	return q.ReadRegexp("i", validate.StartOfString+validate.RE_RFC_EMAIL+validate.EndOfString)
}

// == Utility

// defaultToPrint returns the default value.
func defaultToPrint(val interface{}) string {
	return fmt.Sprintf(" [%s%v%s]", setBold, val, setOff)
}

// defaultBoolToPrint returns the default value for a boolean.
func (q *Question) defaultBoolToPrint(val bool) string {
	if val {
		return fmt.Sprintf(" [%s%s%s/%s]", setBold, q.trueString, setOff, q.falseString)
	}
	return fmt.Sprintf(" [%s/%s%s%s]", q.trueString, setBold, q.falseString, setOff)
}

// newLine gets a line type ready to show questions.
func (q *Question) newLine() *editline.Line {
	var extraChars int
	prompt := ""

	if !q.isMultiple {
		prompt = fmt.Sprintf("%s%s%s", q.prefix, q.prompt, q.defString)

		// Add spaces
		if strings.HasSuffix(prompt, "?") {
			prompt += " "
		} else if !q.isMultiple {
			prompt += ": "
		}

		// The default value uses ANSI characters.
		if q.defValue != nil {
			extraChars = lenAnsi
		}
	} else {
		prompt = q_MULTIPLE_PREFIX
	}

	ln, err := editline.NewLine(prompt, q.errPrefix, extraChars, nil) // No history.
	if err != nil {
		panic(err)
	}

	return ln
}
