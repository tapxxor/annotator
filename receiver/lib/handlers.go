package receiver

import (
	"annotator/types"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
func alertToFile(m *types.Message) {
	alertFilename := sha256.Sum256([]byte(m.Alerts[0].StartsAt.String() + m.GroupKey))
	absPath := filepath.Join(
		filepath.Join(AlertsPaths[m.Alerts[0].Status]),
		fmt.Sprintf("%x", alertFilename))

	log.Printf("Saving to %s\n", absPath)

	if err := WriteToFile(absPath, m.String()); err != nil {
		log.Panicf("Error writing to %x: %v", alertFilename, err)
	}
}

// GetAlert received the json payload from alertmanager and stores it locally
func GetAlert(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s %s\n", r.Method, r.URL, r.Proto, r.RemoteAddr)
	switch r.Method {
	case "POST":
		// Decode the JSON in the body
		d := json.NewDecoder(r.Body)
		defer r.Body.Close()

		alert := &types.Message{}
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
