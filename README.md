# 🐣 esa-cli

esaの記事をローカルで作成・更新できるCLIツールです。

## 機能

- 記事の一覧表示
- 記事のダウンロード（Markdown形式）
- 記事の作成・更新（ローカルのMarkdownファイルから）
- 記事の一括ダウンロード・一括更新・一括移動
- 更新時の競合チェック（esa側が先に更新されていたら上書き前に確認）
- Front Matter形式でのメタデータ管理（タイトル、カテゴリ、タグ、WIP状態）

## 公式ツールとの使い分け

esa-cliは、記事をローカルのMarkdownファイルとして編集し、esaへ戻す作業に特化しています。

- ローカルで編集する、複数の記事をまとめて扱う場合: esa-cli
- コメント、添付ファイルなど、APIを1操作ずつ使う場合: 公式の [esa CLI](https://github.com/esaio/esa-cli)
- AIアシスタントからesaを直接操作する場合: 公式の [esa MCP Server](https://github.com/esaio/esa-mcp-server)

## 必要条件

- Go 1.21以上（ソースからビルドする場合）
- esaのアカウントとアクセストークン

## インストール方法

### 方法1: Homebrew（独自tap）でインストール（推奨）

```bash
brew tap shellme/esa-cli
brew install esa-cli
```

### 方法2: ソースコードからビルド

1. Go 1.21以上をインストール
2. リポジトリをクローン
```bash
git clone https://github.com/shellme/esa-cli.git
cd esa-cli
```
3. ビルド
```bash
go build -o esa-cli ./cmd/esa-cli
```
4. ビルドしたバイナリをパスの通ったディレクトリに移動
```bash
sudo mv esa-cli /usr/local/bin/
```

## 初期設定

インストール後、最初に1回だけ設定が必要です：

```bash
esa-cli setup
```

設定手順：

1. チーム名（サブドメイン）を入力
   - 例：`my-team`（URLが `https://my-team.esa.io` の場合）
2. アクセストークンを入力
   - https://{your-team}.esa.io/user/applications にアクセス
   - 'Personal access tokens' セクションの 'Generate new token' をクリック
   - Token description に 'esa-cli' と入力
   - Scopes で 'read' と 'write' にチェック
   - 'Generate token' をクリック
   - 表示されたトークンをコピー

設定は `~/.esa-cli-config.json` に保存されます。

## 使用方法

### 記事一覧の表示

```bash
# 最新10件の記事を表示
esa-cli list

# 表示件数を指定
esa-cli list 20

# カテゴリでフィルタリング
esa-cli list -c 開発
esa-cli list --category 開発

# タグでフィルタリング
esa-cli list -t API
esa-cli list --tag API

# 検索ワードでフィルタリング
esa-cli list -q 認証
esa-cli list --query 認証

# 作成者でフィルタリング
esa-cli list -u 自分のユーザー名
esa-cli list --user 自分のユーザー名

# JSON形式で出力（本文は含まない）
esa-cli list 50 --json
```

`-q` にはesaの検索構文をそのまま書けます。`list` で確認した条件は、`fetch-all` にも同じ形で渡せます。

```bash
esa-cli list 50 -q "wip:true"
esa-cli fetch-all -q "wip:true" --dry-run
```

### 記事のダウンロード

```bash
# 記事番号を指定してダウンロード
esa-cli fetch 123

# カテゴリの最新記事をダウンロード
esa-cli fetch -c 開発 -l
esa-cli fetch --category 開発 --latest

# タグの最新記事をダウンロード
esa-cli fetch -t API -l
esa-cli fetch --tag API --latest

# 作成者の最新記事をダウンロード
esa-cli fetch -u 自分のユーザー名 -l
esa-cli fetch --user 自分のユーザー名 --latest

# 検索ワードに一致する最新記事をダウンロード
esa-cli fetch -q 認証 -l

# ファイルに保存せず標準出力に表示
esa-cli fetch 123 --print

# 既存のローカルファイルを上書き
esa-cli fetch 123 --overwrite
```

ダウンロードされたファイルは以下の形式で保存されます：
```
123-article-title.md
```

ファイルの内容：
```markdown
---
title: 記事のタイトル
category: カテゴリ
tags: [tag1, tag2]
wip: false
remote_updated_at: "2025-06-21T09:32:41+09:00"
---

記事の本文...
```

### 記事の更新

```bash
# ローカルのMarkdownファイルから記事を更新
esa-cli update 123-article-title.md

# WIP状態を解除して更新
esa-cli update 123-article-title.md -n
esa-cli update 123-article-title.md --no-wip

# カテゴリを変更して更新
esa-cli update 123-article-title.md -c 開発
esa-cli update 123-article-title.md --category 開発

# タグを追加して更新
esa-cli update 123-article-title.md -a API,認証
esa-cli update 123-article-title.md --add-tags API,認証

# タグを削除して更新
esa-cli update 123-article-title.md -r API
esa-cli update 123-article-title.md --remove-tags API

# メッセージを付けて更新
esa-cli update 123-article-title.md -m API仕様を更新
esa-cli update 123-article-title.md --message API仕様を更新
```

### 記事の作成

```bash
# 記事を作成（デフォルトはWIP）
esa-cli create "記事タイトル" -c 開発 -g API,認証

# ローカルにテンプレートだけを作成
esa-cli create "記事タイトル" -T

# 編集したファイルから作成
esa-cli create -f draft-記事タイトル.md

# 公開状態で作成
esa-cli create "記事タイトル" --no-wip
```

### ローカル画像の自動アップロード

`update`、`update-all`、`create -f` は、本文の `![説明](./images/a.png)` のようなローカル画像をesaへアップロードし、URLに置き換えます。ローカルファイルも書き換わります。

- 対象: png / jpg / jpeg / gif / webp / svg（URLとコード内の記述は対象外）
- パスは記事ファイルからの相対パスで指定します
- 1枚でも失敗すると記事は更新しません。成功した画像はURLに置換済みなので、再実行しても重複しません
- `update-all --dry-run` で対象の画像を確認できます

### 記事の一括ダウンロード・一括更新

```bash
# 対象を確認してから一括ダウンロード（既存ファイルはスキップ）
esa-cli fetch-all -c 開発 --dry-run
esa-cli fetch-all -c 開発

# パターンに一致するファイルをまとめて更新
esa-cli update-all "123-*.md" --dry-run
esa-cli update-all "*.md" -m 一括更新
```

`--force` を付けても、競合時の確認は省略されません。

### 差分と状態の確認

```bash
# ローカルとesaの差分を表示
esa-cli diff 123-article-title.md

# カレントディレクトリの記事ファイルの状態を一覧表示
esa-cli status
esa-cli status "123-*.md"
```

状態は「変更なし」「ローカル変更」「要確認（esa側に更新あり）」のいずれかです。要確認はesa側が更新されていて内容も異なる状態で、ローカルにも変更があるとは限りません。`update-all --dry-run` でも同じ状態を確認できます。

### バックアップ

`fetch --overwrite` と `fetch-all --overwrite` は、上書き前のファイルを `.esa-cli-backup/` に保存します。

### コメント

```bash
# 記事とコメントを標準出力に表示（要約など読み取り用）
esa-cli fetch 123 --print --comments

# コメントを 123-title.comments.md に保存
esa-cli fetch 123 --comments

# コメントを投稿
esa-cli comment 123 --body "コメント本文"
esa-cli comment 123 -f reply.md
cat summary.md | esa-cli comment 123 -f - --force

# 投稿せず内容だけを確認
esa-cli comment 123 -f reply.md --dry-run
```

投稿前に内容を表示して確認します。標準入力から本文を渡す場合は確認に回答できないため、`--force` が必要です。コメントはトークンの持ち主の名義で投稿され、通知も届きます。編集と削除は公式CLIかWeb画面を使ってください。

`--print --comments` の出力は読み取り用です。ファイルに保存して `update` に渡すと、コメントが本文に混ざります。

### 終了コード

| コード | 意味 |
|---|---|
| 0 | 成功 |
| 1 | エラー |
| 2 | `update` / `update-all` / `comment` を確認で中止 |

入力のできない環境では、上書きの確認は「いいえ」として扱います。

### Claude Codeから使う

リポジトリの `skills/esa-cli` をスキルとして導入すると、Claudeが取得、差分確認、更新の流れに沿って操作します。

```bash
mkdir -p ~/.claude/skills && cp -R skills/esa-cli ~/.claude/skills/
```

### 記事の一括移動

```bash
# 特定のカテゴリで自分が作成した記事を一括移動
esa-cli move -c 開発 -o デザイン -u 自分のユーザー名
esa-cli move --category 開発 --to デザイン --user 自分のユーザー名

# 移動せず対象の記事だけを表示
esa-cli move -c 開発 -o デザイン -u 自分のユーザー名 --dry-run

# 確認なしで強制実行
esa-cli move -c 開発 -o デザイン -u 自分のユーザー名 -f
esa-cli move --category 開発 --to デザイン --user 自分のユーザー名 --force

# カスタムメッセージ付きで移動
esa-cli move -c 開発 -o デザイン -u 自分のユーザー名 -m リファクタリング完了
esa-cli move --category 開発 --to デザイン --user 自分のユーザー名 --message リファクタリング完了

# 複数条件で絞り込み
esa-cli move -c 開発 -t API -u 自分のユーザー名 -o ドキュメント
esa-cli move --category 開発 --tag API --user 自分のユーザー名 --to ドキュメント
```

### ヘルプの表示

```bash
esa-cli help
```

## macOSの通知

`fetch-all` と `update-all` が完了すると、macOSの通知が表示されます。

## トラブルシューティング

### よくある質問

Q: インストール時にエラーが発生しました
A: 以下のコマンドで再インストールを試してください：
```bash
brew uninstall esa-cli
brew install shellme/esa-cli
```

Q: 記事の更新が反映されません
A: 以下の点を確認してください：
- ファイル名が正しいか（記事番号-タイトル.md）
- Front Matterの形式が正しいか
- インターネット接続が安定しているか

Q: アクセストークンが無効になりました
A: 新しいトークンを生成し、`esa-cli setup` を再実行してください。

Q: `esa-cli` コマンドが見つかりません
A: `which esa-cli` で確認し、インストール先（Homebrewなら `brew --prefix` の `bin`）を `PATH` に追加してください。

### サポート

問題が解決しない場合は、以下の方法でサポートを受けられます：

1. GitHub Issuesで報告
2. プルリクエストの送信

## ライセンス

MIT License

## 作者

[shellme](https://github.com/shellme)

## 開発者向け

開発ルール、コミット規約、リリース手順は [CLAUDE.md](CLAUDE.md) にまとめています。

```bash
go test ./...
go build -o esa-cli ./cmd/esa-cli
```
