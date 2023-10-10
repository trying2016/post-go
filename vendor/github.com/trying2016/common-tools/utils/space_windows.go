package utils

import (
	"syscall"
	"unsafe"
)

type DiskStatus struct {
	All uint64 `json:"all"`

	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	_, _, _ = c.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&disk.Free)), uintptr(unsafe.Pointer(&disk.All)), uintptr(unsafe.Pointer(&disk.Used)))
	return
}
