#!/bin/sh

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Gitフックのディレクトリを作成
mkdir -p "$PROJECT_ROOT/.git/hooks"

# commit-msgフックをコピー
cp "$SCRIPT_DIR/git-hooks/commit-msg" "$PROJECT_ROOT/.git/hooks/commit-msg"
chmod +x "$PROJECT_ROOT/.git/hooks/commit-msg"

echo "✅ Gitフックがセットアップされました"
echo "コミットメッセージは以下の形式に従ってください："
echo "<絵文字> <型>(<スコープ>): <タイトル>"
echo ""
echo "例: 🐛 fix(api): エラーメッセージの詳細化" 