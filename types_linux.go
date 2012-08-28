// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// +build ignore

/* Input to "go tool cgo -godefs" */

package console

// #include <asm-generic/ioctls.h>
// #include <asm-generic/termbits.h>
// #include <asm-generic/termios.h>
import "C"

const (
	TCGETS  = C.TCGETS
	TCSETS  = C.TCSETS
	TCSETSW = C.TCSETSW
	TCSETSF = C.TCSETSF

	TIOCGWINSZ = C.TIOCGWINSZ
	TIOCSWINSZ = C.TIOCSWINSZ
)

type termios C.struct_termios

type winsize C.struct_winsize

const (
	BRKINT = C.BRKINT
	ICRNL  = C.ICRNL
	IGNBRK = C.IGNBRK
	IGNCR  = C.IGNCR
	INLCR  = C.INLCR
	ISTRIP = C.ISTRIP
	IXON   = C.IXON
	PARMRK = C.PARMRK

	OPOST = C.OPOST

	ECHO   = C.ECHO
	ECHONL = C.ECHONL
	ICANON = C.ICANON
	IEXTEN = C.IEXTEN
	ISIG   = C.ISIG

	PARENB = C.PARENB
	CS8    = C.CS8
	CSIZE  = C.CSIZE

	VMIN  = C.VMIN
	VTIME = C.VTIME

	TCSADRAIN = C.TCSADRAIN
	TCSAFLUSH = C.TCSAFLUSH
	TCSANOW   = C.TCSANOW
)
