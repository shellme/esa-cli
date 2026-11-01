package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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

const moveLimit = 100 // 一度に移動できる記事数の上限

const commentFileSuffix = ".comments.md"

const exitCanceled = 2 // 競合の確認などで中止したときの終了コード

var (
	articleFilePattern = regexp.MustCompile(`^(\d+)-.+\.md$`)
	errUpdateCanceled  = errors.New("更新を中止しました")
)

type updateOptions struct {
	noWip      bool
	category   string
	addTags    string
	removeTags string
	message    string
}

type fetchAllOptions struct {
	category  string
	tag       string
	user      string
	query     string
	limit     int
	overwrite bool
	dryRun    bool
}

type updateAllOptions struct {
	updateOptions
	patterns []string
	force    bool
	dryRun   bool
}

// 設定を読み込んでAPIクライアントを作成
func loadClient() *api.Client {
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

	return newAPIClient(cfg.TeamName, cfg.AccessToken)
}

// 標準入力でy/Nを確認
func confirmPrompt(prompt string) bool {
	fmt.Print(prompt)
	var answer string
	if _, err := fmt.Scanln(&answer); errors.Is(err, io.EOF) {
		fmt.Println("\n⚠️  入力できないため「いいえ」として扱います")
		return false
	}
	return strings.ToLower(answer) == "y"
}

// 記事番号に対応するローカルの記事ファイルを探す（コメントファイルは除く）
func existingArticles(postNumber int) ([]string, error) {
	matches, err := filepath.Glob(fmt.Sprintf("%d-*.md", postNumber))
	if err != nil {
		return nil, err
	}
	var files []string
	for _, match := range matches {
		if !strings.HasSuffix(match, commentFileSuffix) {
			files = append(files, match)
		}
	}
	return files, nil
}

// ファイル名から記事番号を取得
func articleNumber(path string) (int, error) {
	match := articleFilePattern.FindStringSubmatch(filepath.Base(path))
	if match == nil || strings.HasSuffix(path, commentFileSuffix) {
		return 0, fmt.Errorf("無効なファイル名です。'記事番号-タイトル.md'の形式である必要があります: %s", path)
	}
	return strconv.Atoi(match[1])
}

