package routines

import (
	"annotator/db"
	"annotator/types"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"annotator/server/lib"
)

// Update Updates the grafana region annotation with end timestamp
func Update(ch chan<- string) {

	for {
		// get the alert entry from database
		lib.Mu.Lock()
		// create connection to database
		c, err := db.Connect(lib.SqliteDB)
		if err != nil {
			lib.Mu.Unlock()
			log.Printf("Error in creating db connection : %v", err)
			ch <- fmt.Sprintf("%s", "Update")
			return
		}

		t := map[string]string{"status": fmt.Sprintf("%q", "resolved")}
		// execute the query
		rs, err := db.Select(c, t)
		if err != nil {
			lib.Mu.Unlock()
			db.Close(c)
			log.Printf("Error in SELECT : %v", err)
			ch <- fmt.Sprintf("%s", "Update")
			return
		}
		//fmt.Printf("\n%#v", rs)
		lib.Mu.Unlock()
		db.Close(c)
		for _, r := range rs {

			ap := types.AnnotationsPatch{Time: r.EndsAt}
			apVal, err := json.Marshal(ap)
			if err != nil {
				panic(err)
			}
			idURL := lib.GrafanaAnnotationsURL + "/" + strconv.FormatInt(r.EndsID, 10)

			req, err := http.NewRequest("PATCH", idURL, bytes.NewBuffer(apVal))
			req.Header.Add("Authorization", "Bearer "+lib.Config.Server.Settings.Apikey)
			req.Header.Set("Content-Type", "application/json")

			// Send req using http Client

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error in PATCH %s: %v", idURL, err)
				ch <- fmt.Sprintf("%s", "Update")
				return
			}

			// Decode the JSON in the body
			var apr types.AnnotationsMethodResponse
			if err := json.NewDecoder(resp.Body).Decode(&apr); err != nil {
				resp.Body.Close()
				ch <- fmt.Sprintf("%s", "Update")
				return
			}
			apVal, client, req = nil, nil, nil

			fmt.Printf("\n%#v\n", apr)
			if resp.StatusCode != 200 && resp.StatusCode != 404 {
				resp.Body.Close()
				log.Printf("Error at PATCH %s : %d received", idURL, resp.StatusCode)
				ch <- fmt.Sprintf("%s", "Update")
				return
			}
			log.Printf("%s: Patched annotation %d\n", r.AlertHash, r.EndsID)

			lib.Mu.Lock()
			// create connection to database
			c, err := db.Connect(lib.SqliteDB)
			if err != nil {
				lib.Mu.Unlock()
				db.Close(c)
				log.Printf("Error in creating db connection : %v", err)
				ch <- fmt.Sprintf("%s", "Update")
				return
			}

			var t map[string]string
			if apr.Message == "Annotation patched" {
				// Update database with starts_id and ends_id
				// get the alert entry from database
				t = map[string]string{
					"status": "'created'"}
			} else if apr.Message == "Not found" {
				t = map[string]string{
					"status": "'error'"}
			}

			// execute the query
			err = db.UpdateWithHash(c, t, r.AlertHash)
			if err != nil {
				lib.Mu.Unlock()
				db.Close(c)
				log.Printf("Error in UpdateWithHash : %v", err)
				ch <- fmt.Sprintf("%s", "Update")
				return
			}
			//fmt.Printf("\n%#v", rs)
			lib.Mu.Unlock()
			db.Close(c)
			log.Printf("%s : Status changed to %s\n", r.AlertHash, t["status"])
			t = nil
		}
		time.Sleep(1 * time.Second)
	}
}
