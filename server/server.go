package main

import (
	"annotator/db"
	"flag"
	"fmt"
	"log"
	"net/url"
	"path"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"annotator/server/lib"
	"annotator/server/routines"
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
	lib.ConfigFile = flag.String("config", "/etc/annotator/annotator.yml", "annotator config file")
	flag.Parse()

	// read configuration
	log.Printf("Loading configuration from %s\n", *(lib.ConfigFile))
	if err := lib.Config.Fread(lib.ConfigFile); err != nil {
		log.Fatalf("Error in reading configuration file %s: %v", *(lib.ConfigFile), err)
	}

	// check configuration
	if err := lib.Config.Validate(); err != nil {
		log.Fatalf("Configuration checks failed: %v", err)
	}

	// set application variables
	u, err := url.Parse(lib.Config.Server.Settings.GrafanaURL)
	if err != nil {
		log.Fatalf("Error in parsing url %s: %v", lib.Config.Server.Settings.GrafanaURL, err)
	}

	u.Path = path.Join(u.Path, lib.AnnotationsAPI)
	lib.GrafanaAnnotationsURL = u.String()
	lib.SqliteHome = lib.Config.Server.Settings.SqliteHome
	lib.SqliteDB = filepath.Join(lib.SqliteHome, "annotations.db")
	lib.AlertsPath = lib.Config.Server.Settings.AlertsPath
	lib.AlertsFiringPath = filepath.Join(lib.AlertsPath, "firing")
	lib.AlertsResolvedPath = filepath.Join(lib.AlertsPath, "resolved")
	lib.AlertsPaths = map[string]string{"firing": lib.AlertsFiringPath, "resolved": lib.AlertsResolvedPath}

	// initialize database
	lib.C, err = db.Connect(lib.SqliteDB)
	if err := db.Init(lib.C); err != nil {
		panic(err)
	}
	db.Close(lib.C)
}

func main() {

	// start prometheus server
	if lib.Config.Server.Settings.Metrics {
		go lib.ServeMetrics()
	}

	// start routines
	go routines.ScanF(lib.Ch)
	go routines.Post(lib.Ch)
	go routines.ScanR(lib.Ch)
	go routines.Update(lib.Ch)
	go routines.Delete(lib.Ch)

	// restart routines that exited unexpectedly
	for {
		res := fmt.Sprintf("%s", <-lib.Ch)
		switch res {
		case "ScanF":
			log.Printf("Restarting %s routine", res)
			go routines.ScanF(lib.Ch)
		case "Post":
			log.Printf("Restarting %s routine", res)
			go routines.Post(lib.Ch)
		case "ScanR":
			log.Printf("Restarting %s routine", res)
			go routines.ScanR(lib.Ch)
		case "Update":
			log.Printf("Restarting %s routine", res)
			go routines.Update(lib.Ch)
		case "Delete":
			log.Printf("Restarting %s routine", res)
			go routines.Delete(lib.Ch)
		}
	}

}
