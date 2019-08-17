package routines

import (
	"annotator/db"
	"annotator/server/lib"
	"annotator/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// ScanR routine checks for resolved alerts from receiver
func ScanR(ch chan<- string) {

	for {
		// scan for resolved alerts
		var files []string
		files, err := lib.ScanFolder(lib.AlertsPaths["resolved"])
		if err != nil {
			ch <- fmt.Sprintf("%s", "ScanR")
			return
		}

		for _, file := range files {

			alert := &types.Message{}

			jsonFile, err := os.Open(file)
			// if we os.Open returns an error then panic
			if err != nil {
				log.Printf("%s: Error in reading : %v\n", filepath.Base(file), err)
				ch <- fmt.Sprintf("%s", "ScanR")
				return
			}
			log.Printf("%s : Read file\n", filepath.Base(file))

			d := json.NewDecoder(jsonFile)
			err = d.Decode(alert)
			if err != nil {
				log.Printf("Error in decoding %s: %v\n", filepath.Base(file), err)
				ch <- fmt.Sprintf("%s", "ScanR")
				return
			}
			log.Printf("%s : Decoded JSON\n", filepath.Base(file))

			tmpEndsAt := (alert.Alerts[0].EndsAt.UnixNano()) / 1000000
			t := map[string]string{
				"ends_at": strconv.FormatInt(tmpEndsAt, 10),
				"status":  "'resolved'"}

			lib.Mu.Lock()
			// create connection to database
			c, err := db.Connect(lib.SqliteDB)
			if err != nil {
				lib.Mu.Unlock()
				db.Close(c)
				log.Printf("Error in creating db connection : %v", err)
				ch <- fmt.Sprintf("%s", "ScanR")
				return
			}

			// execute the query
			err = db.UpdateWithHash(c, t, filepath.Base(file))
			if err != nil {
				lib.Mu.Unlock()
				db.Close(c)
				log.Printf("Error in UpdateWithHash : %v", err)
				ch <- fmt.Sprintf("%s", "ScanR")
				return
			}

			log.Printf("%s : Status changed to %q\n", filepath.Base(file), "resolved")
			//fmt.Printf("\n%#v", rs)
			lib.Mu.Unlock()
			db.Close(c)
			c, t = nil, nil

			// delete file from disk
			if err = os.Remove(file); err != nil {
				log.Printf("Error in deleting file %s : %v", file, err)
				panic(err)
			}
			log.Printf("%s : Deleted file\n", file)

		}
		files = nil
	}
}
