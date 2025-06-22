package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/api/mock"
	"github.com/shellme/esa-cli/internal/config"
	"github.com/shellme/esa-cli/internal/testutil"
)

func TestMain(m *testing.M) {
	// テストのセットアップ
	os.Exit(m.Run())
}

func TestShowHelp(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "ヘルプを表示",
			args:    []string{"help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append([]string{"esa-cli"}, tt.args...)
			main()
		})
	}
}

func TestVersion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "バージョンを表示",
			args:    []string{"version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append([]string{"esa-cli"}, tt.args...)
			main()
		})
	}
}

func TestList(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// テスト用の設定ファイルを作成
	configPath := testutil.CreateTestConfigFile(t, tmpDir)

	// 設定ファイルのパスを設定
	origConfigFile := config.ConfigFile
	config.ConfigFile = configPath
	defer func() { config.ConfigFile = origConfigFile }()

	// モッククライアントを作成
	mockClient := mock.NewMockHTTPClient()

	// newAPIClientを差し替え
	origNewAPIClient := newAPIClient
	newAPIClient = func(team, token string) *api.Client {
		return api.NewClient(team, token, mockClient)
	}
	defer func() { newAPIClient = origNewAPIClient }()

	tests := []struct {
		name    string
		args    []string
		mockRes string
		wantErr bool
	}{
		{
			name:    "記事一覧を表示",
			args:    []string{"list"},
			mockRes: testutil.CreateTestPostsResponse(t),
			wantErr: false,
		},
		{
			name:    "カテゴリーでフィルタリング",
			args:    []string{"list", "--category", "test"},
			mockRes: testutil.CreateTestPostsResponse(t),
			wantErr: false,
		},
		{
			name:    "タグでフィルタリング",
			args:    []string{"list", "--tag", "test"},
			mockRes: testutil.CreateTestPostsResponse(t),
			wantErr: false,
		},
		{
			name:    "クエリでフィルタリング",
			args:    []string{"list", "--query", "テスト"},
			mockRes: testutil.CreateTestPostsResponse(t),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, tt.mockRes), nil)
			os.Args = append([]string{"esa-cli"}, tt.args...)
			main()
		})
	}
}

func TestFetch(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// テスト用の設定ファイルを作成
	configPath := testutil.CreateTestConfigFile(t, tmpDir)

	// 設定ファイルのパスを設定
	origConfigFile := config.ConfigFile
	config.ConfigFile = configPath
	defer func() { config.ConfigFile = origConfigFile }()

	// モッククライアントを作成
	mockClient := mock.NewMockHTTPClient()

	// newAPIClientを差し替え
	origNewAPIClient := newAPIClient
	newAPIClient = func(team, token string) *api.Client {
		return api.NewClient(team, token, mockClient)
	}
	defer func() { newAPIClient = origNewAPIClient }()

	tests := []struct {
		name    string
		args    []string
		mockRes string
		wantErr bool
	}{
		{
			name:    "記事を取得",
			args:    []string{"fetch", "1"},
			mockRes: `{"number": 1, "name": "テスト記事", "category": "test", "tags": ["test"], "body_md": "# テスト記事\n\nこれはテスト記事です。"}`,
			wantErr: false,
		},
		{
			name:    "カテゴリーを指定",
			args:    []string{"fetch", "1", "--category", "test"},
			mockRes: `{"number": 1, "name": "テスト記事", "category": "test", "tags": ["test"], "body_md": "# テスト記事\n\nこれはテスト記事です。"}`,
			wantErr: false,
		},
		{
			name:    "タグを指定",
			args:    []string{"fetch", "1", "--tag", "test"},
			mockRes: `{"number": 1, "name": "テスト記事", "category": "test", "tags": ["test"], "body_md": "# テスト記事\n\nこれはテスト記事です。"}`,
			wantErr: false,
		},
		{
			name:    "クエリを指定",
			args:    []string{"fetch", "1", "--query", "テスト"},
			mockRes: `{"number": 1, "name": "テスト記事", "category": "test", "tags": ["test"], "body_md": "# テスト記事\n\nこれはテスト記事です。"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, tt.mockRes), nil)
			os.Args = append([]string{"esa-cli"}, tt.args...)
			main()
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Skip("API認証が必要なためCIではスキップ")
	// 一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)
	defer os.RemoveAll(tmpDir)

	// テスト用の設定ファイルを作成
	testutil.CreateTestConfigFile(t, tmpDir)

	// 設定ファイルのパスを設定
	originalConfigFile := config.ConfigFile
	config.ConfigFile = filepath.Join(tmpDir, "config.json")
	defer func() {
		config.ConfigFile = originalConfigFile
	}()

	// テスト用の記事ファイルを作成
	postFile := testutil.CreateTestPostFile(t, tmpDir, 1, "テスト記事")
	// ファイル名のみを取得（パスを除く）
	postFileName := filepath.Base(postFile)

	// 現在のディレクトリを一時的に変更
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 基本的なupdateコマンドのテスト（エラーが発生することを期待）
	os.Args = []string{"esa-cli", "update", postFileName}
	main()
}
