//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// BQTableData represents the JSON structure exported from BigQuery REST API.
type BQTableData struct {
	Rows []struct {
		F []struct {
			V string `json:"v"`
		} `json:"f"`
	} `json:"rows"`
}

// MemberReading defines the family/given name separation and hiragana reading.
type MemberReading struct {
	FamilyName     string
	GivenName      string
	FamilyNameKana string
	GivenNameKana  string
}

// Fixed namespace for deterministic TypeID generation
var (
	nsGroup  = uuid.MustParse("018f3a00-0000-7000-8000-000000000001")
	nsColor  = uuid.MustParse("018f3a00-0000-7000-8000-000000000002")
	nsMember = uuid.MustParse("018f3a00-0000-7000-8000-000000000003")
)

// Readings dictionary for all 73 members across Hinatazaka46 and Sakurazaka46.
var memberReadings = map[string]MemberReading{
	// 日向坂46 (39名)
	"上村ひなの":  {"上村", "ひなの", "かみむら", "ひなの"},
	"下田衣珠季":  {"下田", "衣珠季", "しもだ", "いずき"},
	"丹生明里":   {"丹生", "明里", "にぶ", "あかり"},
	"佐々木久美":  {"佐々木", "久美", "ささき", "くみ"},
	"佐々木美玲":  {"佐々木", "美玲", "ささき", "みれい"},
	"佐藤優羽":   {"佐藤", "優羽", "さとう", "ゆう"},
	"加藤史帆":   {"加藤", "史帆", "かとう", "しほ"},
	"坂井新奈":   {"坂井", "新奈", "さかい", "にいな"},
	"大田美月":   {"大田", "美月", "おおた", "みづき"},
	"大野愛実":   {"大野", "愛実", "おおの", "まなみ"},
	"宮地すみれ":  {"宮地", "すみれ", "みやち", "すみれ"},
	"富田鈴花":   {"富田", "鈴花", "とみた", "すずか"},
	"小坂菜緒":   {"小坂", "菜緒", "こさか", "なお"},
	"小西夏菜実":  {"小西", "夏菜実", "こにし", "ななみ"},
	"山下葉留花":  {"山下", "葉留花", "やました", "はるか"},
	"山口陽世":   {"山口", "陽世", "やまぐち", "はるよ"},
	"平尾帆夏":   {"平尾", "帆夏", "ひらお", "ほのか"},
	"平岡海月":   {"平岡", "海月", "ひらおか", "みつき"},
	"東村芽依":   {"東村", "芽依", "ひがしむら", "めい"},
	"松尾桜":    {"松尾", "桜", "まつお", "さくら"},
	"松田好花":   {"松田", "好花", "まつだ", "このか"},
	"森本茉莉":   {"森本", "茉莉", "もりもと", "まりぃ"},
	"正源司陽子":  {"正源司", "陽子", "しょうげんじ", "ようこ"},
	"河田陽菜":   {"河田", "陽菜", "かわた", "ひな"},
	"清水理央":   {"清水", "理央", "しみず", "りお"},
	"渡辺莉奈":   {"渡辺", "莉奈", "わたなべ", "りな"},
	"濱岸ひより":  {"濱岸", "ひより", "はまぎし", "ひより"},
	"片山紗希":   {"片山", "紗希", "かたやま", "さき"},
	"石塚瑶季":   {"石塚", "瑶季", "いしづか", "たまき"},
	"竹内希来里":  {"竹内", "希来里", "たけうち", "きらり"},
	"蔵盛妃那乃":  {"蔵盛", "妃那乃", "くらもり", "ひなの"},
	"藤嶌果歩":   {"藤嶌", "果歩", "ふじしま", "かほ"},
	"金村美玖":   {"金村", "美玖", "かねむら", "みく"},
	"高井俐香":   {"高井", "俐香", "たかい", "りか"},
	"高本彩花":   {"高本", "彩花", "たかもと", "あやか"},
	"高瀬愛奈":   {"高瀬", "愛奈", "たかせ", "まな"},
	"髙橋未来虹":  {"髙橋", "未来虹", "たかはし", "みくに"},
	"鶴崎仁香":   {"鶴崎", "仁香", "つるさき", "にこ"},
	"齊藤京子":   {"齊藤", "京子", "さいとう", "きょうこ"},

	// 櫻坂46 (34名)
	"中嶋優月":  {"中嶋", "優月", "なかしま", "ゆづき"},
	"中川智尋":  {"中川", "智尋", "なかがわ", "ちひろ"},
	"井上梨名":  {"井上", "梨名", "いのうえ", "りな"},
	"佐藤愛桜":  {"佐藤", "愛桜", "さとう", "ねお"},
	"勝又春":   {"勝又", "春", "かつまた", "はる"},
	"向井純葉":  {"向井", "純葉", "むかい", "いとは"},
	"増本綺良":  {"増本", "綺良", "ますもと", "きら"},
	"大園玲":   {"大園", "玲", "おおぞの", "れい"},
	"大沼晶保":  {"大沼", "晶保", "おおぬま", "あきほ"},
	"守屋麗奈":  {"守屋", "麗奈", "もりや", "れな"},
	"小島凪紗":  {"小島", "凪紗", "こじま", "なぎさ"},
	"小池美波":  {"小池", "美波", "こいけ", "みなみ"},
	"小田倉麗奈": {"小田倉", "麗奈", "おだくら", "れいな"},
	"山下瞳月":  {"山下", "瞳月", "やました", "しづき"},
	"山川宇衣":  {"山川", "宇衣", "やまかわ", "うい"},
	"山田桃実":  {"山田", "桃実", "やまだ", "ももみ"},
	"山﨑天":   {"山﨑", "天", "やまさき", "てん"},
	"幸阪茉里乃": {"幸阪", "茉里乃", "こうさか", "まりの"},
	"村井優":   {"村井", "優", "むらい", "ゆう"},
	"村山美羽":  {"村山", "美羽", "むらやま", "みう"},
	"松本和子":  {"松本", "和子", "まつもと", "わこ"},
	"松田里奈":  {"松田", "里奈", "まつだ", "りな"},
	"森田ひかる": {"森田", "ひかる", "もりた", "ひかる"},
	"武元唯衣":  {"武元", "唯衣", "たけもと", "ゆい"},
	"浅井恋乃未": {"浅井", "恋乃未", "あさい", "このみ"},
	"田村保乃":  {"田村", "保乃", "たむら", "ほの"},
	"的野美青":  {"的野", "美青", "まとの", "みお"},
	"目黒陽色":  {"目黒", "陽色", "めぐろ", "ひいろ"},
	"石森璃花":  {"石森", "璃花", "いしもり", "りか"},
	"稲熊ひな":  {"稲熊", "ひな", "いなぐま", "ひな"},
	"藤吉夏鈴":  {"藤吉", "夏鈴", "ふじよし", "かりん"},
	"谷口愛季":  {"谷口", "愛季", "たにぐち", "あいり"},
	"遠藤光莉":  {"遠藤", "光莉", "えんどう", "ひかり"},
	"遠藤理子":  {"遠藤", "理子", "えんどう", "りこ"},
}

