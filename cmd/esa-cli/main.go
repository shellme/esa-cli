package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/config"
	"github.com/shellme/esa-cli/internal/markdown"
	"github.com/shellme/esa-cli/pkg/types"
)

var (
	version = "dev" // ビルド時に上書き

	// APIクライアント生成用の関数変数（テスト時に差し替え可能）
	newAPIClient = func(team, token string) *api.Client {
		return api.NewClient(team, token, http.DefaultClient)
	}
)

func main() {
	// バージョン表示
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("esa-cli version %s\n", version)
		return
	}

	// コマンドライン引数の解析
	setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)

	// listコマンドのオプション
	var category string
	var tag string
	var query string
	listCmd.StringVar(&category, "category", "", "カテゴリでフィルタリング")
	listCmd.StringVar(&tag, "tag", "", "タグでフィルタリング")
	listCmd.StringVar(&query, "query", "", "検索ワードでフィルタリング")

	// fetchコマンドのオプション
	var fetchCategory string
	var fetchTag string
	var fetchQuery string
	var fetchLatest bool
	fetchCmd.StringVar(&fetchCategory, "category", "", "カテゴリでフィルタリング")
	fetchCmd.StringVar(&fetchTag, "tag", "", "タグでフィルタリング")
	fetchCmd.StringVar(&fetchQuery, "query", "", "検索ワードでフィルタリング")
	fetchCmd.BoolVar(&fetchLatest, "latest", false, "最新の記事をダウンロード")

	// updateコマンドのオプション
	var noWip bool
	var updateCategory string
	var addTags string
	var removeTags string
	var message string
	updateCmd.BoolVar(&noWip, "no-wip", false, "WIP状態を解除")
	updateCmd.StringVar(&updateCategory, "category", "", "カテゴリを変更")
	updateCmd.StringVar(&addTags, "add-tags", "", "タグを追加（カンマ区切り）")
	updateCmd.StringVar(&removeTags, "remove-tags", "", "タグを削除（カンマ区切り）")
	updateCmd.StringVar(&message, "message", "", "更新メッセージ")

	// 引数が指定されていない場合はヘルプを表示
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	// コマンドの実行
	switch os.Args[1] {
	case "setup":
		setupCmd.Parse(os.Args[2:])
		runSetup()
	case "list":
		listCmd.Parse(os.Args[2:])
		runList(listCmd, category, tag, query)
	case "fetch":
		fetchCmd.Parse(os.Args[2:])
		runFetch(fetchCmd, fetchCategory, fetchTag, fetchQuery, fetchLatest)
	case "update":
		updateCmd.Parse(os.Args[2:])
		runUpdate(updateCmd, noWip, updateCategory, addTags, removeTags, message)
	case "help":
		showHelp()
	default:
		fmt.Printf("不明なコマンド: %s\n", os.Args[1])
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf("esa-cli %s - esaの記事をローカルで編集するCLIツール\n\n", version)
	fmt.Println("使用方法:")
	fmt.Println("  esa-cli setup                 初期設定")
	fmt.Println("  esa-cli list [件数]            記事一覧を表示（デフォルト10件）")
	fmt.Println("    オプション:")
	fmt.Println("      --category <カテゴリ>      カテゴリでフィルタリング")
	fmt.Println("      --tag <タグ>              タグでフィルタリング")
	fmt.Println("      --query <検索ワード>       検索ワードでフィルタリング")
	fmt.Println("  esa-cli fetch <記事番号>       記事をダウンロード")
	fmt.Println("    オプション:")
	fmt.Println("      --category <カテゴリ>      カテゴリでフィルタリング")
	fmt.Println("      --tag <タグ>              タグでフィルタリング")
	fmt.Println("      --query <検索ワード>       検索ワードでフィルタリング")
	fmt.Println("      --latest                  最新の記事をダウンロード")
	fmt.Println("  esa-cli update <ファイル名>    記事を更新")
	fmt.Println("    オプション:")
	fmt.Println("      --no-wip                   WIP状態を解除")
	fmt.Println("      --category <カテゴリ>      カテゴリを変更")
	fmt.Println("      --add-tags <タグ>          タグを追加（カンマ区切り）")
	fmt.Println("      --remove-tags <タグ>       タグを削除（カンマ区切り）")
	fmt.Println("      --message <メッセージ>      更新メッセージ")
	fmt.Println("  esa-cli version                バージョン表示")
	fmt.Println("  esa-cli help                   このヘルプを表示")
	fmt.Println("")
	fmt.Println("例:")
	fmt.Println("  esa-cli setup                  # 初回設定")
	fmt.Println("  esa-cli list                   # 最新10件の記事一覧")
	fmt.Println("  esa-cli list --category 開発    # 開発カテゴリの記事一覧")
	fmt.Println("  esa-cli list --tag API         # APIタグの記事一覧")
	fmt.Println("  esa-cli list --query 認証      # 認証を含む記事一覧")
	fmt.Println("  esa-cli fetch 123              # 記事123をダウンロード")
	fmt.Println("  esa-cli fetch --category 開発 --latest  # 開発カテゴリの最新記事をダウンロード")
	fmt.Println("  esa-cli fetch --tag API --latest       # APIタグの最新記事をダウンロード")
	fmt.Println("  esa-cli fetch --query 認証 --latest    # 認証を含む最新記事をダウンロード")
	fmt.Println("  esa-cli update 123-title.md    # 記事を更新")
	fmt.Println("  esa-cli update 123-title.md --no-wip  # WIP状態を解除して更新")
	fmt.Println("  esa-cli update 123-title.md --category 開発  # カテゴリを変更して更新")
	fmt.Println("  esa-cli update 123-title.md --add-tags API,認証  # タグを追加して更新")
	fmt.Println("  esa-cli update 123-title.md --message API仕様を更新  # メッセージを付けて更新")
	fmt.Println("")
	fmt.Println("💡 初回利用時は 'esa-cli setup' で設定を行ってください")
}

