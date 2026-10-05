package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/quiz"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
)

func main() {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=======================================================")
	fmt.Println("  🌸 坂道ペンライトクイズ v2 (CLI 自由回答実戦モード) ☀️")
	fmt.Println("=======================================================")

	// 1. Initialize DB and Repository
	dbPath := filepath.Join("data", "penlight.db")
	_ = os.MkdirAll("data", 0755)

	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "データベース接続エラー: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Ensure tables and seeds are present
	if err := ensureDatabase(repo); err != nil {
		fmt.Fprintf(os.Stderr, "初期化エラー: %v\n", err)
		os.Exit(1)
	}

	// 2. Fetch master data
	groups, err := repo.ListGroups(ctx)
	if err != nil || len(groups) == 0 {
		fmt.Fprintf(os.Stderr, "グループ取得エラー: %v\n", err)
		os.Exit(1)
	}

	allColors, err := repo.ListColors(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "カラー取得エラー: %v\n", err)
		os.Exit(1)
	}
	colorMap := make(map[model.ID]model.Color)
	for _, c := range allColors {
		colorMap[c.ID] = c
	}

	allMembers, err := repo.ListMembers(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "メンバー取得エラー: %v\n", err)
		os.Exit(1)
	}

	// 3. Select Group
	fmt.Println("\n出題グループを選択してください:")
	for i, g := range groups {
		fmt.Printf("  [%d] %s\n", i+1, g.Name)
	}
	fmt.Printf("  [%d] 全グループ合同\n", len(groups)+1)
	fmt.Print("\n選択 (番号を入力): ")

	inputStr, _ := reader.ReadString('\n')
	choice, _ := strconv.Atoi(strings.TrimSpace(inputStr))

	var selectedGroupID *model.ID
	groupName := "全グループ合同"
	if choice >= 1 && choice <= len(groups) {
		selectedGroupID = &groups[choice-1].ID
		groupName = groups[choice-1].Name
	}

	// 4. Filter members using pkg/quiz/filter.go
	filter := model.QuizFilter{
		GroupID:          selectedGroupID,
		IncludeGraduated: false,
	}
	pool, err := quiz.FilterMembers(allMembers, filter)
	if err != nil {
		fmt.Printf("候補者抽出エラー: %v\n", err)
		return
	}

	// Filter colors for display
	var displayColors []model.Color
	for _, c := range allColors {
		if selectedGroupID == nil || c.GroupID == nil || *c.GroupID == *selectedGroupID {
			displayColors = append(displayColors, c)
		}
	}

	// 5. Build blended deck using pkg/quiz/deck.go
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	deckSize := 10
	deck := quiz.BuildBlendedDeck(pool, nil, deckSize, rng)

	fmt.Printf("\n✨ 【%s】の出題デッキ (全 %d 問) を生成しました！スタート！\n", groupName, len(deck))
	time.Sleep(800 * time.Millisecond)

	// 6. Play Session
	score := 0
	startTime := time.Now()

	for i, target := range deck {
		fmt.Println("\n-------------------------------------------------------")
		fmt.Printf("【第 %d 問 / 全 %d 問】\n", i+1, len(deck))

		img := quiz.SelectQuestionImage(target, filter, rng)
		costumeTitle := "通常"
		if img != nil && img.PhotoType != nil {
			costumeTitle = img.PhotoType.Name
		}

		fmt.Printf("👤 メンバー : \033[1;33m%s %s\033[0m (%d期生)\n", target.FamilyName, target.GivenName, target.Generation)
		fmt.Printf("👘 写真衣装 : %s\n\n", costumeTitle)

		fmt.Println("🎨 カラーパレット:")
		printColorPalette(displayColors)

		fmt.Println("\n左手色と右手色の番号をスペース区切りで入力してください (例: 1 3):")
		fmt.Println("※両手同色 (白×白など) は同じ番号を2回入力 (例: 2 2)")
		fmt.Print("> ")

		ansLine, _ := reader.ReadString('\n')
		parts := strings.Fields(ansLine)

		if len(parts) < 2 {
			fmt.Println("⚠️  2つの番号を入力してください (スキップされました)")
			continue
		}

		idx1, _ := strconv.Atoi(parts[0])
		idx2, _ := strconv.Atoi(parts[1])

		if idx1 < 1 || idx1 > len(displayColors) || idx2 < 1 || idx2 > len(displayColors) {
			fmt.Println("⚠️  無効な番号です")
			continue
		}

		c1 := displayColors[idx1-1]
		c2 := displayColors[idx2-1]

		// Judge answer using pkg/quiz/judge.go (tolerant of hand orientation)
		isCorrect := quiz.JudgeAnswerPair(c1.ID, c2.ID, target)

		correctL := colorMap[target.Penlight.LeftColorID]
		correctR := colorMap[target.Penlight.RightColorID]

		if isCorrect {
			score++
			fmt.Printf("\n🎉 \033[1;32m大正解！\033[0m [%s × %s]\n",
				formatColorBlock(c1), formatColorBlock(c2))
			fmt.Println("   (左右持ち替え順不同 OK!)")
		} else {
			fmt.Printf("\n😢 \033[1;31m不正解...\033[0m\n")
			fmt.Printf("   あなたの回答 : %s × %s\n",
				formatColorBlock(c1), formatColorBlock(c2))
			fmt.Printf("   正解のカラー : %s × %s\n",
				formatColorBlock(correctL), formatColorBlock(correctR))
		}

		time.Sleep(600 * time.Millisecond)
	}

	duration := time.Since(startTime).Truncate(time.Second)

	// 7. Results
	fmt.Println("\n=======================================================")
	fmt.Println("                   🏆 結 果 発 表 🏆                  ")
	fmt.Println("=======================================================")
	fmt.Printf("  スコア    : %d / %d 点 (正答率 %.1f%%)\n",
		score, len(deck), float64(score)/float64(len(deck))*100)
	fmt.Printf("  所要時間  : %v\n", duration)
	fmt.Println("=======================================================")
}

