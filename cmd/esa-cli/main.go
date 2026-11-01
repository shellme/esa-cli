package main

import (
	"context"
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
	"github.com/spf13/pflag"
)

var (
	version = "dev" // ビルド時に上書き

	// APIクライアント生成用の関数変数（テスト時に差し替え可能）
	newAPIClient = func(team, token string) *api.Client {
		return api.NewClient(team, token, &http.Client{Timeout: 60 * time.Second})
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
	var fetchPrint bool
	var fetchOverwrite bool
	fetchCmd.StringVarP(&fetchCategory, "category", "c", "", "カテゴリでフィルタリング")
	fetchCmd.StringVarP(&fetchTag, "tag", "t", "", "タグでフィルタリング")
	fetchCmd.StringVarP(&fetchQuery, "query", "q", "", "検索ワードでフィルタリング")
	fetchCmd.StringVarP(&fetchUser, "user", "u", "", "作成者でフィルタリング")
	fetchCmd.BoolVarP(&fetchLatest, "latest", "l", false, "最新の記事をダウンロード")
	fetchCmd.BoolVarP(&fetchPrint, "print", "p", false, "ファイルに保存せず標準出力に表示")
	var fetchComments bool
	fetchCmd.BoolVar(&fetchComments, "comments", false, "コメントも取得")
	fetchCmd.BoolVar(&fetchOverwrite, "overwrite", false, "既存のローカルファイルを上書き（元のファイルは.esa-cli-backupに保存）")

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
	var moveDryRun bool
	moveCmd.BoolVar(&moveDryRun, "dry-run", false, "移動せず対象の記事だけを表示")

	// fetch-allコマンドのオプション
	fetchAllCmd := pflag.NewFlagSet("fetch-all", pflag.ExitOnError)
	var fetchAll fetchAllOptions
	fetchAllCmd.StringVarP(&fetchAll.category, "category", "c", "", "カテゴリでフィルタリング")
	fetchAllCmd.StringVarP(&fetchAll.tag, "tag", "t", "", "タグでフィルタリング")
	fetchAllCmd.StringVarP(&fetchAll.user, "user", "u", "", "作成者でフィルタリング")
	fetchAllCmd.StringVarP(&fetchAll.query, "query", "q", "", "検索ワードでフィルタリング")
	fetchAllCmd.IntVarP(&fetchAll.limit, "limit", "l", 100, "取得件数の上限（0で無制限）")
	fetchAllCmd.BoolVar(&fetchAll.overwrite, "overwrite", false, "既存のローカルファイルを上書き（元のファイルは.esa-cli-backupに保存）")
	fetchAllCmd.BoolVar(&fetchAll.dryRun, "dry-run", false, "保存せず対象の記事だけを表示")

	// update-allコマンドのオプション
	updateAllCmd := pflag.NewFlagSet("update-all", pflag.ExitOnError)
	var updateAll updateAllOptions
	updateAllCmd.BoolVarP(&updateAll.noWip, "no-wip", "n", false, "WIP状態を解除")
	updateAllCmd.StringVarP(&updateAll.category, "category", "c", "", "カテゴリを変更")
	updateAllCmd.StringVarP(&updateAll.addTags, "add-tags", "a", "", "タグを追加（カンマ区切り）")
	updateAllCmd.StringVarP(&updateAll.removeTags, "remove-tags", "r", "", "タグを削除（カンマ区切り）")
	updateAllCmd.StringVarP(&updateAll.message, "message", "m", "", "更新メッセージ")
	updateAllCmd.BoolVarP(&updateAll.force, "force", "f", false, "一覧の確認なしで実行（競合時の確認は省略されません）")
	updateAllCmd.BoolVar(&updateAll.dryRun, "dry-run", false, "更新せず各ファイルの状態と画像を表示")

	// commentコマンドのオプション
	commentCmd := pflag.NewFlagSet("comment", pflag.ExitOnError)
	var commentText, commentFile string
	var commentDryRun, commentForce bool
	commentCmd.StringVarP(&commentText, "body", "b", "", "コメント本文")
	commentCmd.StringVarP(&commentFile, "file", "f", "", "本文のファイル（-で標準入力）")
	commentCmd.BoolVar(&commentDryRun, "dry-run", false, "投稿せず内容だけを表示")
	commentCmd.BoolVar(&commentForce, "force", false, "確認なしで投稿")

	// createコマンドのオプション
	createCmd := pflag.NewFlagSet("create", pflag.ExitOnError)
	var createTitle string
	var createCategory string
	var createTags string
	var createMessage string
	var createWip bool
	var createNoWip bool
	var createFile string
	var createTemplate bool
	createCmd.StringVarP(&createTitle, "title", "t", "", "記事のタイトル")
	createCmd.StringVarP(&createCategory, "category", "c", "", "カテゴリ")
	createCmd.StringVarP(&createTags, "tags", "g", "", "タグ（カンマ区切り）")
	createCmd.StringVarP(&createMessage, "message", "m", "", "作成メッセージ")
	createCmd.BoolVarP(&createWip, "wip", "w", false, "WIP状態で作成（デフォルト）")
	createCmd.BoolVar(&createNoWip, "no-wip", false, "公開状態で作成")
	createCmd.StringVarP(&createFile, "file", "f", "", "既存のMarkdownファイルから作成")
	createCmd.BoolVarP(&createTemplate, "template", "T", false, "esa.ioにアップロードせず、ローカルにテンプレートファイルのみ生成")

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
		runFetch(fetchCmd, fetchCategory, fetchTag, fetchQuery, fetchUser, fetchLatest, fetchPrint, fetchOverwrite, fetchComments)
	case "update":
		updateCmd.Parse(os.Args[2:])
		runUpdate(updateCmd, updateOptions{noWip, updateCategory, addTags, removeTags, message})
	case "move":
		moveCmd.Parse(os.Args[2:])
		runMove(moveCmd, moveCategory, moveUser, moveQuery, moveTag, moveToCategory, moveMessage, moveForce, moveDryRun)
	case "fetch-all":
		fetchAllCmd.Parse(os.Args[2:])
		runFetchAll(fetchAll)
	case "update-all":
		updateAllCmd.Parse(os.Args[2:])
		updateAll.patterns = updateAllCmd.Args()
		runUpdateAll(updateAll)
	case "comment":
		commentCmd.Parse(os.Args[2:])
		runComment(commentCmd, commentText, commentFile, commentDryRun, commentForce)
	case "diff":
		runDiff(os.Args[2:])
	case "status":
		runStatus(os.Args[2:])
	case "create":
		createCmd.Parse(os.Args[2:])
		runCreate(createCmd, createTitle, createCategory, createTags, createMessage, createWip, createNoWip, createFile, createTemplate)
	case "help", "-h", "--help":
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
	fmt.Println("      -p, --print               ファイルに保存せず標準出力に表示")
	fmt.Println("      --overwrite               既存のローカルファイルを上書き（元は.esa-cli-backupに保存）")
	fmt.Println("      --comments                コメントも取得（-pなら続けて表示、それ以外は.comments.mdに保存）")
	fmt.Println("  esa-cli update <ファイル名>    記事を更新（競合時は確認、ローカル画像は自動アップロード）")
	fmt.Println("    オプション:")
	fmt.Println("      -n, --no-wip              WIP状態を解除")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリを変更")
	fmt.Println("      -a, --add-tags <タグ>     タグを追加（カンマ区切り）")
	fmt.Println("      -r, --remove-tags <タグ>  タグを削除（カンマ区切り）")
	fmt.Println("      -m, --message <メッセージ> 更新メッセージ")
	fmt.Println("  esa-cli fetch-all              記事を一括ダウンロード")
	fmt.Println("    オプション:")
	fmt.Println("      -c, --category / -t, --tag / -q, --query / -u, --user  フィルタリング")
	fmt.Println("      -l, --limit <件数>         取得件数の上限（デフォルト100、0で無制限）")
	fmt.Println("      --overwrite               既存のファイルを上書き（デフォルトはスキップ、元は.esa-cli-backupに保存）")
	fmt.Println("      --dry-run                 保存せず対象の記事だけを表示")
	fmt.Println("  esa-cli update-all <パターン>  ローカルファイルから記事を一括更新（ローカル画像は自動アップロード）")
	fmt.Println("    オプション:")
	fmt.Println("      -n, --no-wip / -c, --category / -a, --add-tags / -r, --remove-tags / -m, --message")
	fmt.Println("      -f, --force               一覧の確認なしで実行（競合時の確認は省略されません）")
	fmt.Println("      --dry-run                 更新せず各ファイルの状態と画像を表示")
	fmt.Println("  esa-cli comment <記事番号>     記事にコメントを投稿")
	fmt.Println("    オプション:")
	fmt.Println("      -b, --body <本文>          コメント本文")
	fmt.Println("      -f, --file <ファイル>      本文のファイル（-で標準入力）")
	fmt.Println("      --dry-run                 投稿せず内容だけを表示")
	fmt.Println("      --force                   確認なしで投稿")
	fmt.Println("  esa-cli diff <ファイル名>      ローカルとesaの差分を表示")
	fmt.Println("  esa-cli status [パターン]      記事ファイルの状態を一覧表示（デフォルトは*.md）")
	fmt.Println("  esa-cli move                  記事を一括移動（絞り込み条件が1つ以上必要）")
	fmt.Println("    オプション:")
	fmt.Println("      -c, --category <移動元カテゴリ> 移動元のカテゴリ")
	fmt.Println("      -u, --user <作成者>       作成者でフィルタリング")
	fmt.Println("      -q, --query <検索ワード>   検索ワードでフィルタリング")
	fmt.Println("      -t, --tag <タグ>          タグでフィルタリング")
	fmt.Println("      -o, --to <移動先カテゴリ>  移動先のカテゴリ（必須）")
	fmt.Println("      -m, --message <メッセージ> 移動メッセージ")
	fmt.Println("      -f, --force               確認なしで実行")
	fmt.Println("      --dry-run                 移動せず対象の記事だけを表示")
	fmt.Println("  esa-cli create                 新しい記事を作成")
	fmt.Println("    オプション:")
	fmt.Println("      -t, --title <記事のタイトル>  記事のタイトル")
	fmt.Println("      -c, --category <カテゴリ>  カテゴリ")
	fmt.Println("      -g, --tags <タグ>          タグ（カンマ区切り）")
	fmt.Println("      -m, --message <作成メッセージ> 作成メッセージ")
	fmt.Println("      -w, --wip                 WIP状態で作成（デフォルト）")
	fmt.Println("      --no-wip                  公開状態で作成")
	fmt.Println("      -f, --file <既存のMarkdownファイル> 既存のMarkdownファイルから作成（ローカル画像は自動アップロード）")
	fmt.Println("      -T, --template            ローカルにテンプレートファイルのみ生成（esa.ioにアップロードしない）")
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
	fmt.Println("  esa-cli fetch 123 -p            # 記事123を標準出力に表示（ファイル保存なし）")
	fmt.Println("  esa-cli fetch -c 開発 -l        # 開発カテゴリの最新記事をダウンロード")
	fmt.Println("  esa-cli fetch -c 開発 -l -p     # 開発カテゴリの最新記事を標準出力に表示")
	fmt.Println("  esa-cli fetch -t API -l         # APIタグの最新記事をダウンロード")
	fmt.Println("  esa-cli fetch -q 認証 -l        # 認証を含む最新記事をダウンロード")
	fmt.Println("  esa-cli update 123-title.md    # 記事を更新")
	fmt.Println("  esa-cli update 123-title.md -n # WIP状態を解除して更新")
	fmt.Println("  esa-cli update 123-title.md -c 開発  # カテゴリを変更して更新")
	fmt.Println("  esa-cli update 123-title.md -a API,認証  # タグを追加して更新")
	fmt.Println("  esa-cli update 123-title.md -m API仕様を更新  # メッセージを付けて更新")
	fmt.Println("  esa-cli fetch-all -c 開発 --dry-run  # ダウンロード対象を確認")
	fmt.Println("  esa-cli update-all \"123-*.md\" -m 一括更新  # パターンに一致する記事を更新")
	fmt.Println("  esa-cli fetch 123 -p --comments  # 記事とコメントを標準出力に表示")
	fmt.Println("  esa-cli comment 123 -f reply.md --dry-run  # 投稿内容を確認")
	fmt.Println("  esa-cli diff 123-title.md       # esaとの差分を確認")
	fmt.Println("  esa-cli status                  # 記事ファイルの状態を一覧表示")
	fmt.Println("  esa-cli move -c 開発 -o デザイン -u 自分のユーザー名  # 一括移動")
	fmt.Println("  esa-cli move -c 開発 -o デザイン -u 自分のユーザー名 -f  # 確認なしで移動")
	fmt.Println("  esa-cli create \"新機能の説明\" -c 開発 -g API,新機能  # 新しい記事を作成")
	fmt.Println("  esa-cli create \"API仕様書\" -c 開発/API -g API,ドキュメント -w  # WIP状態で記事を作成")
	fmt.Println("  esa-cli create -f draft.md -c 開発/ドキュメント  # 既存ファイルから記事を作成")
	fmt.Println("  esa-cli create \"下書き記事\" -T  # ローカルにテンプレートファイルのみ生成")
	fmt.Println("  esa-cli create \"技術記事\" -c 技術/Go -g Go,技術記事 -T  # カテゴリ・タグ付きテンプレートを生成")
	fmt.Println("")
	fmt.Println("終了コード: 0=成功 / 1=エラー / 2=update・update-all・commentを確認で中止")
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
	limit := 10 // デフォルト値
	if len(cmd.Args()) > 0 {
		l, err := strconv.Atoi(cmd.Args()[0])
		if err != nil || l <= 0 {
			fmt.Printf("❌ 無効な件数です: %s\n", cmd.Args()[0])
			os.Exit(1)
		}
		limit = l
	}

	client := loadClient()

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

	posts, total, err := client.ListAllPosts(context.Background(), api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
		User:     user,
	}, limit)
	if err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}

	// 記事一覧を表示
	if len(posts) == 0 {
		fmt.Println("📭 条件に一致する記事が見つかりませんでした。")
		return
	}

	fmt.Printf("📋 記事一覧 (%d件 / 全%d件):\n", len(posts), total)
	for _, post := range posts {
		fmt.Printf("  [%d] %s\n", post.Number, post.FullName)
	}
}

