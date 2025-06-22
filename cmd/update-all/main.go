package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/config"
	"github.com/shellme/esa-cli/internal/mac"
	"github.com/shellme/esa-cli/internal/markdown"
	"github.com/shellme/esa-cli/pkg/types"
	"github.com/spf13/pflag"
)

func main() {
	// フラグの定義
	var (
		pattern    = pflag.StringP("pattern", "p", "*.md", "ファイルパターン（例: 123-*.md）")
		message    = pflag.StringP("message", "m", "", "更新メッセージ")
		noWip      = pflag.BoolP("no-wip", "n", false, "WIP状態を解除")
		category   = pflag.StringP("category", "c", "", "カテゴリを変更")
		addTags    = pflag.StringP("add-tags", "a", "", "タグを追加（カンマ区切り）")
		removeTags = pflag.StringP("remove-tags", "r", "", "タグを削除（カンマ区切り）")
		force      = pflag.BoolP("force", "f", false, "確認なしで実行")
	)
	pflag.Parse()

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	if cfg.AccessToken == "" || cfg.TeamName == "" {
		fmt.Println("❌ 設定が完了していません")
		fmt.Println("💡 'esa-cli setup' で初期設定を行ってください")
		os.Exit(1)
	}

	// APIクライアントの作成
	client := api.NewClient(cfg.TeamName, cfg.AccessToken, http.DefaultClient)

	// ファイルパターンの処理
	patternStr := *pattern
	if len(pflag.Args()) > 0 {
		patternStr = pflag.Args()[0]
	}

	// ファイルの検索
	fmt.Printf("🔍 ファイルを検索中...\n")
	fmt.Printf("   パターン: %s\n\n", patternStr)

	files, err := findMarkdownFiles(patternStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ファイルの検索に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("📭 条件に一致するファイルが見つかりませんでした。")
		return
	}

	// 更新対象の表示
	fmt.Printf("📝 記事の更新を開始します...\n")
	fmt.Printf("   対象ファイル数: %d件\n", len(files))
	for _, file := range files {
		fmt.Printf("   - %s\n", file)
	}
	fmt.Println()

	// 確認プロンプト
	if !*force {
		fmt.Print("上記のファイルを更新しますか？ (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			fmt.Println("🚫 更新を中止しました。")
			return
		}
		fmt.Println()
	}

	// 記事の更新
	successCount := 0
	for _, filename := range files {
		fmt.Printf("📝 更新中: %s\n", filename)

		if err := updateArticle(client, filename, *message, *noWip, *category, *addTags, *removeTags); err != nil {
			fmt.Printf("   ❌ エラー: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 更新完了: %s\n", filename)
		successCount++
	}

	// 結果の表示
	fmt.Println()
	fmt.Printf("✅ 更新完了 (%d件):\n", successCount)
	if successCount > 0 {
		// macOSの場合は通知を表示
		if err := mac.SendNotification("esa-cli", fmt.Sprintf("%d件の記事を更新しました", successCount)); err != nil {
			// 通知エラーは無視
		}
	}
}

// Markdownファイルを検索
func findMarkdownFiles(pattern string) ([]string, error) {
	var files []string

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// ファイルを検索
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if info.IsDir() {
			return nil
		}

		// ファイル名を取得
		filename := filepath.Base(path)

		// パターンマッチング
		if pattern == "*.md" {
			// デフォルトパターン: .mdファイルで記事番号-タイトル.mdの形式
			if strings.HasSuffix(filename, ".md") && isValidArticleFilename(filename) {
				files = append(files, filename)
			}
		} else {
			// カスタムパターン
			matched, err := filepath.Match(pattern, filename)
			if err != nil {
				return err
			}
			if matched && strings.HasSuffix(filename, ".md") && isValidArticleFilename(filename) {
				files = append(files, filename)
			}
		}

		return nil
	})

	return files, err
}

// 記事ファイル名の形式をチェック
func isValidArticleFilename(filename string) bool {
	// 記事番号-タイトル.mdの形式をチェック
	re := regexp.MustCompile(`^\d+-.+\.md$`)
	return re.MatchString(filename)
}

// 記事を更新
func updateArticle(client *api.Client, filename, message string, noWip bool, category, addTags, removeTags string) error {
	// ファイル名から記事番号を取得
	postNumberStr := strings.Split(filename, "-")[0]
	postNumber, err := strconv.Atoi(postNumberStr)
	if err != nil {
		return fmt.Errorf("無効なファイル名です: %s", filename)
	}

	// ファイルを読み込む
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗: %v", err)
	}

	// Markdownコンテンツを解析
	fm, body, err := markdown.ParseContent(content)
	if err != nil {
		return fmt.Errorf("ファイルの解析に失敗: %v", err)
	}

	// リモートの更新日時をチェック
	if fm.RemoteUpdatedAt != "" {
		remotePost, err := client.FetchPost(context.Background(), postNumber)
		if err != nil {
			// 記事が存在しない場合はチェックをスキップ
			if !strings.Contains(err.Error(), "404") {
				fmt.Printf("   ⚠️  リモート記事の取得に失敗しました: %v\n", err)
			}
		} else {
			localUpdatedAt, _ := time.Parse(time.RFC3339, fm.RemoteUpdatedAt)
			if remotePost.UpdatedAt.After(localUpdatedAt) {
				fmt.Printf("   ⚠️  警告: リモートの記事はローカルで編集を始めてから更新されています。\n")
				fmt.Printf("      リモート: %s\n", remotePost.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
				fmt.Printf("      ローカル: %s\n", localUpdatedAt.Local().Format("2006-01-02 15:04:05"))
				fmt.Print("      このまま上書きしますか？ (y/N): ")

				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" {
					return fmt.Errorf("更新を中止しました")
				}
			}
		}
	}

	// 更新リクエストの作成
	updateReq := types.UpdatePostBody{
		Name:    fm.Title,
		BodyMd:  body,
		Message: message,
		Wip:     fm.Wip,
	}

	// カテゴリの設定
	if category != "" {
		updateReq.Category = category
	} else {
		updateReq.Category = fm.Category
	}

	// タグの設定
	tags := fm.Tags
	if addTags != "" {
		tags = append(tags, strings.Split(addTags, ",")...)
	}
	if removeTags != "" {
		removeTagList := strings.Split(removeTags, ",")
		for _, removeTag := range removeTagList {
			for i, tag := range tags {
				if tag == removeTag {
					tags = append(tags[:i], tags[i+1:]...)
					break
				}
			}
		}
	}
	updateReq.Tags = tags

	// WIP状態の設定
	if noWip {
		updateReq.Wip = false
	}

	// 記事の更新
	updatedPost, err := client.UpdatePost(context.Background(), postNumber, updateReq)
	if err != nil {
		return fmt.Errorf("記事の更新に失敗: %v", err)
	}

	// ローカルファイルを更新後の内容で書き換える
	newFm := types.FrontMatter{
		Title:           updatedPost.Name,
		Category:        updatedPost.Category,
		Tags:            updatedPost.Tags,
		Wip:             updatedPost.Wip,
		RemoteUpdatedAt: updatedPost.UpdatedAt.Format(time.RFC3339),
	}
	newContent, err := markdown.GenerateContent(newFm, updatedPost.BodyMd)
	if err != nil {
		return fmt.Errorf("ローカルファイルの更新に失敗: %v", err)
	}

	if err := os.WriteFile(filename, newContent, 0644); err != nil {
		return fmt.Errorf("ローカルファイルの書き込みに失敗: %v", err)
	}

	return nil
}
