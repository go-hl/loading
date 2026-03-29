package loading

import (
	"syscall"
	"unsafe"
)

type winSize struct {
	rows uint16
	cols uint16
}

func getWinSize() (int, int) {
	var ws winSize
	syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)
	return int(ws.rows), int(ws.cols)
}
