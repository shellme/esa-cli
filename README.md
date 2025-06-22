# 🐣 esa-cli

esaの記事をローカルで作成・更新できるCLIツールです。

📚 [オンラインドキュメント](https://shellme.github.io/esa-cli/)

## 機能

- 記事の一覧表示
- 記事のダウンロード（Markdown形式）
- 記事の更新（ローカルのMarkdownファイルから）
- Front Matter形式でのメタデータ管理（タイトル、カテゴリ、タグ、WIP状態）

## 必要条件

- Go 1.16以上
- esaのアカウントとアクセストークン

## インストール方法

### 方法1: Homebrew（独自tap）でインストール（推奨）

```bash
brew tap shellme/esa-cli
brew install esa-cli
```

### 方法2: 直接バイナリをダウンロード

1. [GitHubのリリースページ](https://github.com/shellme/esa-cli/releases)から最新のリリースをダウンロード
2. お使いのOSに合わせて適切なバイナリを選択
   - macOS (Intel): `esa-cli-darwin-amd64`
   - macOS (Apple Silicon): `esa-cli-darwin-arm64`
   - Linux (Intel): `esa-cli-linux-amd64`
   - Linux (ARM): `esa-cli-linux-arm64`
3. ダウンロードしたバイナリを実行可能にして、パスの通ったディレクトリに移動

```bash
# macOS (Apple Silicon) の場合
curl -L -o esa-cli https://github.com/shellme/esa-cli/releases/latest/download/esa-cli-darwin-arm64
chmod +x esa-cli
sudo mv esa-cli /usr/local/bin/
```

#### macOSでの追加設定

macOSでは、デフォルトのシェルがzshであるため、以下の設定が必要な場合があります：

1. パスの確認
```bash
echo $PATH
```

2. 必要に応じて、`~/.zshrc` または `~/.zprofile` に以下を追加
```bash
export PATH="/usr/local/bin:$PATH"
```

3. 設定の反映
```bash
source ~/.zshrc  # または source ~/.zprofile
```

4. インストールの確認
```bash
which esa-cli
```

#### fishシェルを使用している場合

fishシェルを使用している場合は、以下の手順で設定を行ってください：

1. パスの確認
```fish
echo $PATH
```

2. 必要に応じて、`~/.config/fish/config.fish` に以下を追加
```fish
set -gx PATH /usr/local/bin $PATH
```

3. 設定の反映
```fish
source ~/.config/fish/config.fish
```

4. インストールの確認
```fish
which esa-cli
```

### 方法3: ソースコードからビルド

1. Go 1.16以上をインストール
2. リポジトリをクローン
```bash
git clone https://github.com/shellme/esa-cli.git
cd esa-cli
```
3. ビルド
```bash
go build -o esa-cli
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

# メッセージを付けて更新
esa-cli update 123-article-title.md -m API仕様を更新
esa-cli update 123-article-title.md --message API仕様を更新
```

### 記事の一括移動

```bash
# 特定のカテゴリで自分が作成した記事を一括移動
esa-cli move -c 開発 -o デザイン -u 自分のユーザー名
esa-cli move --category 開発 --to デザイン --user 自分のユーザー名

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

## Mac特有の便利な使い方

### Finderとの連携

```bash
# 記事をダウンロードしてFinderで開く
esa-cli fetch 123 && open .
```

### お気に入りエディタで直接開く

```bash
# VS Codeで開く
esa-cli fetch 123 && code 123-*.md

# Typoraで開く
esa-cli fetch 123 && open -a Typora 123-*.md
```

### 通知機能

記事のダウンロードや更新が完了すると、macOSの通知が表示されます。

### Spotlight検索の活用

ダウンロードした記事には、Spotlight検索用のメタデータが自動的に追加されます。
Finderの検索バーで記事のタイトルやカテゴリを検索できます。

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
A: 以下の点を確認してください：

1. パスが正しく設定されているか確認
```bash
echo $PATH
```

2. `/usr/local/bin` がパスに含まれているか確認

3. 使用しているシェルに応じて、以下の設定ファイルにパスを追加：

- bashの場合（`~/.bashrc` または `~/.bash_profile`）:
```bash
export PATH="/usr/local/bin:$PATH"
```

- zshの場合（`~/.zshrc` または `~/.zprofile`）:
```bash
export PATH="/usr/local/bin:$PATH"
```

- fishの場合（`~/.config/fish/config.fish`）:
```fish
set -gx PATH /usr/local/bin $PATH
```

4. 設定を反映
```bash
# bashの場合
source ~/.bashrc  # または source ~/.bash_profile

# zshの場合
source ~/.zshrc  # または source ~/.zprofile

# fishの場合
source ~/.config/fish/config.fish
```

### サポート

問題が解決しない場合は、以下の方法でサポートを受けられます：

1. GitHub Issuesで報告
2. プルリクエストの送信

## ライセンス

MIT License

## 作者

[shellme](https://github.com/shellme)

## 開発者向け

### 開発環境のセットアップ

1. リポジトリのクローン
```bash
git clone https://github.com/shellme/esa-cli.git
cd esa-cli
```

2. 依存関係のインストール
```bash
go mod download
```

3. Gitフックのセットアップ
```bash
./scripts/setup-git-hooks.sh
```

### コミットメッセージのテンプレート設定

このプロジェクトでは、コミットメッセージのテンプレートを提供しています。
以下のコマンドで設定できます：

```bash
# テンプレートの設定
git config --local commit.template .git/commit-template.txt

# 設定の確認
git config --local commit.template
```

テンプレートには以下の情報が含まれています：
- コミットメッセージの形式
- 使用可能な絵文字と型の組み合わせ
- 使用可能なスコープ
- コミットメッセージの書き方の例
- 言語に関する注意

### コミットメッセージのルール

コミットメッセージは以下の形式に従ってください：

```
<絵文字> <型>(<スコープ>): <タイトル>
```

#### 使用可能な絵文字と型の組み合わせ

| 型 | 絵文字 | 説明 | 例 |
|----|--------|------|-----|
| feat | ✨ | 新機能の追加 | ✨ feat(cli): 新しいコマンドを追加 |
| fix | 🐛 | バグ修正 | 🐛 fix(api): エラーメッセージの詳細化 |
| docs | 📝 | ドキュメントの更新 | 📝 docs: READMEの更新 |
| style | 🎨 | コードスタイルの修正 | 🎨 style: インデントの修正 |
| refactor | ♻️ | リファクタリング | ♻️ refactor(internal): コードの整理 |
| perf | 🚀 | パフォーマンス改善 | 🚀 perf(api): レスポンス時間の改善 |
| test | 🧪 | テストの追加・修正 | 🧪 test: ユニットテストの追加 |
| build | 🏗️ | ビルド関連の変更 | 🏗️ build: 依存関係の更新 |
| ci | 🔧 | CI設定の変更 | 🔧 ci: GitHub Actionsの設定更新 |
| chore | 📦 | その他の変更 | 📦 chore: パッケージの更新 |
| revert | ⏪ | 変更の取り消し | ⏪ revert: 前回の変更を取り消し |

#### 使用可能なスコープ

- api: API関連
- config: 設定関連
- cli: CLI関連
- cmd: コマンド関連
- internal: 内部実装
- pkg: パッケージ

#### 注意事項

- コミットメッセージは日本語で記述してください
- タイトルは簡潔に、本文は必要に応じて詳細を記述してください
- ルールに従わないコミットメッセージはGitフックによって拒否されます

詳細なルールは [.cursor/rules/commit-message.mdc](.cursor/rules/commit-message.mdc) を参照してください。 