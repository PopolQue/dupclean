cask "dupclean" do
  version "0.4.3.1"

  if Hardware::CPU.intel?
    sha256 :no_check # Placeholder - update with actual SHA256 after release
    url "https://github.com/PopolQue/dupclean/releases/download/v#{version}/dupclean-darwin-amd64.app.zip"
  else
    sha256 :no_check # Placeholder - update with actual SHA256 after release
    url "https://github.com/PopolQue/dupclean/releases/download/v#{version}/dupclean-darwin-arm64.app.zip"
  end

  name "DupClean"
  desc "Content-aware duplicate file scanner and disk analyzer"
  homepage "https://github.com/PopolQue/dupclean"

  app "DupClean.app"
  binary "#{appdir}/DupClean.app/Contents/MacOS/dupclean"

  postflight do
    system_command "xattr",
                   args: ["-rd", "com.apple.quarantine", "#{appdir}/DupClean.app"],
                   sudo: false
  end

  zap trash: [
    "~/Library/Caches/dupclean.log",
  ]
end
