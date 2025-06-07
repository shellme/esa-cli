package mock

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/shellme/esa-cli/internal/testutil"
)

func TestMockHTTPClient(t *testing.T) {
	client := NewMockHTTPClient()

	tests := []struct {
		name     string
		response *http.Response
		err      error
	}{
		{
			name:     "正常なレスポンス",
			response: testutil.CreateMockResponse(t, http.StatusOK, `{"message": "success"}`),
			err:      nil,
		},
		{
			name:     "エラーレスポンス",
			response: testutil.CreateMockResponse(t, http.StatusInternalServerError, `{"message": "error"}`),
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.SetResponse(tt.response, tt.err)

			resp, err := client.Do(&http.Request{})
			if err != tt.err {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.err)
			}
			if resp != tt.response {
				t.Errorf("Do() response = %v, want %v", resp, tt.response)
			}
		})
	}
}

func TestMockHTTPClient_Do(t *testing.T) {
	tests := []struct {
		name     string
		client   *MockHTTPClient
		req      *http.Request
		wantResp *http.Response
		wantErr  bool
	}{
		{
			name: "正常なレスポンス",
			client: &MockHTTPClient{
				response: testutil.CreateMockResponse(t, http.StatusOK, `{"message": "success"}`),
				err:      nil,
			},
			req: &http.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/v1/teams/test-team/posts"},
			},
			wantResp: testutil.CreateMockResponse(t, http.StatusOK, `{"message": "success"}`),
			wantErr:  false,
		},
		{
			name: "エラーレスポンス",
			client: &MockHTTPClient{
				response: nil,
				err:      errors.New("mock error"),
			},
			req: &http.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/v1/teams/test-team/posts"},
			},
			wantResp: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Do(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MockHTTPClient.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.StatusCode != tt.wantResp.StatusCode {
					t.Errorf("StatusCode = %v, want %v", got.StatusCode, tt.wantResp.StatusCode)
				}
				gotBody, _ := io.ReadAll(got.Body)
				wantBody, _ := io.ReadAll(tt.wantResp.Body)
				if string(gotBody) != string(wantBody) {
					t.Errorf("Body = %v, want %v", string(gotBody), string(wantBody))
				}
			}
		})
	}
}

func TestMockHTTPClient_SetResponse(t *testing.T) {
	client := &MockHTTPClient{}
	resp := testutil.CreateMockResponse(t, http.StatusOK, `{"message": "success"}`)
	err := errors.New("mock error")

	client.SetResponse(resp, err)

	if client.response != resp {
		t.Errorf("response = %v, want %v", client.response, resp)
	}
	if client.err != err {
		t.Errorf("err = %v, want %v", client.err, err)
	}
}
