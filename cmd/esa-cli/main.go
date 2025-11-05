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
	"github.com/shellme/esa-cli/internal/markdown"
	"github.com/shellme/esa-cli/pkg/types"
	"github.com/spf13/pflag"
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
	setupCmd := pflag.NewFlagSet("setup", pflag.ExitOnError)
	listCmd := pflag.NewFlagSet("list", pflag.ExitOnError)
	fetchCmd := pflag.NewFlagSet("fetch", pflag.ExitOnError)
	updateCmd := pflag.NewFlagSet("update", pflag.ExitOnError)
	fetchAllCmd := pflag.NewFlagSet("fetch-all", pflag.ExitOnError)
	updateAllCmd := pflag.NewFlagSet("update-all", pflag.ExitOnError)

	// listコマンドのオプション
	var category string
	var tag string
	var query string
	var user string
	listCmd.StringVarP(&category, "category", "c", "", "カテゴリでフィルタリング")
	listCmd.StringVarP(&tag, "tag", "t", "", "タグでフィルタリング")
	listCmd.StringVarP(&query, "query", "q", "", "検索ワードでフィルタリング")
	listCmd.StringVarP(&user, "user", "u", "", "作成者でフィルタリング")

	// fetchコマンドのオプション
	var fetchCategory string
	var fetchTag string
	var fetchQuery string
	var fetchUser string
	var fetchLatest bool
	fetchCmd.StringVarP(&fetchCategory, "category", "c", "", "カテゴリでフィルタリング")
	fetchCmd.StringVarP(&fetchTag, "tag", "t", "", "タグでフィルタリング")
	fetchCmd.StringVarP(&fetchQuery, "query", "q", "", "検索ワードでフィルタリング")
	fetchCmd.StringVarP(&fetchUser, "user", "u", "", "作成者でフィルタリング")
	fetchCmd.BoolVarP(&fetchLatest, "latest", "l", false, "最新の記事をダウンロード")

	// updateコマンドのオプション
	var noWip bool
	var updateCategory string
	var addTags string
	var removeTags string
	var message string
	updateCmd.BoolVarP(&noWip, "no-wip", "n", false, "WIP状態を解除")
	updateCmd.StringVarP(&updateCategory, "category", "c", "", "カテゴリを変更")
	updateCmd.StringVarP(&addTags, "add-tags", "a", "", "タグを追加（カンマ区切り）")
	updateCmd.StringVarP(&removeTags, "remove-tags", "r", "", "タグを削除（カンマ区切り）")
	updateCmd.StringVarP(&message, "message", "m", "", "更新メッセージ")

	// moveコマンドのオプション
	moveCmd := pflag.NewFlagSet("move", pflag.ExitOnError)
	var moveCategory string
	var moveUser string
	var moveQuery string
	var moveTag string
	var moveToCategory string
	var moveMessage string
	var moveForce bool
	moveCmd.StringVarP(&moveCategory, "category", "c", "", "移動元のカテゴリ")
	moveCmd.StringVarP(&moveUser, "user", "u", "", "作成者でフィルタリング")
	moveCmd.StringVarP(&moveQuery, "query", "q", "", "検索ワードでフィルタリング")
	moveCmd.StringVarP(&moveTag, "tag", "t", "", "タグでフィルタリング")
	moveCmd.StringVarP(&moveToCategory, "to", "o", "", "移動先のカテゴリ（必須）")
	moveCmd.StringVarP(&moveMessage, "message", "m", "", "移動メッセージ")
	moveCmd.BoolVarP(&moveForce, "force", "f", false, "確認なしで実行")

	// fetch-allコマンドのオプション
	var fetchAllCategory string
	var fetchAllTag string
	var fetchAllQuery string
	var fetchAllUser string
	var fetchAllLimit int
	fetchAllCmd.StringVarP(&fetchAllCategory, "category", "c", "", "カテゴリでフィルタ")
	fetchAllCmd.StringVarP(&fetchAllTag, "tag", "t", "", "タグでフィルタ")
	fetchAllCmd.StringVarP(&fetchAllQuery, "query", "q", "", "検索ワードでフィルタ")
	fetchAllCmd.StringVarP(&fetchAllUser, "user", "u", "", "作成者でフィルタ")
	fetchAllCmd.IntVarP(&fetchAllLimit, "limit", "l", 10, "取得件数制限")

	// update-allコマンドのオプション
	var updateAllPattern string
	var updateAllMessage string
	var updateAllNoWip bool
	var updateAllCategory string
	var updateAllAddTags string
	var updateAllRemoveTags string
	var updateAllForce bool
	updateAllCmd.StringVarP(&updateAllPattern, "pattern", "p", "*.md", "ファイルパターン（例: 123-*.md）")
	updateAllCmd.StringVarP(&updateAllMessage, "message", "m", "", "更新メッセージ")
	updateAllCmd.BoolVarP(&updateAllNoWip, "no-wip", "n", false, "WIP状態を解除")
	updateAllCmd.StringVarP(&updateAllCategory, "category", "c", "", "カテゴリを変更")
	updateAllCmd.StringVarP(&updateAllAddTags, "add-tags", "a", "", "タグを追加（カンマ区切り）")
	updateAllCmd.StringVarP(&updateAllRemoveTags, "remove-tags", "r", "", "タグを削除（カンマ区切り）")
	updateAllCmd.BoolVarP(&updateAllForce, "force", "f", false, "確認なしで実行")

	// createコマンドのオプション
	createCmd := pflag.NewFlagSet("create", pflag.ExitOnError)
	var createTitle string
	var createCategory string
	var createTags string
	var createMessage string
	var createWip bool
	var createFile string
	createCmd.StringVarP(&createTitle, "title", "t", "", "記事のタイトル")
	createCmd.StringVarP(&createCategory, "category", "c", "", "カテゴリ")
	createCmd.StringVarP(&createTags, "tags", "g", "", "タグ（カンマ区切り）")
	createCmd.StringVarP(&createMessage, "message", "m", "", "作成メッセージ")
	createCmd.BoolVarP(&createWip, "wip", "w", false, "WIP状態で作成")
	createCmd.StringVarP(&createFile, "file", "f", "", "既存のMarkdownファイルから作成")

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
		runList(listCmd, category, tag, query, user)
	case "fetch":
		fetchCmd.Parse(os.Args[2:])
		runFetch(fetchCmd, fetchCategory, fetchTag, fetchQuery, fetchUser, fetchLatest)
	case "update":
		updateCmd.Parse(os.Args[2:])
		runUpdate(updateCmd, noWip, updateCategory, addTags, removeTags, message)
	case "move":
		moveCmd.Parse(os.Args[2:])
		runMove(moveCmd, moveCategory, moveUser, moveQuery, moveTag, moveToCategory, moveMessage, moveForce)
	case "fetch-all":
		fetchAllCmd.Parse(os.Args[2:])
		runFetchAll(fetchAllCmd, fetchAllCategory, fetchAllTag, fetchAllQuery, fetchAllUser, fetchAllLimit)
	case "update-all":
		updateAllCmd.Parse(os.Args[2:])
		runUpdateAll(updateAllCmd, updateAllPattern, updateAllMessage, updateAllCategory, updateAllAddTags, updateAllRemoveTags, updateAllNoWip, updateAllForce)
	case "create":
		createCmd.Parse(os.Args[2:])
		runCreate(createCmd, createTitle, createCategory, createTags, createMessage, createWip, createFile)
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
	fmt.Println("    例: esa-cli list 20          # 最新20件を表示")
	fmt.Println("    オプション:")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリでフィルタリング")
	fmt.Println("      -t, --tag <タグ>          タグでフィルタリング")
	fmt.Println("      -q, --query <検索ワード>   検索ワードでフィルタリング")
	fmt.Println("      -u, --user <作成者>       作成者でフィルタリング")
	fmt.Println("  esa-cli fetch <記事番号>       記事をダウンロード")
	fmt.Println("    オプション:")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリでフィルタリング")
	fmt.Println("      -t, --tag <タグ>          タグでフィルタリング")
	fmt.Println("      -q, --query <検索ワード>   検索ワードでフィルタリング")
	fmt.Println("      -u, --user <作成者>       作成者でフィルタリング")
	fmt.Println("      -l, --latest              最新の記事をダウンロード")
	fmt.Println("  esa-cli update <ファイル名>    記事を更新")
	fmt.Println("    オプション:")
	fmt.Println("      -n, --no-wip              WIP状態を解除")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリを変更")
	fmt.Println("      -a, --add-tags <タグ>     タグを追加（カンマ区切り）")
	fmt.Println("      -r, --remove-tags <タグ>  タグを削除（カンマ区切り）")
	fmt.Println("      -m, --message <メッセージ> 更新メッセージ")
	fmt.Println("  esa-cli move                  記事を一括移動")
	fmt.Println("    オプション:")
	fmt.Println("      -c, --category <移動元カテゴリ> 移動元のカテゴリ")
	fmt.Println("      -u, --user <作成者>       作成者でフィルタリング")
	fmt.Println("      -q, --query <検索ワード>   検索ワードでフィルタリング")
	fmt.Println("      -t, --tag <タグ>          タグでフィルタリング")
	fmt.Println("      -o, --to <移動先カテゴリ>  移動先のカテゴリ（必須）")
	fmt.Println("      -m, --message <メッセージ> 移動メッセージ")
	fmt.Println("      -f, --force               確認なしで実行")
	fmt.Println("  esa-cli fetch-all               記事を一括ダウンロード")
	fmt.Println("    オプション:")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリでフィルタ")
	fmt.Println("      -t, --tag <タグ>          タグでフィルタ")
	fmt.Println("      -q, --query <検索ワード>   検索ワードでフィルタ")
	fmt.Println("      -u, --user <作成者>       作成者でフィルタ")
	fmt.Println("      -l, --limit <取得件数>     取得件数制限")
	fmt.Println("  esa-cli update-all               記事を一括更新")
	fmt.Println("    オプション:")
	fmt.Println("      -p, --pattern <ファイルパターン> ファイルパターン（例: 123-*.md）")
	fmt.Println("      -m, --message <更新メッセージ> 更新メッセージ")
	fmt.Println("      -n, --no-wip              WIP状態を解除")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリを変更")
	fmt.Println("      -a, --add-tags <タグ>     タグを追加（カンマ区切り）")
	fmt.Println("      -r, --remove-tags <タグ>  タグを削除（カンマ区切り）")
	fmt.Println("      -f, --force               確認なしで実行")
	fmt.Println("  esa-cli create                 新しい記事を作成")
	fmt.Println("    オプション:")
	fmt.Println("      -t, --title <記事のタイトル>  記事のタイトル")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリ")
	fmt.Println("      -g, --tags <タグ>          タグ（カンマ区切り）")
	fmt.Println("      -m, --message <作成メッセージ> 作成メッセージ")
	fmt.Println("      -w, --wip                 WIP状態で作成")
	fmt.Println("      -f, --file <既存のMarkdownファイル> 既存のMarkdownファイルから作成")
	fmt.Println("  esa-cli version                バージョン表示")
	fmt.Println("  esa-cli help                   このヘルプを表示")
	fmt.Println("")
	fmt.Println("例:")
	fmt.Println("  esa-cli setup                  # 初回設定")
	fmt.Println("  esa-cli list                   # 最新10件の記事一覧")
	fmt.Println("  esa-cli list -c 開発            # 開発カテゴリの記事一覧")
	fmt.Println("  esa-cli list -t API             # APIタグの記事一覧")
	fmt.Println("  esa-cli list -q 認証            # 認証を含む記事一覧")
	fmt.Println("  esa-cli list -u 自分のユーザー名 # 自分が作成した記事一覧")
	fmt.Println("  esa-cli fetch 123              # 記事123をダウンロード")
	fmt.Println("  esa-cli fetch -c 開発 -l        # 開発カテゴリの最新記事をダウンロード")
	fmt.Println("  esa-cli fetch -t API -l         # APIタグの最新記事をダウンロード")
	fmt.Println("  esa-cli fetch -q 認証 -l        # 認証を含む最新記事をダウンロード")
	fmt.Println("  esa-cli update 123-title.md    # 記事を更新")
	fmt.Println("  esa-cli update 123-title.md -n # WIP状態を解除して更新")
	fmt.Println("  esa-cli update 123-title.md -c 開発  # カテゴリを変更して更新")
	fmt.Println("  esa-cli update 123-title.md -a API,認証  # タグを追加して更新")
	fmt.Println("  esa-cli update 123-title.md -m API仕様を更新  # メッセージを付けて更新")
	fmt.Println("  esa-cli move -c 開発 -o デザイン -u 自分のユーザー名  # 一括移動")
	fmt.Println("  esa-cli move -c 開発 -o デザイン -u 自分のユーザー名 -f  # 確認なしで移動")
	fmt.Println("  esa-cli fetch-all -c 開発 -u 自分のユーザー名  # 開発カテゴリの自分の記事を一括ダウンロード")
	fmt.Println("  esa-cli fetch-all -t API -l 5  # APIタグの最新5件を一括ダウンロード")
	fmt.Println("  esa-cli update-all  # 現在のディレクトリの全記事を一括更新")
	fmt.Println("  esa-cli update-all \"123-*.md\" -m 一括更新  # 特定パターンの記事を一括更新")
	fmt.Println("  esa-cli create \"新機能の説明\" -c 開発 -g API,新機能  # 新しい記事を作成")
	fmt.Println("  esa-cli create \"API仕様書\" -c 開発/API -g API,ドキュメント -w  # WIP状態で記事を作成")
	fmt.Println("  esa-cli create -f draft.md -c 開発/ドキュメント  # 既存ファイルから記事を作成")
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

