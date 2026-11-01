---
name: release
description: esa-cliの新しいバージョンをリリースする。バージョン更新、タグ作成、リリースワークフローとHomebrew Formulaの確認までを行う。
disable-model-invocation: true
---

# リリース手順

引数がバージョン（例: `0.7.0`）です。なければ前回タグからの変更を見て提案し、確認を取ります。機能追加はminor、修正のみはpatchです。

## 事前確認

1. `git status --short` が空で、`main` が `origin/main` と一致していることを確認する。違えば止めて報告する
2. `git log <前回タグ>..HEAD --oneline` で含まれる変更を一覧する
3. `go test ./...` と `go vet ./...` が通ることを確認する

## 実行

連続実行せず、各コマンドの結果を確認してから次へ進む。

1. `internal/version/version.go` を更新する
2. `git status --short` で変更がそのファイルだけであることを確認する
3. `🏗️ build: バージョンをx.y.zに更新` でコミットする
4. `git tag -a vX.Y.Z -m "Release vX.Y.Z"` を作成する
5. `git push origin main --follow-tags` を実行する

## 確認

1. Releaseワークフローの完了を `gh run watch <id> --exit-status` で待つ
2. `gh release view vX.Y.Z` で、下書きでなく、バイナリ4種類と `checksums.txt` があることを確認する
3. `gh api repos/shellme/homebrew-esa-cli/contents/Formula/esa-cli.rb -H 'Accept: application/vnd.github.raw'` のバージョンを確認する
4. ローカルの `ROADMAP.local.md` があれば、該当バージョンをリリース済みにする

## 報告

リリースURL、確認結果、利用者に伝えるべき動作の変更を報告する。手元の更新コマンドは次のとおり。

```bash
brew update && yes | brew upgrade esa-cli
```
