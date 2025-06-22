package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/shellme/esa-cli/pkg/types"
)

// ListPostsOptions 記事一覧取得のオプション
type ListPostsOptions struct {
	Category string
	Tag      string
	Query    string
	User     string
	Limit    int
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client esa APIクライアント
type Client struct {
	teamName    string
	accessToken string
	client      HTTPDoer
}

// NewClient クライアントを作成
func NewClient(teamName, accessToken string, client HTTPDoer) *Client {
	return &Client{
		teamName:    teamName,
		accessToken: accessToken,
		client:      client,
	}
}

// 接続テスト
func (c *Client) TestConnection() error {
	url := "https://api.esa.io/v1/teams"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// デバッグログ
	fmt.Printf("🔍 リクエストURL: %s\n", url)
	fmt.Printf("🔍 リクエストヘッダー: %v\n", req.Header)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("ネットワークエラー: %v", err)
	}

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close() // 読み取り後にクローズ
	if err != nil {
		return fmt.Errorf("レスポンスの読み取りに失敗: %v", err)
	}

	// デバッグログ
	fmt.Printf("🔍 レスポンスステータス: %d\n", resp.StatusCode)
	fmt.Printf("🔍 レスポンスヘッダー: %v\n", resp.Header)
	fmt.Printf("🔍 レスポンスボディ: %s\n", string(body))

	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("認証エラー: アクセストークンが無効です (レスポンス: %s)", string(body))
	case 404:
		return fmt.Errorf("チームが見つかりません: '%s' は存在しないか、アクセス権限がありません (レスポンス: %s)", c.teamName, string(body))
	default:
		return fmt.Errorf("API エラー (ステータス: %d, レスポンス: %s)", resp.StatusCode, string(body))
	}
}

func (c *Client) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("https://api.esa.io/v1%s", path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

// ListPosts 記事一覧を取得
func (c *Client) ListPosts(ctx context.Context, options *ListPostsOptions) ([]*types.Post, error) {
	baseURL := "https://api.esa.io/v1"
	path := "/teams/" + c.teamName + "/posts"

	// クエリパラメータを構築
	params := []string{}
	if options != nil {
		if options.Category != "" {
			params = append(params, "category="+options.Category)
		}
		if options.Tag != "" {
			params = append(params, "tag="+options.Tag)
		}
		if options.Query != "" {
			params = append(params, "q="+options.Query)
		}
		if options.User != "" {
			params = append(params, "user="+options.User)
		}
		if options.Limit > 0 {
			params = append(params, "per_page="+strconv.Itoa(options.Limit))
		}
	}
	url := baseURL + path
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result struct {
		Posts []*types.Post `json:"posts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Posts, nil
}

// FetchPost 記事を取得
func (c *Client) FetchPost(ctx context.Context, postNum int) (*types.Post, error) {
	baseURL := "https://api.esa.io/v1"
	path := fmt.Sprintf("/teams/%s/posts/%d", c.teamName, postNum)
	url := baseURL + path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var post types.Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, err
	}
	return &post, nil
}

// UpdatePost updates a post on esa.io.
func (c *Client) UpdatePost(ctx context.Context, postNumber int, post types.UpdatePostBody) (*types.Post, error) {
	path := fmt.Sprintf("/teams/%s/posts/%d", c.teamName, postNumber)

	reqBody := types.PostRequest{Post: post}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.makeRequest(http.MethodPatch, path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	var updatedPost types.Post
	if err := json.NewDecoder(resp.Body).Decode(&updatedPost); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &updatedPost, nil
}

// CreatePost creates a new post on esa.io.
func (c *Client) CreatePost(ctx context.Context, post types.CreatePostBody) (*types.Post, error) {
	path := fmt.Sprintf("/teams/%s/posts", c.teamName)

	reqBody := types.CreatePostRequest{Post: post}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := c.makeRequest(http.MethodPost, path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	var createdPost types.Post
	if err := json.NewDecoder(resp.Body).Decode(&createdPost); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &createdPost, nil
}

// BulkUpdateCategory 複数の記事のカテゴリを一括更新
func (c *Client) BulkUpdateCategory(ctx context.Context, postNumbers []int, newCategory string, message string) ([]*types.Post, error) {
	var updatedPosts []*types.Post

	for _, postNumber := range postNumbers {
		post, err := c.UpdatePost(ctx, postNumber, types.UpdatePostBody{
			Category: newCategory,
			Message:  message,
		})
		if err != nil {
			return updatedPosts, fmt.Errorf("記事 %d の更新に失敗: %w", postNumber, err)
		}
		updatedPosts = append(updatedPosts, post)
	}

	return updatedPosts, nil
}

func sanitizeFilename(name string) string {
	// ファイル名に使えない文字を置換
	replacements := map[string]string{
		"/": "-", "\\": "-", ":": "-", "*": "-",
		"?": "-", "\"": "-", "<": "-", ">": "-", "|": "-",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 長すぎる場合は切り詰め
	if len(result) > 100 {
		result = result[:100]
	}

	return result
}
