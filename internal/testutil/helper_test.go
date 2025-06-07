package testutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTempDir(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := CreateTempDir(t)

	// ディレクトリが存在することを確認
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("CreateTempDir() created directory does not exist: %s", tmpDir)
	}

	// ディレクトリが空であることを確認
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Errorf("CreateTempDir() failed to read directory: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("CreateTempDir() created directory is not empty: %s", tmpDir)
	}
}

func TestCreateTestPost(t *testing.T) {
	// テスト記事を作成
	content := CreateTestPost(t)

	// 記事の内容を確認
	expectedContent := `---
title: テスト記事
category: test
tags: [test]
wip: false
---

テスト本文
`
	if content != expectedContent {
		t.Errorf("CreateTestPost() content = %v, want %v", content, expectedContent)
	}

	// 必要なフィールドが含まれていることを確認
	requiredFields := []string{
		"title: テスト記事",
		"category: test",
		"tags: [test]",
		"wip: false",
		"テスト本文",
	}
	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("CreateTestPost() content does not contain required field: %s", field)
		}
	}
}

func TestCreateTestPostFile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := CreateTempDir(t)

	// テスト記事ファイルを作成
	filename := CreateTestPostFile(t, tmpDir, 1, "テスト記事")

	// ファイルが存在することを確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("CreateTestPostFile() created file does not exist: %s", filename)
	}

	// ファイルの内容を確認
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("CreateTestPostFile() failed to read file: %v", err)
	}
	expectedContent := CreateTestPost(t)
	if string(content) != expectedContent {
		t.Errorf("CreateTestPostFile() content = %v, want %v", string(content), expectedContent)
	}

	// ファイル名が正しいことを確認
	expectedFilename := filepath.Join(tmpDir, "1-テスト記事.md")
	if filename != expectedFilename {
		t.Errorf("CreateTestPostFile() filename = %v, want %v", filename, expectedFilename)
	}
}

func TestCreateTestConfigFile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := CreateTempDir(t)

	// テスト設定ファイルを作成
	filename := CreateTestConfigFile(t, tmpDir)

	// ファイルが存在することを確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("CreateTestConfigFile() created file does not exist: %s", filename)
	}

	// ファイルの内容を確認
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("CreateTestConfigFile() failed to read file: %v", err)
	}

	// 必要なフィールドが含まれていることを確認
	requiredFields := []string{
		"access_token",
		"team_name",
	}
	for _, field := range requiredFields {
		if !strings.Contains(string(content), field) {
			t.Errorf("CreateTestConfigFile() content does not contain required field: %s", field)
		}
	}
}

func TestCreateMockResponse(t *testing.T) {
	// モックレスポンスを作成
	resp := CreateMockResponse(t, 200, `{"message": "success"}`)

	// レスポンスが正しく作成されていることを確認
	if resp == nil {
		t.Error("CreateMockResponse() returned nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("CreateMockResponse() StatusCode = %v, want %v", resp.StatusCode, 200)
	}

	// レスポンスボディを確認
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("CreateMockResponse() failed to read body: %v", err)
	}
	expectedBody := `{"message": "success"}`
	if string(body) != expectedBody {
		t.Errorf("CreateMockResponse() body = %v, want %v", string(body), expectedBody)
	}
}

func TestCreateTestPostsResponse(t *testing.T) {
	// テスト記事一覧レスポンスを作成
	content := CreateTestPostsResponse(t)

	// レスポンスが空でないことを確認
	if content == "" {
		t.Error("CreateTestPostsResponse() returned empty string")
	}

	// 必要なフィールドが含まれていることを確認
	requiredFields := []string{
		"posts",
		"number",
		"name",
		"category",
		"tags",
	}
	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("CreateTestPostsResponse() content does not contain required field: %s", field)
		}
	}
}
