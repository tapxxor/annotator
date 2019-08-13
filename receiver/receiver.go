package main

import (
	"annotator/lib"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Configuration is a map tha holds application configuration
type Configuration map[string]string

// config variable for application configuration
var (
	config      lib.ReceiverConf
	configFile  *string
	alertsPath  string
	alertsPaths map[string]string
)

// WriteToFile writes a string to a given path
func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}

	// Issue a Sync to flush writes to stable storage.
	return file.Sync()
}

// storeAlert saves alert json representation to file with name md5256sum(startsAt,groupKey)
func alertToFile(m *lib.Message) {
	alertFilename := sha256.Sum256([]byte(m.Alerts[0].StartsAt.String() + m.GroupKey))
	absPath := filepath.Join(
		filepath.Join(alertsPaths[m.Alerts[0].Status]),
		fmt.Sprintf("%x", alertFilename))

	log.Printf("Saving to %s\n", absPath)

	if err := WriteToFile(absPath, m.String()); err != nil {
		log.Panicf("Error writing to %x: %v", alertFilename, err)
	}
}

// getAlert received the json payload from alertmanager and stores it locally
func getAlert(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s %s\n", r.Method, r.URL, r.Proto, r.RemoteAddr)
	switch r.Method {
	case "POST":
		// Decode the JSON in the body
		d := json.NewDecoder(r.Body)
		defer r.Body.Close()

		alert := &lib.Message{}
		err := d.Decode(alert)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		log.Printf("\n%s\n", alert)
		alertToFile(alert)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", "OK")
	default:
		http.Error(w, "Only POST method is allowed", http.StatusInternalServerError)
	}
}

func init() {
	// set logging proeperties
	// bit 0: Ldate, the date in the local time zone: 2009/01/23
	// bit 1: Ltime, the time in the local time zone: 01:23:23
	// bit 2: Lmicroseconds, the microsecond resolution: 01:23:23.123123.  assumes Ltime.
	// bit 4: Lshortfile, the final file name element and line number: d.go:23. overrides Llong
	// bit 5: LUTC, if Ldate or Ltime is set, use UTC rather than the local time zone
	log.SetFlags(log.LUTC | log.Lshortfile | log.Lmicroseconds | log.Ltime | log.Ldate)

	// read flags
	configFile = flag.String("config", "/etc/annotator/annotator.yml", "annotator config file")
	flag.Parse()

	// read configuration
	log.Printf("Loading configuration from %s\n", *configFile)
	if err := config.Fread(configFile); err != nil {
		log.Fatalf("Error in reading configuration file %s: %v", *configFile, err)
	}

	// check configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration checks failed: %v", err)
	}

	// create data folder
	alertsPath = config.Receiver.Settings.AlertsPath
	alertsPaths =
		map[string]string{
			"firing":   filepath.Join(alertsPath, "firing"),
			"resolved": filepath.Join(alertsPath, "resolved"),
		}

	// create alerts path
	if _, err := os.Stat(alertsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(alertsPath, 0775); err != nil {
			log.Fatalf("Configuration error: %s",
				fmt.Errorf("could not create folder \"%s\"", alertsPath))
		}
	}

	// create alerts firing path
	if _, err := os.Stat(alertsPaths["firing"]); os.IsNotExist(err) {
		if err := os.MkdirAll(alertsPaths["firing"], 0775); err != nil {
			log.Fatalf("Configuration error: %s",
				fmt.Errorf("could not create folder \"%s\"", alertsPaths["firing"]))
		}
	}

	// create alerts resolved path
	if _, err := os.Stat(alertsPaths["resolved"]); os.IsNotExist(err) {
		if err := os.MkdirAll(alertsPaths["resolved"], 0775); err != nil {
			log.Fatalf("Configuration error: %s",
				fmt.Errorf("could not create folder \"%s\"", alertsPaths["resolved"]))
		}
	}
}

func main() {

	// start serving requests
	log.Println("Receiver is ready to get alerts from alertmanager")
	http.HandleFunc("/", getAlert)

	// expose promtheus metrics
	if config.Receiver.Settings.Metrics {
		http.Handle(config.Receiver.Settings.MetricsPath, promhttp.Handler())
	}

	// start receiver
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(config.Receiver.Settings.Port)), nil))
}
