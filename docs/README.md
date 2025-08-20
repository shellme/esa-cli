# esa-cli Documentation

esa-cliの公式ドキュメントサイトです。

## 🚀 プロジェクト構造

このAstro + Starlightプロジェクトでは、以下のフォルダとファイルが使用されています：

```
.
├── public/          # 静的アセット（favicon等）
├── src/
│   ├── assets/      # 画像やスタイルシート
│   ├── content/     # ドキュメントコンテンツ
│   │   ├── docs/    # メインドキュメント
│   └── content.config.ts
├── astro.config.mjs
├── package.json
└── tsconfig.json
```

## 📚 ドキュメントの構成

Starlightは`src/content/docs/`ディレクトリ内の`.md`または`.mdx`ファイルを探します。各ファイルはファイル名に基づいてルートとして公開されます。

画像は`src/assets/`に追加でき、Markdown内で相対リンクで埋め込むことができます。

## 🧞 コマンド

すべてのコマンドは、プロジェクトのルートからターミナルで実行します：

| コマンド                   | アクション                                           |
| :------------------------ | :----------------------------------------------- |
| `npm install`             | 依存関係をインストール                            |
| `npm run dev`             | ローカル開発サーバーを`localhost:4321`で開始      |
| `npm run build`           | 本番サイトを`./dist/`にビルド                      |
| `npm run preview`         | デプロイ前にローカルでビルドをプレビュー           |
| `npm run astro ...`       | `astro add`、`astro check`などのCLIコマンドを実行 |
| `npm run astro -- --help` | Astro CLIのヘルプを表示                           |

## 🔗 関連リンク

- [esa-cli GitHub リポジトリ](https://github.com/shellme/esa-cli)
- [Starlight ドキュメント](https://starlight.astro.build/)
- [Astro ドキュメント](https://docs.astro.build)
