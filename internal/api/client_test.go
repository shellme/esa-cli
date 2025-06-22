package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/shellme/esa-cli/internal/api/mock"
	"github.com/shellme/esa-cli/internal/testutil"
	"github.com/shellme/esa-cli/pkg/types"
)

func TestListPosts(t *testing.T) {
	tests := []struct {
		name  string
		given struct {
			teamName string
			options  *ListPostsOptions
			response *http.Response
			err      error
		}
		when struct {
			ctx context.Context
		}
		then struct {
			wantPosts []*types.Post
			wantErr   bool
		}
	}{
		{
			name: "正常系：記事一覧を取得できる",
			given: struct {
				teamName string
				options  *ListPostsOptions
				response *http.Response
				err      error
			}{
				teamName: "test-team",
				options:  &ListPostsOptions{},
				response: testutil.CreateMockResponse(t, http.StatusOK, testutil.CreateTestPostsResponse(t)),
				err:      nil,
			},
			when: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			then: struct {
				wantPosts []*types.Post
				wantErr   bool
			}{
				wantPosts: []*types.Post{
					{
						Number:   1,
						Name:     "テスト記事1",
						FullName: "テスト記事1",
						Category: "test",
						Tags:     []string{"test"},
						Wip:      false,
						BodyMd:   "テスト本文1",
					},
					{
						Number:   2,
						Name:     "テスト記事2",
						FullName: "テスト記事2",
						Category: "test",
						Tags:     []string{"test"},
						Wip:      true,
						BodyMd:   "テスト本文2",
					},
				},
				wantErr: false,
			},
		},
		{
			name: "異常系：APIエラーが発生した場合",
			given: struct {
				teamName string
				options  *ListPostsOptions
				response *http.Response
				err      error
			}{
				teamName: "test-team",
				options:  &ListPostsOptions{},
				response: nil,
				err:      errors.New("API error"),
			},
			when: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			then: struct {
				wantPosts []*types.Post
				wantErr   bool
			}{
				wantPosts: nil,
				wantErr:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockClient := mock.NewMockHTTPClient()
			mockClient.SetResponse(tt.given.response, tt.given.err)
			client := NewClient("test-team", "test-token", mockClient)

			// When
			got, err := client.ListPosts(tt.when.ctx, tt.given.options)

			// Then
			if (err != nil) != tt.then.wantErr {
				t.Errorf("ListPosts() error = %v, wantErr %v", err, tt.then.wantErr)
				return
			}
			if !tt.then.wantErr {
				if len(got) != len(tt.then.wantPosts) {
					t.Errorf("ListPosts() got %d posts, want %d", len(got), len(tt.then.wantPosts))
					return
				}
				for i, want := range tt.then.wantPosts {
					if got[i].Number != want.Number {
						t.Errorf("ListPosts()[%d].Number = %v, want %v", i, got[i].Number, want.Number)
					}
					if got[i].Name != want.Name {
						t.Errorf("ListPosts()[%d].Name = %v, want %v", i, got[i].Name, want.Name)
					}
				}
			}
		})
	}
}

func TestFetchPost(t *testing.T) {
	tests := []struct {
		name  string
		given struct {
			teamName string
			postNum  int
			response *http.Response
			err      error
		}
		when struct {
			ctx context.Context
		}
		then struct {
			wantPost *types.Post
			wantErr  bool
		}
	}{
		{
			name: "正常系：記事を取得できる",
			given: struct {
				teamName string
				postNum  int
				response *http.Response
				err      error
			}{
				teamName: "test-team",
				postNum:  1,
				response: testutil.CreateMockResponse(t, http.StatusOK, `{
					"number": 1,
					"name": "テスト記事",
					"full_name": "テスト記事",
					"category": "test",
					"tags": ["test"],
					"wip": false,
					"body_md": "テスト本文"
				}`),
				err: nil,
			},
			when: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			then: struct {
				wantPost *types.Post
				wantErr  bool
			}{
				wantPost: &types.Post{
					Number:   1,
					Name:     "テスト記事",
					FullName: "テスト記事",
					Category: "test",
					Tags:     []string{"test"},
					Wip:      false,
					BodyMd:   "テスト本文",
				},
				wantErr: false,
			},
		},
		{
			name: "異常系：記事が存在しない場合",
			given: struct {
				teamName string
				postNum  int
				response *http.Response
				err      error
			}{
				teamName: "test-team",
				postNum:  999,
				response: testutil.CreateMockResponse(t, http.StatusNotFound, `{"message": "Not Found"}`),
				err:      nil,
			},
			when: struct {
				ctx context.Context
			}{
				ctx: context.Background(),
			},
			then: struct {
				wantPost *types.Post
				wantErr  bool
			}{
				wantPost: nil,
				wantErr:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockClient := mock.NewMockHTTPClient()
			mockClient.SetResponse(tt.given.response, tt.given.err)
			client := NewClient("test-team", "test-token", mockClient)

			// When
			got, err := client.FetchPost(tt.when.ctx, tt.given.postNum)

			// Then
			if (err != nil) != tt.then.wantErr {
				t.Errorf("FetchPost() error = %v, wantErr %v", err, tt.then.wantErr)
				return
			}
			if !tt.then.wantErr {
				if got.Number != tt.then.wantPost.Number {
					t.Errorf("FetchPost().Number = %v, want %v", got.Number, tt.then.wantPost.Number)
				}
				if got.Name != tt.then.wantPost.Name {
					t.Errorf("FetchPost().Name = %v, want %v", got.Name, tt.then.wantPost.Name)
				}
			}
		})
	}
}
