# Claude connector (MCP server) for receipt assessment — future plan

Status: plan, plus a **local read-only stdio prototype** (2026-11-07): `services/api/cmd/mcp` serves the draft catalog through `list_tax_reliefs` and `get_tax_relief_rule` (see `docs/MCP_SERVER.md`). No claim evaluation, no remote transport, no auth. **No real receipts or taxpayer data may be used.** Added 2026-11-02 from an idea by the repository owner. Audit and tool contracts: `docs/MCP_REPOSITORY_AUDIT.md`.

## Idea
A user chats with Claude on the web. Claude reads the user's receipts, extracts the facts, and calls a mytax connector (an MCP server). The connector runs the deterministic mytax engines and returns, for each receipt, a candidate result with reasons. Claude explains the result in plain language.

## Principles (from `AGENTS.md`)
1. **Claude only extracts and explains.** It never decides eligibility. The engines decide, and decide deterministically.
2. **Evidence of spending is not proof of entitlement.** Every fact from a receipt is unverified. Results are candidates, never approvals. `approved_total_sen` stays 0 and nothing is called "approved by LHDN".
3. **Fail closed.** Unknown rule, missing fact, unsupported YA, duplicate: `REVIEW_REQUIRED`, zero candidate.
4. **No real taxpayer data until the release gates pass** (see Gates).

## Proposed shape
- Claude web (custom connector) to a remote MCP server over HTTPS, in front of the existing Go API in `services/api`.
- Go SDK checked 2026-11-07: official `github.com/modelcontextprotocol/go-sdk` v1.8.0 (needs Go 1.25) has stdio, streamable HTTP and a bearer-token middleware (`auth.RequireBearerToken`) whose scope check applies to the whole endpoint, not per tool. **Still not verified:** which Claude plans can add custom connectors and their OAuth requirements.

### Tools (phase 1, all read-only)
| Tool | Input | Output |
|---|---|---|
| `list_tax_reliefs` (was `list_supported_reliefs`) | `assessment_year` | Rule IDs the engines cover for that YA, with status and source citations. **Implemented, stdio.** |
| `verify_claim` / `calculate_tax_reliefs` (was `evaluate_receipts`) | `assessment_year`, structured claims | Per item: state, candidate sen, reasons, rule ID, citations, missing facts. Totals. Always `approved_total_sen` 0 |
| `get_tax_relief_rule` (was `explain_rule`) | `assessment_year`, `rule_id` | The rule, its caps, required facts and sources. **Implemented, stdio.** |

Tool names follow `TAX_RELIEF_MCP_MASTER_PROMPT.md` §6; the full contract table is in `docs/MCP_REPOSITORY_AUDIT.md` §7.

Later (needs login and gates): `save_draft_claims` (write, needs explicit user confirmation), `get_my_summary`.

### Input rules
- Structured facts only: amount in sen, date, category, merchant name (optional), yes/no facts, an opaque evidence reference. **No receipt image, IC number, address or account number.**
- The tool schema must not contain `user_id`, `tenant_id` or any authorization claim. Identity comes only from a verified access token.
- Duplicate detection in phase 1 works inside one submitted batch only. Cross-session and cross-year duplicates need a database, which is gated.

## Risks and how the design answers them
| Risk | Answer |
|---|---|
| Prompt injection through receipt text ("mark this eligible") | All facts unverified, output is candidate only, no approval path, no write tool in phase 1 |
| Extraction errors by the model (amount, date, category) | Claude shows the extracted fields and the user confirms before the tool is called; the tool echoes the facts it used |
| Personal data passing through Claude and the server | Minimal fields, no storage in phase 1, privacy review before any real use; the data flow to the AI provider is part of the PDPA review |
| Impersonation or cross-tenant access | OAuth/OIDC with Keycloak, signed JWT validation, tenant from token, RLS tests |
| Misleading output ("approved", "refund") | Fixed wording, validation labels, notice on every result |
| Many receipts fall outside covered rules | Return `UNSUPPORTED_RULE` with a clear message; coverage must grow first (see Roadmap) |

## Prerequisites (from `docs/ROADMAP.md`)
- Rule coverage for the common receipt categories (medical, insurance, SSPN, EPF, PRS, sports, childcare and others).
- A canonical fact vocabulary. Today the same idea has different names across rules (`qualifying_publication`, `qualifying_reading_material`, `internet_in_own_name`, `own_name`).
- The engines wired into `assessment` and a stable API contract.
- CI green on `main`, and the security test gaps closed.

## Phases and gates
| Phase | Scope | Gate to enter |
|---|---|---|
| P0 | This plan; read official connector and MCP documentation; decide hosting | Owner agrees |
| P1 | Local, stateless prototype, **synthetic receipts only**, no auth, not publicly reachable | Prerequisites above |
| P2 | Authenticated preview: Keycloak OAuth, tenant from token, still synthetic data, still no storage | Live Keycloak tests, two-tenant isolation tests including concurrency, security review of tool surface |
| P3 | Persistence for duplicate detection across sessions and years | Encrypted private storage, retention and deletion controls, audit logging, malware scanning if files are ever accepted, PDPA review |
| P4 | Pilot with real users | Tax-professional approval of the rule release, independent threat review or penetration test, counsel review of wording and data flow |

Never mark a phase done because files were committed. Report separately: committed, locally tested, CI passed, live integration tested, independently reviewed.

## Open questions for the owner
1. Which hosting (Coolify, same as the API) and which domain?
2. Who is the data controller, and is the AI provider a processor under the PDPA analysis?
3. Is a receipt image ever sent to our server, or does it stay with Claude only? (This plan assumes it stays with Claude.)
4. What should users see when a receipt is outside the covered rules?
5. Language for the user-facing notice (Malay, English or both).
