package libgobuster

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// WordlistType represents the type of wordlist source
type WordlistType int

const (
	// WordlistTypeFile is a local file
	WordlistTypeFile WordlistType = iota
	// WordlistTypeStdin is standard input
	WordlistTypeStdin
	// WordlistTypeURL is a remote URL
	WordlistTypeURL
)

// GetWordlistType returns the type of wordlist based on the path
func GetWordlistType(path string) WordlistType {
	if path == "-" {
		return WordlistTypeStdin
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return WordlistTypeURL
	}
	return WordlistTypeFile
}

// OpenWordlist opens the wordlist source and returns a ReadCloser
func OpenWordlist(ctx context.Context, path string) (io.ReadCloser, error) {
	switch GetWordlistType(path) {
	case WordlistTypeStdin:
		return io.NopCloser(os.Stdin), nil
	case WordlistTypeURL:
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for wordlist: %w", err)
		}
		req.Header.Set("User-Agent", DefaultUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch wordlist: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch wordlist: received status code %d", resp.StatusCode)
		}
		return resp.Body, nil
	case WordlistTypeFile:
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open wordlist: %w", err)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unknown wordlist type for path %q", path)
	}
}
