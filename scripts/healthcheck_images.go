//go:build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ImageSource struct {
	ImageID       string `json:"image_id"`
	MemberID      string `json:"member_id"`
	Name          string `json:"name"`
	Group         string `json:"group"`
	PhotoTypeSlug string `json:"photo_type_slug"`
	URL           string `json:"url"`
	DestKey       string `json:"dest_key"`
}

type HealthIssue struct {
	MemberName string `json:"member_name"`
	MemberID   string `json:"member_id"`
	Group      string `json:"group"`
	PhotoType  string `json:"photo_type"`
	URL        string `json:"url"`
	IssueType  string `json:"issue_type"` // HTTP_ERROR, MIME_ERROR, DECODE_ERROR, ASPECT_RATIO_ANOMALY, FILE_TOO_SMALL
	Details    string `json:"details"`
}

type HealthCheckReport struct {
	TotalChecked int           `json:"total_checked"`
	ValidCount   int           `json:"valid_count"`
	IssueCount   int           `json:"issue_count"`
	Issues       []HealthIssue `json:"issues"`
}

func checkImage(ctx context.Context, client *http.Client, src ImageSource) *HealthIssue {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
	if err != nil {
		return &HealthIssue{
			MemberName: src.Name,
			MemberID:   src.MemberID,
			Group:      src.Group,
			PhotoType:  src.PhotoTypeSlug,
			URL:        src.URL,
			IssueType:  "HTTP_ERROR",
			Details:    fmt.Sprintf("create request error: %v", err),
		}
	}
	// ADR-0033: Do not send Referer to upstream CDN
	req.Header.Del("Referer")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PenlightHealthCheck/2.0)")

	resp, err := client.Do(req)
	if err != nil {
		return &HealthIssue{
			MemberName: src.Name,
			MemberID:   src.MemberID,
			Group:      src.Group,
			PhotoType:  src.PhotoTypeSlug,
			URL:        src.URL,
			IssueType:  "HTTP_ERROR",
			Details:    fmt.Sprintf("fetch error: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &HealthIssue{
			MemberName: src.Name,
			MemberID:   src.MemberID,
			Group:      src.Group,
			PhotoType:  src.PhotoTypeSlug,
			URL:        src.URL,
			IssueType:  "HTTP_ERROR",
			Details:    fmt.Sprintf("status code: %d", resp.StatusCode),
		}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 15*1024*1024))
	if err != nil {
		return &HealthIssue{
			MemberName: src.Name,
			MemberID:   src.MemberID,
			Group:      src.Group,
			PhotoType:  src.PhotoTypeSlug,
			URL:        src.URL,
			IssueType:  "HTTP_ERROR",
			Details:    fmt.Sprintf("read body error: %v", err),
		}
	}

	if len(body) < 1024 {
		return &HealthIssue{
			MemberName: src.Name,
			MemberID:   src.MemberID,
			Group:      src.Group,
			PhotoType:  src.PhotoTypeSlug,
			URL:        src.URL,
			IssueType:  "FILE_TOO_SMALL",
			Details:    fmt.Sprintf("file size %d bytes (< 1KB)", len(body)),
		}
	}

	detectedMIME := http.DetectContentType(body)
	if !strings.HasPrefix(detectedMIME, "image/") {
		return &HealthIssue{
			MemberName: src.Name,
			MemberID:   src.MemberID,
			Group:      src.Group,
			PhotoType:  src.PhotoTypeSlug,
			URL:        src.URL,
			IssueType:  "MIME_ERROR",
			Details:    fmt.Sprintf("unexpected MIME type %q", detectedMIME),
		}
	}

	// Decode dimensions
	cfg, _, decodeErr := image.DecodeConfig(bytes.NewReader(body))
	if decodeErr == nil {
		// Normal portrait photos are square (width == height) or portrait (height > width).
		// A landscape banner (width > height * 1.2) is abnormal for an idol profile picture.
		if cfg.Width > int(float64(cfg.Height)*1.2) {
			return &HealthIssue{
				MemberName: src.Name,
				MemberID:   src.MemberID,
				Group:      src.Group,
				PhotoType:  src.PhotoTypeSlug,
				URL:        src.URL,
				IssueType:  "ASPECT_RATIO_ANOMALY",
				Details:    fmt.Sprintf("landscape banner image (%dx%d, aspect ratio %.2f)", cfg.Width, cfg.Height, float64(cfg.Width)/float64(cfg.Height)),
			}
		}
	}

	return nil
}

func main() {
	sourcesFile := "seeds/data/image_sources.json"
	data, err := os.ReadFile(sourcesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", sourcesFile, err)
		os.Exit(1)
	}

	var sources []ImageSource
	if err := json.Unmarshal(data, &sources); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse %s: %v\n", sourcesFile, err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	concurrency := 8
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	report := HealthCheckReport{
		TotalChecked: len(sources),
		Issues:       make([]HealthIssue, 0),
	}

	for _, src := range sources {
		wg.Add(1)
		sem <- struct{}{}
		go func(s ImageSource) {
			defer wg.Done()
			defer func() { <-sem }()

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			issue := checkImage(ctx, client, s)

			mu.Lock()
			if issue != nil {
				report.Issues = append(report.Issues, *issue)
			} else {
				report.ValidCount++
			}
			mu.Unlock()
		}(src)
	}

	wg.Wait()
	report.IssueCount = len(report.Issues)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(report)

	if report.IssueCount > 0 {
		os.Exit(2)
	}
}