func runList(cmd *pflag.FlagSet, category, tag, query, user string) {
	options := &api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
		User:     user,
		Limit:    10, // デフォルト値
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

	// 検索条件の表示
	fmt.Println("🔍 記事を検索中...")
	if category != "" {
		fmt.Printf("   カテゴリ: %s\n", category)
	}
	if tag != "" {
		fmt.Printf("   タグ: %s\n", tag)
	}
	if user != "" {
		fmt.Printf("   作成者: %s\n", user)
	}
	if query != "" {
		fmt.Printf("   検索ワード: %s\n", query)
	}
	fmt.Printf("   取得件数: %d件\n", options.Limit)
	fmt.Println()

	// カテゴリでフィルタリング（クライアント側で追加フィルタリング）
	// esa.ioのAPIがカテゴリで正しくフィルタリングしていない場合の対処
	if category != "" {
		filteredPosts := []*types.Post{}
		for _, post := range posts {
			// FullNameは "カテゴリ/記事名" の形式なので、カテゴリ部分をチェック
			// 完全一致または、指定したカテゴリ配下の記事をフィルタリング
			if strings.HasPrefix(post.FullName, category+"/") || post.FullName == category {
				filteredPosts = append(filteredPosts, post)
			}
		}
		posts = filteredPosts
	}

	// 記事一覧を表示
	if len(posts) == 0 {
		fmt.Println("📭 条件に一致する記事が見つかりませんでした。")
		return
	}

	fmt.Printf("📋 記事一覧 (%d件):\n", len(posts))
	for _, post := range posts {
		fmt.Printf("  [%d] %s\n", post.Number, post.FullName)
	}
}

