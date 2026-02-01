class ArcToZen < Formula
  desc "Go CLI tool that imports Arc browser data into Zen browser"
  homepage "https://github.com/rkw6086/arc-to-zen"
  url "https://github.com/rkw6086/arc-to-zen/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "350543eba690000f26b3b0d772b926c76f564011c2b78cedb923cf8b928b88b7"
  license "MIT"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "build/arc-to-zen"
  end

  test do
    # Test that the binary exists and can show help
    assert_match "arc-to-zen", shell_output("#{bin}/arc-to-zen -h 2>&1", 1)
  end
end
