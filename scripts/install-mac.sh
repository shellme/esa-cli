#!/bin/bash

set -e

echo "🐣 esa-cli for Mac - インストーラー"
echo "=================================="

# 管理者権限チェック（不要にする）
echo "✅ 管理者権限は不要です"

# アーキテクチャの判定
ARCH=$(uname -m)
if [[ "$ARCH" == "arm64" ]]; then
    BINARY="esa-cli-darwin-arm64"
    INSTALL_DIR="/opt/homebrew/bin"
else
    BINARY="esa-cli-darwin-amd64"
    INSTALL_DIR="/usr/local/bin"
fi

# インストール先のディレクトリが存在しない場合は作成
if [ ! -d "$INSTALL_DIR" ]; then
    echo "📁 インストール先のディレクトリを作成中..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo chown $(whoami) "$INSTALL_DIR"
fi

# バイナリのダウンロード
echo "📦 esa-cliをダウンロード中..."
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

curl -L -o "$BINARY" "https://github.com/shellme/esa-cli/releases/latest/download/$BINARY"
chmod +x "$BINARY"

# インストール
echo "📥 インストール中..."
mv "$BINARY" "$INSTALL_DIR/esa-cli"

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