// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

/*
To get the header file, if you have not this system:

Ver="1699.26.8"

wget -c http://opensource.apple.com/tarballs/xnu/xnu-${Ver}.tar.gz
tar zxf xnu*.tar.gz
find ./xnu* -type f -name "*.h" | xargs grep -i TIOCGWINSZ
*/

// #include <sys/ttycom.h>
import "C"

package console

const _TIOCGWINSZ = C.TIOCGWINSZ
