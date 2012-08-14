// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

/*
To get some header file in Darwin, if you have not that system:

Ver="1699.26.8"

wget -c http://opensource.apple.com/tarballs/xnu/xnu-${Ver}.tar.gz
tar zxf xnu*.tar.gz
find ./xnu* -type f -name "*.h" | xargs grep -i [NAME]
*/

// +build darwin freebsd netbsd openbsd

package console

// #include <sys/ttycom.h>
import "C"

const _TIOCGWINSZ = C.TIOCGWINSZ
