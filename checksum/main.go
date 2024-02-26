package main

import (
	"context"
	"fmt"
	"strings"
)

type Checksum struct{}

// CalculateChecksum will calculate the checksum of a given URL
func (c *Checksum) CalculateChecksum(ctx context.Context, url string) (string, error) {
	str, err := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{
			"curl", "-slo", "/jumppad", url,
		}).
		WithExec([]string{
			"sha256sum", "-b", "/jumppad",
		}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return strings.TrimSpace(str), nil
}
