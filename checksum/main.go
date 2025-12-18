package main

import (
	"context"
	"fmt"
	"main/internal/dagger"
	"strings"
)

type Checksum struct{}

// CalculateChecksum will calculate the checksum of a given URL
func (c *Checksum) CalculateFromUrl(ctx context.Context, url string) (string, error) {
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

func (c *Checksum) CalculateFromFile(ctx context.Context, file *dagger.File) (string, error) {
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
