package hwinfoShMem

import "encoding/binary"

// Header is a parsed HWiNFO_SENSORS_SHARED_MEM2 structure. All accessors
// read directly from the backing byte slice; no copying is performed.
//
// Field layout (#pragma pack(1), little-endian):
//
//	offset  0  dwSignature              DWORD      (4 bytes, ASCII "HWiS"/"DEAD")
//	offset  4  dwVersion                DWORD      (4 bytes)
//	offset  8  dwRevision               DWORD      (4 bytes)
//	offset 12  poll_time                __time64_t (8 bytes)
//	offset 20  dwOffsetOfSensorSection  DWORD      (4 bytes)
//	offset 24  dwSizeOfSensorElement    DWORD      (4 bytes)
//	offset 28  dwNumSensorElements      DWORD      (4 bytes)
//	offset 32  dwOffsetOfReadingSection DWORD      (4 bytes)
//	offset 36  dwSizeOfReadingElement   DWORD      (4 bytes)
//	offset 40  dwNumReadingElements     DWORD      (4 bytes)
type Header struct {
	data []byte
}

// NewHeader wraps the given byte slice as a Header. The slice must be at
// least headerLength bytes; accessor calls will panic otherwise.
func NewHeader(data []byte) Header {
	return Header{data: data}
}

// Signature returns the 4-byte ASCII signature: "HWiS" when HWiNFO is
// actively publishing or "DEAD" when it has shut down.
func (h *Header) Signature() string {
	return decodeCString(h.data[0:4])
}

// Version returns the shared memory protocol version.
func (h *Header) Version() int {
	return int(binary.LittleEndian.Uint32(h.data[4:8]))
}

// Revision returns the shared memory protocol revision.
func (h *Header) Revision() int {
	return int(binary.LittleEndian.Uint32(h.data[8:12]))
}

// PollTime returns HWiNFO's last polling time as a raw __time64_t value
// (seconds since Unix epoch).
func (h *Header) PollTime() uint64 {
	return binary.LittleEndian.Uint64(h.data[12:20])
}

// OffsetOfSensorSection returns the byte offset of the sensor array from the
// start of the shared memory region.
func (h *Header) OffsetOfSensorSection() int {
	return int(binary.LittleEndian.Uint32(h.data[20:24]))
}

// SizeOfSensorElement returns the size of a single sensor element in bytes.
func (h *Header) SizeOfSensorElement() int {
	return int(binary.LittleEndian.Uint32(h.data[24:28]))
}

// NumSensorElements returns the number of sensors in the sensor array.
func (h *Header) NumSensorElements() int {
	return int(binary.LittleEndian.Uint32(h.data[28:32]))
}

// OffsetOfReadingSection returns the byte offset of the reading array.
func (h *Header) OffsetOfReadingSection() int {
	return int(binary.LittleEndian.Uint32(h.data[32:36]))
}

// SizeOfReadingElement returns the size of a single reading element in bytes.
func (h *Header) SizeOfReadingElement() int {
	return int(binary.LittleEndian.Uint32(h.data[36:40]))
}

// NumReadingElements returns the number of readings in the reading array.
func (h *Header) NumReadingElements() int {
	return int(binary.LittleEndian.Uint32(h.data[40:44]))
}

// TotalSize is the total byte length of the HWiNFO shared memory region:
// header + sensor section + reading section.
func (h *Header) TotalSize() int {
	return h.OffsetOfReadingSection() + h.NumReadingElements()*h.SizeOfReadingElement()
}
