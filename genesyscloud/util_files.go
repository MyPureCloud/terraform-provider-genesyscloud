package genesyscloud

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func downloadOrOpenFile(path string) (io.Reader, *os.File, error) {
	var reader io.Reader
	var file *os.File

	_, err := os.Stat(path)
	if err != nil {
		_, err = url.Parse(path)
		if err != nil {
			resp, err := http.Get(path)
			if err != nil {
				return nil, nil, err
			}
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				return nil, nil, fmt.Errorf("HTTP Error downloading file: %v", resp.StatusCode)
			}
			reader = resp.Body
		} else {
			return nil, nil, fmt.Errorf("Invalid file path or URL: %v", path)
		}
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		reader = file
	}

	return reader, file, nil
}
