package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shellme/esa-cli/internal/markdown"
	"github.com/shellme/esa-cli/pkg/types"
)

// ListPostsOptions 記事一覧取得のオプション
type ListPostsOptions struct {
	Category string
	Tag      string
	Query    string
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

	req.Header.Set("Authorization", "Bearer "+c.teamName)

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

	req.Header.Set("Authorization", "Bearer "+c.teamName)
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

// 記事を取得してMarkdownファイルとして保存
func (c *Client) FetchPostOld(postNumber int) error {
	path := fmt.Sprintf("/teams/%s/posts/%d", c.teamName, postNumber)

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("記事番号 %d が見つかりません", postNumber)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var post types.PostResponse
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// ファイル名を生成
	filename := fmt.Sprintf("%d-%s.md", post.Number, sanitizeFilename(post.Name))

	// Markdownファイルとして保存
	content := markdown.GenerateContent(post.Post)
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("✅ 記事を保存しました: %s\n", filename)
	fmt.Printf("📖 タイトル: %s\n", post.FullName)
	fmt.Printf("🔗 URL: https://%s.esa.io/posts/%d\n", c.teamName, post.Number)
	return nil
}

// UpdatePostOptions 記事更新のオプション
type UpdatePostOptions struct {
	NoWip      bool
	Category   string
	AddTags    []string
	RemoveTags []string
	Message    string
}

// Markdownファイルから記事を更新
func (c *Client) UpdatePost(filename string, options *UpdatePostOptions) error {
	// ファイル存在確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが見つかりません: %s", filename)
	}

	// ファイル名から記事番号を抽出
	base := filepath.Base(filename)
	parts := strings.SplitN(base, "-", 2)
	if len(parts) < 2 {
		return fmt.Errorf("ファイル名の形式が正しくありません。期待する形式: {記事番号}-{タイトル}.md")
	}

	postNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("ファイル名から記事番号を読み取れません: %v", err)
	}

	// ファイルを読み込み
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %v", err)
	}

	// Markdownからメタデータと本文を分離
	post, err := markdown.ParseContent(string(content))
	if err != nil {
		return fmt.Errorf("Markdownの解析に失敗しました: %v", err)
	}

	// 更新オプションを適用
	if options != nil {
		if options.NoWip {
			post.Wip = false
		}
		if options.Category != "" {
			post.Category = options.Category
		}
		if len(options.AddTags) > 0 {
			post.Tags = append(post.Tags, options.AddTags...)
		}
		if len(options.RemoveTags) > 0 {
			tags := make([]string, 0, len(post.Tags))
			for _, tag := range post.Tags {
				remove := false
				for _, removeTag := range options.RemoveTags {
					if tag == removeTag {
						remove = true
						break
					}
				}
				if !remove {
					tags = append(tags, tag)
				}
			}
			post.Tags = tags
		}
	}

	// 記事を更新
	path := fmt.Sprintf("/teams/%s/posts/%d", c.teamName, postNumber)
	body := struct {
		Post struct {
			Name     string   `json:"name"`
			Category string   `json:"category"`
			Tags     []string `json:"tags"`
			Wip      bool     `json:"wip"`
			BodyMD   string   `json:"body_md"`
			Message  string   `json:"message,omitempty"`
		} `json:"post"`
	}{
		Post: struct {
			Name     string   `json:"name"`
			Category string   `json:"category"`
			Tags     []string `json:"tags"`
			Wip      bool     `json:"wip"`
			BodyMD   string   `json:"body_md"`
			Message  string   `json:"message,omitempty"`
		}{
			Name:     post.Name,
			Category: post.Category,
			Tags:     post.Tags,
			Wip:      post.Wip,
			BodyMD:   post.BodyMD,
			Message:  options.Message,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("JSONの生成に失敗しました: %v", err)
	}

	resp, err := c.makeRequest("PATCH", path, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	return nil
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
