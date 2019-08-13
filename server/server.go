package main

import (
	"annotator/db"
	"annotator/lib"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	annotationsAPI string = "/api/annotations"
)

var (
	config                lib.ServerConf
	grafanaAPIAnnotations lib.AnnotationsResponse
	regions               lib.Regions
	configFile            *string
	grafanaAnnotationsURL string
	sqliteHome            string
	sqliteDB              string
	alertsPath            string
	alertsFiringPath      string
	alertsResolvedPath    string
	alertsPaths           map[string]string
	c                     *sql.DB
	mu                    sync.Mutex
	ch                    = make(chan string)
)

func serveMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Start exposing prometheus metrics on port %d\n", config.Server.Settings.Port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(config.Server.Settings.Port)), nil))
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

	// set application variables
	u, err := url.Parse(config.Server.Settings.GrafanaURL)
	if err != nil {
		log.Fatalf("Error in parsing url %s: %v", config.Server.Settings.GrafanaURL, err)
	}

	u.Path = path.Join(u.Path, annotationsAPI)
	grafanaAnnotationsURL = u.String()
	sqliteHome = config.Server.Settings.SqliteHome
	sqliteDB = filepath.Join(sqliteHome, "annotations.db")
	alertsPath = config.Server.Settings.AlertsPath
	alertsFiringPath = filepath.Join(alertsPath, "firing")
	alertsResolvedPath = filepath.Join(alertsPath, "resolved")
	alertsPaths = map[string]string{"firing": alertsFiringPath, "resolved": alertsResolvedPath}

	// initialize database
	c, err = db.Connect(sqliteDB)
	if err := db.Init(c); err != nil {
		panic(err)
	}
	db.Close(c)
}

func scanFolder(t string) (err error) {
	// read alert files from disk and save the names to files slice
	var files []string

	err = filepath.Walk(t, func(fpath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, fpath)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if len(files) == 0 {
		return err
	}

	// for every alert populate the map seenFiles make(map[string]*lib.Message)
	for _, file := range files {

		alert := &lib.Message{}

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then panic
		if err != nil {
			log.Printf("Error in reading %s: %v\n", file, err)
			return err
		}
		log.Printf("Read %s\n", file)

		d := json.NewDecoder(jsonFile)
		err = d.Decode(alert)
		if err != nil {
			log.Printf("Error in decoding %s: %v\n", file, err)
			return err
		}
		log.Printf("Decoded %s\n", file)

		if filepath.Base(t) == "firing" {
			// prepare query payload
			tmpStartsAt := (alert.Alerts[0].StartsAt.UnixNano()) / 1000000
			t := map[string]string{
				"starts_hash": fmt.Sprintf("%q", filepath.Base(file)),
				"starts_at":   strconv.FormatInt(tmpStartsAt, 10),
				"alertname":   fmt.Sprintf("%q", alert.Alerts[0].Annotations["summary"]),
				"description": fmt.Sprintf("%q", alert.Alerts[0].Annotations["Description"]),
				"status":      fmt.Sprintf("%q", "init")}

			mu.Lock()
			// create connection to database
			c, err = db.Connect(sqliteDB)
			if err != nil {
				mu.Unlock()
				log.Printf("Error in creating db connection : %v", err)
				return err
			}

			// execute the query
			if err := db.Insert(c, t); err != nil {
				mu.Unlock()
				log.Printf("Error in INSERT : %v", err)
				return err
			}

			db.Close(c)
			mu.Unlock()
			c, t = nil, nil
			log.Printf("Insert query succeeded\n")

			// delete file from disk
			if err = os.Remove(file); err != nil {
				log.Printf("Error in deleting file %s : %v", file, err)
				panic(err)
			}
			log.Printf("Delete succeeded for %s\n", file)
		} else {
			// get the alert entry from database
			mu.Lock()
			// create connection to database
			c, err = db.Connect(sqliteDB)
			if err != nil {
				mu.Unlock()
				log.Printf("Error in creating db connection : %v", err)
				return err
			}

			t := map[string]string{"starts_hash": fmt.Sprintf("%q", filepath.Base(file))}
			// execute the query
			r, err := db.Select(c, t)
			if err != nil {
				mu.Unlock()
				log.Printf("Error in SELECT : %v", err)
				return err
			}
			fmt.Printf("\n%#v", *r)
		}
	}
	return
}

// Scan routine checks for new alerts from receiver
func Scan(ch chan<- string) {

	for {
		// scan for firing alerts
		if err := scanFolder(alertsPaths["firing"]); err != nil {
			ch <- fmt.Sprintf("%s", "run")
			return
		}
		time.Sleep(1 * time.Second)

		// scan for resolved alerts
		if err := scanFolder(alertsPaths["resolved"]); err != nil {
			ch <- fmt.Sprintf("%s", "run")
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// Post creates the grafana region annotation
func Post(ch chan<- string) {
	i := 0
	for {
		mu.Lock()
		fmt.Printf("Post: %d\n", i)
		i++
		mu.Unlock()
		if i == 10 {
			ch <- fmt.Sprintf("%s", "run")
		}
		time.Sleep(10 * time.Second)
	}
}

// Delete removes annotations older than retention period
func Delete(ch chan<- string) {
	i := 0
	for {
		mu.Lock()
		fmt.Printf("Delete: %d\n", i)
		i++
		mu.Unlock()
		if i == 10 {
			ch <- fmt.Sprintf("%s", "run")
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {

	// start prometheus server
	if config.Server.Settings.Metrics {
		go serveMetrics()
	}

	go Scan(ch)
	go Post(ch)
	go Delete(ch)
	// update annotations
	for {
		res := fmt.Sprintf("%s", <-ch)
		switch res {
		case "Scan":
			log.Printf("Restarting %s routine", res)
			go Scan(ch)
		case "Post":
			log.Printf("Restarting %s routine", res)
			go Post(ch)
		case "Delete":
			log.Printf("Restarting %s routine", res)
			go Delete(ch)
		}
	}

}

// grafanaAPIAnnotations.Load(grafanaAPIAnnotationsURL, config.Server.Settings.Apikey)

// update the regions map using grafanaAPIAnnotations contents
// err = regions.Load(&grafanaAPIAnnotations)
// if err != nil {
// 	log.Fatalf("Error in making Regions : %v\n", err)
// }

// log.Printf("\n%s\n", regions.String())
