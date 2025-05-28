package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadWithRedirects downloads a file from a URL with redirect handling
func DownloadWithRedirects(url string, destinationPath string) error {
	// Create HTTP client with redirect policy
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	// Download the file
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}
	// Write the response to the file
	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}
