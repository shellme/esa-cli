package types

// Post 記事の型定義
type Post struct {
	Number   int      `json:"number"`
	Name     string   `json:"name"`
	FullName string   `json:"full_name"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Wip      bool     `json:"wip"`
	BodyMD   string   `json:"body_md"`
}

// PostResponse APIレスポンス用の構造体
type PostResponse struct {
	Post
}

// PostRequest APIリクエスト用の構造体
type PostRequest struct {
	Post Post `json:"post"`
}
