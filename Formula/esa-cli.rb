class EsaCli < Formula
  desc "Command line tool for esa.io"
  homepage "https://github.com/shellme/esa-cli"
  version "0.1.0"
  license "MIT"

  if OS.mac?
    url "https://github.com/shellme/esa-cli/releases/download/v0.1.0/esa-cli-darwin-universal.tar.gz"
    sha256 "71ad50e126b9aa1fa4c6f0e7319cb1fffcb50ed67cef14beeeec315f19cb157c"
  end

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