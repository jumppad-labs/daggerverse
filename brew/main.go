package main

import (
	"fmt"
	"strings"
)

type Brew struct{}

func (b *Brew) Formula(
	repository,
	// +optional
	darwinX86URL string,
	// +optional
	darwinArm64URL string,
	// +optional
	linuxX86URL string,
	// +optional
	linux_arm64_url string,
) error {
	template := &strings.Builder{}

	h := fmt.Sprintf(header, version)
	template.WriteString(h)
}

var header = `
# typed: false
# frozen_string_literal: true

class Jumppad < Formula
  desc ""
  homepage "https://jumppad.dev/"
  version "%s"

`

var macArm = `
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/%s/releases/download/%s/jumppad_%s_darwin_x86_64.zip"
    sha256 "%s"
  end
`

var t = `

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/jumppad-labs/jumppad/releases/download/${VERSION}/jumppad_${VERSION}_darwin_arm64.zip"
    sha256 "${DARWIN_ARM64_SHA}"
  end
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/jumppad-labs/jumppad/releases/download/${VERSION}/jumppad_${VERSION}_linux_x86_64.tar.gz"
    sha256 "${LINUX_x86_SHA}"
  end
  if OS.linux? && Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
    url "https://github.com/jumppad-labs/jumppad/releases/download/${VERSION}/jumppad_${VERSION}_linux_arm64.tar.gz"
    sha256 "${LINUX_ARM64_SHA}"
  end

  def install
    bin.install "jumppad"
  end
end
`
