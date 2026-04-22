package hwinfo

import (
	_ "embed"
	"fmt"
	"os/exec"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	shmem "github.com/alexieff-io/hwinfo-telegraf-plugin/plugins/inputs/hwinfo/hwinfoShMem"
)

//go:embed sample.conf
var sampleConfig string

type HWiNFOInputPlugin struct {
	Log telegraf.Logger `toml:"-"`

	hwinfoVersion string
	pluginVersion string
	shmemVersion  string
}

func init() {
	inputs.Add("hwinfo", func() telegraf.Input {
		return &HWiNFOInputPlugin{
			pluginVersion: PluginVersion(),
		}
	})
}

func (input *HWiNFOInputPlugin) SampleConfig() string {
	return sampleConfig
}

func (input *HWiNFOInputPlugin) Init() error {
	input.hwinfoVersion = queryHWiNFOVersion(input.Log)
	return nil
}

func (input *HWiNFOInputPlugin) Gather(a telegraf.Accumulator) error {
	data, err := input.gather()
	if err != nil {
		return fmt.Errorf("gather from HWiNFO shared memory: %w", err)
	}

	writeCount := 0
	for _, datum := range data {
		for _, metric := range input.buildFieldsAndTags(datum) {
			a.AddFields("hwinfo", metric.fields, metric.tags)
			writeCount++
		}
	}
	if input.Log != nil {
		input.Log.Debugf("wrote %d metrics from %d sensors", writeCount, len(data))
	}
	return nil
}

// queryHWiNFOVersion queries the running HWiNFO64 process for its version
// string via PowerShell. Returns "unknown" if HWiNFO isn't running or the
// query fails for any reason.
func queryHWiNFOVersion(l telegraf.Logger) string {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		"(Get-Process HWiNFO64 | Select-Object Path | Get-Item).VersionInfo.ProductVersion")
	out, err := cmd.Output()
	if err != nil {
		if l != nil {
			l.Debugf("failed to query HWiNFO64 version: %v", err)
		}
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// PluginVersion returns the module version embedded in the binary's build
// info, or "unknown" if it can't be resolved.
func PluginVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	i := slices.IndexFunc(bi.Deps, func(m *debug.Module) bool {
		return m.Path == "github.com/alexieff-io/hwinfo-telegraf-plugin"
	})
	if i == -1 {
		return "unknown"
	}
	return bi.Deps[i].Version
}

type Metric struct {
	fields map[string]interface{}
	tags   map[string]string
}

type SensorReadings struct {
	sensor   shmem.Sensor
	readings []shmem.Reading
}

func (input *HWiNFOInputPlugin) gather() ([]SensorReadings, error) {
	rawData, err := shmem.Read()
	if err != nil {
		return nil, err
	}
	input.shmemVersion = rawData.Version()

	sensors := rawData.Sensors()
	data := make([]SensorReadings, 0, len(sensors))
	for _, s := range sensors {
		data = append(data, SensorReadings{sensor: s})
	}

	for _, r := range rawData.Readings() {
		sensorIndex := int(r.SensorIndex())
		if sensorIndex >= len(data) {
			if input.Log != nil {
				input.Log.Errorf("sensor index out of range: reading references sensor %d, but only %d sensors exist", sensorIndex, len(data))
			}
			continue
		}
		data[sensorIndex].readings = append(data[sensorIndex].readings, r)
	}

	return data, nil
}

func (input *HWiNFOInputPlugin) buildFieldsAndTags(sr SensorReadings) []Metric {
	sensor := sr.sensor
	metrics := make([]Metric, 0, len(sr.readings))

	for _, reading := range sr.readings {
		fields := map[string]interface{}{
			reading.Type().String(): reading.Value(),
		}
		tags := map[string]string{
			"hwinfoVersion": input.hwinfoVersion,
			"pluginVersion": input.pluginVersion,
			"shmemVersion":  input.shmemVersion,

			"sensorId":       sensor.ID(),
			"sensorInst":     strconv.FormatUint(sensor.SensorInst(), 10),
			"sensorType":     string(sensor.SensorType()),
			"sensorNameOrig": sensor.NameOrig(),
			"sensorName":     sensor.NameUser(),

			"readingId":       strconv.FormatInt(int64(reading.ID()), 10),
			"readingNameOrig": reading.LabelOrig(),
			"readingName":     reading.LabelUser(),
			"unit":            reading.Unit(),
		}
		metrics = append(metrics, Metric{fields: fields, tags: tags})
	}
	return metrics
}
