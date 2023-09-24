package codec

import (
	"reflect"
	"unsafe"
)

// stringToBytes converts a string to a byte slice without copying the underlying data.
// IMPORTANT: The returned byte slice must not be modified!
// This is a low-level function and should be used carefully.
func stringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

// bytesToString converts a byte slice to a string without copying the underlying data.
// This is a low-level function and should be used carefully.
func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