// GroupConfig defines a group to import.
type GroupConfig struct {
	Slug          string
	Name          string
	ShortName     string
	ThemeColorHex string
	DisplayOrder  int
	ColorsFile    string
	MembersFile   string
}

func parseGeneration(genStr string) int {
	cleaned := strings.TrimSuffix(genStr, "st")
	cleaned = strings.TrimSuffix(cleaned, "nd")
	cleaned = strings.TrimSuffix(cleaned, "rd")
	cleaned = strings.TrimSuffix(cleaned, "th")
	val, err := strconv.Atoi(cleaned)
	if err != nil {
		return 1
	}
	return val
}

func normalizeHex(hex string) string {
	hex = strings.TrimSpace(hex)
	if !strings.HasPrefix(hex, "#") {
		hex = "#" + hex
	}
	// Pad if 5 chars like #2bdd6 -> #02bdd6
	if len(hex) == 6 {
		hex = "#0" + hex[1:]
	}
	return strings.ToUpper(hex)
}

func main() {
	groups := []GroupConfig{
		{
			Slug:          "hinatazaka46",
			Name:          "日向坂46",
			ShortName:     "日向坂",
			ThemeColorHex: "#7CC7E8",
			DisplayOrder:  1,
			ColorsFile:    "data/bq_export/hinatazaka_penlight.json",
			MembersFile:   "data/bq_export/hinatazaka_member_master.json",
		},
		{
			Slug:          "sakurazaka46",
			Name:          "櫻坂46",
			ShortName:     "櫻坂",
			ThemeColorHex: "#F19DB5",
			DisplayOrder:  2,
			ColorsFile:    "data/bq_export/sakurazaka_penlight.json",
			MembersFile:   "data/bq_export/sakurazaka_member_master.json",
		},
	}

	var sqlStatements []string
	sqlStatements = append(sqlStatements, "-- Generated by scripts/import_bq_seeds.go")
	sqlStatements = append(sqlStatements, "-- Source: BigQuery sakamichipenlightquiz")
	sqlStatements = append(sqlStatements, "")

	type ImageSource struct {
		MemberID string `json:"member_id"`
		Name     string `json:"name"`
		Group    string `json:"group"`
		URL      string `json:"url"`
		Type     string `json:"type"`
		DestKey  string `json:"dest_key"`
	}
	var imageSources []ImageSource

	nowUTC := "2026-10-04T00:00:00Z"

	for _, g := range groups {
		groupID := fmt.Sprintf("grp_%s", strings.ReplaceAll(uuid.NewSHA1(nsGroup, []byte(g.Slug)).String(), "-", ""))

		sqlStatements = append(sqlStatements, fmt.Sprintf(
			"INSERT OR IGNORE INTO groups (id, name, short_name, slug, theme_color_hex, display_order, is_active, created_at, updated_at)\nVALUES ('%s', '%s', '%s', '%s', '%s', %d, 1, '%s', '%s');",
			groupID, g.Name, g.ShortName, g.Slug, g.ThemeColorHex, g.DisplayOrder, nowUTC, nowUTC,
		))

		// 1. Process Colors
		colorsDataBytes, err := os.ReadFile(g.ColorsFile)
		if err != nil {
			log.Fatalf("failed to read colors file %s: %v", g.ColorsFile, err)
		}
		var colorsData BQTableData
		if err := json.Unmarshal(colorsDataBytes, &colorsData); err != nil {
			log.Fatalf("failed to unmarshal colors json: %v", err)
		}

		colorIDMap := make(map[string]string) // BQ color int ID -> col_<uuid>
		type colorItem struct {
			bqID  int
			name  string
			hex   string
			order int
		}
		var colorList []colorItem
		for _, row := range colorsData.Rows {
			bqID, _ := strconv.Atoi(row.F[0].V)
			nameJa := row.F[1].V
			hexVal := normalizeHex(row.F[3].V)
			colorList = append(colorList, colorItem{
				bqID:  bqID,
				name:  nameJa,
				hex:   hexVal,
				order: bqID,
			})
		}
		sort.Slice(colorList, func(i, j int) bool {
			return colorList[i].order < colorList[j].order
		})

		sqlStatements = append(sqlStatements, fmt.Sprintf("\n-- Colors for %s", g.Name))
		for _, col := range colorList {
			colUUIDKey := fmt.Sprintf("%s:%s", g.Slug, col.name)
			colID := fmt.Sprintf("col_%s", strings.ReplaceAll(uuid.NewSHA1(nsColor, []byte(colUUIDKey)).String(), "-", ""))
			colorIDMap[strconv.Itoa(col.bqID)] = colID

			sqlStatements = append(sqlStatements, fmt.Sprintf(
				"INSERT OR IGNORE INTO colors (id, group_id, name, hex_code, display_order, created_at, updated_at)\nVALUES ('%s', '%s', '%s', '%s', %d, '%s', '%s');",
				colID, groupID, col.name, col.hex, col.order, nowUTC, nowUTC,
			))
		}

		// 2. Process Members
		membersDataBytes, err := os.ReadFile(g.MembersFile)
		if err != nil {
			log.Fatalf("failed to read members file %s: %v", g.MembersFile, err)
		}
		var membersData BQTableData
		if err := json.Unmarshal(membersDataBytes, &membersData); err != nil {
			log.Fatalf("failed to unmarshal members json: %v", err)
		}

		type memberRowData struct {
			id       string
			name     string
			gen      string
			grad     bool
			p1       string
			p2       string
			imgType  string
			imageURL string
		}

		memberMap := make(map[string][]memberRowData)
		for _, row := range membersData.Rows {
			mid := row.F[0].V
			name := row.F[1].V
			gen := row.F[4].V
			grad := row.F[5].V == "true"
			p1 := row.F[6].V
			p2 := row.F[7].V
			imgType := row.F[8].V
			imageURL := row.F[9].V

			memberMap[name] = append(memberMap[name], memberRowData{
				id:       mid,
				name:     name,
				gen:      gen,
				grad:     grad,
				p1:       p1,
				p2:       p2,
				imgType:  imgType,
				imageURL: imageURL,
			})
		}

		var memberNames []string
		for name := range memberMap {
			memberNames = append(memberNames, name)
		}
		sort.Strings(memberNames)

		sqlStatements = append(sqlStatements, fmt.Sprintf("\n-- Members for %s", g.Name))
		for _, name := range memberNames {
			rows := memberMap[name]
			first := rows[0]

			// Pick preferred image
			bestRow := first
			typePreference := []string{"13thSingle", "5thMember", "2ndAlbum", "5thHinatansai", "4thTowel"}
			for _, pref := range typePreference {
				found := false
				for _, r := range rows {
					if r.imgType == pref {
						bestRow = r
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			memUUIDKey := fmt.Sprintf("%s:%s", g.Slug, name)
			memID := fmt.Sprintf("mem_%s", strings.ReplaceAll(uuid.NewSHA1(nsMember, []byte(memUUIDKey)).String(), "-", ""))

			reading, ok := memberReadings[name]
			if !ok {
				log.Fatalf("missing reading for member: %s", name)
			}

			genInt := parseGeneration(first.gen)
			status := "active"
			if first.grad {
				status = "graduated"
			}

			leftColID, ok := colorIDMap[first.p1]
			if !ok {
				log.Fatalf("missing color %s for member %s", first.p1, name)
			}
			rightColID, ok := colorIDMap[first.p2]
			if !ok {
				log.Fatalf("missing color %s for member %s", first.p2, name)
			}

			imageKey := fmt.Sprintf("%s.webp", memID)

			imageSources = append(imageSources, ImageSource{
				MemberID: memID,
				Name:     name,
				Group:    g.Slug,
				URL:      bestRow.imageURL,
				Type:     bestRow.imgType,
				DestKey:  imageKey,
			})

			sqlStatements = append(sqlStatements, fmt.Sprintf(
				"INSERT OR IGNORE INTO members (id, group_id, family_name, given_name, family_name_kana, given_name_kana, generation, status, left_color_id, right_color_id, ordered, image_key, created_at, updated_at)\nVALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d, '%s', '%s', '%s', 0, '%s', '%s', '%s');",
				memID, groupID, reading.FamilyName, reading.GivenName, reading.FamilyNameKana, reading.GivenNameKana, genInt, status, leftColID, rightColID, imageKey, nowUTC, nowUTC,
			))
		}
	}

	// Write SQL seed
	if err := os.MkdirAll("seeds", 0o755); err != nil {
		log.Fatalf("failed to create seeds dir: %v", err)
	}
	seedSQLPath := filepath.Join("seeds", "seed.sql")
	if err := os.WriteFile(seedSQLPath, []byte(strings.Join(sqlStatements, "\n")+"\n"), 0o644); err != nil {
		log.Fatalf("failed to write seed sql: %v", err)
	}
	fmt.Printf("Successfully generated %s (%d lines)\n", seedSQLPath, len(sqlStatements))

	// Write image sources
	imgSrcJSON, err := json.MarshalIndent(imageSources, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal image sources: %v", err)
	}
	imgSrcPath := filepath.Join("data", "image_sources.json")
	if err := os.WriteFile(imgSrcPath, imgSrcJSON, 0o644); err != nil {
		log.Fatalf("failed to write image sources json: %v", err)
	}
	fmt.Printf("Successfully generated %s (%d items)\n", imgSrcPath, len(imageSources))
}
