package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

var (
	ghnd C.HANDLE
	imut sync.Mutex
)

// LockMutex acquires the package-level Go mutex and opens the HWiNFO Windows
// mutex handle. On failure the Go mutex is released and the error is returned;
// the caller must NOT call UnlockMutex in that case.
func LockMutex() error {
	imut.Lock()
	lpName := C.CString(C.HWiNFO_SENSORS_SM2_MUTEX)
	defer C.free(unsafe.Pointer(lpName))

	ghnd = C.OpenMutex(C.READ_CONTROL, C.FALSE, lpName)
	if ghnd == C.HANDLE(C.NULL) {
		imut.Unlock()
		return handleLastError(uint64(C.GetLastError()))
	}
	return nil
}

// UnlockMutex releases the Windows handle and the Go mutex. Must be paired
// with a successful LockMutex call.
func UnlockMutex() {
	if ghnd != C.HANDLE(C.NULL) {
		C.CloseHandle(ghnd)
		ghnd = C.HANDLE(C.NULL)
	}
	imut.Unlock()
}

var (
	ErrFileNotFound  = errors.New("could not find HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")
	ErrInvalidHandle = errors.New("could not read HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")
)

type UnknownError struct {
	Code uint64
}

func (e UnknownError) Error() string {
	return fmt.Sprintf("unknown error code: %d", e.Code)
}

func handleLastError(code uint64) error {
	switch code {
	case 2: // ERROR_FILE_NOT_FOUND
		return ErrFileNotFound
	case 6: // ERROR_INVALID_HANDLE
		return ErrInvalidHandle
	default:
		return UnknownError{Code: code}
	}
}
