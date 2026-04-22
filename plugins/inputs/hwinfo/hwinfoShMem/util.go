package hwinfoShMem

import "C"

import (
	"strings"
	"unsafe"

	"golang.org/x/text/encoding/charmap"
)

func goStringFromPtr(ptr unsafe.Pointer, length int) string {
	s := C.GoStringN((*C.char)(ptr), C.int(length))
	if idx := strings.IndexByte(s, 0); idx >= 0 {
		return s[:idx]
	}
	return s
}

// DecodeCharPtr decodes an ISO-8859-1 C string to UTF-8.
// ISO-8859-1 maps every possible byte 1:1, so decoding never fails in practice;
// on the theoretical error path we return the raw input unchanged.
func DecodeCharPtr(ptr unsafe.Pointer, length int) string {
	s := goStringFromPtr(ptr, length)
	ds, err := isodecoder.String(s)
	if err != nil {
		return s
	}
	return ds
}

var isodecoder = charmap.ISO8859_1.NewDecoder()

func StartsWithLower(str, substr string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(substr))
}
