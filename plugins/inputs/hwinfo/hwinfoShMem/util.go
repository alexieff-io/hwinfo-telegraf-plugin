package hwinfoShMem

import (
	"bytes"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

var isoDecoder = charmap.ISO8859_1.NewDecoder()

// decodeCString reads a fixed-length C string encoded as ISO-8859-1 from the
// given byte slice. Bytes up to the first null (or end of slice) are decoded
// to UTF-8. ISO-8859-1 maps every byte 1:1, so decoding never fails in
// practice; on the theoretical error path we return the raw string unchanged.
func decodeCString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	decoded, err := isoDecoder.Bytes(b)
	if err != nil {
		return string(b)
	}
	return string(decoded)
}

// StartsWithLower reports whether str starts with substr, comparing in a
// case-insensitive manner.
func StartsWithLower(str, substr string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(substr))
}
