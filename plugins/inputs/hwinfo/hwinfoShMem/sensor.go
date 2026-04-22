package hwinfoShMem

import (
	"encoding/binary"
	"fmt"
)

// Sensor is a parsed HWiNFO_SENSORS_SENSOR_ELEMENT.
//
// Field layout (#pragma pack(1), little-endian):
//
//	offset   0  dwSensorID         DWORD    (4 bytes)
//	offset   4  dwSensorInst       DWORD    (4 bytes)
//	offset   8  szSensorNameOrig   char[128]
//	offset 136  szSensorNameUser   char[128]
type Sensor struct {
	data []byte
}

// SensorType identifies the kind of hardware a sensor represents.
type SensorType string

const (
	System        SensorType = "system"
	CPU           SensorType = "cpu"
	SMART         SensorType = "smart"
	Drive         SensorType = "drive"
	GPU           SensorType = "gpu"
	Network       SensorType = "network"
	Windows       SensorType = "windows"
	MemoryTimings SensorType = "memory-timings"
	Unknown       SensorType = "unknown"
)

// NewSensor wraps the given byte slice as a Sensor. The slice must be at
// least sensorSize bytes.
func NewSensor(data []byte) Sensor {
	return Sensor{data: data}
}

// SensorID returns the sensor's unique ID within HWiNFO.
func (s *Sensor) SensorID() uint64 {
	return uint64(binary.LittleEndian.Uint32(s.data[0:4]))
}

// SensorInst returns the sensor's instance number. Combined with SensorID
// it uniquely identifies a sensor.
func (s *Sensor) SensorInst() uint64 {
	return uint64(binary.LittleEndian.Uint32(s.data[4:8]))
}

// ID returns a unique identifier for this sensor combining SensorID and
// SensorInst, formatted as "sensorID-sensorInst". The previous format
// (SensorID*100 + SensorInst) collided whenever any sensor had 100 or more
// instances; this one does not.
func (s *Sensor) ID() string {
	return fmt.Sprintf("%d-%d", s.SensorID(), s.SensorInst())
}

// NameOrig returns the original sensor name as assigned by HWiNFO.
func (s *Sensor) NameOrig() string {
	return decodeCString(s.data[8 : 8+stringLen])
}

// NameUser returns the displayed sensor name, which may have been renamed
// by the user in HWiNFO's settings.
func (s *Sensor) NameUser() string {
	return decodeCString(s.data[8+stringLen : 8+2*stringLen])
}

// SensorType classifies the sensor by matching its original name prefix.
// HWiNFO does not expose sensor type directly, so this uses best-effort
// string matching.
func (s *Sensor) SensorType() SensorType {
	name := s.NameOrig()
	switch {
	case StartsWithLower(name, "system"):
		return System
	case StartsWithLower(name, "cpu"):
		return CPU
	case StartsWithLower(name, "s.m.a.r.t."):
		return SMART
	case StartsWithLower(name, "drive"):
		return Drive
	case StartsWithLower(name, "gpu"):
		return GPU
	case StartsWithLower(name, "network"):
		return Network
	case StartsWithLower(name, "windows"):
		return Windows
	case StartsWithLower(name, "memory timings"):
		return MemoryTimings
	default:
		return Unknown
	}
}
