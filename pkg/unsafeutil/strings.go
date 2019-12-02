package unsafeutil

import (
	"reflect"
	"unsafe"
)

// From: https://github.com/valyala/fastjson
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// From: https://github.com/valyala/fastjson
func S2B(s string) []byte {
	strh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	var sh reflect.SliceHeader
	sh.Data = strh.Data
	sh.Len = strh.Len
	sh.Cap = strh.Len
	return *(*[]byte)(unsafe.Pointer(&sh))
}
