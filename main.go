package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// downloadFile downloads a file from a given URL and saves it to a specified destination.
func downloadFile(url, destination string) error {
	// Make an HTTP GET request to the specified URL
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the HTTP response status code is OK (200)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP response error: %s", response.Status)
	}

	// Create a new file at the destination
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the contents of the HTTP response body to the file
	_, err = io.Copy(out, response.Body)
	return err
}

func main() {
	// Check if the correct number of command-line arguments is provided
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . [URL]")
		os.Exit(1)
	}

	// Get the URL from the command-line arguments
	url := os.Args[1]

	// Extract the filename from the URL
	filename := filepath.Base(url)

	// Display a message indicating the start of the download
	fmt.Printf("Downloading %s...\n", filename)

	// Call the downloadFile function to download the file
	err := downloadFile(url, filename)
	if err != nil {
		// Display an error message and exit if there's an error during download
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display a message indicating the successful download
	fmt.Printf("Downloaded %s\n", filename)
}
