// package main

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// )

// // downloadFile downloads a file from a given URL and saves it to a specified destination.
// func downloadFile(url, destination string) error {
// 	// Make an HTTP GET request to the specified URL
// 	response, err := http.Get(url)
// 	if err != nil {
// 		return err
// 	}
// 	defer response.Body.Close()

// 	// Check if the HTTP response status code is OK (200)
// 	if response.StatusCode != http.StatusOK {
// 		return fmt.Errorf("HTTP response error: %s", response.Status)
// 	}

// 	// Create a new file at the destination
// 	out, err := os.Create(destination)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()

// 	// Copy the contents of the HTTP response body to the file
// 	_, err = io.Copy(out, response.Body)
// 	return err
// }

// func main() {
// 	// Check if the correct number of command-line arguments is provided
// 	if len(os.Args) != 2 {
// 		fmt.Println("Usage: go run . [URL]")
// 		os.Exit(1)
// 	}

// 	// Get the URL from the command-line arguments
// 	url := os.Args[1]

// 	// Extract the filename from the URL
// 	filename := filepath.Base(url)

// 	// Display a message indicating the start of the download
// 	fmt.Printf("Downloading %s...\n", filename)

// 	// Call the downloadFile function to download the file
// 	err := downloadFile(url, filename)
// 	if err != nil {
// 		// Display an error message and exit if there's an error during download
// 		fmt.Printf("Error: %v\n", err)
// 		os.Exit(1)
// 	}

//		// Display a message indicating the successful download
//		fmt.Printf("Downloaded %s\n", filename)
//	}
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// savePage saves the content of a web page to a file in the specified folder with the given filename.
func savePage(content []byte, folder, filename string) error {
	path := filepath.Join(folder, filename)
	return os.WriteFile(path, content, 0o644)
}

// mirrorDownload downloads a page and its resources (images, scripts, stylesheets) for mirroring purposes.
func mirrorDownload(urlStr, folder string) error {
	// Parse the URL to extract relevant information
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Create local folder and filename based on the URL
	pathParts := strings.Split(parsedURL.Path, "/")
	localFolder := filepath.Join(folder, path.Clean(path.Join(pathParts...)))
	localFilename := filepath.Join(localFolder, path.Base(urlStr))

	// Ensure the local folder exists
	os.MkdirAll(localFolder, os.ModePerm)

	// Make an HTTP GET request to the URL
	response, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the request was successful (HTTP status code 200)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download %s", urlStr)
	}

	// Create a local file and copy the content from the HTTP response body
	file, err := os.Create(localFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// downloadPage downloads a page, extracts links to images, scripts, and stylesheets,
// and saves the main HTML page with updated links.
func downloadPage(urlStr, folder string, reject, exclude []string) error {
	// Make an HTTP GET request to the URL
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create an HTML tokenizer to parse the response body
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		// Get the next token from the HTML tokenizer
		tokenType := tokenizer.Next()

		// Check if the token is an error token (end of HTML document)
		if tokenType == html.ErrorToken {
			return nil
		}

		// Check if the token is a start tag or self-closing tag
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			// Get the token information
			token := tokenizer.Token()

			// Check attributes of the tag for "href" or "src" attributes
			for _, attr := range token.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					// Join the URL with the attribute value to get the absolute URL
					linkURL := urljoin(urlStr, attr.Val)

					// Check if the link should be excluded based on reject and exclude criteria
					if shouldExclude(linkURL, reject, exclude) {
						continue
					}

					// Attempt to download the linked resource
					err := mirrorDownload(linkURL, folder)
					if err != nil {
						fmt.Printf("Failed to download %s: %s\n", linkURL, err)
					}
				}
			}
		}
	}
}

// shouldExclude checks whether a given link URL should be excluded based on reject and exclude criteria.
func shouldExclude(linkURL string, reject, exclude []string) bool {
	for _, r := range reject {
		if strings.HasSuffix(linkURL, r) {
			return true
		}
	}
	for _, e := range exclude {
		if strings.HasPrefix(linkURL, e) {
			return true
		}
	}
	return false
}

func main() {
	// Define command-line flags
	urlFlag := flag.String("url", "", "URL to the file to be downloaded")
	mirrorFlag := flag.Bool("mirror", false, "Mirror a website")
	rejectFlag := flag.String("reject", "", "List of file suffixes to avoid")
	excludeFlag := flag.String("exclude", "", "List of paths to avoid")
	folderFlag := flag.String("folder", ".", "Folder to save the downloaded files")

	// Parse command-line flags
	flag.Parse()

	// Check if a URL is provided
	if *urlFlag == "" {
		fmt.Println("Please provide a URL")
		return
	}

	// Check if the mirror flag is set
	if *mirrorFlag {
		// Split reject and exclude criteria from comma-separated strings
		reject := strings.Split(*rejectFlag, ",")
		exclude := strings.Split(*excludeFlag, ",")

		// Attempt to mirror the website
		err := downloadPage(*urlFlag, *folderFlag, reject, exclude)
		if err != nil {
			fmt.Printf("Error mirroring website: %s\n", err)
		}
		return
	}

	// Download a single file from the specified URL
	fmt.Printf("Downloading file from %s...\n", *urlFlag)
	err := mirrorDownload(*urlFlag, *folderFlag)
	if err != nil {
		fmt.Printf("Failed to download file: %s\n", err)
		return
	}

	fmt.Printf("File downloaded successfully to %s\n", *folderFlag)
}
