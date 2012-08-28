// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_linux.go

package console

const (
	TCGETS  = 0x5401
	TCSETS  = 0x5402
	TCSETSW = 0x5403
	TCSETSF = 0x5404

	TIOCGWINSZ = 0x5413
	TIOCSWINSZ = 0x5414
)

type termios struct {
	Iflag uint32
	Oflag uint32
	Cflag uint32
	Lflag uint32
	Line  uint8
	Cc    [19]uint8
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

const (
	BRKINT = 0x2
	ICRNL  = 0x100
	IGNBRK = 0x1
	IGNCR  = 0x80
	INLCR  = 0x40
	ISTRIP = 0x20
	IXON   = 0x400
	PARMRK = 0x8

	OPOST = 0x1

	ECHO   = 0x8
	ECHONL = 0x40
	ICANON = 0x2
	IEXTEN = 0x8000
	ISIG   = 0x1

	PARENB = 0x100
	CS8    = 0x30
	CSIZE  = 0x30

	VMIN  = 0x6
	VTIME = 0x5

	TCSADRAIN = 0x1
	TCSAFLUSH = 0x2
	TCSANOW   = 0x0
)
