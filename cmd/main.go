package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/telegraf/plugins/common/shim"

	_ "github.com/alexieff-io/hwinfo-telegraf-plugin/plugins/inputs/hwinfo"
)

var (
	pollInterval         = flag.Duration("poll_interval", 1*time.Second, "how often to gather metrics")
	pollIntervalDisabled = flag.Bool("poll_interval_disabled", false, "disable the poll interval; gather metrics only when the agent requests them")
	configFile           = flag.String("config", "", "path to the config file for this plugin")
)

func main() {
	flag.Parse()

	interval := *pollInterval
	if *pollIntervalDisabled {
		interval = shim.PollIntervalDisabled
	}

	s := shim.New()

	if *configFile != "" {
		if err := s.LoadConfig(configFile); err != nil {
			fatalf("error loading config: %v", err)
		}
	}

	if err := s.Run(interval); err != nil {
		fatalf("hwinfo input plugin failed: %v", err)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}
