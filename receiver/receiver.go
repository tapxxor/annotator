package main

import (
	receiver "annotator/receiver/lib"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	// set logging proeperties
	// bit 0: Ldate, the date in the local time zone: 2009/01/23
	// bit 1: Ltime, the time in the local time zone: 01:23:23
	// bit 2: Lmicroseconds, the microsecond resolution: 01:23:23.123123.  assumes Ltime.
	// bit 4: Lshortfile, the final file name element and line number: d.go:23. overrides Llong
	// bit 5: LUTC, if Ldate or Ltime is set, use UTC rather than the local time zone
	log.SetFlags(log.LUTC | log.Lshortfile | log.Lmicroseconds | log.Ltime | log.Ldate)

	// read flags
	receiver.ConfigFile = flag.String("config", "/etc/annotator/annotator.yml", "annotator config file")
	flag.Parse()

	// read configuration
	log.Printf("Loading configuration from %s\n", *(receiver.ConfigFile))
	if err := receiver.Config.Fread(receiver.ConfigFile); err != nil {
		log.Fatalf("Error in reading configuration file %s: %v", *(receiver.ConfigFile), err)
	}

	// check configuration
	if err := receiver.Config.Validate(); err != nil {
		log.Fatalf("Configuration checks failed: %v", err)
	}

	// create data folder
	receiver.AlertsPath = receiver.Config.Receiver.Settings.AlertsPath

	receiver.AlertsPaths =
		map[string]string{
			"firing":   filepath.Join(receiver.AlertsPath, "firing"),
			"resolved": filepath.Join(receiver.AlertsPath, "resolved"),
		}

	// create alerts path
	if _, err := os.Stat(receiver.AlertsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(receiver.AlertsPath, 0775); err != nil {
			log.Fatalf("Configuration error: %s",
				fmt.Errorf("could not create folder \"%s\"", receiver.AlertsPath))
		}
	}

	// create alerts firing path
	if _, err := os.Stat(receiver.AlertsPaths["firing"]); os.IsNotExist(err) {
		if err := os.MkdirAll(receiver.AlertsPaths["firing"], 0775); err != nil {
			log.Fatalf("Configuration error: %s",
				fmt.Errorf("could not create folder \"%s\"", receiver.AlertsPaths["firing"]))
		}
	}

	// create alerts resolved path
	if _, err := os.Stat(receiver.AlertsPaths["resolved"]); os.IsNotExist(err) {
		if err := os.MkdirAll(receiver.AlertsPaths["resolved"], 0775); err != nil {
			log.Fatalf("Configuration error: %s",
				fmt.Errorf("could not create folder \"%s\"", receiver.AlertsPaths["resolved"]))
		}
	}
}

func main() {

	// start serving requests
	log.Println("Receiver is ready to get alerts from alertmanager")

	http.HandleFunc("/", http.HandlerFunc(receiver.GetAlert))

	// expose promtheus metrics
	if receiver.Config.Receiver.Settings.Metrics {
		http.Handle(receiver.Config.Receiver.Settings.MetricsPath, promhttp.Handler())
	}

	// start receiver
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(receiver.Config.Receiver.Settings.Port)), nil))
}
