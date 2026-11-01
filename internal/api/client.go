package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shellme/esa-cli/pkg/types"
)

// ListPostsOptions 記事一覧取得のオプション
type ListPostsOptions struct {
	Category string
	Tag      string
	Query    string
	User     string
	Limit    int // per_pageパラメータ（最大100）
	Page     int // pageパラメータ（1から始まる）
}

const maxPerPage = 100

type listResponse struct {
	Posts      []*types.Post `json:"posts"`
	NextPage   *int          `json:"next_page"`
	TotalCount int           `json:"total_count"`
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
		client:      &retryDoer{inner: client, sleep: time.Sleep, now: time.Now},
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
	result, err := c.listPage(ctx, options)
	if err != nil {
		return nil, err
	}
	return result.Posts, nil
}

// ListAllPosts 全ページをたどって記事を取得し、取得件数とは別に該当総数も返す
// maxPosts件に達したら取得を打ち切る（0以下は無制限）
func (c *Client) ListAllPosts(ctx context.Context, options ListPostsOptions, maxPosts int) ([]*types.Post, int, error) {
	options.Limit = maxPerPage
	options.Page = 1

	var posts []*types.Post
	for {
		result, err := c.listPage(ctx, &options)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, result.Posts...)
		if maxPosts > 0 && len(posts) >= maxPosts {
			return posts[:maxPosts], result.TotalCount, nil
		}
		if result.NextPage == nil {
			return posts, result.TotalCount, nil
		}
		options.Page = *result.NextPage
	}
}

// InCategory 記事が指定カテゴリまたはそのサブカテゴリにあるかを判定
func InCategory(post *types.Post, category string) bool {
	return post.Category == category || strings.HasPrefix(post.Category, category+"/")
}

func categoryQuery(category string) string {
	if strings.ContainsAny(category, " \t") {
		return fmt.Sprintf("in:%q", category)
	}
	return "in:" + category
}

func (c *Client) listPage(ctx context.Context, options *ListPostsOptions) (*listResponse, error) {
	baseURL := "https://api.esa.io/v1"
	path := "/teams/" + c.teamName + "/posts"

	// クエリパラメータを構築
	queryParams := url.Values{}
	if options != nil {
		// カテゴリ、タグ、ユーザー、クエリを組み合わせてqパラメータに設定
		// esa.ioのAPIでは:
		// - カテゴリ検索は in:カテゴリ名 の形式（サブカテゴリを含む）
		// - タグ検索は #タグ名 の形式
		// - ユーザー検索は @ユーザー名 の形式
		qParts := []string{}
		if options.Category != "" {
			qParts = append(qParts, categoryQuery(options.Category))
		}
		if options.Tag != "" {
			qParts = append(qParts, "#"+options.Tag)
		}
		if options.User != "" {
			qParts = append(qParts, "@"+options.User)
		}
		if options.Query != "" {
			qParts = append(qParts, options.Query)
		}
		if len(qParts) > 0 {
			queryParams.Set("q", strings.Join(qParts, " "))
		}
		if options.Limit > 0 {
			queryParams.Set("per_page", strconv.Itoa(options.Limit))
		}
		if options.Page > 0 {
			queryParams.Set("page", strconv.Itoa(options.Page))
		}
	}
	apiURL := baseURL + path
	if len(queryParams) > 0 {
		apiURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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

	var result listResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
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

// Comment 記事のコメント
type Comment struct {
	BodyMd    string    `json:"body_md"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy struct {
		ScreenName string `json:"screen_name"`
	} `json:"created_by"`
}

// FetchComments 記事のコメントを古い順にすべて取得する
func (c *Client) FetchComments(ctx context.Context, postNumber int) ([]*Comment, error) {
	var comments []*Comment
	for page := 1; ; {
		apiURL := fmt.Sprintf("https://api.esa.io/v1/teams/%s/posts/%d/comments?page=%d&per_page=%d&sort=created&order=asc", c.teamName, postNumber, page, maxPerPage)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.accessToken)

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("API request failed: %w", err)
		}
		var result struct {
			Comments []*Comment `json:"comments"`
			NextPage *int       `json:"next_page"`
		}
		if resp.StatusCode == http.StatusOK {
			err = json.NewDecoder(resp.Body).Decode(&result)
		} else {
			err = fmt.Errorf("API returned status: %s", resp.Status)
		}
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		comments = append(comments, result.Comments...)
		if result.NextPage == nil {
			return comments, nil
		}
		page = *result.NextPage
	}
}

// CreateComment 記事にコメントを投稿する
func (c *Client) CreateComment(ctx context.Context, postNumber int, body string) error {
	jsonBody, err := json.Marshal(map[string]interface{}{"comment": map[string]string{"body_md": body}})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	path := fmt.Sprintf("/teams/%s/posts/%d/comments", c.teamName, postNumber)
	resp, err := c.makeRequest(http.MethodPost, path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status: %s", resp.Status)
	}
	return nil
}

// UploadAttachment ファイルをアップロードし、記事に貼れるURLを返す
func (c *Client) UploadAttachment(ctx context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)
	part, err := form.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	if err := form.Close(); err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("https://api.esa.io/v1/teams/%s/attachments", c.teamName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("API returned status: %s（トークンに添付ファイルの書き込み権限が必要です）", resp.Status)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("API returned status: %s", resp.Status)
	}

	var result struct {
		Attachment struct {
			URL string `json:"url"`
		} `json:"attachment"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if result.Attachment.URL == "" {
		return "", fmt.Errorf("レスポンスに添付ファイルのURLがありません")
	}
	return result.Attachment.URL, nil
}

// BulkUpdateCategory 複数の記事のカテゴリを一括更新
func (c *Client) BulkUpdateCategory(ctx context.Context, postNumbers []int, newCategory string, message string) ([]*types.Post, error) {
	var updatedPosts []*types.Post

	for _, postNumber := range postNumbers {
		post, err := c.UpdatePost(ctx, postNumber, types.UpdatePostBody{
			// wipやtagsは送らず、カテゴリ以外の状態を変えない
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
