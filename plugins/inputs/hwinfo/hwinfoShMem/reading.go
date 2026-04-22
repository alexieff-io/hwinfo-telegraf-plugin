package hwinfoShMem

import (
	"encoding/binary"
	"math"
)

// ReadingType is the enum describing what kind of value a Reading holds.
type ReadingType int

const (
	ReadingTypeNone ReadingType = iota
	ReadingTypeTemp
	ReadingTypeVolt
	ReadingTypeFan
	ReadingTypeCurrent
	ReadingTypePower
	ReadingTypeClock
	ReadingTypeUsage
	ReadingTypeOther
)

func (t ReadingType) String() string {
	return [...]string{"none", "temp", "volt", "fan", "current", "power", "clock", "usage", "other"}[t]
}

// Reading is a parsed HWiNFO_SENSORS_READING_ELEMENT.
//
// Field layout (#pragma pack(1), little-endian):
//
//	offset   0  tReading        SENSOR_READING_TYPE (4-byte enum)
//	offset   4  dwSensorIndex   DWORD               (4)
//	offset   8  dwReadingID     DWORD               (4)
//	offset  12  szLabelOrig     char[128]
//	offset 140  szLabelUser     char[128]
//	offset 268  szUnit          char[16]
//	offset 284  Value           double              (8)
//	offset 292  ValueMin        double              (8)
//	offset 300  ValueMax        double              (8)
//	offset 308  ValueAvg        double              (8)
type Reading struct {
	data []byte
}

// Byte offsets of the fields, derived from the struct layout above.
const (
	readingOffsetType        = 0
	readingOffsetSensorIndex = 4
	readingOffsetReadingID   = 8
	readingOffsetLabelOrig   = 12
	readingOffsetLabelUser   = readingOffsetLabelOrig + stringLen
	readingOffsetUnit        = readingOffsetLabelUser + stringLen
	readingOffsetValue       = readingOffsetUnit + unitLen
	readingOffsetValueMin    = readingOffsetValue + 8
	readingOffsetValueMax    = readingOffsetValueMin + 8
	readingOffsetValueAvg    = readingOffsetValueMax + 8
)

// NewReading wraps the given byte slice as a Reading. The slice must be at
// least readingSize bytes.
func NewReading(data []byte) Reading {
	return Reading{data: data}
}

// ID returns the reading's unique ID within its parent sensor. Retained as
// int32 for backward compatibility; the underlying field is an unsigned DWORD.
func (r *Reading) ID() int32 {
	return int32(binary.LittleEndian.Uint32(r.data[readingOffsetReadingID : readingOffsetReadingID+4]))
}

// Type returns the reading's value category (temperature, voltage, etc).
func (r *Reading) Type() ReadingType {
	return ReadingType(binary.LittleEndian.Uint32(r.data[readingOffsetType : readingOffsetType+4]))
}

// SensorIndex returns the index of the parent sensor within the Sensors
// array.
func (r *Reading) SensorIndex() uint64 {
	return uint64(binary.LittleEndian.Uint32(r.data[readingOffsetSensorIndex : readingOffsetSensorIndex+4]))
}

// ReadingID is an alias for ID that preserves the wider return type.
func (r *Reading) ReadingID() uint64 {
	return uint64(binary.LittleEndian.Uint32(r.data[readingOffsetReadingID : readingOffsetReadingID+4]))
}

// LabelOrig returns the original reading label as assigned by HWiNFO.
func (r *Reading) LabelOrig() string {
	return decodeCString(r.data[readingOffsetLabelOrig : readingOffsetLabelOrig+stringLen])
}

// LabelUser returns the displayed reading label, which may have been
// renamed by the user.
func (r *Reading) LabelUser() string {
	return decodeCString(r.data[readingOffsetLabelUser : readingOffsetLabelUser+stringLen])
}

// Unit returns the reading's unit of measurement (e.g. "RPM", "°C").
func (r *Reading) Unit() string {
	return decodeCString(r.data[readingOffsetUnit : readingOffsetUnit+unitLen])
}

// Value returns the current reading value.
func (r *Reading) Value() float64 {
	return r.readFloat64(readingOffsetValue)
}

// ValueMin returns the minimum observed reading value since HWiNFO started.
func (r *Reading) ValueMin() float64 {
	return r.readFloat64(readingOffsetValueMin)
}

// ValueMax returns the maximum observed reading value since HWiNFO started.
func (r *Reading) ValueMax() float64 {
	return r.readFloat64(readingOffsetValueMax)
}

// ValueAvg returns the average reading value since HWiNFO started.
func (r *Reading) ValueAvg() float64 {
	return r.readFloat64(readingOffsetValueAvg)
}

func (r *Reading) readFloat64(offset int) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(r.data[offset : offset+8]))
}
