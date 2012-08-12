// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

/*
Get header file:

  Ver="1699.22.81"

  curl -s -O http://opensource.apple.com/tarballs/xnu/xnu-${Ver}.tar.gz
  tar zxf xnu-*.tar.gz

  find ./xnu-* -type f -name "*.h" | xargs grep -i TIOCGWINSZ

  def2go -s darwin ./xnu-1699.22.81/bsd/sys/ttycom.h
*/

package console

// ttycom.h
const (
	_TIOCGWINSZ = 0x40087468
	_TIOCSWINSZ = 0x80087467
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}
