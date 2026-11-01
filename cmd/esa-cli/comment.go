package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/spf13/pflag"
)

// コメントをMarkdownに整形する
func formatComments(comments []*api.Comment) string {
	var result strings.Builder
	fmt.Fprintf(&result, "## コメント（%d件）\n", len(comments))
	for _, comment := range comments {
		fmt.Fprintf(&result, "\n### %s（%s）\n\n%s\n",
			comment.CreatedBy.ScreenName,
			comment.CreatedAt.Local().Format("2006-01-02 15:04"),
			strings.TrimSpace(comment.BodyMd))
	}
	return result.String()
}

// 引数またはファイルからコメント本文を読み込む
func commentBody(body, file string) (string, error) {
	if file != "" {
		var data []byte
		var err error
		if file == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(file)
		}
		if err != nil {
			return "", fmt.Errorf("本文の読み込みに失敗しました: %w", err)
		}
		body = string(data)
	}
	if body = strings.TrimSpace(body); body == "" {
		return "", errors.New("コメント本文が空です")
	}
	return body, nil
}

// 内容を表示し、確認のうえコメントを投稿する
func postComment(ctx context.Context, client *api.Client, postNumber int, body string, dryRun bool, confirm func(string) bool) error {
	post, err := client.FetchPost(ctx, postNumber)
	if err != nil {
		return fmt.Errorf("記事の取得に失敗しました: %w", err)
	}

	fmt.Printf("📝 投稿先: [%d] %s\n\n%s\n\n", postNumber, post.FullName, body)
	if dryRun {
		return nil
	}
	if !confirm("このコメントを投稿しますか？ (y/N): ") {
		return errUpdateCanceled
	}
	return client.CreateComment(ctx, postNumber, body)
}

func runComment(cmd *pflag.FlagSet, body, file string, dryRun, force bool) {
	if len(cmd.Args()) < 1 {
		fmt.Println("❌ 記事番号を指定してください")
		fmt.Println("💡 使用例: esa-cli comment 123 --body コメント本文")
		os.Exit(1)
	}
	postNumber, err := strconv.Atoi(cmd.Args()[0])
	if err != nil {
		fmt.Printf("❌ 無効な記事番号です: %s\n", cmd.Args()[0])
		os.Exit(1)
	}
	text, err := commentBody(body, file)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	confirm := confirmPrompt
	if force {
		confirm = func(string) bool { return true }
	}
	err = postComment(context.Background(), loadClient(), postNumber, text, dryRun, confirm)
	if errors.Is(err, errUpdateCanceled) {
		fmt.Println("🚫 投稿を中止しました")
		os.Exit(exitCanceled)
	}
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	if !dryRun {
		fmt.Println("✅ コメントを投稿しました")
	}
}
