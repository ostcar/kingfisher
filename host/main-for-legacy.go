//go:build legacy

package main

//#cgo CFLAGS: -Wno-main-return-type
import "C"

//export main
func main() {
	entry()
}