func runSetup() {
	// 一時的なクライアントを作成
	client := api.NewClient("", "", http.DefaultClient)
	if err := config.Setup(client); err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}
}

func runList(cmd *flag.FlagSet, category, tag, query string) {
	options := &api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
	}
	if len(cmd.Args()) > 0 {
		if l, err := strconv.Atoi(cmd.Args()[0]); err == nil && l > 0 {
			options.Limit = l
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 設定の読み込みに失敗しました: %v\n", err)
		fmt.Println("💡 'esa-cli setup' で初期設定を行ってください")
		os.Exit(1)
	}

	if cfg.AccessToken == "" || cfg.TeamName == "" {
		fmt.Println("❌ 設定が完了していません")
		fmt.Println("💡 'esa-cli setup' で初期設定を行ってください")
		os.Exit(1)
	}

	client := newAPIClient(cfg.TeamName, cfg.AccessToken)

	// 記事一覧を表示
	posts, err := client.ListPosts(context.Background(), options)
	if err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}

	// 記事一覧を表示
	for _, post := range posts {
		fmt.Printf("%d: %s\n", post.Number, post.FullName)
	}
}

func runFetch(cmd *flag.FlagSet, category, tag, query string, latest bool) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 設定の読み込みに失敗しました: %v\n", err)
		fmt.Println("💡 'esa-cli setup' で初期設定を行ってください")
		os.Exit(1)
	}

	if cfg.AccessToken == "" || cfg.TeamName == "" {
		fmt.Println("❌ 設定が完了していません")
		fmt.Println("💡 'esa-cli setup' で初期設定を行ってください")
		os.Exit(1)
	}

	client := newAPIClient(cfg.TeamName, cfg.AccessToken)

	if latest {
		// 最新の記事を取得
		options := &api.ListPostsOptions{
			Category: category,
			Tag:      tag,
			Query:    query,
			Limit:    1,
		}
		posts, err := client.ListPosts(context.Background(), options)
		if err != nil {
			fmt.Printf("❌ エラー: %v\n", err)
			os.Exit(1)
		}
		if len(posts) == 0 {
			fmt.Println("❌ 条件に一致する記事が見つかりません")
			os.Exit(1)
		}
		post := posts[0]
		// 最新記事の番号で後続の処理を行う
		fetchArticle(client, post.Number)
		return
	}

	// 記事番号が指定されていない場合
	if len(cmd.Args()) < 1 {
		fmt.Println("❌ 記事番号を指定してください")
		fmt.Println("💡 使用例: esa-cli fetch 123")
		os.Exit(1)
	}

	postNumber, err := strconv.Atoi(cmd.Args()[0])
	if err != nil {
		fmt.Printf("❌ 無効な記事番号です: %s\n", cmd.Args()[0])
		os.Exit(1)
	}

	fetchArticle(client, postNumber)
}

