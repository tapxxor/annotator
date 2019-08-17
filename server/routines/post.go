package routines

import (
	"annotator/db"
	"annotator/server/lib"
	"annotator/types"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Post routine post a new annotation to grafana based on firing alert
func Post(ch chan<- string) {

	for {
		// get the alert entry from database
		lib.Mu.Lock()
		// create connection to database
		c, err := db.Connect(lib.SqliteDB)
		if err != nil {
			lib.Mu.Unlock()
			log.Printf("Error in creating db connection : %v", err)
			ch <- fmt.Sprintf("%s", "Post")
			return
		}

		t := map[string]string{"status": fmt.Sprintf("%q", "init")}
		// execute the query
		rs, err := db.Select(c, t)
		if err != nil {
			lib.Mu.Unlock()
			log.Printf("Error in SELECT where status='init' : %v", err)
			ch <- fmt.Sprintf("%s", "Post")
			return
		}
		//fmt.Printf("\n%#v", rs)
		lib.Mu.Unlock()
		db.Close(c)

		for _, r := range rs {

			ap := types.AnnotationsPost{Time: r.StartsAt,
				IsRegion: true,
				TimeEnd:  r.StartsAt + 10000, Tags: []string{r.Alertname},
				Text: r.Description}

			apVal, err := json.Marshal(ap)
			if err != nil {
				panic(err)
			}

			log.Printf("%s : POST annotation\n", r.AlertHash)
			req, err := http.NewRequest("POST", lib.GrafanaAnnotationsURL, bytes.NewBuffer(apVal))
			req.Header.Add("Authorization", "Bearer "+lib.Config.Server.Settings.Apikey)
			req.Header.Set("Content-Type", "application/json")

			// Send req using http Client
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error in POST %s: %v", lib.GrafanaAnnotationsURL, err)
				ch <- fmt.Sprintf("%s", "Post")
				return
			}

			// Decode the JSON in the body
			var apr types.AnnotationsMethodResponse
			if err := json.NewDecoder(resp.Body).Decode(&apr); err != nil {
				resp.Body.Close()
				ch <- fmt.Sprintf("%s", "Post")
				return
			}

			log.Printf("%s : POST code : %d\n", r.AlertHash, resp.StatusCode)
			if resp.StatusCode != 200 {
				resp.Body.Close()
				log.Printf("Error at POST %s : %d received", lib.GrafanaAnnotationsURL, resp.StatusCode)
				ch <- fmt.Sprintf("%s", "Post")
				return
			}

			if apr.Message == "Annotation added" {
				// Update database with starts_id and ends_id
				// get the alert entry from database
				log.Printf("%s : %s : (%d, %d)\n", r.AlertHash, apr.Message, apr.ID, apr.EndID)

				lib.Mu.Lock()
				// create connection to database
				c, err := db.Connect(lib.SqliteDB)
				if err != nil {
					lib.Mu.Unlock()
					log.Printf("Error in creating db connection : %v", err)
					ch <- fmt.Sprintf("%s", "Post")
					return
				}

				t := map[string]string{
					"starts_id": strconv.FormatInt(apr.ID, 10),
					"ends_id":   strconv.FormatInt(apr.EndID, 10),
					"status":    "'firing'"}

				// execute the query
				err = db.UpdateWithHash(c, t, r.AlertHash)
				if err != nil {
					lib.Mu.Unlock()
					db.Close(c)
					log.Printf("Error in UpdateWithHash : %v", err)
					ch <- fmt.Sprintf("%s", "Post")
					return
				}
				//fmt.Printf("\n%#v", rs)
				lib.Mu.Unlock()
				db.Close(c)
				log.Printf("%s : Status changed to %q\n", r.AlertHash, "firing")
			} else {
				log.Printf("Error at POST %s : message %s received", lib.GrafanaAnnotationsURL, apr.Message)
				ch <- fmt.Sprintf("%s", "Post")
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}
