# AGENTS.md - Nais CLI

Kjøreregler for AI-agenter som jobber med dette prosjektet.

## Nais MCP Server

Se `MCP.md` for dokumentasjon om Nais MCP server som lar AI-agenter interagere med Nais-plattformen via GraphQL.

## Prosjektstruktur

- **Go-prosjekt** med Cobra for CLI-kommandoer
- Kommandoer i `internal/` organisert etter domene
- GraphQL-klient generert med genqlient
- Schema i `schema.graphql`

## Vanlige kommandoer

| Oppgave | Kommando |
|---------|----------|
| Bygg | `mise run build` |
| Kjør tester | `mise run test` |
| Formater kode | `mise run fmt` |
| Alle sjekker | `mise run check` |
| Statisk analyse | `mise run check:staticcheck` |
| Generer GraphQL | `mise run generate:graphql` |
| Oppdater schema (lokal API) | `mise run update:graphql-schema:local` |
| Oppdater schema (live API) | `mise run update:graphql-schema:live` |
| Kjør lokalt mot lokal API | `mise run local` |

## Kjøre CLI lokalt

```bash
mise run build
./bin/nais --version
./bin/nais <kommando>
```

## Arbeidsflyt

1. **Før du endrer kode**: Les relevante filer for å forstå eksisterende mønstre
2. **Etter endringer i GraphQL queries**: Kjør `mise run generate:graphql`
3. **Ved schema-endringer i API**: Kjør `mise run update:graphql-schema:local`
4. **Etter alle endringer**: Kjør `mise run test` og `mise run fmt`

## GraphQL

- Schema ligger i `schema.graphql` (hentes fra nais-api)
- Klient genereres med genqlient (konfig i `genqlient.yaml`)
- Queries defineres i Go-filer med `# @genqlient` kommentarer

## Viktige konvensjoner

- **Kommandostruktur**: Hver kommando i egen pakke under `internal/`
- **Formattering**: `gofumpt`
- **Commit-meldinger**: Conventional Commits (se `script/semantic-commit-hook.sh`)

## Lokal utvikling

```bash
mise install
mise run build
./bin/nais --version
```

For shell completion:
```bash
source <(./bin/nais completion zsh)
```
