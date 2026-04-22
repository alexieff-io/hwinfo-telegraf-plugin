package hwinfoShMem

import (
	"encoding/binary"
	"testing"
)

// buildSensor constructs a byte slice matching HWiNFO_SENSORS_SENSOR_ELEMENT
// with #pragma pack(1) layout (DWORD=4, char[128]=128).
func buildSensor(sensorID, sensorInst uint32, nameOrig, nameUser string) []byte {
	const nameLen = 128
	buf := make([]byte, 4+4+nameLen+nameLen)
	binary.LittleEndian.PutUint32(buf[0:4], sensorID)
	binary.LittleEndian.PutUint32(buf[4:8], sensorInst)
	copy(buf[8:8+nameLen], nameOrig)
	copy(buf[8+nameLen:], nameUser)
	return buf
}

func TestSensor_ID_Format(t *testing.T) {
	tests := []struct {
		name       string
		sensorID   uint32
		sensorInst uint32
		want       string
	}{
		{"simple", 1, 0, "1-0"},
		{"high_instance", 1, 150, "1-150"},
		// Previous format ID*100+Inst would have made (1,150) and (2,50)
		// collide at "250". Confirm the new format keeps them distinct.
		{"no_collision_neighbor", 2, 50, "2-50"},
		{"large_id", 4026532608, 0, "4026532608-0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSensor(buildSensor(tc.sensorID, tc.sensorInst, "", ""))
			if got := s.ID(); got != tc.want {
				t.Errorf("ID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSensor_IDs(t *testing.T) {
	s := NewSensor(buildSensor(42, 7, "", ""))
	if got, want := s.SensorID(), uint64(42); got != want {
		t.Errorf("SensorID() = %d, want %d", got, want)
	}
	if got, want := s.SensorInst(), uint64(7); got != want {
		t.Errorf("SensorInst() = %d, want %d", got, want)
	}
}

func TestSensor_Names(t *testing.T) {
	s := NewSensor(buildSensor(1, 0, "CPU [#0]: AMD Ryzen 5 5600X", "My CPU"))
	if got, want := s.NameOrig(), "CPU [#0]: AMD Ryzen 5 5600X"; got != want {
		t.Errorf("NameOrig() = %q, want %q", got, want)
	}
	if got, want := s.NameUser(), "My CPU"; got != want {
		t.Errorf("NameUser() = %q, want %q", got, want)
	}
}

func TestSensor_SensorType(t *testing.T) {
	tests := []struct {
		name     string
		nameOrig string
		want     SensorType
	}{
		{"system", "System: MSI MS-7C92", System},
		{"cpu", "CPU [#0]: AMD Ryzen 5 5600X", CPU},
		{"smart", "S.M.A.R.T.: Samsung SSD 970 EVO", SMART},
		{"drive", "Drive [0]: Samsung SSD", Drive},
		{"gpu", "GPU [#0]: NVIDIA GeForce RTX 3080", GPU},
		{"network", "Network: Intel Ethernet", Network},
		{"windows", "Windows Performance Counters", Windows},
		{"memory_timings", "Memory Timings", MemoryTimings},
		{"case_insensitive_cpu", "cpu [#0]", CPU},
		{"unknown_prefix", "Motherboard [#0]", Unknown},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSensor(buildSensor(1, 0, tc.nameOrig, ""))
			if got := s.SensorType(); got != tc.want {
				t.Errorf("SensorType() for %q = %q, want %q", tc.nameOrig, got, tc.want)
			}
		})
	}
}