func runFetch(cmd *pflag.FlagSet, category, tag, query, user string, latest bool) {
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
			User:     user,
			Limit:    1,
		}

		// 検索条件の表示
		fmt.Println("🔍 記事を検索中...")
		if category != "" {
			fmt.Printf("   カテゴリ: %s\n", category)
		}
		if tag != "" {
			fmt.Printf("   タグ: %s\n", tag)
		}
		if user != "" {
			fmt.Printf("   作成者: %s\n", user)
		}
		if query != "" {
			fmt.Printf("   検索ワード: %s\n", query)
		}
		fmt.Println()

		posts, err := client.ListPosts(context.Background(), options)
		if err != nil {
			fmt.Printf("❌ エラー: %v\n", err)
			os.Exit(1)
		}

		// カテゴリでフィルタリング（クライアント側で追加フィルタリング）
		if category != "" {
			filteredPosts := []*types.Post{}
			for _, post := range posts {
				if strings.HasPrefix(post.FullName, category+"/") || post.FullName == category {
					filteredPosts = append(filteredPosts, post)
				}
			}
			posts = filteredPosts
		}

		if len(posts) == 0 {
			fmt.Println("❌ 条件に一致する記事が見つかりません")
			os.Exit(1)
		}
		post := posts[0]
		fmt.Printf("📥 最新記事をダウンロード中: [%d] %s\n", post.Number, post.FullName)
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
	fmt.Printf("📄 ファイル名: %s\n", fileName)
	fmt.Printf("📝 タイトル: %s\n", post.Name)
	if post.Category != "" {
		fmt.Printf("📁 カテゴリ: %s\n", post.Category)
	}
	if len(post.Tags) > 0 {
		fmt.Printf("🏷️  タグ: %s\n", strings.Join(post.Tags, ", "))
	}
}

