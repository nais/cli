# Nais Command Line Interface

This repository contains the source code for the [Nais](https://nais.io) command line interface (CLI).

See the [CLI documentation](https://cli.nais.io) for usage instructions.

## Local Development

### Adding a command that uses nais-relay

Follow the [shared relay integration requirements](https://github.com/nais/relay/blob/main/SERVICE-INTEGRATION.md) and fill in the [integration template](https://github.com/nais/relay/blob/main/SERVICE-INTEGRATION-TEMPLATE.md) in the product's design or rollout documentation. A new command needs a product-specific authenticated NAIS API mutation and an approved target policy; the CLI does not create `RelayAccess` resources.

### Install development tools

```bash
mise install
```

### Build Nais CLI

```
mise run build
```

### Run tests

```
mise run test
```

### Verify Nais CLI

```
./bin/nais --version
```

### Set up shell completion for local builds

For example, enable Zsh completion in the current session:

```zsh
source <(./bin/nais completion zsh)
```

Supported shells: Zsh, Bash, Fish, and PowerShell. Run `./bin/nais completion <shell> --help` for shell-specific setup instructions.

## Instrumentation

We use [otel](https://opentelemetry.io) for instrumentation and record the user interaction with the CLI to see if we can
optimize the user experience.

We respect your privacy and do not collect telemetry if you set the environment variable `DO_NOT_TRACK=1`,
following the [Do Not Track](https://consoledonottrack.com) convention.

There's a [grafana dashboard](https://monitoring.nais.io/d/ce2c9sehbbbwgd/nais-cli?orgId=1&from=now-24h&to=now).

## Contributing

This repo uses [Conventional Commits](https://www.conventionalcommits.org/). Please read up on how to format your commit messages. Please see the [commit-msg hook](script/semantic-commit-hook.sh) to see which types we allow.
