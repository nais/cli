# nais-cli - A Nais command-line interface

See Nais doc for usage instructions: [nais-cli](https://docs.nais.io/how-to-guides/nais-cli/install/)

## Local Development

### Install the required go version:

```bash
mise install
```

### Build nais cli

```
mise run build
```

### Run tests

```
mise run test
```

### Verify nais cli

```
./bin/nais --version
```

## Instrumentation

We use [otel](https://opentelemetry.io) for instrumentation and record the user interaction with the CLI to see if we can
optimize the user experience.

We respect your privacy and do not collect any data if you set the environment variable `DO_NOT_TRACK`
according to the [Do Not Track](https://consoledonottrack.com).

There's a [grafana dashboard](https://monitoring.nais.io/d/ce2c9sehbbbwgd/nais-cli?orgId=1&from=now-24h&to=now)
