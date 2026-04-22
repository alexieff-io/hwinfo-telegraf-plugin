package hwinfoShMem

// #include "hwisenssm2.h"
import "C"

import (
	"errors"
	"fmt"

	"github.com/hidez8891/shm"
)

// ErrAccessDenied is returned when Windows denies access to the HWiNFO
// shared memory region. Typically means the plugin isn't running as
// Administrator.
var ErrAccessDenied = errors.New("access denied opening HWiNFO shared memory; is the plugin running as Administrator?")

type HWiNFOShMem struct {
	header   *Header
	sensors  []Sensor
	readings []Reading
}

func Read() (*HWiNFOShMem, error) {
	if err := LockMutex(); err != nil {
		return nil, err
	}
	defer UnlockMutex()

	header, err := ReadHeader()
	if err != nil {
		return nil, err
	}

	sensors, err := ReadSensors(header)
	if err != nil {
		return nil, err
	}

	readings, err := ReadReadings(header)
	if err != nil {
		return nil, err
	}

	return &HWiNFOShMem{header: header, sensors: sensors, readings: readings}, nil
}

func ReadHeader() (*Header, error) {
	bytes, err := readSharedMemory(0, headerLength)
	if err != nil {
		return nil, err
	}
	header := NewHeader(bytes)
	return &header, nil
}

func ReadSensors(header *Header) ([]Sensor, error) {
	offset := header.OffsetOfSensorSection()
	num := header.NumSensorElements()
	size := header.SizeOfSensorElement()

	bytes, err := readSharedMemory(offset, num*size)
	if err != nil {
		return nil, err
	}

	sensors := make([]Sensor, 0, num)
	for i := 0; i < num; i++ {
		start := i * size
		sensors = append(sensors, NewSensor(bytes[start:start+size]))
	}
	return sensors, nil
}

func ReadReadings(header *Header) ([]Reading, error) {
	offset := header.OffsetOfReadingSection()
	num := header.NumReadingElements()
	size := header.SizeOfReadingElement()

	bytes, err := readSharedMemory(offset, num*size)
	if err != nil {
		return nil, err
	}

	readings := make([]Reading, 0, num)
	for i := 0; i < num; i++ {
		start := i * size
		readings = append(readings, NewReading(bytes[start:start+size]))
	}
	return readings, nil
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
