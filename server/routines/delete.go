package routines

import (
	"annotator/db"
	"fmt"
	"log"
	"strconv"
	"time"

	"annotator/server/lib"
)

// Delete removes annotations older than retention period
func Delete(ch chan<- string) {

	for {

		t := map[string]string{
			"status": "'staled'"}

		lib.Mu.Lock()
		// create connection to database
		c, err := db.Connect(lib.SqliteDB)
		if err != nil {
			lib.Mu.Unlock()
			db.Close(c)
			log.Printf("Error in creating db connection : %v", err)
			ch <- fmt.Sprintf("%s", "Delete")
			return
		}

		// execute the query
		err = db.UpdateWithEndsAt(c, t)
		if err != nil {
			lib.Mu.Unlock()
			db.Close(c)
			log.Printf("Error in UpdateWithEndsAt : %v", err)
			ch <- fmt.Sprintf("%s", "Delete")
			return
		}
		//fmt.Printf("\n%#v", rs)
		lib.Mu.Unlock()

		// give mutex back
		time.Sleep(5 * time.Second)

		// get annotations that are marked
		lib.Mu.Lock()
		// execute the query
		rs, err := db.Select(c, t)
		if err != nil {
			lib.Mu.Unlock()
			db.Close(c)
			log.Printf("Error in SELECT : %v", err)
			ch <- fmt.Sprintf("%s", "Delete")
			return
		}
		//fmt.Printf("\n%#v", rs)
		lib.Mu.Unlock()

		for _, r := range rs {

			// make the api/annotations/<start id> url
			log.Printf("%s : Delete starts_id (%d)\n", r.AlertHash, r.StartsID)
			idURL := lib.GrafanaAnnotationsURL + "/" + strconv.FormatInt(r.StartsID, 10)

			res, err := lib.DeleteAnnotationWithID(idURL)
			if err != nil {
				ch <- fmt.Sprintf("%s", "Delete")
				return
			}
			log.Printf("%s : Message (%s)\n", r.AlertHash, res)

			// make the api/annotations/<end id> url
			log.Printf("%s : Delete ends_id (%d)\n", r.AlertHash, r.EndsID)
			idURL = lib.GrafanaAnnotationsURL + "/" + strconv.FormatInt(r.EndsID, 10)

			res, err = lib.DeleteAnnotationWithID(idURL)
			if err != nil {
				ch <- fmt.Sprintf("%s", "Delete")
				return
			}

			log.Printf("%s : Message (%s)\n", r.AlertHash, res)
			lib.Mu.Lock()
			if err = db.DeleteWithHash(c, r.AlertHash); err != nil {
				log.Printf("Error in DeleteWithHash : %v", err)
			}
			lib.Mu.Unlock()
		}

		db.Close(c)
		time.Sleep(5 * time.Second)
	}
}
