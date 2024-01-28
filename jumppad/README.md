# Jumppad Dagger Module

This module enables running [jumppad](https://jumppad.dev) functional tests inside Docker or Podman
containers inside dagger.

## Usage

The following example shows how to run the jumppad functional tests using
a Docker engine. The Docker engine is started using Docker-in-Docker (DinD)
inside the Dagger engine.

```shell
dagger call --focus=false test-blueprint-with-binary \
  --binary $(which jumppad) \
  --src ./examples/container
  --runtime docker
```

The following example shows how to run the jumppad functional tests using
a Podman engine. The Podman engine is started using Podman-in-Docker (PinD)
inside the Dagger engine.

```shell
dagger call --focus=false test-blueprint-with-binary \
  --binary $(which jumppad) \
  --src ./examples/container
  --runtime podman
```