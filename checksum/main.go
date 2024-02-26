package main

import (
	"context"
)

type Checksum struct{}

// CalculateChecksum will calculate the checksum of a given URL
func (c *Checksum) CalculateChecksum(ctx context.Context, url string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{
			"curl", "-slo", "/jumppad", url,
		}).
		WithExec([]string{
			"sha256sum", "-b", "/jumppad",
		}).
		Stdout(ctx)
}
