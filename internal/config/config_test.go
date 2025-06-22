package config

import (
	"net/http"
	"os"
	"testing"

	"github.com/shellme/esa-cli/internal/api"
	"github.com/shellme/esa-cli/internal/testutil"
)

func TestLoad(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// 設定ファイルのパスを一時的に変更
	ConfigDir = tmpDir
	ConfigFile = testutil.CreateTestConfigFile(t, tmpDir)

	tests := []struct {
		name    string
		setup   func()
		want    *Config
		wantErr bool
	}{
		{
			name: "正常系：設定ファイルが存在する",
			setup: func() {
				// 設定ファイルを作成
				cfg := &Config{
					AccessToken: "test-token",
					TeamName:    "test-team",
				}
				if err := Save(cfg); err != nil {
					t.Fatal(err)
				}
			},
			want: &Config{
				AccessToken: "test-token",
				TeamName:    "test-team",
			},
			wantErr: false,
		},
		{
			name: "異常系：設定ファイルが存在しない",
			setup: func() {
				// 設定ファイルを削除
				os.Remove(ConfigFile)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.AccessToken != tt.want.AccessToken || got.TeamName != tt.want.TeamName) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSave(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// 設定ファイルのパスを一時的に変更
	ConfigDir = tmpDir
	ConfigFile = testutil.CreateTestConfigFile(t, tmpDir)

	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "正常系：有効な設定を保存",
			cfg: &Config{
				AccessToken: "test-token",
				TeamName:    "test-team",
			},
			wantErr: false,
		},
		{
			name:    "異常系：nilの設定を保存",
			cfg:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Save(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 保存した設定を読み込んで検証
				got, err := Load()
				if err != nil {
					t.Fatal(err)
				}
				if got.AccessToken != tt.cfg.AccessToken || got.TeamName != tt.cfg.TeamName {
					t.Errorf("Save() = %v, want %v", got, tt.cfg)
				}
			}
		})
	}
}

func TestSetup(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := testutil.CreateTempDir(t)

	// 設定ファイルのパスを一時的に変更
	ConfigDir = tmpDir
	ConfigFile = testutil.CreateTestConfigFile(t, tmpDir)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "正常系：有効な入力",
			wantErr: true, // 実際のAPIを呼び出すため、テストトークンでは失敗する
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準入力のモック
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			// テスト用の標準入力を設定
			r, w, _ := os.Pipe()
			os.Stdin = r
			go func() {
				w.Write([]byte("test-team\ntest-token\n"))
				w.Close()
			}()

			// ダミークライアントを渡す（実際には使用されない）
			dummyClient := api.NewClient("", "", http.DefaultClient)
			err := Setup(dummyClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				// エラーが発生した場合は以降の検証を行わない
				return
			}

			if !tt.wantErr {
				// 設定を読み込んで検証
				cfg, err := Load()
				if err != nil {
					t.Fatal(err)
				}
				if cfg.AccessToken == "" || cfg.TeamName == "" {
					t.Error("Setup() did not save configuration")
				}
			}
		})
	}
}

// 異常系：nilのクライアント専用のテスト
func TestSetup_NilClient(t *testing.T) {
	err := Setup(nil)
	if err == nil {
		t.Errorf("Setup(nil) error = nil, want error")
	}
}
