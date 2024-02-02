package main

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	fmt.Println("start at", strconv.Itoa(year)+"-"+month.String()+"-"+strconv.Itoa(day)+" "+strconv.Itoa(hour)+":"+strconv.Itoa(min)+":"+strconv.Itoa(sec))

	// sending request, awaiting response...
	print("sending request, awaiting response... ")
	lien, options := getArgs()
	resp, err := http.Get(lien)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("status", resp.Status)
	fmt.Println(options)

	// Determine the file extension from the Content-Type header
	fileExt := getFileExtension(resp.Header.Get("Content-Type"))

	switch fileExt {
	case ".png", ".jpg", ".jpeg":
		// Read the image
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			fmt.Println("Error decoding image:", err)
			return
		}

		// You can now work with the 'img' variable which contains the image.
		// For example, you can save it to a file or perform other operations.
		fmt.Println("Image loaded successfully.")

	case ".pdf":
		// You can use a PDF library to handle PDF files (e.g., github.com/unidoc/unipdf).
		fmt.Println("PDF file received. You can handle it using a PDF library.")
	default:
		// For other file types, read the content as a binary file
		saveBinaryFile(resp.Body, "output"+fileExt)
		fmt.Println("Binary file saved successfully.")
	}
}

func getArgs() (string, []string) {
	lien := ""
	var options []string
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			options = append(options, os.Args[i])
		} else if strings.Contains(os.Args[i], "http") {
			lien = os.Args[i]
		}
	}
	return lien, options
}

func getFileExtension(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpeg"
	case "application/pdf":
		return ".pdf"
	// Add more cases for other content types as needed
	default:
		// If the content type is unknown, assume it's a binary file
		return ".bin"
	}
}

func saveBinaryFile(reader io.Reader, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}
