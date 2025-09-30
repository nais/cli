# nais apply poc

notes from poc experiments

## arguments

- `environment` is required for all commands
    - (should be a configuration from the user's profile and moved to being a flag instead)
- `file` points to a base `nais.toml` configuration file

## flags

- `team` is required for all commands
    - (should be a configuration from the user's profile)
- `mixin` is optional
    - auto-detected if matching `*<environment>.toml` existing next to the base config file
    - e.g. `nais apply dev nais.toml --team devteam` will use `nais.dev.toml` if it exists

## usage

```shell
nais alpha apply <environment> <config-file> --team <team> [--mixin <override-file>]
```

```shell
nais alpha apply dev nais.toml --mixin nais.dev.toml --team devteam
```

```shell
nais alpha apply prod nais.toml --mixin nais.prod.toml --team devteam
```

## schema

uses jsonschema for validation, though tooling support for various IDEs and editors is kinda iffy?

## possible future additions?

```shell
nais alpha apply dev nais.toml --mixin nais.dev.toml --team devteam --set application.bar.image=example.com/app:1.2.3
```

```shell
nais alpha apply prod nais.toml --mixin nais.prod.toml --team devteam --set application.bar.image=example.com/app:1.2.3
```

```shell
nais <resource> set image <environment> <name> <image>
```

```shell
nais application set image dev bar example.com/app:1.2.3
```

```shell
nais job set image dev bar example.com/app:1.2.3
```
