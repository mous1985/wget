package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func main() {
	year, month, day := time.Now().Date()
	hour, min, sec := time.Now().Clock()
	fmt.Println("start at", strconv.Itoa(year)+"-"+month.String()+"-"+strconv.Itoa(day)+" "+strconv.Itoa(hour)+":"+strconv.Itoa(min)+":"+strconv.Itoa(sec))

	// sending request, awaiting response...
	print("sending request, awaiting response... ")
	resp, err := http.Get(lien)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("status", resp.Status)
	if newFilename == "" {
		// Obtenir le nom du fichier à partir de l'URL
		newFilename = filepath.Base(lien)
	}
	// Determine the file extension from the Content-Type header
	contentType := resp.Header.Get("Content-Type")
	fileExt := getFileExtension(contentType)

	switch {
	case strings.HasPrefix(contentType, "image"):
		// Handle image files
		handleImage(resp.Body, fileExt, newFilename, outputFolder)
	case strings.HasPrefix(contentType, "application/pdf"):
		// Handle PDF files
		fmt.Println("PDF file received. You can handle it using a PDF library.")
	case strings.HasPrefix(contentType, "text/html"):
		// Handle HTML content
		outputFolder := "website_clone"
		err := cloneWebsite(resp.Body, lien, outputFolder)
		if err != nil {
			fmt.Println("Error cloning website:", err)
		}
	default:
		// Handle other file types as binary data
		saveBinaryFile(resp.Body, "output"+fileExt)
		fmt.Println("Binary file saved successfully.")
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getFileExtension(contentType string) string {
	if strings.Contains(contentType, "image/png") {
		return ".png"
	} else if strings.Contains(contentType, "image/jpeg") {
		return ".jpeg"
	} else if strings.Contains(contentType, "application/pdf") {
		return ".pdf"
	} else if strings.Contains(contentType, "text/html") {
		return ".html"
	} else {
		return ".bin"
	}
}

func handleImage(reader io.Reader, fileExt string, newFilename string, outputFolder string) {
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}
	// Example of saving the image
	saveImage(img, newFilename, outputFolder)
}

func saveImage(img image.Image, newFilename string, outputFolder string) {
	file, err := os.Create(newFilename)
	if err != nil {
		fmt.Println("Error creating image file:", err)
		return
	}
	defer file.Close()

	switch filepath.Ext(newFilename) {
	case ".png":
		png.Encode(file, img)
	case ".jpeg", ".jpg":
		jpeg.Encode(file, img, nil)
	default:
		fmt.Println("Unsupported image format for saving")
	}
}

func cloneWebsite(body io.Reader, baseURL, outputFolder string) error {
	htmlContent, err := parseHTML(body)
	if err != nil {
		return err
	}
	downloadAssets(baseURL, htmlContent, outputFolder)
	return nil
}

// Add the functions `saveBinaryFile` and any other helper functions here as they are from your original code or new implementations.
func saveBinaryFile(reader io.Reader, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func parseHTML(body io.Reader) (string, error) {
	tokenizer := html.NewTokenizer(body)
	var htmlContent strings.Builder

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return htmlContent.String(), nil
		case html.TextToken, html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			htmlContent.WriteString(tokenizer.Token().String())
		}
	}
}

func downloadAssets(baseURL, htmlContent, outputFolder string) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// Create the output folder if it doesn't exist
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		os.Mkdir(outputFolder, os.ModePerm)
	}

	var downloadFile func(node *html.Node)
	downloadFile = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "link" || node.Data == "script" {
				for _, attr := range node.Attr {
					if attr.Key == "href" || attr.Key == "src" {
						fileURL := attr.Val
						if !strings.HasPrefix(fileURL, "http") {
							fileURL = baseURL + "/" + fileURL
						}

						resp, err := http.Get(fileURL)
						if err != nil {
							fmt.Println("Error downloading file:", err)
							return
						}
						defer resp.Body.Close()

						// Extract the file name from the URL
						fileName := filepath.Base(fileURL)

						// Save the file to the output folder
						outputPath := filepath.Join(outputFolder, fileName)
						file, err := os.Create(outputPath)
						if err != nil {
							fmt.Println("Error creating file:", err)
							return
						}
						defer file.Close()

						_, err = io.Copy(file, resp.Body)
						if err != nil {
							fmt.Println("Error copying file content:", err)
							return
						}

						fmt.Println("Downloaded:", fileURL)
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			downloadFile(child)
		}
	}

	downloadFile(doc)
}
