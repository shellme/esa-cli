class EsaCli < Formula
  desc "CLI tool for managing esa.io articles"
  homepage "https://github.com/shellme/esa-cli"
  license "MIT"
  version "0.2.0"
  url "https://github.com/shellme/esa-cli/releases/download/v#{version}/esa-cli-darwin-universal.tar.gz"
  sha256 "0806c4ecfb149560c77e6685b5c11dce2e8fe7d2a13ec8a01f718b28d8f7efd4"

  def install
    bin.install "esa-cli-darwin-universal" => "esa-cli"
    
    # 設定ファイルのテンプレートを作成
    (etc/"esa-cli").mkpath
    (etc/"esa-cli/config.template").write <<~EOS
      {
        "team_name": "your-team-name",
        "access_token": ""
      }
    EOS
  end

  def caveats
    <<~EOS
      🎉 esa-cli がインストールされました！

      📋 次のステップ:
      1. esa-cli setup で初期設定
      2. esa-cli list で記事一覧を確認
      3. esa-cli fetch 123 で記事をダウンロード

      💡 詳しい使い方: https://github.com/shellme/esa-cli
      🆘 サポート: GitHub Issues
    EOS
  end

  test do
    system "#{bin}/esa-cli", "version"
  end
end 