// 記事を取得してファイルに書き込む共通関数
func fetchArticle(client *api.Client, postNumber int) {
	// 記事を取得
	post, err := client.FetchPost(context.Background(), postNumber)
	if err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}

	fm := types.FrontMatter{
		Title:           post.Name,
		Category:        post.Category,
		Tags:            post.Tags,
		Wip:             post.Wip,
		RemoteUpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}

	content, err := markdown.GenerateContent(fm, post.BodyMd)
	if err != nil {
		fmt.Printf("❌ ファイル内容の生成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fileName := fmt.Sprintf("%d-%s.md", post.Number, post.Name)
	if err := os.WriteFile(fileName, content, 0644); err != nil {
		fmt.Printf("❌ ファイルの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 記事をダウンロードしました: %s\n", fileName)
}

func runUpdate(cmd *flag.FlagSet, noWip bool, category, addTags, removeTags, message string) {
	if len(cmd.Args()) < 1 {
		fmt.Println("❌ ファイル名を指定してください")
		fmt.Println("💡 使用例: esa-cli update 123-title.md")
		os.Exit(1)
	}
	fileName := cmd.Args()[0]

	// ファイル名から記事番号を取得
	postNumberStr := strings.Split(fileName, "-")[0]
	postNumber, err := strconv.Atoi(postNumberStr)
	if err != nil {
		fmt.Printf("❌ 無効なファイル名です。'記事番号-タイトル.md'の形式である必要があります: %s\n", fileName)
		os.Exit(1)
	}

	// ファイルを読み込む
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("❌ ファイルの読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fm, body, err := markdown.ParseContent(content)
	if err != nil {
		fmt.Printf("❌ ファイルの解析に失敗しました: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}
	client := newAPIClient(cfg.TeamName, cfg.AccessToken)

	// リモートの更新日時をチェック
	if fm.RemoteUpdatedAt != "" {
		remotePost, err := client.FetchPost(context.Background(), postNumber)
		if err != nil {
			// 記事が存在しない場合はチェックをスキップ
			if !strings.Contains(err.Error(), "404") {
				fmt.Printf("⚠️  リモート記事の取得に失敗しました: %v\n", err)
			}
		} else {
			localUpdatedAt, _ := time.Parse(time.RFC3339, fm.RemoteUpdatedAt)
			if remotePost.UpdatedAt.After(localUpdatedAt) {
				fmt.Println("⚠️  警告: リモートの記事はローカルで編集を始めてから更新されています。")
				fmt.Printf("  リモート: %s\n", remotePost.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
				fmt.Printf("  ローカル: %s\n", localUpdatedAt.Local().Format("2006-01-02 15:04:05"))
				fmt.Print("このまま上書きしますか？ (y/N): ")

				var confirm string
				fmt.Scanln(&confirm)
				if strings.ToLower(confirm) != "y" {
					fmt.Println("🚫 更新を中止しました。")
					os.Exit(0)
				}
			}
		}
	}

	updateReq := types.UpdatePostBody{
		Name:    fm.Title,
		BodyMd:  body,
		Message: message,
		Wip:     fm.Wip,
	}
	if category != "" {
		updateReq.Category = category
	} else {
		updateReq.Category = fm.Category
	}
	if addTags != "" {
		updateReq.Tags = append(fm.Tags, strings.Split(addTags, ",")...)
	} else {
		updateReq.Tags = fm.Tags
	}
	// TODO: removeTagsの処理を追加する

	updatedPost, err := client.UpdatePost(context.Background(), postNumber, updateReq)
	if err != nil {
		fmt.Printf("❌ 記事の更新に失敗しました: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("❌ ローカルファイルの更新に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(fileName, newContent, 0644); err != nil {
		fmt.Printf("❌ ローカルファイルの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 記事を更新しました: %s\n", fileName)
}
