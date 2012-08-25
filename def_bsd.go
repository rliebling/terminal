// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

/* The Darwin headers can be got from http://opensource.apple.com/tarballs/xnu/ */

// +build darwin freebsd netbsd openbsd

package console

// #include <sys/ttycom.h>
import "C"

const _TIOCGWINSZ = C.TIOCGWINSZ
