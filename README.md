# 🍎 esa-cli

esaの記事をローカルで作成・更新できるCLIツールです。

## 機能

- 記事の一覧表示
- 記事のダウンロード（Markdown形式）
- 記事の更新（ローカルのMarkdownファイルから）
- Front Matter形式でのメタデータ管理（タイトル、カテゴリ、タグ、WIP状態）

## 必要条件

- Go 1.16以上
- esaのアカウントとアクセストークン

## インストール方法

### 方法1: Homebrewでインストール（推奨）

```bash
brew install shellme/esa-cli
```

### 方法2: ワンライナーでインストール

```bash
curl -sSL https://raw.githubusercontent.com/shellme/esa-cli/main/scripts/install-mac.sh | bash
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
```

### 記事のダウンロード

```bash
# 記事番号を指定してダウンロード
esa-cli fetch 123
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
---

記事の本文...
```

### 記事の更新

```bash
# ローカルのMarkdownファイルから記事を更新
esa-cli update 123-article-title.md
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

### サポート

問題が解決しない場合は、以下の方法でサポートを受けられます：

1. GitHub Issuesで報告
2. プルリクエストの送信

## ライセンス

MIT License

## 作者

[shellme](https://github.com/shellme) 