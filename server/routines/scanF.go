package routines

import (
	"annotator/db"
	"annotator/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"annotator/server/lib"
)

// ScanF routine checks for new alerts from receiver
func ScanF(ch chan<- string) {

	for {
		// scan for firing alerts
		var files []string
		files, err := lib.ScanFolder(lib.AlertsPaths["firing"])
		if err != nil {
			ch <- fmt.Sprintf("%s", "ScanF")
			return
		}

		for _, file := range files {

			alert := &types.Message{}

			jsonFile, err := os.Open(file)
			// if we os.Open returns an error then panic
			if err != nil {
				log.Printf("%s : Error in reading : %v\n", filepath.Base(file), err)
				ch <- fmt.Sprintf("%s", "ScanF")
				return
			}
			log.Printf("%s : Read file\n", filepath.Base(file))

			d := json.NewDecoder(jsonFile)
			err = d.Decode(alert)
			if err != nil {
				log.Printf("%s : Error in decoding : %v\n", filepath.Base(file), err)
				ch <- fmt.Sprintf("%s", "ScanF")
				return
			}
			log.Printf("%s : Decoded JSON\n", filepath.Base(file))

			// prepare query payload
			tmpStartsAt := (alert.Alerts[0].StartsAt.UnixNano()) / 1000000

			evalPeriod := alert.Alerts[0].Labels["for"]
			evalPeriodDur, err := time.ParseDuration(evalPeriod)
			if err == nil {
				tmpStartsAt -= int64(evalPeriodDur / time.Millisecond)
			}

			t := map[string]string{
				"alert_hash":  fmt.Sprintf("%q", filepath.Base(file)),
				"starts_at":   strconv.FormatInt(tmpStartsAt, 10),
				"alertname":   fmt.Sprintf("%q", alert.Alerts[0].Annotations["summary"]),
				"description": fmt.Sprintf("%q", alert.Alerts[0].Annotations["Description"]),
				"status":      fmt.Sprintf("%q", "init")}

			lib.Mu.Lock()
			// create connection to database
			c, err := db.Connect(lib.SqliteDB)
			if err != nil {
				lib.Mu.Unlock()
				log.Printf("Error in creating db connection : %v", err)
				ch <- fmt.Sprintf("%s", "ScanF")
				return
			}

			// execute the query
			if err := db.Insert(c, t); err != nil {
				lib.Mu.Unlock()
				db.Close(c)
				log.Printf("Error in INSERT : %v", err)
				ch <- fmt.Sprintf("%s", "ScanF")
				return
			}

			db.Close(c)
			lib.Mu.Unlock()
			c, t, d = nil, nil, nil
			jsonFile, alert = nil, nil
			log.Printf("%s : Insert to DB\n", filepath.Base(file))

			// delete file from disk
			if err = os.Remove(file); err != nil {
				log.Printf("%s : Error in deleting file : %v", filepath.Base(file), err)
				panic(err)
			}
			log.Printf("%s : Deleted file\n", filepath.Base(file))

		}
		files = nil
	}
}
