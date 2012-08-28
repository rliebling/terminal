// Any copyright is dedicated to the Public Domain.
// http://creativecommons.org/publicdomain/zero/1.0/

// Maker create Go code from C header files.
// Run: go run Maker/make.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("FAIL: ")

	srcFilename := fmt.Sprintf("types_%s.go", runtime.GOOS)
	dstFilename := fmt.Sprintf("ztypes_%s_%s.go", runtime.GOOS, runtime.GOARCH)

	cmd := exec.Command("go", "tool", "cgo", "-godefs", srcFilename)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(dstFilename, out, 0644); err != nil {
		log.Fatal(err)
	}

	if err = exec.Command("gofmt", "-w", dstFilename).Run(); err != nil {
		log.Fatal(err)
	}
}
