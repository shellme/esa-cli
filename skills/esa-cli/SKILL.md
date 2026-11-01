---
name: esa-cli
description: esa.ioの記事をローカルのMarkdownファイルとして取得・編集・更新する。esaの記事の編集、複数記事の一括取得や一括更新、カテゴリの一括移動、画像付き記事の作成、記事の議論の要約、コメント投稿を頼まれたときに使う。
---

# esa-cli

esaの記事をローカルファイルとして編集するCLIです。`esa-cli help` で全オプションを確認できます。

## 基本の流れ

1. `esa-cli fetch <番号>` または `esa-cli fetch-all -c <カテゴリ> --dry-run` の後に `--dry-run` なしで取得する
2. `<番号>-<タイトル>.md` を編集する。Front Matterの `remote_updated_at` は変更しない
3. `esa-cli diff <ファイル>` で変更内容を確認する
4. ユーザーが明示的に依頼している場合のみ `esa-cli update <ファイル> -m "<変更内容>"` を実行する

複数ファイルは `esa-cli status` で状態を確認し、`esa-cli update-all "<パターン>" --dry-run` で対象を確認してから `esa-cli update-all "<パターン>" -f -m "<変更内容>"` で更新します。

内容を読むだけなら、ファイルを作らない `esa-cli fetch <番号> --print` を使います。

## 記事を探す

`esa-cli list <件数> -q "<esaの検索構文>" --json` で候補を探します。`-c`、`-t`、`-u` と組み合わせられます。同じ条件を `esa-cli fetch-all` に渡すと、まとめて取得できます。

## 議論の要約とコメント

記事の議論を要約するときは `esa-cli fetch <番号> --print --comments` で本文とコメントを読みます。

コメントの投稿は `esa-cli comment <番号> -f <ファイル>` です。先に `--dry-run` で内容を表示してユーザーに見せ、了承を得てから `--force` を付けて実行します。投稿はユーザー本人の名義になり、通知も届きます。

## 守ること

- `update`、`update-all`、`create`、`move`、`comment` はチームに公開される操作です。実行前に対象をユーザーに示す
- 上書きの確認には自動で回答しない。`yes |` などのパイプで確認を通過しない
- 終了コード2は競合などによる中止です。`esa-cli diff` の結果をユーザーに見せて判断を仰ぐ
- `--overwrite` はローカルの編集を置き換えます。必要な場合だけ使う（元のファイルは `.esa-cli-backup/` に保存される）
- 移動は `esa-cli move ... --dry-run` で対象を確認してから、ユーザーの了承を得て `-f` で実行する

## 状態の読み方

- `変更なし`: esaと同じ内容
- `ローカル変更`: `update` で反映できる
- `要確認（esa側に更新あり）`: esa側が更新されていて内容が異なる。`diff` で確認する

## 画像

本文の `![説明](./images/a.png)` のようなローカル画像は、`update`、`update-all`、`create -f` の実行時に自動でアップロードされ、URLに置き換わります。パスは記事ファイルからの相対パスです。
