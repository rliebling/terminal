/*
Package console provides a portable interface for console I/O.

Console manages input and output (I/O) for character-mode applications.
The high-level functions enable an application to read from standard input to
retrieve keyboard input stored in a console's input buffer. They also enable
an application to write to standard output or standard error to display text
in the console's screen buffer. And they also support redirection of standard
handles and control of console modes for different I/O functionality.

The low-level functions enable applications to receive detailed input about
keyboard. They also enable greater control of output to the screen.

Usage:

  term, err := console.New(0)
  if err != nil {
  	panic(err)
  }
  defer term.Restore()

Important

The "go test" tool runs tests with standard input connected to standard
error. So whatever program that uses the file descriptor of "/dev/stdin"
(which is 0), then it is going to fail. The solution is to use the standard
error.
*/
package console
