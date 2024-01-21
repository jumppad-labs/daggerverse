package main

import (
	"context"
	"fmt"
	"strings"
)

type Jumppad struct {
	binary *File
}

// WithVersion installs a specific version of jumppad from GitHub releases
func (m *Jumppad) WithVersion(version, architecture string) *Jumppad {
	// remove the v if it exists
	version = strings.TrimPrefix(version, "v")

	jumppadArch := ""
	if architecture == "amd64" {
		jumppadArch = "x86_64"
	}

	m.binary = dag.Container().
		From("alpine:latest").
		WithWorkdir("/setup").
		WithExec([]string{
			"wget",
			fmt.Sprintf("https://github.com/jumppad-labs/jumppad/releases/download/v%s/jumppad_%s_linux_%s.tar.gz", version, version, jumppadArch),
			"-O", "./jumppad.tar.gz",
		}).
		WithExec([]string{"tar", "-xzf", "./jumppad.tar.gz"}).
		WithExec([]string{"mv", "./jumppad", "/usr/local/bin/jumppad"}).File("/usr/local/bin/jumppad")

	return m
}

// WithFile installs a specific version of jumppad from the provided file
func (m *Jumppad) WithFile(file *File) *Jumppad {
	m.binary = file
	return m
}

// TestBlueprint tests a blueprint using either docker or podman,
// this method is designed to be used with the Dagger API not the CLI
func (m *Jumppad) TestBlueprint(ctx context.Context, src *Directory, architecture, runtime Optional[string]) error {
	arch := architecture.GetOr("amd64")
	run := runtime.GetOr("docker")

	var testBase *Container
	if run == "docker" {
		testBase = m.dockerBase(ctx, arch)
	} else {
		testBase = m.podmanBase(ctx, arch)
	}

	testBase = testBase.WithFile("/usr/local/bin/jumppad", m.binary)

	_, err := testBase.
		WithEntrypoint([]string{"/scripts/entrypoint.sh"}).
		WithDirectory("/test/src", src).
		WithWorkdir("/test/src").
		WithExec([]string{"jumppad", "test", "."}, ContainerWithExecOpts{InsecureRootCapabilities: true}).
		Sync(ctx)

	return err
}

// TestBlueprintWithVersion tests a blueprint with a specific version of jumppad installed from GitHub releases
//
// example usage: "dagger call test --src ./examples/multiple_k3s_clusters --version v0.5.59"
func (m *Jumppad) TestBlueprintWithVersion(ctx context.Context, src *Directory, version string, architecture, runtime Optional[string]) error {
	arch := architecture.GetOr("amd64")

	// fetch the binary
	m.WithVersion(version, arch)
	return m.TestBlueprint(ctx, src, architecture, runtime)
}

func (m *Jumppad) TestBlueprintWithBinary(ctx context.Context, src *Directory, binary *File, architecture, runtime Optional[string]) error {
	m.WithFile(binary)
	return m.TestBlueprint(ctx, src, architecture, runtime)
}

// dockerBase creates a Docker engine in docker container
func (m *Jumppad) dockerBase(ctx context.Context, architecture string) *Container {
	testBase := dag.Container(ContainerOpts{Platform: Platform(fmt.Sprintf("linux/%s", architecture))}).
		From("ghcr.io/jumppad-labs/dind:v1.0.0").
		WithoutEntrypoint().
		WithUser("root").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "git"}).
		WithEnvVariable("DOCKER_TLS_CERTDIR", "") // disable TLS

	return testBase.
		WithNewFile("/scripts/entrypoint.sh", ContainerWithNewFileOpts{Contents: dnidEntrypoint}).
		WithExec([]string{"chmod", "+x", "/scripts/entrypoint.sh"})
}

var dnidEntrypoint = `#!/bin/bash
set -e

# start docker d
/usr/local/bin/dockerd.sh > /var/log/docker.log 2>&1 &

# Loop until 'docker version' exits with 0.
until docker version > /dev/null 2>&1
do
  sleep 1
done

$@
`

// podmanBase creates a Podman engine in docker container
func (m *Jumppad) podmanBase(ctx context.Context, architecture string) *Container {
	testBase := dag.Container(ContainerOpts{Platform: Platform(fmt.Sprintf("linux/%s", architecture))}).
		From("quay.io/podman/stable:v4.8.3").
		WithoutEntrypoint().
		WithUser("root").
		WithExec([]string{"dnf", "install", "-y", "git"}).
		WithEnvVariable("DOCKER_TLS_CERTDIR", "").                       // disable TLS
		WithEnvVariable("DOCKER_HOST", "unix:///run/podman/podman.sock") // add the podman sock

	return testBase.
		WithNewFile("/etc/containers/containers.conf", ContainerWithNewFileOpts{Contents: podmanConf}).
		WithNewFile("/scripts/entrypoint.sh", ContainerWithNewFileOpts{Contents: podmanEntrypoint}).
		WithExec([]string{"chmod", "+x", "/scripts/entrypoint.sh"})
}

var podmanEntrypoint = `#!/bin/bash
set -e

# start podman sock
podman system service -t 0 > /var/log/podman.log 2>&1 &
sleep 10
chmod +x /run/podman
chmod 777 /run/podman/podman.sock

# Loop until 'docker version' exits with 0.
until podman version > /dev/null 2>&1
do
  sleep 1
done

$@
`

var podmanConf = `[containers]
netns="private"
userns="host"
ipcns="host"
utsns="private"
cgroupns="host"
cgroups="disabled"
log_driver = "k8s-file"
[engine]
cgroup_manager = "cgroupfs"
events_logger="file"
runtime="crun"
`
