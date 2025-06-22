package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellme/esa-cli/internal/api/mock"
	"github.com/shellme/esa-cli/internal/testutil"
	"github.com/shellme/esa-cli/pkg/types"
)

func TestClient_ListPosts(t *testing.T) {
	mockClient := mock.NewMockHTTPClient()
	client := NewClient("test-team", "test-token", mockClient)

	tests := []struct {
		name    string
		setup   func()
		options *ListPostsOptions
		wantErr bool
	}{
		{
			name: "正常なレスポンス",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, testutil.CreateTestPostsResponse(t)), nil)
			},
			options: &ListPostsOptions{},
			wantErr: false,
		},
		{
			name: "クエリパラメータ付き",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, testutil.CreateTestPostsResponse(t)), nil)
			},
			options: &ListPostsOptions{Query: "test"},
			wantErr: false,
		},
		{
			name: "エラーレスポンス",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusInternalServerError, `{"error": "Internal Server Error"}`), nil)
			},
			options: &ListPostsOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			posts, err := client.ListPosts(context.Background(), tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(posts) == 0 {
				t.Error("ListPosts() returned empty posts")
			}
		})
	}
}

func TestClient_FetchPost(t *testing.T) {
	mockClient := mock.NewMockHTTPClient()
	client := NewClient("test-team", "test-token", mockClient)

	tests := []struct {
		name    string
		setup   func()
		number  int
		wantErr bool
	}{
		{
			name: "正常なレスポンス",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, `{
					"number": 1,
					"name": "テスト記事",
					"full_name": "テスト記事",
					"wip": false,
					"body_md": "テスト本文",
					"body_html": "<p>テスト本文</p>",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z",
					"message": "テストメッセージ",
					"url": "https://test-team.esa.io/posts/1",
					"tags": ["test"],
					"category": "test"
				}`), nil)
			},
			number:  1,
			wantErr: false,
		},
		{
			name: "記事が見つからない",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusNotFound, `{"error": "Not Found"}`), nil)
			},
			number:  999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			post, err := client.FetchPost(context.Background(), tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchPost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && post == nil {
				t.Error("FetchPost() returned nil post")
			}
		})
	}
}

func TestClient_UpdatePost(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// テスト用の記事ファイルを作成
	filename := testutil.CreateTestPostFile(t, tmpDir, 1, "テスト記事")

	mockClient := mock.NewMockHTTPClient()
	client := NewClient("test-team", "test-token", mockClient)

	tests := []struct {
		name    string
		setup   func()
		file    string
		wantErr bool
	}{
		{
			name: "正常なレスポンス",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusOK, `{
					"number": 1,
					"name": "テスト記事",
					"full_name": "テスト記事",
					"wip": false,
					"body_md": "テスト本文",
					"body_html": "<p>テスト本文</p>",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z",
					"message": "テストメッセージ",
					"url": "https://test-team.esa.io/posts/1",
					"tags": ["test"],
					"category": "test"
				}`), nil)
			},
			file:    filename,
			wantErr: false,
		},
		{
			name: "記事が見つからない",
			setup: func() {
				mockClient.SetResponse(testutil.CreateMockResponse(t, http.StatusNotFound, `{"error": "Not Found"}`), nil)
			},
			file:    "999-存在しない記事.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			_, err := client.UpdatePost(context.Background(), 1, types.UpdatePostBody{})
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
