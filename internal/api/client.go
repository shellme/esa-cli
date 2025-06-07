package api

import (
	"bytes"
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

type Client struct {
	BaseURL     string
	AccessToken string
	TeamName    string
	HTTPClient  *http.Client
}

func NewClient(accessToken, teamName string) *Client {
	return &Client{
		BaseURL:     "https://api.esa.io/v1",
		AccessToken: accessToken,
		TeamName:    teamName,
		HTTPClient:  &http.Client{},
	}
}

// 接続テスト
func (c *Client) TestConnection() error {
	url := fmt.Sprintf("%s/teams/%s", c.BaseURL, c.TeamName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ネットワークエラー: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("認証エラー: アクセストークンが無効です")
	case 404:
		return fmt.Errorf("チームが見つかりません: '%s' は存在しないか、アクセス権限がありません", c.TeamName)
	default:
		return fmt.Errorf("API エラー (ステータス: %d)", resp.StatusCode)
	}
}

func (c *Client) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	return c.HTTPClient.Do(req)
}

// 記事一覧を表示
func (c *Client) ListPosts(limit int) error {
	path := fmt.Sprintf("/teams/%s/posts?per_page=%d&sort=updated&order=desc", c.TeamName, limit)

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Posts []types.Post `json:"posts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	fmt.Printf("📚 最新の記事一覧（最大%d件）\n\n", limit)
	for _, post := range result.Posts {
		wipStatus := ""
		if post.WIP {
			wipStatus = " [WIP]"
		}
		fmt.Printf("%d: %s%s\n", post.Number, post.FullName, wipStatus)
	}

	fmt.Printf("\n💡 記事をダウンロード: esa-cli fetch <記事番号>\n")
	return nil
}

// 記事を取得してMarkdownファイルとして保存
func (c *Client) FetchPost(postNumber int) error {
	path := fmt.Sprintf("/teams/%s/posts/%d", c.TeamName, postNumber)

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
	fmt.Printf("🔗 URL: https://%s.esa.io/posts/%d\n", c.TeamName, post.Number)
	return nil
}

// Markdownファイルから記事を更新
func (c *Client) UpdatePost(filename string) error {
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

	// API リクエストを作成
	postReq := types.PostRequest{Post: post}
	jsonData, err := json.Marshal(postReq)
	if err != nil {
		return fmt.Errorf("JSON作成に失敗しました: %v", err)
	}

	path := fmt.Sprintf("/teams/%s/posts/%d", c.TeamName, postNumber)
	resp, err := c.makeRequest("PATCH", path, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("記事番号 %d が見つかりません", postNumber)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("更新に失敗しました (ステータス: %d): %s", resp.StatusCode, string(body))
	}

	var updatedPost types.PostResponse
	if err := json.NewDecoder(resp.Body).Decode(&updatedPost); err != nil {
		return fmt.Errorf("レスポンスの解析に失敗しました: %v", err)
	}

	fmt.Printf("✅ 記事を更新しました: %s\n", updatedPost.FullName)
	fmt.Printf("📅 更新日時: %s\n", updatedPost.UpdatedAt)
	fmt.Printf("🔗 URL: https://%s.esa.io/posts/%d\n", c.TeamName, updatedPost.Number)
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
