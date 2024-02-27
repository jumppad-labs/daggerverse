package main

import (
	"context"
	"fmt"
	"strings"
)

type Checksum struct{}

// CalculateChecksum will calculate the checksum of a given URL
func (c *Checksum) CalculateFromURL(ctx context.Context, url string) (string, error) {
	str, err := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{
			"curl", "-sLo", "/jumppad", url,
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

func (c *Checksum) CalculateFromFile(ctx context.Context, file *File) (string, error) {
	str, err := dag.Container().
		From("alpine:latest").
		WithFile("/jumppad", file).
		WithExec([]string{
			"sha256sum", "-b", "/jumppad",
		}).
		Stdout(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return strings.TrimSpace(str), nil
}
