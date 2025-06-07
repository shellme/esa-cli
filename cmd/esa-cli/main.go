package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/config"
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
		runList(category, tag, query)
	case "fetch":
		fetchCmd.Parse(os.Args[2:])
		runFetch(fetchCategory, fetchTag, fetchQuery, fetchLatest)
	case "update":
		updateCmd.Parse(os.Args[2:])
		runUpdate(noWip, updateCategory, addTags, removeTags, message)
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

func runList(category, tag, query string) {
	options := &api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
	}
	if len(os.Args) > 2 {
		if l, err := strconv.Atoi(os.Args[2]); err == nil && l > 0 {
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

func runFetch(category, tag, query string, latest bool) {
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
		fmt.Printf("タイトル: %s\n", post.FullName)
		fmt.Printf("本文:\n%s\n", post.BodyMD)
		return
	}

	// 記事番号が指定されていない場合
	if len(os.Args) < 3 {
		fmt.Println("❌ 記事番号を指定してください")
		fmt.Println("💡 使用例: esa-cli fetch 123")
		os.Exit(1)
	}

	postNumber, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("❌ 無効な記事番号です: %s\n", os.Args[2])
		os.Exit(1)
	}

	// 記事を取得
	post, err := client.FetchPost(context.Background(), postNumber)
	if err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}

	// 記事の内容を表示
	fmt.Printf("タイトル: %s\n", post.FullName)
	fmt.Printf("本文:\n%s\n", post.BodyMD)
}

func runUpdate(noWip bool, category, addTags, removeTags, message string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ ファイル名を指定してください")
		fmt.Println("💡 使用例: esa-cli update 123-article-title.md")
		os.Exit(1)
	}

	filename := os.Args[2]
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

	// 更新オプションを設定
	options := &api.UpdatePostOptions{
		NoWip:      noWip,
		Category:   category,
		AddTags:    strings.Split(addTags, ","),
		RemoveTags: strings.Split(removeTags, ","),
		Message:    message,
	}

	if err := client.UpdatePost(filename, options); err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 記事を更新しました")
}
