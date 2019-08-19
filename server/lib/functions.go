package lib

import (
	"annotator/types"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ScanFolder returns a slice of the files of the folder
func ScanFolder(t string) ([]string, error) {
	// read alert files from disk and save the names to files slice
	var files []string

	err := filepath.Walk(t, func(fpath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, fpath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, err
	}

	return files, err
}

// DeleteAnnotationWithID deletes annotation from grafana using ID
func DeleteAnnotationWithID(idURL string) (res string, err error) {

	req, err := http.NewRequest("DELETE", idURL, bytes.NewBuffer(nil))
	req.Header.Add("Authorization", "Bearer "+Config.Server.Settings.Apikey)
	req.Header.Set("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error at DELETE %s: %v", idURL, err)
		return "", err
	}

	// Decode the JSON in the body
	var apr types.AnnotationsMethodResponse
	if err := json.NewDecoder(resp.Body).Decode(&apr); err != nil {
		resp.Body.Close()
		return "", err
	}

	// 200 : "message": "Annotation deleted"
	// 500 : "message": "Could not find annotation to update"
	if resp.StatusCode != 200 && resp.StatusCode != 500 {
		resp.Body.Close()
		log.Printf("Error at DELETE %s : %d received (%s)", idURL, resp.StatusCode, apr.Message)
		return "", errors.New("http code not allowed")
	}

	resp.Body.Close()

	client, req = nil, nil

	return apr.Message, err
}
