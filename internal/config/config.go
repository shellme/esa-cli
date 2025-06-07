package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AccessToken string `json:"access_token"`
	TeamName    string `json:"team_name"`
}

// 設定ファイルのパスを取得
func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".esa-cli-config.json")
}

// 設定を読み込み
func Load() (*Config, error) {
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// 設定を保存
func Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0600)
}

// APIクライアントのインターフェース
type APIClient interface {
	TestConnection() error
}

// 初期設定コマンド
func Setup(client APIClient) error {
	config, err := Load()
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗しました: %v", err)
	}

	fmt.Println("🔧 esa-cli 初期設定")
	fmt.Println("")
	fmt.Println("📋 以下の手順でアクセストークンを取得してください：")
	fmt.Println("1. https://{your-team}.esa.io/user/applications にアクセス")
	fmt.Println("2. 'Personal access tokens' セクションの 'Generate new token' をクリック")
	fmt.Println("3. Token description に 'esa-cli' と入力")
	fmt.Println("4. Scopes で 'read' と 'write' にチェック")
	fmt.Println("5. 'Generate token' をクリック")
	fmt.Println("6. 表示されたトークンをコピー（画面を閉じると再表示できません）")
	fmt.Println("")

	fmt.Print("チーム名（サブドメイン）を入力: ")
	fmt.Scanln(&config.TeamName)

	fmt.Print("アクセストークンを入力: ")
	fmt.Scanln(&config.AccessToken)

	// 設定をテスト
	fmt.Println("")
	fmt.Println("🧪 設定をテスト中...")
	if err := client.TestConnection(); err != nil {
		return fmt.Errorf("接続テストに失敗しました: %v\n\nトークンやチーム名を確認してください", err)
	}

	if err := Save(config); err != nil {
		return fmt.Errorf("設定の保存に失敗しました: %v", err)
	}

	fmt.Println("✅ 設定が完了しました！")
	fmt.Println("")
	fmt.Println("🚀 使用方法:")
	fmt.Println("  esa-cli fetch 123      # 記事番号123をダウンロード")
	fmt.Println("  esa-cli update file.md # 記事を更新")
	fmt.Println("  esa-cli list           # 記事一覧を表示")
	
	return nil
} 