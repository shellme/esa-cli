package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/markdown"
	"github.com/shellme/esa-cli/pkg/types"
)

const backupDir = ".esa-cli-backup"

const (
	stateClean        = "変更なし"
	stateLocalChanged = "ローカル変更"
	stateRemoteNewer  = "要確認（esa側に更新あり）"
)

// 比較用の内容を作る（タグの順序、改行コード、remote_updated_atの差は無視する）
func comparableContent(fm types.FrontMatter, body string) (string, error) {
	fm.RemoteUpdatedAt = ""
	fm.Tags = append([]string{}, fm.Tags...)
	sort.Strings(fm.Tags)
	content, err := markdown.GenerateContent(fm, strings.ReplaceAll(body, "\r\n", "\n"))
	return string(content), err
}

// ローカルファイルとesaの記事を比較し、状態と比較用の内容を返す
func compareArticle(ctx context.Context, client *api.Client, path string) (state, local, remote string, err error) {
	postNumber, err := articleNumber(path)
	if err != nil {
		return "", "", "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}
	fm, body, err := markdown.ParseContent(content)
	if err != nil {
		return "", "", "", fmt.Errorf("ファイルの解析に失敗しました: %w", err)
	}
	post, err := client.FetchPost(ctx, postNumber)
	if err != nil {
		return "", "", "", fmt.Errorf("リモート記事の取得に失敗しました: %w", err)
	}

	if local, err = comparableContent(fm, body); err != nil {
		return "", "", "", err
	}
	remote, err = comparableContent(types.FrontMatter{
		Title: post.Name, Category: post.Category, Tags: post.Tags, Wip: post.Wip,
	}, post.BodyMd)
	if err != nil {
		return "", "", "", err
	}

	if local == remote {
		return stateClean, local, remote, nil
	}
	if fm.RemoteUpdatedAt != "" {
		localUpdatedAt, err := time.Parse(time.RFC3339, fm.RemoteUpdatedAt)
		if err != nil {
			return "", "", "", fmt.Errorf("remote_updated_atの形式が不正です: %w", err)
		}
		if !post.UpdatedAt.After(localUpdatedAt) {
			return stateLocalChanged, local, remote, nil
		}
	}
	return stateRemoteNewer, local, remote, nil
}

// diffコマンドで差分を表示する
func showDiff(local, remote, path string) error {
	dir, err := os.MkdirTemp("", "esa-cli-diff")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	localPath, remotePath := filepath.Join(dir, "local"), filepath.Join(dir, "remote")
	if err := os.WriteFile(localPath, []byte(local), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(remotePath, []byte(remote), 0600); err != nil {
		return err
	}

	cmd := exec.Command("diff", "-u", "--label", "esa: "+path, "--label", "local: "+path, remotePath, localPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	// 終了コード1は差分ありを表す
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return nil
	}
	return err
}

func runDiff(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ ファイル名を指定してください")
		fmt.Println("💡 使用例: esa-cli diff 123-title.md")
		os.Exit(1)
	}

	state, local, remote, err := compareArticle(context.Background(), loadClient(), args[0])
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📌 状態: %s\n", state)
	if state == stateClean {
		return
	}
	if err := showDiff(local, remote, args[0]); err != nil {
		fmt.Printf("❌ 差分の表示に失敗しました: %v\n", err)
		os.Exit(1)
	}
}

// ファイルごとの状態を表示し、取得に失敗した件数を返す
func printStatuses(ctx context.Context, client *api.Client, files []string) int {
	failed := 0
	for _, file := range files {
		state, _, _, err := compareArticle(ctx, client, file)
		if err != nil {
			fmt.Printf("  ❌ %s: %v\n", file, err)
			failed++
			continue
		}
		fmt.Printf("  [%s] %s\n", state, file)
		printLocalImages(file)
	}
	return failed
}

func runStatus(patterns []string) {
	if len(patterns) == 0 {
		patterns = []string{"*.md"}
	}
	files, err := findArticleFiles(patterns)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("📭 記事ファイルが見つかりませんでした。")
		return
	}
	if printStatuses(context.Background(), loadClient(), files) > 0 {
		os.Exit(1)
	}
}

// 上書き前のファイルを.esa-cli-backupへコピーする
func backupFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dir := filepath.Join(filepath.Dir(path), backupDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	name := fmt.Sprintf("%s.%s", filepath.Base(path), time.Now().Format("20060102-150405"))
	backupPath := filepath.Join(dir, name)
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "💾 バックアップ: %s\n", backupPath)
	return nil
}
