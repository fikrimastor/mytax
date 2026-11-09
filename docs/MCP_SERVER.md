# mytax MCP server (local, read-only)

Status: **developer preview, milestone 3a.** Local stdio only. Draft rules only. No taxpayer data. Not affiliated with LHDN. Not tax advice.

The server is a thin adapter over `services/api/internal/policy`. It has no tax logic of its own and adds no amounts. The REST endpoint `GET /v1/policy/catalog?ya=2025` serves the same draft data.

## Run

```bash
cd services/api
go run ./cmd/mcp
```

It speaks MCP over stdin/stdout and logs to stderr. It opens no network port. Built with the official Go SDK, `github.com/modelcontextprotocol/go-sdk` v1.8.0, which needs Go 1.25.

For an MCP client, build a binary outside the repository and point the client at it:

```bash
cd services/api
go build -o "$HOME/.local/bin/mytax-mcp" ./cmd/mcp
```

```json
{
  "mcpServers": {
    "mytax": { "command": "/Users/you/.local/bin/mytax-mcp" }
  }
}
```

## Tools

Both tools are read-only and idempotent. Amounts are integer sen. Every result has `package_status` (`DRAFT` today), `filing_ready: false` and a notice.

| Tool | Input | Output |
|---|---|---|
| `list_tax_reliefs` | `assessment_year` | Package version and status, shared groups with caps, the rules the engine can evaluate (`id`, `group`, `exclusive_family`, `limit_sen`), and the 31 descriptive registry entries (not evaluable) |
| `get_tax_relief_rule` | `assessment_year`, `rule_id` | The rule (cap, required facts, sources), the shared group it counts against if any |

Errors (returned as MCP tool errors, `isError: true`):

| Case | Text starts with |
|---|---|
| Any year except 2025 | `RULE_UNAVAILABLE` |
| Rule ID not in the package | `UNKNOWN_RELIEF_RULE` |
| Extra or wrongly typed argument | SDK schema validation error |

No schema has a user, tenant or authorization field.

## What it does not do

- No claim evaluation. `verify_claim` comes with milestone 5 and will call `policy.Evaluate` unchanged.
- No guidance search. It needs a curated source store first (`docs/MCP_REPOSITORY_AUDIT.md` §6).
- No HTTP transport and no auth. A remote transport needs gate P2 in `docs/MCP_CONNECTOR_PLAN.md`.
- No approved rules exist. Nothing here is an entitlement or a filing figure.

## Tests

`go test ./internal/mcpserver/...` runs a real MCP client against the server over in-memory transports: tool list, YA 2025 catalog, unsupported years, a rule with a shared group, a rule in an exclusive family, unknown rules.
