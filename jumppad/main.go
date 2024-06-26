package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/log"
)

type Jumppad struct {
	Binary *File
	Cache  *CacheVolume
}

// WithVersion installs a specific version of jumppad from GitHub releases
func (m *Jumppad) WithVersion(
	// the version of jumppad to install
	version,
	// the architecture to install jumppad for, i.e. amd64, arm64.
	architecture string,
) *Jumppad {
	// remove the v if it exists
	version = strings.TrimPrefix(version, "v")

	jumppadArch := ""
	if architecture == "amd64" {
		jumppadArch = "x86_64"
	}

	m.Binary = dag.Container().
		From("alpine:latest").
		WithWorkdir("/setup").
		WithExec([]string{
			"wget",
			fmt.Sprintf("https://github.com/jumppad-labs/jumppad/releases/download/%s/jumppad_%s_linux_%s.tar.gz", version, version, jumppadArch),
			"-O", "./jumppad.tar.gz",
		}).
		WithExec([]string{"tar", "-xzf", "./jumppad.tar.gz"}).
		WithExec([]string{"mv", "./jumppad", "/usr/local/bin/jumppad"}).File("/usr/local/bin/jumppad")

	return m
}

// WithFile installs a specific version of jumppad from the provided file
func (m *Jumppad) WithFile(
	// the file to use as the jumppad binary
	file *File,
) *Jumppad {
	m.Binary = file
	return m
}

// WithCache uses a specifies cache volume for docker or podman server
func (m *Jumppad) WithCache(
	// the cache volume to use
	cache *CacheVolume,
) *Jumppad {
	m.Cache = cache

	return m
}

// TestBlueprint tests a blueprint using either docker or podman,
// this method is designed to be used with the Dagger API not the CLI
func (m *Jumppad) TestBlueprint(
	ctx context.Context,
	// the directory containing the blueprint to test
	src *Directory,
	// the working directory to run the test in, this is relative to the src directory
	// +optional
	workingDirectory string,
	// the architecture to test the blueprint on, i.e. amd64, arm64.
	// +optional
	// +default="amd64"
	architecture,
	// the runtime to use, either docker or podman
	// +optional
	// +default="docker"
	runtime string) error {
	var testBase *Container
	if runtime == "docker" {
		testBase = m.dockerBase(ctx, architecture)
	} else {
		testBase = m.podmanBase(ctx, architecture)
	}

	testBase = testBase.WithFile("/usr/local/bin/jumppad", m.Binary)

	wd := path.Join("/test/src", workingDirectory)

	ctn, err := testBase.
		WithEntrypoint([]string{"/scripts/entrypoint.sh"}).
		WithDirectory("/test/src", src).
		WithWorkdir(wd).
		WithExec([]string{"jumppad", "test", "."}, ContainerWithExecOpts{InsecureRootCapabilities: true}).
		Sync(ctx)

	out, _ := ctn.Stderr(ctx)
	log.Debug(out)

	out, _ = ctn.Stdout(ctx)
	log.Debug(out)

	return err
}

// TestBlueprintWithVersion tests a blueprint with a specific version of jumppad installed from GitHub releases
//
// example usage: "dagger call test-blueprint-with-version --src ./examples/multiple_k3s_clusters --version v0.5.59"
func (m *Jumppad) TestBlueprintWithVersion(
	ctx context.Context,
	// the directory containing the blueprint to test
	src *Directory,
	// the version of jumppad to install
	version string,
	// the working directory to run the test in, this is relative to the src directory
	// +optional
	// +default="amd64"
	workingDirectory string,
	// the architecture to test the blueprint on, i.e. amd64, arm64.
	// +optional
	architecture string,
	// the runtime to use, either docker or podman
	// +optional
	// +default="docker"
	runtime string,
	// the cache volume to use
	// +optional
	cache string,
) error {
	log.SetLevel(log.DebugLevel)

	// fetch the binary
	m.WithVersion(version, architecture)

	if cache != "" {
		m.WithCache(dag.CacheVolume(cache))
	}

	return m.TestBlueprint(ctx, src, workingDirectory, architecture, runtime)
}

// TestBlueprintWithVersion tests a blueprint with an existing binary
//
// example usage: "dagger call test-blueprint-with-binary --src ./examples/multiple_k3s_clusters --binary $(which jumppad)
func (m *Jumppad) TestBlueprintWithBinary(
	ctx context.Context,
	// the directory containing the blueprint to test
	src *Directory,
	// the path to the jumppad binary
	binary *File,
	// the working directory to run the test in, this is relative to the src directory
	// +optional
	workingDirectory string,
	// the architecture to test the blueprint on, i.e. amd64, arm64.
	// +optional
	// +default="amd64"
	architecture string,
	// the runtime to use, either docker or podman
	// +optional
	// +default="docker"
	runtime string,

	// +optional
	cache string,
) error {
	log.SetLevel(log.DebugLevel)

	m.WithFile(binary)

	if cache != "" {
		m.WithCache(dag.CacheVolume(cache))
	}

	return m.TestBlueprint(ctx, src, workingDirectory, architecture, runtime)
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

	if m.Cache != nil {
		testBase = testBase.WithMountedCache("/var/lib/docker", m.Cache)
	}

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
		WithExec([]string{"dnf", "install", "-y", "git", "unzip"}).
		WithEnvVariable("DOCKER_TLS_CERTDIR", "").                       // disable TLS
		WithEnvVariable("DOCKER_HOST", "unix:///run/podman/podman.sock") // add the podman sock

	if m.Cache != nil {
		testBase = testBase.WithMountedCache("/var/lib/containers", m.Cache)
	}

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
