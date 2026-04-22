package hwinfoShMem

import (
	"errors"
	"fmt"
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

// HWiNFOShMem is a parsed snapshot of the HWiNFO shared memory region.
type HWiNFOShMem struct {
	header   *Header
	sensors  []Sensor
	readings []Reading
}

// Read takes the HWiNFO mutex, copies the full shared memory region into a
// Go-owned buffer, releases the mutex, and parses sensors and readings. The
// mutex is held only for the duration of the memory copy, per the HWiNFO
// SDK's "release as quick as possible" contract.
func Read() (*HWiNFOShMem, error) {
	body, err := snapshotSharedMemory()
	if err != nil {
		return nil, err
	}

	if len(body) < headerLength {
		return nil, fmt.Errorf("HWiNFO shared memory too small: got %d bytes, need at least %d", len(body), headerLength)
	}
	h := NewHeader(body[:headerLength])
	header := &h

	if sig := header.Signature(); sig != signatureActive {
		return nil, InactiveError{Signature: sig}
	}

	if want := header.TotalSize(); len(body) < want {
		return nil, fmt.Errorf("HWiNFO shared memory truncated: got %d bytes, header claims %d", len(body), want)
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
