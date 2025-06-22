package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
		category = pflag.StringP("category", "c", "", "カテゴリでフィルタ")
		tag      = pflag.StringP("tag", "t", "", "タグでフィルタ")
		user     = pflag.StringP("user", "u", "", "作成者でフィルタ")
		query    = pflag.StringP("query", "q", "", "検索ワードでフィルタ")
		limit    = pflag.IntP("limit", "l", 10, "取得件数制限")
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

	// 検索条件の表示
	fmt.Println("🔍 記事を検索中...")
	if *category != "" {
		fmt.Printf("   カテゴリ: %s\n", *category)
	}
	if *tag != "" {
		fmt.Printf("   タグ: %s\n", *tag)
	}
	if *user != "" {
		fmt.Printf("   作成者: %s\n", *user)
	}
	if *query != "" {
		fmt.Printf("   検索ワード: %s\n", *query)
	}
	fmt.Printf("   制限: %d件\n", *limit)
	fmt.Println()

	// 記事一覧の取得
	options := &api.ListPostsOptions{
		Limit:    *limit,
		Category: *category,
		Tag:      *tag,
		User:     *user,
		Query:    *query,
	}

	posts, err := client.ListPosts(context.Background(), options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "記事一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("📭 条件に一致する記事が見つかりませんでした。")
		return
	}

	// ダウンロード対象の表示
	fmt.Printf("📥 記事のダウンロードを開始します...\n")
	fmt.Printf("   対象記事数: %d件\n\n", len(posts))

	// 記事のダウンロード
	successCount := 0
	for _, post := range posts {
		fmt.Printf("📥 ダウンロード中: [%d] %s\n", post.Number, post.Name)

		// 記事の詳細取得
		detail, err := client.FetchPost(context.Background(), post.Number)
		if err != nil {
			fmt.Printf("   ❌ エラー: %v\n", err)
			continue
		}

		// Front Matterの作成
		fm := types.FrontMatter{
			Title:           detail.Name,
			Category:        detail.Category,
			Tags:            detail.Tags,
			Wip:             detail.Wip,
			RemoteUpdatedAt: detail.UpdatedAt.Format(time.RFC3339),
		}

		// Markdownコンテンツの生成
		content, err := markdown.GenerateContent(fm, detail.BodyMd)
		if err != nil {
			fmt.Printf("   ❌ ファイル内容の生成に失敗しました: %v\n", err)
			continue
		}

		// ファイル名の生成
		filename := fmt.Sprintf("%d-%s.md", post.Number, post.Name)

		// ファイルの保存
		if err := os.WriteFile(filename, content, 0644); err != nil {
			fmt.Printf("   ❌ 保存エラー: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 保存完了: %s\n", filename)
		successCount++
	}

	// 結果の表示
	fmt.Println()
	fmt.Printf("✅ ダウンロード完了 (%d件):\n", successCount)
	if successCount > 0 {
		// macOSの場合は通知を表示
		if err := mac.SendNotification("esa-cli", fmt.Sprintf("%d件の記事をダウンロードしました", successCount)); err != nil {
			// 通知エラーは無視
		}
	}
}
