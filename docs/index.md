---
title: Home
nav_order: 1
---

# Nais CLI

The Nais CLI (`nais`) is a command-line tool for interacting with the [Nais platform](https://nais.io).

## Installation

### macOS

```bash
brew tap nais/tap
brew install nais
```

### Ubuntu / Debian

```bash
NAIS_GPG_KEY="/etc/apt/keyrings/nav_nais_gar.asc"
curl -sfSL "https://europe-north1-apt.pkg.dev/doc/repo-signing-key.gpg" | sudo dd of="$NAIS_GPG_KEY"
echo "deb [arch=amd64 signed-by=$NAIS_GPG_KEY] https://europe-north1-apt.pkg.dev/projects/nais-io nais-ppa main" | sudo tee /etc/apt/sources.list.d/nav_nais_gar.list
sudo apt update
sudo apt install nais
```

### Windows

```powershell
scoop bucket add nais https://github.com/nais/scoop-bucket
scoop install nais-cli
```

### Manual download

Download the archive for your platform from [GitHub Releases](https://github.com/nais/cli/releases/latest) and extract the `nais` binary to a directory on your `$PATH`.

## Getting started

After installation, log in to the Nais platform:

```bash
nais login
```

Then explore available commands:

```bash
nais --help
```

## Shell completion

Enable shell completion for faster navigation:

```bash
# Bash
source <(nais completion bash)

# Zsh
source <(nais completion zsh)

# Fish
nais completion fish | source
```

Add the appropriate line to your shell profile to make it persistent.
