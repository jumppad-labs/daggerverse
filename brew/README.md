# Brew

## Formula

Enabled a homebrew repository to be updated with the details of the formula.

Example:

```shell
dagger call formula \
  --homepage="https://jumppad.dev" \
  --repository=jumppad-labs \
  --repository=homebrew-repo \
  --version=0.7.1 \
  --commiter-name=Jumppad \
  --commiter-email="hello@jumppad.dev" \
  --binary-name=jumppad \
  --git-token=GITHUB_TOKEN \
  --darwin-x-86-url=https://github.com/jumppad-labs/jumppad/releases/download/0.7.0/jumppad_0.7.0_darwin_x86_64.zip \
  --darwin-arm-64-url=https://github.com/jumppad-labs/jumppad/releases/download/0.7.0/jumppad_0.7.0_darwin_arm64.zip \
  --linux-x-86-url=https://github.com/jumppad-labs/jumppad/releases/download/0.7.0/jumppad_0.7.0_linux_x86_64.zip \
  --linux-arm-64-url=https://github.com/jumppad-labs/jumppad/releases/download/0.7.0/jumppad_0.7.0_linux_arm64.zip
```


