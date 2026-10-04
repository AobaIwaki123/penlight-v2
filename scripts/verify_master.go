//go:build ignore

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	_ "modernc.org/sqlite"
)

type ImageSourceEntry struct {
	ImageKey string `json:"dest_key"`
	URL      string `json:"url"`
}

func main() {
	fmt.Println("=== [Master Data Integrity Verification] ===")

	dbPath := filepath.Join("data", "penlight.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ Database file not found: %s\n", dbPath)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	hasErrors := false

	// 1. Verify master version
	var currentVersion string
	err = db.QueryRow("SELECT version FROM master_versions WHERE id = 'current';").Scan(&currentVersion)
	if err != nil {
		fmt.Printf("❌ Failed to query master_versions: %v\n", err)
		hasErrors = true
	} else if currentVersion != model.CurrentMasterVersion {
		fmt.Printf("❌ Master version mismatch: DB has %q, expected %q\n", currentVersion, model.CurrentMasterVersion)
		hasErrors = true
	} else {
		fmt.Printf("✅ Master version: %s (matches canonical model)\n", currentVersion)
	}

	// 2. Verify primary images: must have 0 duplicates
	var totalPrimary, distinctMembers int
	err = db.QueryRow("SELECT count(*), count(distinct member_id) FROM member_images WHERE is_primary = 1;").Scan(&totalPrimary, &distinctMembers)
	if err != nil {
		fmt.Printf("❌ Failed to count primary images: %v\n", err)
		hasErrors = true
	} else if totalPrimary != distinctMembers {
		fmt.Printf("❌ Duplicate primary images detected! Total primary: %d, Unique members: %d\n", totalPrimary, distinctMembers)
		hasErrors = true
	} else {
		fmt.Printf("✅ Primary images: %d unique (0 duplicates)\n", totalPrimary)
	}

	// 3. Verify counts
	var groupCount, colorCount, memberCount, activeCount int
	_ = db.QueryRow("SELECT count(*) FROM groups WHERE is_active = 1;").Scan(&groupCount)
	_ = db.QueryRow("SELECT count(*) FROM colors;").Scan(&colorCount)
	_ = db.QueryRow("SELECT count(*) FROM members;").Scan(&memberCount)
	_ = db.QueryRow("SELECT count(*) FROM members WHERE status != 'graduated';").Scan(&activeCount)
	fmt.Printf("✅ Counts: %d groups, %d colors, %d total members (%d active)\n", groupCount, colorCount, memberCount, activeCount)

	// 4. Verify image keys match image_sources.json
	imgSrcPath := filepath.Join("seeds", "data", "image_sources.json")
	srcBytes, err := os.ReadFile(imgSrcPath)
	if err != nil {
		fmt.Printf("❌ Failed to read %s: %v\n", imgSrcPath, err)
		hasErrors = true
	} else {
		var entries []ImageSourceEntry
		if err := json.Unmarshal(srcBytes, &entries); err != nil {
			fmt.Printf("❌ Failed to parse %s: %v\n", imgSrcPath, err)
			hasErrors = true
		} else {
			knownKeys := make(map[string]bool)
			for _, e := range entries {
				knownKeys[e.ImageKey] = true
			}

			rows, err := db.Query("SELECT image_key FROM member_images;")
			if err != nil {
				fmt.Printf("❌ Failed to query member_images: %v\n", err)
				hasErrors = true
			} else {
				defer rows.Close()
				missingCount := 0
				for rows.Next() {
					var k string
					_ = rows.Scan(&k)
					if !knownKeys[k] {
						missingCount++
					}
				}
				if missingCount > 0 {
					fmt.Printf("❌ %d member images not found in image_sources.json\n", missingCount)
					hasErrors = true
				} else {
					fmt.Printf("✅ All %d DB image keys exist in image_sources.json\n", len(entries))
				}
			}
		}
	}

	// 5. Optional live server check if port 8080 is reachable
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8080/healthz")
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		fmt.Println("=== [Live Server Check (http://localhost:8080)] ===")
		
		bootResp, err := client.Get("http://localhost:8080/api/v1/sync/bootstrap")
		if err != nil {
			fmt.Printf("❌ Bootstrap endpoint failed: %v\n", err)
			hasErrors = true
		} else {
			defer bootResp.Body.Close()
			etag := bootResp.Header.Get("ETag")
			expectedEtag := fmt.Sprintf(`"%s"`, model.CurrentMasterVersion)
			if etag != expectedEtag {
				fmt.Printf("❌ Server ETag mismatch: got %s, expected %s\n", etag, expectedEtag)
				hasErrors = true
			} else {
				fmt.Printf("✅ Bootstrap ETag: %s\n", etag)
			}
		}
	}

	if hasErrors {
		fmt.Println("❌ Master data verification failed!")
		os.Exit(1)
	}

	fmt.Println("✅ All master data verification checks passed successfully!")
}
