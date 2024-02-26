package main

import (
	"context"
	"fmt"
	"strings"
)

type Brew struct{}

// example usage:
func (b *Brew) Formula(
	ctx context.Context,
	homepage,
	version,
	gitToken,
	binaryName string,
	// +optional
	darwinX86URL string,
	// +optional
	darwinArm64URL string,
	// +optional
	linuxX86URL string,
	// +optional
	linuxArm64URL string,
) (string, error) {
	template := &strings.Builder{}

	// Write header
	h := fmt.Sprintf(header, homepage, version)
	template.WriteString(h)

	// do we need to add darwin intel
	if darwinX86URL != "" {
		checksum, err := dag.Checksum().CalculateChecksum(ctx, darwinX86URL)
		if err != nil {
			return "", fmt.Errorf("failed to calculate checksum: %w", err)
		}

		parts := strings.Split(checksum, " ")
		h = fmt.Sprintf(darwinIntel, darwinX86URL, parts[0])
		template.WriteString(h)
	}

	if darwinArm64URL != "" {
		checksum, err := dag.Checksum().CalculateChecksum(ctx, darwinArm64URL)
		if err != nil {
			return "", fmt.Errorf("failed to calculate checksum: %w", err)
		}

		parts := strings.Split(checksum, " ")
		h = fmt.Sprintf(darwinArm, darwinArm64URL, parts[0])
		template.WriteString(h)
	}

	if linuxX86URL != "" {
		checksum, err := dag.Checksum().CalculateChecksum(ctx, linuxArm64URL)
		if err != nil {
			return "", fmt.Errorf("failed to calculate checksum: %w", err)
		}

		h = fmt.Sprintf(linuxIntel, linuxX86URL, checksum)
		template.WriteString(h)
	}

	if linuxArm64URL != "" {
		checksum, err := dag.Checksum().CalculateChecksum(ctx, linuxArm64URL)
		if err != nil {
			return "", fmt.Errorf("failed to calculate checksum: %w", err)
		}

		parts := strings.Split(checksum, " ")
		h = fmt.Sprintf(linuxArm, linuxArm64URL, parts[0])
		template.WriteString(h)
	}

	// Write footer
	h = fmt.Sprintf(footer, binaryName)
	template.WriteString(h)

	// Commit the template

	return template.String(), nil
}

var header = `
# typed: false
# frozen_string_literal: true

class Jumppad < Formula
  desc ""
  homepage "%s"
  version "%s"
`

var darwinIntel = `
  if OS.mac? && Hardware::CPU.intel?
    url "%s"
    sha256 "%s"
  end
`

var darwinArm = `
  if OS.mac? && Hardware::CPU.arm?
    url "%s"
    sha256 "%s"
  end
`

var linuxArm = `
  if OS.linux? && Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
    url "%s"
    sha256 "%s"
  end
`

var linuxIntel = `
  if OS.linux? && Hardware::CPU.intel?
    url "%s"
    sha256 "%s"
  end
`

var footer = `
  def install
    bin.install "%s"
  end
end
`
