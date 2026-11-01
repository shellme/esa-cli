# esa-cli

esaの記事をローカルのMarkdownファイルとして編集するGo製CLI。単発のAPI操作は公式の esa CLI / MCP に任せ、ローカル編集と一括操作に絞る。

## コマンド

```bash
go test ./...
go vet ./...
go build -o esa-cli ./cmd/esa-cli
```

コマンドやオプションを変更したら、`cmd/esa-cli/main.go` の `showHelp`、`README.md`、`skills/esa-cli/SKILL.md` も更新する。

## コミットメッセージ

日本語で、次の形式で書く。本文は理由や影響を説明する必要があるときだけ付ける。

```
<絵文字> <型>(<スコープ>): <タイトル>
```

| 型 | 絵文字 | 用途 |
|---|---|---|
| feat | ✨ | 新機能 |
| fix | 🐛 | バグ修正 |
| docs | 📝 | ドキュメント |
| style | 🎨 | コードスタイル |
| refactor | ♻️ | リファクタリング |
| perf | 🚀 | パフォーマンス改善 |
| test | 🧪 | テスト |
| build | 🏗️ | ビルド、バージョン更新 |
| ci | 🔧 | CI設定 |
| chore | 📦 | その他 |
| revert | ⏪ | 取り消し |

スコープは `api`、`config`、`cli`、`cmd`、`internal`、`pkg`、`markdown` から選ぶ。省略してもよい。

## テスト

- 標準の `testing` パッケージを使い、新機能とバグ修正にはテストを追加する
- esa APIは本物を呼ばない。`cmd/esa-cli` は `fakeEsa`、`internal/api` は `internal/api/mock` を使う
- ファイルを書くテストは `chdirTemp` か `t.TempDir` で一時ディレクトリを使い、リポジトリにファイルを残さない
- サブテスト名は動作を説明する自然な文にする

## リリース

`/release` スキルの手順に従う。
