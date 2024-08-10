package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func bulkEncode(inputDir string) {
	// Walk through all the files in the directory
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the current item is a file (not a directory)
		if !info.IsDir() {
			// Get the file extension
			ext := strings.ToLower(filepath.Ext(path))

			// Check if the file extension is either .mp4 or .mkv
			if ext == ".mp4" || ext == ".mkv" {
				log.Printf("Encoding file: %s\n", path)
				encode(path) // Call the encode function for each file
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("bulkEncode() failed with %s\n", err)
	}
}
