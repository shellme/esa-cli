.PHONY: build clean test release

# 変数
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

# デフォルトターゲット
all: build

# ビルド
build:
	go build $(LDFLAGS) -o esa-cli cmd/esa-cli/main.go

# クリーン
clean:
	rm -f esa-cli
	rm -f esa-cli-*

# テスト
test:
	go test -v ./...

# リリース用バイナリのビルド
release: clean
	# Mac (Intel)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o esa-cli-darwin-amd64 cmd/esa-cli/main.go
	# Mac (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o esa-cli-darwin-arm64 cmd/esa-cli/main.go
	# ユニバーサルバイナリの作成
	lipo -create -output esa-cli-darwin-universal esa-cli-darwin-amd64 esa-cli-darwin-arm64
	# アーカイブの作成
	tar czf esa-cli-darwin-universal.tar.gz esa-cli-darwin-universal

# インストール
install: build
	cp esa-cli /usr/local/bin/

# アンインストール
uninstall:
	rm -f /usr/local/bin/esa-cli

# ヘルプ
help:
	@echo "利用可能なターゲット:"
	@echo "  build      - バイナリのビルド"
	@echo "  clean      - ビルドファイルの削除"
	@echo "  test       - テストの実行"
	@echo "  release    - リリース用バイナリのビルド"
	@echo "  install    - システムへのインストール"
	@echo "  uninstall  - システムからのアンインストール"
	@echo "  help       - このヘルプメッセージの表示" 