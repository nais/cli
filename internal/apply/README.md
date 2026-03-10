# nais apply

Apply Valkey and OpenSearch resources to a Nais environment.

## Usage

```shell
nais alpha apply <config-file> --environment <env> --team <team>
```

## Arguments

- `config-file` — path to a YAML manifest file (`.yaml` or `.yml`)

## Flags

- `--environment` (`-e`) — target environment, e.g. `dev` or `prod` (required, tab-completion supported)
- `--team` (`-t`) — team slug (required)

## Manifest format

Each file is a standard Kubernetes CRD manifest (`apiVersion: nais.io/v1`).
Multiple resources can be placed in the same file separated by `---`.

### Valkey

```yaml
apiVersion: nais.io/v1
kind: Valkey
metadata:
  name: my-cache
spec:
  tier: SingleNode          # SingleNode | HighAvailability
  memory: 4GB               # 1GB | 4GB | 8GB | 14GB | 28GB | 56GB | 112GB | 200GB
  maxMemoryPolicy: allkeys-lru  # optional; see below
```

**maxMemoryPolicy values:** `allkeys-lfu`, `allkeys-lru`, `allkeys-random`,
`noeviction`, `volatile-lfu`, `volatile-lru`, `volatile-random`, `volatile-ttl`

### OpenSearch

```yaml
apiVersion: nais.io/v1
kind: OpenSearch
metadata:
  name: my-index
spec:
  tier: SingleNode          # SingleNode | HighAvailability
  memory: 4GB               # 2GB | 4GB | 8GB | 16GB | 32GB | 64GB
  version: "2"              # "1" | "2" | "2.19" | "3.3"
  storageGB: 50
```

### Multi-resource file

```yaml
apiVersion: nais.io/v1
kind: Valkey
metadata:
  name: sessions
spec:
  tier: HighAvailability
  memory: 28GB
---
apiVersion: nais.io/v1
kind: OpenSearch
metadata:
  name: search
spec:
  tier: SingleNode
  memory: 8GB
  version: "2.19"
  storageGB: 100
```

## Examples

```shell
nais alpha apply nais.yaml --environment dev --team myteam
nais alpha apply nais.prod.yaml --environment prod --team myteam
```
