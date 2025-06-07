#!/bin/bash

set -e

echo "🍎 esa-cli for Mac - インストーラー"
echo "=================================="

# 管理者権限チェック（不要にする）
echo "✅ 管理者権限は不要です"

# Homebrewの確認・インストール
if ! command -v brew &> /dev/null; then
    echo "🍺 Homebrewをインストール中..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # PATHを更新
    if [[ $(uname -m) == "arm64" ]]; then
        echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
        eval "$(/opt/homebrew/bin/brew shellenv)"
    else
        echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile  
        eval "$(/usr/local/bin/brew shellenv)"
    fi
fi

# リポジトリをクローン
echo "📦 esa-cliをインストール中..."
TEMP_DIR=$(mktemp -d)
git clone https://github.com/shellme/esa-cli.git "$TEMP_DIR"
cd "$TEMP_DIR"

# ビルド
echo "🔨 ビルド中..."
go build -o esa-cli cmd/esa-cli/main.go

# インストール
echo "📥 インストール中..."
sudo mv esa-cli /usr/local/bin/

# クリーンアップ
cd - > /dev/null
rm -rf "$TEMP_DIR"

# 成功メッセージ
echo ""
echo "🎉 インストール完了！"
echo ""
echo "📋 次のステップ:"
echo "1. 新しいターミナルを開く"
echo "2. esa-cli setup で初期設定"
echo "3. esa-cli list で動作確認"
echo ""
echo "💡 詳しい使い方: https://github.com/shellme/esa-cli"
echo "🆘 困ったときは: GitHub Issues"

# ターミナル再起動の促し
echo ""
echo "⚠️  ターミナルを再起動してください"
echo "   または 'source ~/.zprofile' を実行" 