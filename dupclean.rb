class Dupclean < Formula
  desc "Content-aware duplicate file scanner for music producers and DJs"
  homepage "https://github.com/PopolQue/dupclean"
  url "https://github.com/PopolQue/dupclean/archive/refs/tags/v0.2.3.tar.gz"
  sha256 "8d1be8714af84c9844db4e435ccea8732446864d10f2bea0b7ca55f8905f6b67"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args
  end

  test do
    assert_match "DupClean", shell_output("#{bin}/dupclean --help")

    # Functional test: create duplicate files and verify detection
    (testpath/"file1.txt").write("duplicate content")
    (testpath/"file2.txt").write("duplicate content")
    output = shell_output("#{bin}/dupclean #{testpath} 2>&1 || true")
    assert_match "duplicate", output.downcase
  end
end
