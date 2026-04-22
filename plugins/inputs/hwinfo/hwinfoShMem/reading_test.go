package hwinfoShMem

import (
	"encoding/binary"
	"math"
	"testing"
)

// buildReading constructs a byte slice matching HWiNFO_SENSORS_READING_ELEMENT
// with #pragma pack(1) layout:
//
//	SENSOR_READING_TYPE tReading;   // 4 (enum on Windows)
//	DWORD dwSensorIndex;            // 4
//	DWORD dwReadingID;              // 4
//	char  szLabelOrig[128];
//	char  szLabelUser[128];
//	char  szUnit[16];
//	double Value, ValueMin, ValueMax, ValueAvg; // 4 * 8
//
// Total: 316 bytes.
func buildReading(tReading, sensorIndex, readingID uint32,
	labelOrig, labelUser, unit string,
	value, valueMin, valueMax, valueAvg float64) []byte {

	const (
		labelLen = 128
		unitLen  = 16
	)
	size := 4 + 4 + 4 + labelLen + labelLen + unitLen + 4*8
	buf := make([]byte, size)

	binary.LittleEndian.PutUint32(buf[0:4], tReading)
	binary.LittleEndian.PutUint32(buf[4:8], sensorIndex)
	binary.LittleEndian.PutUint32(buf[8:12], readingID)
	copy(buf[12:12+labelLen], labelOrig)
	copy(buf[12+labelLen:12+2*labelLen], labelUser)
	copy(buf[12+2*labelLen:12+2*labelLen+unitLen], unit)

	valOffset := 12 + 2*labelLen + unitLen // 284
	binary.LittleEndian.PutUint64(buf[valOffset:], math.Float64bits(value))
	binary.LittleEndian.PutUint64(buf[valOffset+8:], math.Float64bits(valueMin))
	binary.LittleEndian.PutUint64(buf[valOffset+16:], math.Float64bits(valueMax))
	binary.LittleEndian.PutUint64(buf[valOffset+24:], math.Float64bits(valueAvg))
	return buf
}

// TestReading_ValueChain is the regression test for the pointer-arithmetic
// chain valuePtr -> valueMinPtr -> valueMaxPtr -> valueAvgPtr. Any off-by-one
// in sizeof arithmetic will scramble these four values.
func TestReading_ValueChain(t *testing.T) {
	const (
		value    = 1.111
		valueMin = 2.222
		valueMax = 3.333
		valueAvg = 4.444
	)
	data := buildReading(uint32(ReadingTypeTemp), 0, 42,
		"Core 0", "Core 0", "\xB0C",
		value, valueMin, valueMax, valueAvg)
	r := NewReading(data)

	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"Value", r.Value(), value},
		{"ValueMin", r.ValueMin(), valueMin},
		{"ValueMax", r.ValueMax(), valueMax},
		{"ValueAvg", r.ValueAvg(), valueAvg},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestReading_MetaFields(t *testing.T) {
	data := buildReading(uint32(ReadingTypeFan), 3, 42,
		"Chassis Fan", "My Fan", "RPM", 0, 0, 0, 0)
	r := NewReading(data)

	if got, want := r.Type(), ReadingTypeFan; got != want {
		t.Errorf("Type() = %v, want %v", got, want)
	}
	if got, want := r.SensorIndex(), uint64(3); got != want {
		t.Errorf("SensorIndex() = %v, want %v", got, want)
	}
	if got, want := r.ID(), int32(42); got != want {
		t.Errorf("ID() = %v, want %v", got, want)
	}
	if got, want := r.LabelOrig(), "Chassis Fan"; got != want {
		t.Errorf("LabelOrig() = %q, want %q", got, want)
	}
	if got, want := r.LabelUser(), "My Fan"; got != want {
		t.Errorf("LabelUser() = %q, want %q", got, want)
	}
	if got, want := r.Unit(), "RPM"; got != want {
		t.Errorf("Unit() = %q, want %q", got, want)
	}
}

func TestReadingType_String(t *testing.T) {
	tests := []struct {
		rt   ReadingType
		want string
	}{
		{ReadingTypeNone, "none"},
		{ReadingTypeTemp, "temp"},
		{ReadingTypeVolt, "volt"},
		{ReadingTypeFan, "fan"},
		{ReadingTypeCurrent, "current"},
		{ReadingTypePower, "power"},
		{ReadingTypeClock, "clock"},
		{ReadingTypeUsage, "usage"},
		{ReadingTypeOther, "other"},
	}
	for _, tc := range tests {
		if got := tc.rt.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.rt, got, tc.want)
		}
	}
}
