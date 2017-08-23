class Teresa < Formula
  desc "Teresa client"
  homepage "https://github.com/luizalabs/teresa"
  url "https://github.com/luizalabs/teresa/releases/download/v0.6.0/teresa-darwin-amd64"
  sha256 "189b14f3e548624c2231f8189eca501c58190b944f2ad5af7e011fdcf0cfb915"
  head "https://github.com/luizalabs/teresa.git"
  version 'v0.6.0'

  bottle :unneeded

  def install
    system "cp", "teresa-darwin-amd64", "teresa"
    bin.install "teresa"
  end

  test do
    system "teresa", "version"
  end
end