func splitTags(tags string) []string {
	var result []string
	for _, tag := range strings.Split(tags, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// ファイルの内容で記事を更新し、ローカルファイルを更新後の内容で書き換える
func updateFile(ctx context.Context, client *api.Client, path string, opts updateOptions, confirm func(string) bool) error {
	postNumber, err := articleNumber(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	fm, body, err := markdown.ParseContent(content)
	if err != nil {
		return fmt.Errorf("ファイルの解析に失敗しました: %w", err)
	}

	// リモートの更新日時をチェック（確認できないときは上書きしない）
	var localUpdatedAt time.Time
	if fm.RemoteUpdatedAt != "" {
		localUpdatedAt, err = time.Parse(time.RFC3339, fm.RemoteUpdatedAt)
		if err != nil {
			return fmt.Errorf("remote_updated_atの形式が不正です: %w", err)
		}
	}
	remotePost, err := client.FetchPost(ctx, postNumber)
	if err != nil {
		return fmt.Errorf("リモート記事の確認に失敗したため更新を中止しました: %w", err)
	}
	if fm.RemoteUpdatedAt == "" {
		fmt.Println("⚠️  警告: remote_updated_atがないため、リモートの記事との競合を確認できません。")
		fmt.Printf("  リモート: %s\n", remotePost.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
		if !confirm("このまま上書きしますか？ (y/N): ") {
			return errUpdateCanceled
		}
	} else if remotePost.UpdatedAt.After(localUpdatedAt) {
		fmt.Println("⚠️  警告: リモートの記事はローカルで編集を始めてから更新されています。")
		fmt.Printf("  リモート: %s\n", remotePost.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
		fmt.Printf("  ローカル: %s\n", localUpdatedAt.Local().Format("2006-01-02 15:04:05"))
		if !confirm("このまま上書きしますか？ (y/N): ") {
			return errUpdateCanceled
		}
	}

	body, err = uploadImages(ctx, client, path, fm, body)
	if err != nil {
		return err
	}

	category := fm.Category
	if opts.category != "" {
		category = opts.category
	}

	tags := append([]string{}, fm.Tags...)
	tags = append(tags, splitTags(opts.addTags)...)
	for _, removeTag := range splitTags(opts.removeTags) {
		for i, tag := range tags {
			if tag == removeTag {
				tags = append(tags[:i], tags[i+1:]...)
				break
			}
		}
	}

	wip := fm.Wip && !opts.noWip
	updatedPost, err := client.UpdatePost(ctx, postNumber, types.UpdatePostBody{
		Name:     fm.Title,
		Category: category,
		Tags:     &tags,
		BodyMd:   &body,
		Wip:      &wip,
		Message:  opts.message,
	})
	if err != nil {
		return fmt.Errorf("記事の更新に失敗しました: %w", err)
	}

	newContent, err := postContent(updatedPost)
	if err != nil {
		return fmt.Errorf("ローカルファイルの更新に失敗しました: %w", err)
	}
	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return fmt.Errorf("ローカルファイルの書き込みに失敗しました: %w", err)
	}
	return nil
}

// 本文のローカル画像をアップロードしてURLに置換する。
// 再実行で重複しないよう、1枚ごとにローカルファイルへ書き戻す
func uploadImages(ctx context.Context, client *api.Client, path string, fm types.FrontMatter, body string) (string, error) {
	for _, image := range markdown.LocalImages(body) {
		fmt.Printf("🖼️  画像をアップロード中: %s\n", image)
		imageURL, err := client.UploadAttachment(ctx, markdown.ImageFilePath(path, image))
		if err != nil {
			return "", fmt.Errorf("画像のアップロードに失敗したため中止しました (%s): %w", image, err)
		}

		body = markdown.ReplaceLocalImages(body, func(p string) string {
			if p == image {
				return imageURL
			}
			return p
		})
		content, err := markdown.GenerateContent(fm, body)
		if err == nil {
			err = os.WriteFile(path, content, 0644)
		}
		if err != nil {
			return "", fmt.Errorf("ローカルファイルの書き込みに失敗しました: %w", err)
		}
	}
	return body, nil
}

// ファイルに含まれるローカル画像を一覧表示する
func printLocalImages(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if _, body, err := markdown.ParseContent(content); err == nil {
		for _, image := range markdown.LocalImages(body) {
			fmt.Printf("      🖼️  %s\n", image)
		}
	}
}

func postContent(post *types.Post) ([]byte, error) {
	return markdown.GenerateContent(types.FrontMatter{
		Title:           post.Name,
		Category:        post.Category,
		Tags:            post.Tags,
		Wip:             post.Wip,
		RemoteUpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}, post.BodyMd)
}

func runUpdate(cmd *pflag.FlagSet, opts updateOptions) {
	if len(cmd.Args()) < 1 {
		fmt.Println("❌ ファイル名を指定してください")
		fmt.Println("💡 使用例: esa-cli update 123-title.md")
		os.Exit(1)
	}
	fileName := cmd.Args()[0]

	if _, err := articleNumber(fileName); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	err := updateFile(context.Background(), loadClient(), fileName, opts, confirmPrompt)
	if errors.Is(err, errUpdateCanceled) {
		fmt.Printf("🚫 %v\n", err)
		os.Exit(exitCanceled)
	}
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 記事を更新しました: %s\n", fileName)
}

// 条件に一致する記事を一括ダウンロードし、失敗件数を返す
func fetchAllPosts(ctx context.Context, client *api.Client, opts fetchAllOptions) (int, error) {
	posts, total, err := client.ListAllPosts(ctx, api.ListPostsOptions{
		Category: opts.category,
		Tag:      opts.tag,
		User:     opts.user,
		Query:    opts.query,
	}, opts.limit)
	if err != nil {
		return 0, err
	}

	if len(posts) == 0 {
		fmt.Println("📭 条件に一致する記事が見つかりませんでした。")
		return 0, nil
	}
	if total > len(posts) {
		fmt.Printf("⚠️  該当する%d件のうち%d件のみ処理します（--limitで変更できます）\n", total, len(posts))
	}

	saved, skipped, failed := 0, 0, 0
	for _, post := range posts {
		existing, err := existingArticles(post.Number)
		if err != nil {
			return failed, err
		}
		if len(existing) > 0 && !opts.overwrite {
			fmt.Printf("⏭️  スキップ（既存）: %s\n", existing[0])
			skipped++
			continue
		}

		fileName := fmt.Sprintf("%d-%s.md", post.Number, sanitizeArticleName(post.Name))
		if opts.dryRun {
			fmt.Printf("📄 保存予定: %s\n", fileName)
			continue
		}

		if len(existing) > 0 {
			if err := backupFile(existing[0]); err != nil {
				fmt.Printf("❌ [%d] %s: %v\n", post.Number, post.Name, err)
				failed++
				continue
			}
		}

		content, err := postContent(post)
		if err == nil {
			err = os.WriteFile(fileName, content, 0644)
		}
		if err != nil {
			fmt.Printf("❌ [%d] %s: %v\n", post.Number, post.Name, err)
			failed++
			continue
		}
		fmt.Printf("✅ 保存完了: %s\n", fileName)
		saved++
	}

	fmt.Printf("\n📊 結果: 保存 %d件 / スキップ %d件 / 失敗 %d件\n", saved, skipped, failed)
	if saved > 0 {
		mac.SendNotification("esa-cli", fmt.Sprintf("%d件の記事をダウンロードしました", saved))
	}
	return failed, nil
}

// 記事名に含まれるパス区切りをファイル名用に置換
func sanitizeArticleName(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}

func runFetchAll(opts fetchAllOptions) {
	failed, err := fetchAllPosts(context.Background(), loadClient(), opts)
	if err != nil {
		fmt.Printf("❌ 記事の一括取得に失敗しました: %v\n", err)
		os.Exit(1)
	}
	if failed > 0 {
		os.Exit(1)
	}
}

// パターンに一致する記事ファイルを探す（サブフォルダは探索しない）
func findArticleFiles(patterns []string) ([]string, error) {
	var files []string
	seen := map[string]bool{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("無効なパターンです: %s", pattern)
		}
		for _, match := range matches {
			if _, err := articleNumber(match); err != nil || seen[match] {
				continue
			}
			seen[match] = true
			files = append(files, match)
		}
	}
	return files, nil
}

// ファイルを一括更新し、失敗件数と中止したファイルを返す
func updateAllFiles(ctx context.Context, client *api.Client, files []string, opts updateOptions, confirm func(string) bool) (int, []string) {
	updated, failed := 0, 0
	var canceled []string
	for _, file := range files {
		fmt.Printf("📝 更新中: %s\n", file)
		err := updateFile(ctx, client, file, opts, confirm)
		switch {
		case errors.Is(err, errUpdateCanceled):
			fmt.Printf("   🚫 %v\n", err)
			canceled = append(canceled, file)
		case err != nil:
			fmt.Printf("   ❌ %v\n", err)
			failed++
		default:
			fmt.Println("   ✅ 更新完了")
			updated++
		}
	}

	fmt.Printf("\n📊 結果: 更新 %d件 / 中止 %d件 / 失敗 %d件\n", updated, len(canceled), failed)
	for _, file := range canceled {
		fmt.Printf("  🚫 中止: %s\n", file)
	}
	if updated > 0 {
		mac.SendNotification("esa-cli", fmt.Sprintf("%d件の記事を更新しました", updated))
	}
	return failed, canceled
}

func runUpdateAll(opts updateAllOptions) {
	if len(opts.patterns) == 0 {
		fmt.Println("❌ 対象のファイルパターンを指定してください")
		fmt.Println("💡 使用例: esa-cli update-all \"123-*.md\" --dry-run")
		os.Exit(1)
	}

	files, err := findArticleFiles(opts.patterns)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("📭 条件に一致するファイルが見つかりませんでした。")
		return
	}

	fmt.Printf("📋 更新対象のファイル (%d件):\n", len(files))
	if opts.dryRun {
		if printStatuses(context.Background(), loadClient(), files) > 0 {
			os.Exit(1)
		}
		return
	}
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}
	if !opts.force && !confirmPrompt("\n上記のファイルを更新しますか？ (y/N): ") {
		fmt.Println("🚫 更新を中止しました。")
		return
	}

	failed, canceled := updateAllFiles(context.Background(), loadClient(), files, opts.updateOptions, confirmPrompt)
	if failed > 0 {
		os.Exit(1)
	}
	if len(canceled) > 0 {
		os.Exit(exitCanceled)
	}
}

// 移動対象の記事を検索する
func findMovePosts(ctx context.Context, client *api.Client, options api.ListPostsOptions) ([]*types.Post, int, error) {
	posts, total, err := client.ListAllPosts(ctx, options, moveLimit)
	if err != nil {
		return nil, 0, err
	}
	if options.Category == "" {
		return posts, total, nil
	}

	// 検索結果に対象外のカテゴリが混ざっても移動しない
	filtered := []*types.Post{}
	for _, post := range posts {
		if api.InCategory(post, options.Category) {
			filtered = append(filtered, post)
		}
	}
	return filtered, total, nil
}

func runMove(cmd *pflag.FlagSet, category, user, query, tag, toCategory, message string, force, dryRun bool) {
	if toCategory == "" {
		fmt.Println("❌ エラー: 移動先のカテゴリを指定してください (--to オプション)")
		fmt.Println("💡 例: esa-cli move --category 開発 --to デザイン --user 自分のユーザー名")
		os.Exit(1)
	}
	if category == "" && user == "" && query == "" && tag == "" {
		fmt.Println("❌ エラー: 絞り込み条件を1つ以上指定してください (--category / --user / --query / --tag)")
		os.Exit(1)
	}

	client := loadClient()

	fmt.Println("🔍 移動対象の記事を検索中...")
	posts, total, err := findMovePosts(context.Background(), client, api.ListPostsOptions{
		Category: category,
		Tag:      tag,
		Query:    query,
		User:     user,
	})
	if err != nil {
		fmt.Printf("❌ 記事の検索に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("⚠️  移動対象の記事が見つかりませんでした")
		return
	}
	if total > moveLimit {
		fmt.Printf("⚠️  該当する%d件のうち最初の%d件のみ処理します。条件を絞り込むか、再実行してください。\n", total, moveLimit)
	}

	fmt.Printf("\n📋 移動対象の記事 (%d件):\n", len(posts))
	for i, post := range posts {
		fmt.Printf("  %d. [%d] %s (現在のカテゴリ: %s)\n", i+1, post.Number, post.FullName, post.Category)
	}
	fmt.Printf("\n🎯 移動先カテゴリ: %s\n", toCategory)

	if dryRun {
		return
	}
	if !force && !confirmPrompt("\n⚠️  上記の記事を移動しますか？ (y/N): ") {
		fmt.Println("❌ 移動をキャンセルしました")
		return
	}

	var postNumbers []int
	for _, post := range posts {
		postNumbers = append(postNumbers, post.Number)
	}

	if message == "" {
		message = fmt.Sprintf("カテゴリを %s に一括移動", toCategory)
	}

	fmt.Printf("\n🚀 記事の移動を開始します...\n")
	updatedPosts, err := client.BulkUpdateCategory(context.Background(), postNumbers, toCategory, message)
	for _, post := range updatedPosts {
		fmt.Printf("   ✅ [%d] %s\n", post.Number, post.FullName)
	}
	fmt.Printf("\n📊 結果: 移動 %d件 / 対象 %d件\n", len(updatedPosts), len(postNumbers))
	if err != nil {
		fmt.Printf("❌ 記事の移動に失敗しました: %v\n", err)
		os.Exit(1)
	}
}