func runUpdate(cmd *pflag.FlagSet, noWip bool, category, addTags, removeTags, message string) {
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

	// タグの設定
	tags := fm.Tags
	if addTags != "" {
		tags = append(tags, strings.Split(addTags, ",")...)
	}
	if removeTags != "" {
		removeTagList := strings.Split(removeTags, ",")
		for _, removeTag := range removeTagList {
			removeTag = strings.TrimSpace(removeTag)
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

func runMove(cmd *pflag.FlagSet, category, user, query, tag, toCategory, message string, force bool) {
	// 移動先カテゴリの指定をチェック
	if toCategory == "" {
		fmt.Println("❌ エラー: 移動先のカテゴリを指定してください (--to オプション)")
		fmt.Println("💡 例: esa-cli move --category 開発 --to デザイン --user 自分のユーザー名")
		os.Exit(1)
	}

	// 設定の読み込み
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

	// 移動対象の記事を検索
	options := &api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
		User:     user,
		Limit:    100, // 一度に100件まで取得
	}

	fmt.Printf("🔍 移動対象の記事を検索中...\n")
	fmt.Printf("   カテゴリ: %s\n", category)
	fmt.Printf("   作成者: %s\n", user)
	fmt.Printf("   タグ: %s\n", tag)
	fmt.Printf("   検索ワード: %s\n", query)

	posts, err := client.ListPosts(context.Background(), options)
	if err != nil {
		fmt.Printf("❌ 記事の検索に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("⚠️  移動対象の記事が見つかりませんでした")
		os.Exit(0)
	}

	// 移動対象の記事一覧を表示
	fmt.Printf("\n📋 移動対象の記事 (%d件):\n", len(posts))
	for i, post := range posts {
		fmt.Printf("  %d. [%d] %s (現在のカテゴリ: %s)\n", i+1, post.Number, post.FullName, post.Category)
	}

	// 移動先カテゴリを表示
	fmt.Printf("\n🎯 移動先カテゴリ: %s\n", toCategory)

	// 確認プロンプト（--forceが指定されていない場合）
	if !force {
		fmt.Printf("\n⚠️  上記の記事を移動しますか？ (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("❌ 移動をキャンセルしました")
			os.Exit(0)
		}
	}

	// 記事番号のリストを作成
	var postNumbers []int
	for _, post := range posts {
		postNumbers = append(postNumbers, post.Number)
	}

	// 移動メッセージの設定
	if message == "" {
		message = fmt.Sprintf("カテゴリを %s から %s に一括移動", category, toCategory)
	}

	// 一括移動の実行
	fmt.Printf("\n🚀 記事の移動を開始します...\n")
	updatedPosts, err := client.BulkUpdateCategory(context.Background(), postNumbers, toCategory, message)
	if err != nil {
		fmt.Printf("❌ 記事の移動に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 結果の表示
	fmt.Printf("\n✅ 移動が完了しました！\n")
	fmt.Printf("   移動した記事数: %d件\n", len(updatedPosts))
	fmt.Printf("   移動先カテゴリ: %s\n", toCategory)

	for _, post := range updatedPosts {
		fmt.Printf("   ✅ [%d] %s\n", post.Number, post.FullName)
	}
}

func runFetchAll(cmd *pflag.FlagSet, category, tag, query, user string, limit int) {
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

	options := &api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
		User:     user,
		Limit:    limit,
	}

	fmt.Printf("🔍 記事を検索中...\n")
	fmt.Printf("   カテゴリ: %s\n", category)
	fmt.Printf("   タグ: %s\n", tag)
	fmt.Printf("   検索ワード: %s\n", query)

	posts, err := client.ListPosts(context.Background(), options)
	if err != nil {
		fmt.Printf("❌ 記事の検索に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("⚠️  記事が見つかりませんでした")
		os.Exit(0)
	}

	// 記事一覧を表示
	fmt.Printf("\n📋 記事一覧 (%d件):\n", len(posts))
	for i, post := range posts {
		fmt.Printf("  %d. [%d] %s\n", i+1, post.Number, post.FullName)
	}

	// 記事を一括ダウンロード
	fmt.Printf("\n🚀 記事をダウンロード中...\n")
	for _, post := range posts {
		fetchArticle(client, post.Number)
	}

	fmt.Printf("\n✅ 記事のダウンロードが完了しました！\n")
}

func runUpdateAll(cmd *pflag.FlagSet, pattern, message, updateAllCategory, updateAllAddTags, updateAllRemoveTags string, updateAllNoWip, updateAllForce bool) {
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

	// ファイルパターンの処理
	patternStr := pattern
	if len(cmd.Args()) > 0 {
		patternStr = cmd.Args()[0]
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
	if !updateAllForce {
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

		if err := updateArticle(client, filename, message, updateAllNoWip, updateAllCategory, updateAllAddTags, updateAllRemoveTags); err != nil {
			fmt.Printf("   ❌ エラー: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 更新完了: %s\n", filename)
		successCount++
	}

	// 結果の表示
	fmt.Println()
	fmt.Printf("✅ 更新完了 (%d件):\n", successCount)
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

func runCreate(cmd *pflag.FlagSet, title, category, tags, message string, wip bool, file string) {
	// 設定の読み込み
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

	// 位置引数からタイトルを取得
	if len(cmd.Args()) > 0 && title == "" {
		title = cmd.Args()[0]
	}

	// 対話形式での入力（タイトルが指定されていない場合）
	if title == "" && file == "" {
		fmt.Println("📝 新しい記事を作成します")
		fmt.Print("記事のタイトル: ")
		fmt.Scanln(&title)
		if title == "" {
			fmt.Println("❌ タイトルが指定されていません")
			os.Exit(1)
		}
	}

	client := newAPIClient(cfg.TeamName, cfg.AccessToken)

	// タグの処理
	var tagList []string
	if tags != "" {
		tagList = strings.Split(tags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
	}

	// 記事作成リクエストの作成
	createBody := types.CreatePostBody{
		Name:     title,
		Category: category,
		Tags:     tagList,
		BodyMd:   "",
		Wip:      wip,
		Message:  message,
	}

	// ファイルから作成する場合
	if file != "" {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ ファイルの読み込みに失敗しました: %v\n", err)
			os.Exit(1)
		}

		// Markdownコンテンツを解析
		fm, body, err := markdown.ParseContent(content)
		if err != nil {
			fmt.Printf("❌ ファイルの解析に失敗しました: %v\n", err)
			os.Exit(1)
		}

		// ファイルの内容で上書き
		if fm.Title != "" {
			createBody.Name = fm.Title
		}
		if fm.Category != "" {
			createBody.Category = fm.Category
		}
		if len(fm.Tags) > 0 {
			createBody.Tags = fm.Tags
		}
		createBody.Wip = fm.Wip
		createBody.BodyMd = body
	}

	// 新しい記事の作成
	post, err := client.CreatePost(context.Background(), createBody)
	if err != nil {
		fmt.Printf("❌ 記事の作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 作成された記事をローカルファイルとして保存
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

	fmt.Printf("✅ 新しい記事が作成されました: %s\n", post.FullName)
	fmt.Printf("📄 ローカルファイル: %s\n", fileName)
}
