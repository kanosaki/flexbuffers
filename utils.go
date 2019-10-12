package flexbuffers

import "C"
import (
	"bytes"
	"reflect"
	"unsafe"
)

func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func cstringBytesToString(s []byte) string {
	l := bytes.IndexByte(s, 0)
	if l < 0 {
		panic("null not found")
	}
	ss := s[0:l]
	return *(*string)(unsafe.Pointer(&ss))
}

func bytesToString(s []byte) string {
	return *(*string)(unsafe.Pointer(&s))
}

func unsafeBufferString(buf []byte, offset, size int) string {
	var sh reflect.StringHeader
	sh.Len = size
	sh.Data = uintptr(unsafe.Pointer(&buf[offset]))
	return *(*string)(unsafe.Pointer(&sh))
}

func unsafeReadCString(buf []byte, offset int) string {
	size := bytes.IndexByte(buf[offset:], 0)
	if size < 0 {
		panic("null not found")
	}
	return unsafeBufferString(buf, offset, size)
}
