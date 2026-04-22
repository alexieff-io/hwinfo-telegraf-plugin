package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"errors"
	"fmt"

	"github.com/hidez8891/shm"
)

// signatureActive marks an active HWiNFO shared memory region; anything else
// (typically "DEAD") means the region is stale or HWiNFO has shut down.
const signatureActive = "HWiS"

// ErrAccessDenied is returned when Windows denies access to the HWiNFO
// shared memory region. Typically means the plugin isn't running as
// Administrator.
var ErrAccessDenied = errors.New("access denied opening HWiNFO shared memory; is the plugin running as Administrator?")

// InactiveError is returned when the shared memory region exists but is
// marked inactive (signature != "HWiS"). The returned Signature is the raw
// value found in the region.
type InactiveError struct {
	Signature string
}

func (e InactiveError) Error() string {
	return fmt.Sprintf("HWiNFO shared memory is inactive: signature %q, want %q", e.Signature, signatureActive)
}

type HWiNFOShMem struct {
	header   *Header
	sensors  []Sensor
	readings []Reading
}

// Read takes the HWiNFO mutex, reads the full shared memory region, parses
// sensors and readings, then releases the mutex. The mutex is held for the
// minimum possible time per the HWiNFO SDK's "release as quick as possible"
// contract.
//
// The region is opened twice: once with a header-sized bound to learn the
// total size, then once with that total size to read sensors+readings in a
// single shot. A truly-single open would require bypassing the shm library's
// software-side bounds check on ReadAt.
func Read() (*HWiNFOShMem, error) {
	if err := LockMutex(); err != nil {
		return nil, err
	}
	defer UnlockMutex()

	headerBytes, err := readSharedMemory(0, headerLength)
	if err != nil {
		return nil, fmt.Errorf("read HWiNFO header: %w", err)
	}
	h := NewHeader(headerBytes)
	header := &h

	if sig := header.Signature(); sig != signatureActive {
		return nil, InactiveError{Signature: sig}
	}

	body, err := readSharedMemory(0, header.TotalSize())
	if err != nil {
		return nil, fmt.Errorf("read HWiNFO body: %w", err)
	}

	return &HWiNFOShMem{
		header:   header,
		sensors:  parseSensors(header, body),
		readings: parseReadings(header, body),
	}, nil
}

func parseSensors(header *Header, body []byte) []Sensor {
	offset := header.OffsetOfSensorSection()
	num := header.NumSensorElements()
	size := header.SizeOfSensorElement()

	sensors := make([]Sensor, 0, num)
	for i := 0; i < num; i++ {
		start := offset + i*size
		sensors = append(sensors, NewSensor(body[start:start+size]))
	}
	return sensors
}

func parseReadings(header *Header, body []byte) []Reading {
	offset := header.OffsetOfReadingSection()
	num := header.NumReadingElements()
	size := header.SizeOfReadingElement()

	readings := make([]Reading, 0, num)
	for i := 0; i < num; i++ {
		start := offset + i*size
		readings = append(readings, NewReading(body[start:start+size]))
	}
	return readings
}

func isAccessDeniedErr(err error) bool {
	return fmt.Sprintf("%v", err) == "CreateFileMapping: Access is denied."
}

func readSharedMemory(start, size int) ([]byte, error) {
	memory, err := shm.Open(C.HWiNFO_SENSORS_MAP_FILE_NAME2, int32(size))
	if err != nil {
		if isAccessDeniedErr(err) {
			return nil, ErrAccessDenied
		}
		return nil, err
	}
	defer memory.Close()

	bytes := make([]byte, size)
	if _, err := memory.ReadAt(bytes, int64(start)); err != nil {
		return nil, fmt.Errorf("read shared memory at offset %d: %w", start, err)
	}
	return bytes, nil
}

func (s *HWiNFOShMem) Version() string {
	return fmt.Sprintf("v%d rev%d", s.header.Version(), s.header.Revision())
}

func (s *HWiNFOShMem) Header() *Header {
	return s.header
}

func (s *HWiNFOShMem) Sensors() []Sensor {
	return s.sensors
}

func (s *HWiNFOShMem) Readings() []Reading {
	return s.readings
}
