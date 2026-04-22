package hwinfoShMem

// HWiNFO shared memory layout constants. Values come from hwisenssm2.h;
// keeping them as plain Go constants lets the package build without cgo.
const (
	// sharedMemoryName is the named file-mapping object HWiNFO publishes
	// sensor data into.
	sharedMemoryName = "Global\\HWiNFO_SENS_SM2"

	// mutexName is the named mutex HWiNFO expects readers to acquire while
	// reading the shared memory region.
	mutexName = "Global\\HWiNFO_SM2_MUTEX"

	// stringLen is the fixed length of name/label char arrays in the
	// HWiNFO structures (HWiNFO_SENSORS_STRING_LEN2 in the original header).
	stringLen = 128

	// unitLen is the fixed length of unit char arrays (HWiNFO_UNIT_STRING_LEN).
	unitLen = 16

	// headerLength is the size of HWiNFO_SENSORS_SHARED_MEM2 under
	// #pragma pack(1): dwSignature(4) + dwVersion(4) + dwRevision(4) +
	// poll_time(8, __time64_t) + 6*DWORD(24) = 44.
	headerLength = 44

	// sensorSize is the fixed size of HWiNFO_SENSORS_SENSOR_ELEMENT:
	// dwSensorID(4) + dwSensorInst(4) + szSensorNameOrig(128) +
	// szSensorNameUser(128) = 264. Note: actual size is read from the
	// header at runtime; this is a reference for tests.
	sensorSize = 264

	// readingSize is the fixed size of HWiNFO_SENSORS_READING_ELEMENT:
	// tReading(4) + dwSensorIndex(4) + dwReadingID(4) + szLabelOrig(128) +
	// szLabelUser(128) + szUnit(16) + 4*double(32) = 316. Note: actual size
	// is read from the header at runtime; this is a reference for tests.
	readingSize = 316
)
