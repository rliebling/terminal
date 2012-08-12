// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

package console

// /usr/include/asm-generic/ioctls.h
const (
	_TIOCGWINSZ = 0x5413
	_TIOCSWINSZ = 0x5414
)

// /usr/include/asm-generic/termios.h
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}
