package lib

import (
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
