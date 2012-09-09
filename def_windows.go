// Copyright 2012 Jonas mg
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/* The Windows headers can be got from directory "MinGW/include" */

package terminal

// == wincon.h

const (
	ENABLE_LINE_INPUT         = 2
	ENABLE_ECHO_INPUT         = 4
	ENABLE_PROCESSED_INPUT    = 1
	ENABLE_WINDOW_INPUT       = 8
	ENABLE_MOUSE_INPUT        = 16
	ENABLE_INSERT_MODE        = 32
	ENABLE_QUICK_EDIT_MODE    = 64
	ENABLE_EXTENDED_FLAGS     = 128
	ENABLE_AUTO_POSITION      = 256
	ENABLE_PROCESSED_OUTPUT   = 1
	ENABLE_WRAP_AT_EOL_OUTPUT = 2
)

const (
	_KEY_EVENT = 1
)
/*
#define KEY_EVENT 1
#define MOUSE_EVENT 2
#define WINDOW_BUFFER_SIZE_EVENT 4
#define MENU_EVENT 8
#define FOCUS_EVENT 16
#define CAPSLOCK_ON 128
#define ENHANCED_KEY 256
#define RIGHT_ALT_PRESSED 1
#define LEFT_ALT_PRESSED 2
#define RIGHT_CTRL_PRESSED 4
#define LEFT_CTRL_PRESSED 8
#define SHIFT_PRESSED 16
#define NUMLOCK_ON 32
#define SCROLLLOCK_ON 64
#define FROM_LEFT_1ST_BUTTON_PRESSED 1
#define RIGHTMOST_BUTTON_PRESSED 2
#define FROM_LEFT_2ND_BUTTON_PRESSED 4
#define FROM_LEFT_3RD_BUTTON_PRESSED 8
#define FROM_LEFT_4TH_BUTTON_PRESSED 16
#define MOUSE_MOVED	1
#define DOUBLE_CLICK	2
#define MOUSE_WHEELED	4
*/

// typedef struct _CONSOLE_SCREEN_BUFFER_INFO {
//	COORD	dwSize;
//	COORD	dwCursorPosition;
//	WORD	wAttributes;
//	SMALL_RECT srWindow;
//	COORD	dwMaximumWindowSize;
// } CONSOLE_SCREEN_BUFFER_INFO,*PCONSOLE_SCREEN_BUFFER_INFO;

type _CONSOLE_SCREEN_BUFFER_INFO struct {
	dwSize              _COORD
	dwCursorPosition    _COORD
	wAttributes         uint16
	srWindow            _SMALL_RECT
	dwMaximumWindowSize _COORD
}

// typedef struct _INPUT_RECORD {
//	WORD EventType;
//	union {
//		KEY_EVENT_RECORD KeyEvent;
//		MOUSE_EVENT_RECORD MouseEvent;
//		WINDOW_BUFFER_SIZE_RECORD WindowBufferSizeEvent;
//		MENU_EVENT_RECORD MenuEvent;
//		FOCUS_EVENT_RECORD FocusEvent;
//	} Event;
//} INPUT_RECORD,*PINPUT_RECORD;

type _INPUT_RECORD struct {
	EventType uint16
	_Event
}

type _Event struct {
	KeyEvent _KEY_EVENT_RECORD
	MouseEvent _MOUSE_EVENT_RECORD
	WindowBufferSizeEvent _WINDOW_BUFFER_SIZE_RECORD
	MenuEvent _MENU_EVENT_RECORD
	FocusEvent _FOCUS_EVENT_RECORD
}

// * * *

// typedef struct _SMALL_RECT {
//	SHORT Left;
//	SHORT Top;
//	SHORT Right;
//	SHORT Bottom;
// } SMALL_RECT, *PSMALL_RECT;

type _SMALL_RECT struct {
	left, top, right, bottom int16
}

// typedef struct _COORD {
//	SHORT X;
//	SHORT Y;
// } COORD, *PCOORD;

type _COORD struct {
	x, y int16
}

// typedef struct _KEY_EVENT_RECORD {
//	BOOL bKeyDown;
//	WORD wRepeatCount;
//	WORD wVirtualKeyCode;
//	WORD wVirtualScanCode;
//	union {
//		WCHAR UnicodeChar;
//		CHAR AsciiChar;
//	} uChar;
//	DWORD dwControlKeyState;
// }
// #ifdef __GNUC__
// /* gcc's alignment is not what win32 expects */
// PACKED
// #endif
// KEY_EVENT_RECORD;

type _KEY_EVENT_RECORD struct {
	bKeyDown bool
	wRepeatCount uint16
	wVirtualKeyCode uint16
	wVirtualScanCode uint16
	uChar
	dwControlKeyState uint32
}

type uChar struct {
	UnicodeChar rune
	AsciiChar byte
}

// typedef struct _MOUSE_EVENT_RECORD {
//	COORD dwMousePosition;
//	DWORD dwButtonState;
//	DWORD dwControlKeyState;
//	DWORD dwEventFlags;
// } MOUSE_EVENT_RECORD;

type _MOUSE_EVENT_RECORD struct {
	dwMousePosition _COORD
	dwButtonState uint32
	dwControlKeyState uint32
	dwEventFlags uint32
}

// typedef struct _WINDOW_BUFFER_SIZE_RECORD {	COORD dwSize; } WINDOW_BUFFER_SIZE_RECORD;

type _WINDOW_BUFFER_SIZE_RECORD struct {
	dwSize _COORD
}

// typedef struct _MENU_EVENT_RECORD {	UINT dwCommandId; } MENU_EVENT_RECORD,*PMENU_EVENT_RECORD;
type _MENU_EVENT_RECORD struct {
	dwCommandId uint32
}

// typedef struct _FOCUS_EVENT_RECORD {	BOOL bSetFocus; } FOCUS_EVENT_RECORD;

type _FOCUS_EVENT_RECORD struct {
	bSetFocus bool
}
