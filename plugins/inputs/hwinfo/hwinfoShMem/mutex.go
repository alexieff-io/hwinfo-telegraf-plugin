package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// mutexWaitMillis is the longest we wait to acquire the HWiNFO shared-memory
// mutex before giving up on a read cycle. The HWiNFO SDK asks owners to
// release "as quick as possible", so a stuck mutex past ~1 second indicates a
// real problem rather than normal contention.
const mutexWaitMillis = 1000

var (
	ghnd C.HANDLE
	imut sync.Mutex
)

// LockMutex acquires the package-level Go mutex and the HWiNFO Windows mutex
// (via OpenMutex + WaitForSingleObject). On failure the Go mutex is released
// and the error is returned; the caller must NOT call UnlockMutex in that case.
func LockMutex() error {
	imut.Lock()
	lpName := C.CString(C.HWiNFO_SENSORS_SM2_MUTEX)
	defer C.free(unsafe.Pointer(lpName))

	h := C.OpenMutex(C.SYNCHRONIZE, C.FALSE, lpName)
	if h == C.HANDLE(C.NULL) {
		err := handleLastError(uint64(C.GetLastError()))
		imut.Unlock()
		return err
	}

	switch C.WaitForSingleObject(h, C.DWORD(mutexWaitMillis)) {
	case C.WAIT_OBJECT_0:
		ghnd = h
		return nil
	case C.WAIT_ABANDONED:
		// Previous owner crashed without releasing; we technically have
		// ownership now, but the shared memory may be mid-write. Safer to
		// release and ask the caller to retry next tick.
		C.ReleaseMutex(h)
		C.CloseHandle(h)
		imut.Unlock()
		return ErrMutexAbandoned
	case C.WAIT_TIMEOUT:
		C.CloseHandle(h)
		imut.Unlock()
		return ErrMutexTimeout
	default: // WAIT_FAILED or other unexpected value
		err := handleLastError(uint64(C.GetLastError()))
		C.CloseHandle(h)
		imut.Unlock()
		return fmt.Errorf("WaitForSingleObject on HWiNFO mutex failed: %w", err)
	}
}

// UnlockMutex releases the Windows mutex, closes the handle, and releases the
// Go mutex. Must be paired with a successful LockMutex call.
func UnlockMutex() {
	if ghnd != C.HANDLE(C.NULL) {
		C.ReleaseMutex(ghnd)
		C.CloseHandle(ghnd)
		ghnd = C.HANDLE(C.NULL)
	}
	imut.Unlock()
}

var (
	ErrFileNotFound   = errors.New("could not find HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")
	ErrInvalidHandle  = errors.New("could not read HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")
	ErrMutexTimeout   = errors.New("timed out waiting for HWiNFO shared memory mutex")
	ErrMutexAbandoned = errors.New("HWiNFO shared memory mutex was abandoned by its previous owner")
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
