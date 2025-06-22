package types

import "time"

// Post a struct for a post returned by the API
type Post struct {
	Number    int       `json:"number"`
	Name      string    `json:"name"`
	FullName  string    `json:"full_name"`
	Wip       bool      `json:"wip"`
	BodyMd    string    `json:"body_md"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
}

// PostResponse is a struct for API response for getting a post
type PostResponse struct {
	Post
}

// UpdatePostBody is a struct for the body of a post to be updated
type UpdatePostBody struct {
	Name     string   `json:"name,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	BodyMd   string   `json:"body_md,omitempty"`
	Wip      bool     `json:"wip"`
	Message  string   `json:"message,omitempty"`
}

// PostRequest is a struct for API request for updating a post
type PostRequest struct {
	Post UpdatePostBody `json:"post"`
}

// CreatePostBody is a struct for the body of a post to be created
type CreatePostBody struct {
	Name     string   `json:"name,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	BodyMd   string   `json:"body_md,omitempty"`
	Wip      bool     `json:"wip"`
	Message  string   `json:"message,omitempty"`
}

// CreatePostRequest is a struct for API request for creating a post
type CreatePostRequest struct {
	Post CreatePostBody `json:"post"`
}

// FrontMatter a struct for a post's front matter
type FrontMatter struct {
	Title           string   `yaml:"title"`
	Category        string   `yaml:"category"`
	Tags            []string `yaml:"tags"`
	Wip             bool     `yaml:"wip"`
	RemoteUpdatedAt string   `yaml:"remote_updated_at,omitempty"`
}
