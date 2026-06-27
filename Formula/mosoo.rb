class Mosoo < Formula
  desc "Generated CLI for Mosoo integrators"
  homepage "https://github.com/langgenius/mosoo-cli-go"
  url "https://github.com/langgenius/mosoo-cli-go/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "d18a95b6cdf72a6aa27b91649a826b09c84822c459c4f18872a9dac0c3ac3af7"
  license :cannot_represent
  head "https://github.com/langgenius/mosoo-cli-go.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/lathe-cli/lathe/pkg/lathe.Version=v#{version}"
    system "go", "build", "-trimpath", "-ldflags", ldflags, "-o", bin/"mosoo", "./cmd/mosoo"
  end

  test do
    assert_match "mosoo v#{version}", shell_output("#{bin}/mosoo --version")
    system bin/"mosoo", "commands", "--json"
  end
end