func runFetch(cmd *pflag.FlagSet, category, tag, query, user string, latest bool, printToStdout, overwrite, comments bool) {
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
		if printToStdout {
			fmt.Fprintf(os.Stderr, "📥 最新記事を取得中: [%d] %s\n", post.Number, post.FullName)
		} else {
			fmt.Printf("📥 最新記事をダウンロード中: [%d] %s\n", post.Number, post.FullName)
		}
		// 最新記事の番号で後続の処理を行う
		fetchArticle(client, post.Number, printToStdout, overwrite, comments)
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

	fetchArticle(client, postNumber, printToStdout, overwrite, comments)
}

// 記事を取得してファイルに書き込む共通関数
func fetchArticle(client *api.Client, postNumber int, printToStdout, overwrite, comments bool) {
	if !printToStdout && !overwrite {
		existing, err := existingArticles(postNumber)
		if err == nil && len(existing) > 0 {
			fmt.Fprintf(os.Stderr, "❌ ローカルファイルが既に存在します: %s\n", existing[0])
			fmt.Fprintln(os.Stderr, "💡 上書きする場合は --overwrite を指定してください")
			os.Exit(1)
		}
	}

	// 記事を取得
	post, err := client.FetchPost(context.Background(), postNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ エラー: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "❌ ファイル内容の生成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	var commentText string
	if comments {
		list, err := client.FetchComments(context.Background(), postNumber)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ コメントの取得に失敗しました: %v\n", err)
			os.Exit(1)
		}
		commentText = formatComments(list)
	}

	if printToStdout {
		// 標準出力にMarkdownコンテンツを出力
		fmt.Print(string(content))
		if comments {
			fmt.Printf("\n---\n\n%s", commentText)
		}
	} else {
		// ファイルに保存
		if existing, _ := existingArticles(postNumber); len(existing) > 0 {
			if err := backupFile(existing[0]); err != nil {
				fmt.Fprintf(os.Stderr, "❌ バックアップに失敗しました: %v\n", err)
				os.Exit(1)
			}
		}
		fileName := fmt.Sprintf("%d-%s.md", post.Number, sanitizeArticleName(post.Name))
		if err := os.WriteFile(fileName, content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "❌ ファイルの書き込みに失敗しました: %v\n", err)
			os.Exit(1)
		}

		if comments {
			commentFile := strings.TrimSuffix(fileName, ".md") + commentFileSuffix
			if err := os.WriteFile(commentFile, []byte(commentText), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ コメントの書き込みに失敗しました: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("💬 コメントを保存しました: %s\n", commentFile)
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
}

func runCreate(cmd *pflag.FlagSet, title, category, tags, message string, wip, noWip bool, file string, template bool) {
	if wip && noWip {
		fmt.Println("❌ --wip と --no-wip は同時に指定できません")
		os.Exit(1)
	}

	// 設定の読み込み（テンプレートモードでない場合のみ必要）
	if !template {
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
	}

	// 位置引数からタイトルを取得
	if len(cmd.Args()) > 0 && title == "" {
		title = cmd.Args()[0]
	}

	// 対話形式での入力（タイトルが指定されていない場合）
	if title == "" && file == "" {
		if template {
			fmt.Println("📝 新しい記事のテンプレートを作成します")
		} else {
			fmt.Println("📝 新しい記事を作成します")
		}
		fmt.Print("記事のタイトル: ")
		fmt.Scanln(&title)
		if title == "" {
			fmt.Println("❌ タイトルが指定されていません")
			os.Exit(1)
		}
	}

	var client *api.Client
	if !template {
		cfg, _ := config.Load()
		client = newAPIClient(cfg.TeamName, cfg.AccessToken)
	}

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
		Wip:      !noWip,
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
		// フラグの指定がなければファイルのWIP状態に従う
		if !wip && !noWip {
			createBody.Wip = fm.Wip
		}
		createBody.BodyMd = body
	}

	if template {
		// テンプレートモード: ローカルファイルのみ生成
		fm := types.FrontMatter{
			Title:    createBody.Name,
			Category: createBody.Category,
			Tags:     createBody.Tags,
			Wip:      createBody.Wip,
		}

		content, err := markdown.GenerateContent(fm, createBody.BodyMd)
		if err != nil {
			fmt.Printf("❌ ファイル内容の生成に失敗しました: %v\n", err)
			os.Exit(1)
		}

		// ファイル名を生成（記事番号がないので、タイトルベース）
		fileName := sanitizeFilename(title)
		fileName = fmt.Sprintf("draft-%s.md", fileName)

		if err := os.WriteFile(fileName, content, 0644); err != nil {
			fmt.Printf("❌ ファイルの書き込みに失敗しました: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ テンプレートファイルを作成しました: %s\n", fileName)
		fmt.Printf("📝 タイトル: %s\n", createBody.Name)
		if createBody.Category != "" {
			fmt.Printf("📁 カテゴリ: %s\n", createBody.Category)
		}
		if len(createBody.Tags) > 0 {
			fmt.Printf("🏷️  タグ: %s\n", strings.Join(createBody.Tags, ", "))
		}
		fmt.Printf("💡 編集後、'esa-cli create -f %s' でesa.ioにアップロードできます\n", fileName)
		return
	}

	// 通常モード: esa.ioに記事を作成
	if file != "" {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ ファイルの読み込みに失敗しました: %v\n", err)
			os.Exit(1)
		}
		fileFm, _, err := markdown.ParseContent(content)
		if err != nil {
			fmt.Printf("❌ ファイルの解析に失敗しました: %v\n", err)
			os.Exit(1)
		}
		createBody.BodyMd, err = uploadImages(context.Background(), client, file, fileFm, createBody.BodyMd)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			os.Exit(1)
		}
	}

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

	fileName := fmt.Sprintf("%d-%s.md", post.Number, sanitizeArticleName(post.Name))
	if err := os.WriteFile(fileName, content, 0644); err != nil {
		fmt.Printf("❌ ファイルの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 新しい記事が作成されました: %s\n", post.FullName)
	fmt.Printf("📄 ローカルファイル: %s\n", fileName)
}

// sanitizeFilename ファイル名に使用できない文字を置換
func sanitizeFilename(name string) string {
	replacements := map[string]string{
		"/": "-", "\\": "-", ":": "-", "*": "-",
		"?": "-", "\"": "-", "<": "-", ">": "-", "|": "-",
		" ": "-",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 長すぎる場合は切り詰め
	if len(result) > 100 {
		result = strings.ToValidUTF8(result[:100], "")
	}

	return result
}
