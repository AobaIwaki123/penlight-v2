//go:build ignore

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aobaiwaki/penlight-v2/pkg/config"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
	_ "modernc.org/sqlite"
)

type ImageSourceEntry struct {
	DestKey string `json:"dest_key"`
	URL     string `json:"url"`
}

type ImageItem struct {
	PhotoTypeSlug string `json:"photo_type_slug"`
	URL           string `json:"url"`
	IsPrimary     bool   `json:"is_primary"`
}

type MemberExportItem struct {
	GroupSlug      string      `json:"group_slug"`
	FamilyName     string      `json:"family_name"`
	GivenName      string      `json:"given_name"`
	FamilyNameKana string      `json:"family_name_kana"`
	GivenNameKana  string      `json:"given_name_kana"`
	Generation     int         `json:"generation"`
	Status         string      `json:"status"`
	LeftColor      string      `json:"left_color"`
	RightColor     string      `json:"right_color"`
	VerifiedAt     *string     `json:"verified_at,omitempty"`
	Images         []ImageItem `json:"images"`
}

func main() {
	cfg := config.Load()
	dbPath := cfg.DBPath
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("SQLite database not found at %s. Please run server or pass DATA_DIR.", dbPath)
	}

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Load image sources map
	imageSources := make(map[string]string)
	if srcBytes, err := os.ReadFile("seeds/data/image_sources.json"); err == nil {
		var entries []ImageSourceEntry
		if err := json.Unmarshal(srcBytes, &entries); err == nil {
			for _, e := range entries {
				imageSources[e.DestKey] = e.URL
			}
		}
	}

	// Fetch all groups for slug map
	groups, err := repo.ListGroups(ctx)
	if err != nil {
		log.Fatalf("failed to list groups: %v", err)
	}
	groupSlugMap := make(map[string]string) // group_id -> slug
	for _, g := range groups {
		groupSlugMap[string(g.ID)] = g.Slug
	}

	// Fetch all colors for name map
	colors, err := repo.ListColors(ctx)
	if err != nil {
		log.Fatalf("failed to list colors: %v", err)
	}
	colorNameMap := make(map[string]string) // color_id -> name
	for _, c := range colors {
		colorNameMap[string(c.ID)] = c.Name
	}

	// Query all members (including graduated/hiatus)
	const memberQuery = `
		SELECT m.id, m.group_id, m.family_name, m.given_name, m.family_name_kana, m.given_name_kana,
		       m.generation, m.status, m.left_color_id, m.right_color_id, m.verified_at
		FROM members m
		JOIN groups g ON m.group_id = g.id
		ORDER BY g.display_order ASC, m.generation ASC, m.family_name_kana ASC;
	`
	rows, err := repo.DB().QueryContext(ctx, memberQuery)
	if err != nil {
		log.Fatalf("failed to query members: %v", err)
	}
	defer rows.Close()

	var exportList []MemberExportItem

	for rows.Next() {
		var id, groupID, fam, given, famKana, givenKana, status, leftColID, rightColID string
		var gen int
		var verifiedAt sql.NullString

		if err := rows.Scan(&id, &groupID, &fam, &given, &famKana, &givenKana, &gen, &status, &leftColID, &rightColID, &verifiedAt); err != nil {
			log.Fatalf("failed to scan member: %v", err)
		}

		gSlug := groupSlugMap[groupID]
		leftName := colorNameMap[leftColID]
		rightName := colorNameMap[rightColID]

		// Query images for member
		const imgQuery = `
			SELECT mi.image_key, mi.is_primary, pt.slug
			FROM member_images mi
			JOIN photo_types pt ON mi.photo_type_id = pt.id
			WHERE mi.member_id = ?
			ORDER BY mi.display_order ASC;
		`
		imgRows, err := repo.DB().QueryContext(ctx, imgQuery, id)
		if err != nil {
			log.Fatalf("failed to query images for %s: %v", id, err)
		}

		var images []ImageItem
		for imgRows.Next() {
			var imgKey, ptSlug string
			var isPrimary int
			if err := imgRows.Scan(&imgKey, &isPrimary, &ptSlug); err != nil {
				imgRows.Close()
				log.Fatalf("failed to scan image: %v", err)
			}
			url := imageSources[imgKey]
			if url == "" {
				url = fmt.Sprintf("https://cdn.example.com/%s", imgKey)
			}
			images = append(images, ImageItem{
				PhotoTypeSlug: ptSlug,
				URL:           url,
				IsPrimary:     isPrimary == 1,
			})
		}
		imgRows.Close()

		var vStr *string
		if verifiedAt.Valid && verifiedAt.String != "" {
			s := verifiedAt.String
			vStr = &s
		}

		exportList = append(exportList, MemberExportItem{
			GroupSlug:      gSlug,
			FamilyName:     fam,
			GivenName:      given,
			FamilyNameKana: famKana,
			GivenNameKana:  givenKana,
			Generation:     gen,
			Status:         status,
			LeftColor:      leftName,
			RightColor:     rightName,
			VerifiedAt:     vStr,
			Images:         images,
		})
	}

	outPath := filepath.Join("seeds", "data", "members.json")
	outBytes, err := json.MarshalIndent(exportList, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal export json: %v", err)
	}

	if err := os.WriteFile(outPath, append(outBytes, '\n'), 0644); err != nil {
		log.Fatalf("failed to write %s: %v", outPath, err)
	}

	log.Printf("Successfully exported %d members from SQLite to %s", len(exportList), outPath)
}
