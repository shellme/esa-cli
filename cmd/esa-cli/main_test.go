package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/api/mock"
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

	// 設定ファイルのパスを環境変数に設定
	os.Setenv("ESA_CLI_CONFIG", configPath)
	defer os.Unsetenv("ESA_CLI_CONFIG")

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
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// テスト用の記事ファイルを作成
	filename := testutil.CreateTestPostFile(t, tmpDir, 1, "テスト記事")

	// モッククライアントを作成
	mockClient := mock.NewMockHTTPClient()
	mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, `{
		"number": 1,
		"name": "更新された記事",
		"category": "test",
		"tags": ["test", "updated"],
		"body_md": "# 更新された記事\n\nこれは更新された記事です。"
	}`), nil)

	// newAPIClientを差し替え
	origNewAPIClient := newAPIClient
	newAPIClient = func(team, token string) *api.Client {
		return api.NewClient(team, token, mockClient)
	}
	defer func() { newAPIClient = origNewAPIClient }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "記事を更新",
			args:    []string{"update", filename},
			wantErr: false,
		},
		{
			name:    "カテゴリーを指定",
			args:    []string{"update", filename, "--category", "test"},
			wantErr: false,
		},
		{
			name:    "タグを指定",
			args:    []string{"update", filename, "--tag", "test"},
			wantErr: false,
		},
		{
			name:    "クエリを指定",
			args:    []string{"update", filename, "--query", "テスト"},
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
