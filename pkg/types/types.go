package types

// Post esaの記事を表す構造体
type Post struct {
	Number       int      `json:"number"`
	Name         string   `json:"name"`
	BodyMd       string   `json:"body_md"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	WIP          bool     `json:"wip"`
	Message      string   `json:"message"`
	FullName     string   `json:"full_name"`
	UpdatedAt    string   `json:"updated_at"`
}

// PostResponse APIレスポンス用の構造体
type PostResponse struct {
	Post
}

// PostRequest APIリクエスト用の構造体
type PostRequest struct {
	Post Post `json:"post"`
} 