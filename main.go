package main

import (
	"fmt"
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
	lien, options := getArgs()
	resp, err := http.Get(lien)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("status", resp.Status)
	fmt.Println(options)

	// Read the HTML content
	htmlContent, err := parseHTML(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// Download CSS and JavaScript files
	downloadAssets(lien, htmlContent, "output_folder")
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
