package storage

import (
	"fmt"
	"syscall"
	"unsafe"
)

// diskFreeBytes returns the number of free bytes on the volume containing path.
func diskFreeBytes(path string) (int64, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable uint64
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("invalid path: %w", err)
	}

	ret, _, errNo := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		0,
		0,
	)
	if ret == 0 {
		return 0, fmt.Errorf("GetDiskFreeSpaceExW failed: %w", errNo)
	}

	return int64(freeBytesAvailable), nil
}
