package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// CreateMockResponse モックレスポンスを作成
func CreateMockResponse(t *testing.T, statusCode int, body string) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

// CreateTempDir テスト用の一時ディレクトリを作成
func CreateTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "esa-cli-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CreateTestPost テスト用の記事データを作成
func CreateTestPost(t *testing.T) string {
	t.Helper()
	return `---
title: テスト記事
category: test
tags: [test]
wip: false
---

テスト本文
`
}

// CreateTestPostFile テスト用の記事ファイルを作成
func CreateTestPostFile(t *testing.T, dir string, number int, title string) string {
	t.Helper()
	filename := filepath.Join(dir, fmt.Sprintf("%d-%s.md", number, title))
	content := CreateTestPost(t)
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return filename
}

// CreateTestConfigFile テスト用の設定ファイルを作成
func CreateTestConfigFile(t *testing.T, dir string) string {
	t.Helper()
	filename := filepath.Join(dir, "config.json")
	cfg := map[string]string{
		"access_token": "test-token",
		"team_name":    "test-team",
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatal(err)
	}
	return filename
}

// CreateTestPostsResponse テスト用の記事一覧レスポンスを作成
func CreateTestPostsResponse(t *testing.T) string {
	t.Helper()
	posts := []map[string]interface{}{
		{
			"number":   1,
			"name":     "テスト記事1",
			"category": "test",
			"tags":     []string{"test"},
		},
		{
			"number":   2,
			"name":     "テスト記事2",
			"category": "test",
			"tags":     []string{"test"},
		},
	}
	data, err := json.Marshal(map[string]interface{}{
		"posts": posts,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
