// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

/* The Windows headers can be got from directory "MinGW/include" */

package console

const (
	// wincon.h
	_ENABLE_LINE_INPUT         = 2
	_ENABLE_ECHO_INPUT         = 4
	_ENABLE_PROCESSED_INPUT    = 1
	_ENABLE_WINDOW_INPUT       = 8
	_ENABLE_MOUSE_INPUT        = 16
	_ENABLE_INSERT_MODE        = 32
	_ENABLE_QUICK_EDIT_MODE    = 64
	_ENABLE_EXTENDED_FLAGS     = 128
	_ENABLE_AUTO_POSITION      = 256
	_ENABLE_PROCESSED_OUTPUT   = 1
	_ENABLE_WRAP_AT_EOL_OUTPUT = 2
)
