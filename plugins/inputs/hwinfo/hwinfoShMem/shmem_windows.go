//go:build windows

package hwinfoShMem

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// mutexWaitMillis is the longest we wait to acquire the HWiNFO shared-memory
// mutex before giving up on a read cycle. The HWiNFO SDK asks owners to
// release "as quick as possible", so a stuck mutex past ~1 second indicates
// a real problem rather than normal contention.
const mutexWaitMillis = 1000

// maxSharedMemorySize is a sanity bound on the total HWiNFO shared memory
// size. Real-world mappings are around 100 KB; if the header reports a size
// larger than this, we refuse to read rather than risk an access violation.
const maxSharedMemorySize = 64 * 1024 * 1024

// We drive the mutex-state transitions from a single package-level mutex so
// concurrent Gather calls serialize cleanly; within the critical section we
// then perform the Windows-level acquire-open-map-copy-unmap-close dance.
var snapshotLock sync.Mutex

// Mutex errors exposed to callers so they can distinguish recoverable
// contention from hard failures.
var (
	ErrFileNotFound   = errors.New("could not find HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")
	ErrInvalidHandle  = errors.New("could not read HWiNFO shared memory file, is HWiNFO running with Shared Memory Support enabled?")
	ErrMutexTimeout   = errors.New("timed out waiting for HWiNFO shared memory mutex")
	ErrMutexAbandoned = errors.New("HWiNFO shared memory mutex was abandoned by its previous owner")
)

// procOpenFileMappingW is declared manually because golang.org/x/sys/windows
// does not export OpenFileMapping. We need it (rather than CreateFileMapping)
// so we can request FILE_MAP_READ instead of the library's PAGE_READWRITE.
var (
	modKernel32          = windows.NewLazySystemDLL("kernel32.dll")
	procOpenFileMappingW = modKernel32.NewProc("OpenFileMappingW")
)

func openFileMapping(access uint32, inheritHandle bool, name *uint16) (windows.Handle, error) {
	var inherit uintptr
	if inheritHandle {
		inherit = 1
	}
	r0, _, e1 := procOpenFileMappingW.Call(uintptr(access), inherit, uintptr(unsafe.Pointer(name)))
	if r0 == 0 {
		if e1 != nil && e1 != windows.NOERROR {
			return 0, e1
		}
		return 0, windows.GetLastError()
	}
	return windows.Handle(r0), nil
}

// snapshotSharedMemory holds the HWiNFO mutex, opens the named shared
// memory region read-only, copies its contents into a Go-owned byte slice,
// then releases everything. The returned slice is safe to use after this
// function returns.
func snapshotSharedMemory() ([]byte, error) {
	snapshotLock.Lock()
	defer snapshotLock.Unlock()

	mutexHandle, err := acquireHWiNFOMutex()
	if err != nil {
		return nil, err
	}
	defer releaseHWiNFOMutex(mutexHandle)

	mapName, err := windows.UTF16PtrFromString(sharedMemoryName)
	if err != nil {
		return nil, fmt.Errorf("build mapping name: %w", err)
	}

	mapHandle, err := openFileMapping(windows.FILE_MAP_READ, false, mapName)
	if err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return nil, ErrAccessDenied
		}
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("OpenFileMapping: %w", err)
	}
	defer windows.CloseHandle(mapHandle)

	addr, err := windows.MapViewOfFile(mapHandle, windows.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("MapViewOfFile: %w", err)
	}
	defer windows.UnmapViewOfFile(addr)

	// The mapping always contains at least a header; peek one to learn the
	// exact size of the region before we commit to a larger read. The
	// uintptr→unsafe.Pointer conversion here and below triggers vet's
	// unsafeptr check; it's a false positive because MapViewOfFile returns
	// an address to kernel-managed memory, not a Go heap object.
	headerView := unsafe.Slice((*byte)(unsafe.Pointer(addr)), headerLength)
	h := NewHeader(headerView)
	total := h.TotalSize()

	if total < headerLength {
		return nil, fmt.Errorf("HWiNFO header reports impossibly small size %d", total)
	}
	if total > maxSharedMemorySize {
		return nil, fmt.Errorf("HWiNFO header reports unreasonable size %d (cap %d); likely corrupt header", total, maxSharedMemorySize)
	}

	// Copy the full region into a Go-owned buffer so it remains valid after
	// we unmap and release the mutex.
	view := unsafe.Slice((*byte)(unsafe.Pointer(addr)), total)
	copied := make([]byte, total)
	copy(copied, view)
	return copied, nil
}

// acquireHWiNFOMutex opens the named HWiNFO mutex with SYNCHRONIZE access
// and blocks (up to mutexWaitMillis) for ownership.
func acquireHWiNFOMutex() (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(mutexName)
	if err != nil {
		return 0, fmt.Errorf("build mutex name: %w", err)
	}

	h, err := windows.OpenMutex(windows.SYNCHRONIZE, false, name)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return 0, ErrFileNotFound
		}
		if errors.Is(err, windows.ERROR_INVALID_HANDLE) {
			return 0, ErrInvalidHandle
		}
		return 0, fmt.Errorf("OpenMutex: %w", err)
	}

	event, waitErr := windows.WaitForSingleObject(h, mutexWaitMillis)
	switch event {
	case windows.WAIT_OBJECT_0:
		return h, nil
	case windows.WAIT_ABANDONED:
		// Previous owner crashed mid-write; we release and report. The
		// shared memory may be inconsistent so a retry next tick is safer
		// than trusting this snapshot.
		windows.ReleaseMutex(h)
		windows.CloseHandle(h)
		return 0, ErrMutexAbandoned
	case uint32(windows.WAIT_TIMEOUT):
		windows.CloseHandle(h)
		return 0, ErrMutexTimeout
	default:
		windows.CloseHandle(h)
		if waitErr != nil {
			return 0, fmt.Errorf("WaitForSingleObject: %w", waitErr)
		}
		return 0, fmt.Errorf("WaitForSingleObject returned unexpected event %d", event)
	}
}

func releaseHWiNFOMutex(h windows.Handle) {
	windows.ReleaseMutex(h)
	windows.CloseHandle(h)
}
