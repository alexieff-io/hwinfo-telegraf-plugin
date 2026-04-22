//go:build !windows

package hwinfoShMem

import "errors"

// ErrUnsupportedOS is returned by Read on non-Windows platforms. The plugin
// can be compiled for any OS so the package is importable from tests, but
// HWiNFO's shared memory interface only exists on Windows.
var ErrUnsupportedOS = errors.New("HWiNFO shared memory access is only supported on Windows")

func snapshotSharedMemory() ([]byte, error) {
	return nil, ErrUnsupportedOS
}