func printColorPalette(colors []model.Color) {
	for i, c := range colors {
		block := formatColorBlock(c)
		fmt.Printf(" [%2d] %s  ", i+1, block)
		if (i+1)%3 == 0 {
			fmt.Println()
		}
	}
	if len(colors)%3 != 0 {
		fmt.Println()
	}
}

func formatColorBlock(c model.Color) string {
	r, g, b := hexToRGB(c.HexCode)
	// 24-bit ANSI color block + color name
	return fmt.Sprintf("\033[48;2;%d;%d;%dm  \033[0m \033[38;2;%d;%d;%dm%s\033[0m",
		r, g, b, r, g, b, c.Name)
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 255, 255, 255
	}
	return int((v >> 16) & 0xFF), int((v >> 8) & 0xFF), int(v & 0xFF)
}

func ensureDatabase(repo *repository.SQLiteRepository) error {
	ctx := context.Background()

	// Check if groups already exist
	groups, err := repo.ListGroups(ctx)
	if err == nil && len(groups) > 0 {
		return nil
	}

	// Apply migration
	migBytes, err := os.ReadFile("migrations/000001_init.up.sql")
	if err != nil {
		return fmt.Errorf("migration read error: %w", err)
	}
	if _, err := repo.DB().ExecContext(ctx, string(migBytes)); err != nil {
		return fmt.Errorf("migration exec error: %w", err)
	}

	// Apply seed
	seedBytes, err := os.ReadFile("seeds/seed.sql")
	if err != nil {
		return fmt.Errorf("seed read error: %w", err)
	}
	if _, err := repo.DB().ExecContext(ctx, string(seedBytes)); err != nil {
		return fmt.Errorf("seed exec error: %w", err)
	}

	return nil
}
