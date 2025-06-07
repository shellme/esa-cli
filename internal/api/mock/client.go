package mock

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

// MockHTTPClient HTTPクライアントのモック
type MockHTTPClient struct {
	mu       sync.Mutex
	requests []*http.Request
	response *http.Response
	err      error
}

// NewMockHTTPClient モックHTTPクライアントを作成
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{}
}

// SetResponse レスポンスを設定
func (m *MockHTTPClient) SetResponse(resp *http.Response, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = resp
	m.err = err
}

// GetRequests リクエスト履歴を取得
func (m *MockHTTPClient) GetRequests() []*http.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests
}

// Do リクエストを実行（モック）
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)
	return m.response, m.err
}

// NewReadCloser 新しいReadCloserを作成
func NewReadCloser(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewBuffer(data))
}
