package hwinfoShMem

import (
	"encoding/binary"
	"testing"
)

// buildHeader constructs a byte slice matching HWiNFO_SENSORS_SHARED_MEM2 with
// #pragma pack(1) layout. Field sizes assume DWORD=4 and __time64_t=8.
func buildHeader(signature string, version, revision uint32, pollTime uint64,
	offsetSensor, sizeSensor, numSensor uint32,
	offsetReading, sizeReading, numReading uint32) []byte {

	const size = 4 + 4 + 4 + 8 + 4 + 4 + 4 + 4 + 4 + 4 // 44
	buf := make([]byte, size)
	copy(buf[0:4], signature)
	binary.LittleEndian.PutUint32(buf[4:8], version)
	binary.LittleEndian.PutUint32(buf[8:12], revision)
	binary.LittleEndian.PutUint64(buf[12:20], pollTime)
	binary.LittleEndian.PutUint32(buf[20:24], offsetSensor)
	binary.LittleEndian.PutUint32(buf[24:28], sizeSensor)
	binary.LittleEndian.PutUint32(buf[28:32], numSensor)
	binary.LittleEndian.PutUint32(buf[32:36], offsetReading)
	binary.LittleEndian.PutUint32(buf[36:40], sizeReading)
	binary.LittleEndian.PutUint32(buf[40:44], numReading)
	return buf
}

func TestHeader_Accessors(t *testing.T) {
	data := buildHeader("HWiS", 2, 5, 1700000000,
		44, 264, 3,
		836, 316, 100)
	h := NewHeader(data)

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"Signature", h.Signature(), "HWiS"},
		{"Version", h.Version(), 2},
		{"Revision", h.Revision(), 5},
		{"PollTime", h.PollTime(), uint64(1700000000)},
		{"OffsetOfSensorSection", h.OffsetOfSensorSection(), 44},
		{"SizeOfSensorElement", h.SizeOfSensorElement(), 264},
		{"NumSensorElements", h.NumSensorElements(), 3},
		{"OffsetOfReadingSection", h.OffsetOfReadingSection(), 836},
		{"SizeOfReadingElement", h.SizeOfReadingElement(), 316},
		{"NumReadingElements", h.NumReadingElements(), 100},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestHeader_TotalSize(t *testing.T) {
	// TotalSize = OffsetOfReadingSection + NumReadingElements * SizeOfReadingElement
	// = 836 + 100 * 316 = 32436
	data := buildHeader("HWiS", 2, 5, 0, 44, 264, 3, 836, 316, 100)
	h := NewHeader(data)
	if got, want := h.TotalSize(), 32436; got != want {
		t.Errorf("TotalSize() = %d, want %d", got, want)
	}
}

func TestHeader_InactiveSignature(t *testing.T) {
	data := buildHeader("DEAD", 2, 5, 0, 44, 264, 0, 44, 316, 0)
	h := NewHeader(data)
	if got, want := h.Signature(), "DEAD"; got != want {
		t.Errorf("Signature() = %q, want %q", got, want)
	}
}
