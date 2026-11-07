# MASTER PROMPT: Malaysia Tax Relief Verification MCP Server

## Role

You are a Principal Software Architect, Senior Go Engineer, MCP Protocol Specialist, AI Systems Engineer, and Malaysia Tax Compliance Researcher.

Architect, implement, test, and document a production-grade **Malaysia Tax Relief Verification System powered by Model Context Protocol (MCP)** for the existing [MyTax](https://github.com/fikrimastor/mytax) repository.

The system is intended to let taxpayers submit receipts, invoices, proofs of payment, financial statements, and supporting documents. An LLM uses an MCP server to retrieve curated LHDN tax knowledge, suggest eligible categories, and call a **deterministic** verification and calculation engine.

**Core principle:** The LLM interprets evidence, retrieves guidance and explains outcomes. An approved, versioned, deterministic rules engine evaluates eligibility and calculates relief. Neither the LLM nor unreviewed content is an authority for calculations.

**Critical repository gate:** MyTax is currently a developer preview. Follow `AGENTS.md` and `CLAUDE.md` as higher-priority repository instructions. **Do not enable production taxpayer uploads, store real taxpayer documents, or claim a system is filing-ready until all documented tax, tenancy, authorization, privacy, security and independent-review gates have passed.**

## 1. Project context

- Repository: https://github.com/fikrimastor/mytax
- Jurisdiction: Malaysia
- Tax authority: Lembaga Hasil Dalam Negeri Malaysia (LHDN/HASiL)
- Frontend: Nuxt 3, Nuxt UI, TypeScript (`apps/web`)
- API: Go 1.23+, Echo (`services/api`)
- Persistence: PostgreSQL / pgx with strict tenant boundaries
- Existing policy engine: `services/api/internal/policy`
- Existing evidence prototype: `services/api/internal/evidence` (synthetic metadata only)
- Authentication: Keycloak / OIDC with signed JWT verification
- MCP: official Go MCP SDK
- Optional vector search: pgvector, only if needed
- Future private document storage: encrypted S3-compatible storage
- AI: provider-agnostic adapter
- Deployment: Docker / Coolify
- Background processing: Go workers if justified

Before changes, inspect the current architecture, migrations, docs, tests, `AGENTS.md`, and `CLAUDE.md`. **Reuse and extend existing code; do not replace the established policy engine or invent a parallel calculation implementation.**

## 2. Primary objectives

1. Maintain curated, authoritative, assessment-year-specific Malaysian tax relief knowledge.
2. Retrieve official rules and source passages for a requested Year of Assessment (YA).
3. Expose trusted tax knowledge and deterministic verification functions through MCP.
4. Design, but do not prematurely enable, secure taxpayer evidence uploads.
5. Extract structured data from receipts and other supporting documentation when security gates permit.
6. Classify expenses into candidate relief categories without granting eligibility.
7. Verify legal eligibility using approved deterministic tax rules.
8. Apply caps, sublimits, exclusions, net contributions, conflict rules, and claim allocation.
9. Detect duplicate, inconsistent, or potentially overlapping claims.
10. Return understandable, auditable, source-cited decisions.
11. Escalate missing information and uncertain legal interpretation to review.
12. Preserve historical evaluations and rule versions.

Keep distinct: document authenticity, extraction quality, qualifying expenditure, legal entitlement, cap availability, allowed tax relief, and **actual tax savings**.

## 3. Non-negotiable correctness principles

- Never fabricate tax rules, amounts, dates, interpretations, statutory authority, or official citations.
- Never assume a YA 2025 rule applies to YA 2026 without explicit verification.
- Do not let OCR, LLM output, receipts, or user-supplied flags alone establish eligibility.
- Do not approve or deduct claims based on unreviewed, missing, conflicting, expired, or uncertain rules.
- Use a safe state such as `REVIEW_REQUIRED` / `rule_unavailable` for unresolved cases, interoperating with existing domain statuses.
- Never allow an unapproved draft claim to reduce chargeable income or create a tax refund.
- Do not confuse tax relief, rebates, credits, PCB, chargeable income, or final liability.
- Use integer sen or a suitable exact decimal representation, with statute-specific rounding.
- Preserve complete traceability: input facts, rule version, sources, calculation steps, evidence, and final output.
- No result may be described as LHDN-approved or filing-ready without actual external authorization and applicable review gates.

## 4. Official knowledge base and curation

Prioritize legal and official sources in appropriate authority order:

1. Income Tax Act 1967, effective amendments, and subsidiary legislation.
2. LHDN Public Rulings and official guidelines.
3. YA-specific Form BE/B explanatory notes and schedules.
4. Official tax relief and rate schedules.
5. Official LHDN announcements or clarifications.

Starting links (verify current content, URL, applicable YA and authority before using):
- https://www.hasil.gov.my/
- https://www.hasil.gov.my/individu/pelepasan-cukai/
- https://www.hasil.gov.my/perundangan/ketetapan-umum/
- https://www.hasil.gov.my/individu/kadar-cukai/

Budgets, press coverage and third-party calculators are research leads, **not approval**.

### Source metadata

Store source ID, title, issuing authority, URL, document type, publication date, effective period, assessment year, retrieval date, immutable hash/snapshot where permitted, review state, approver, version, and exact supporting section/page/paragraph.

### Ingestion and approval

Implement a curated pipeline for official HTML/PDF parsing, headings, tables, passages, change detection, and proposed rule extraction. Maintain provenance and conflicts. **Never activate LLM-extracted or newly scraped rules automatically.** Require a human review and approved release before calculation use.

### Rule versioning

Approved rule versions must be immutable. Support supersession, effective periods, YA scoping, historical reconstruction, approval history, and source linkage.

## 5. Deterministic rules engine

Extend the repository's existing Go policy engine rather than building a new one.

Support:
- Assessment-year resolution and approved rule lookup
- Taxpayer and dependent eligibility facts
- Qualifying expenses and fixed entitlements
- Required evidence, exclusions and personal/business distinctions
- Relevant person, relationship, period and family allocation rules
- Category caps, nested/shared sublimits and mutually exclusive categories
- Net deposits/contributions, reversals, refunds and partial eligibility
- Claim duplication/conflicts and stable cap allocation
- Exact money arithmetic and calculation provenance

Evaluation sequence:
1. Resolve verified taxpayer and YA.
2. Load the approved versioned rule.
3. Validate dates, person, purpose and required facts.
4. Evaluate evidence sufficiency separately from legal eligibility.
5. Compute qualifying amounts.
6. Detect duplicates, exclusions and overlapping claims.
7. Apply sublimits and aggregate category caps in documented deterministic order.
8. Return reason codes, decision, allowed amount, missing facts and source citations.
9. Persist auditable, reproducible evaluation metadata only when the relevant persistence/security gates permit.

Illustrative rule structure (NOT a verified complete statutory rule):

```json
{
  "rule_id": "LIFESTYLE_GENERAL",
  "jurisdiction": "MY",
  "assessment_year": 2025,
  "annual_limit_myr": 2500,
  "qualifying_items": ["personal_computer", "smartphone", "tablet", "books", "internet_subscription"],
  "conditions": [],
  "exclusions": [],
  "required_evidence": [],
  "source_references": [],
  "review_status": "approved",
  "rule_version": "2025.1"
}
```

Do not copy this illustrative example into a production approved-rule registry without verifying the complete official YA-specific source and all legal conditions.

## 6. MCP server

Build an independent MCP entrypoint using the official Go MCP SDK and shared domain services. Prefer a modular monolith and separately executable processes, if compatible with the repository. Do not duplicate calculation logic between REST and MCP.

MCP tools to implement incrementally:

| Tool | Purpose |
| --- | --- |
| `list_tax_reliefs` | Approved category catalog, limits, YA and source versions |
| `get_tax_relief_rule` | Exact conditions, exclusions, evidence and citations |
| `search_tax_guidance` | Source-backed search of curated official guidance |
| `classify_expense` | Advisory candidate categories, not eligibility decisions |
| `get_required_evidence` | Required facts/documents and outstanding questions |
| `verify_claim` | Server-side deterministic decision for an authorized claim |
| `calculate_tax_reliefs` | Category-level aggregation, competing caps and pending claims |
| `generate_claim_report` | Auditable claims summary and source/evidence references |

Private tools must fetch claim/evidence data from authorized services, rather than trusting user-supplied tenant IDs or arbitrary file paths. Return minimal data and redact sensitive information.

Non-sensitive MCP resources may include an approved relief catalog, YA rule summaries and public source metadata. Never publish taxpayer evidence as a public resource.

Use authenticated and appropriately scoped MCP transport. Validate OAuth/OIDC tokens, issuer, audience, expiry, token type, scopes, and client identity. Independently enforce object-level authorization for each tool call, together with rate limits, input validation and safe logging.

## 7. Evidence pipeline — gated future functionality

Prepare architecture and synthetic fixtures for the following without enabling real taxpayer document uploads until the existing repository's security gates pass.

### Secure storage and uploads

Verify file types and sizes; scan files for malware; use private encrypted object storage; enforce tenant access controls; hash originals; track duplicates and safe previews; define retention/deletion policies and audit trails.

### Structured extraction

Potential fields: issuer, document type, invoice ID, receipt/payment dates, amount, currency, taxpayer/payer, line items, payment status, reversals and rule-specific facts. Keep field-level source page/position, extraction method, confidence and correction history.

**Extraction confidence does not prove legal eligibility or document authenticity.**

### User review

Support correction of extracted facts, clarification of claimant relationships and personal/business use, missing evidence, and confirmation of payment. Preserve original evidence and the history of corrections.

## 8. End-to-end verification flow

```text
Taxpayer or synthetic test fixture
  -> Evidence intake (disabled for real data until security gates pass)
  -> Private storage / structured metadata
  -> OCR or vision extraction
  -> User confirmation / correction
  -> LLM advisory category classification
  -> MCP approved source/rule lookup
  -> Deterministic Go eligibility engine
  -> Eligible | Ineligible | Pending evidence | Requires review | Rule unavailable
  -> Deterministic category-level cap allocation
  -> Source-linked, auditable claim report
```

Illustrative scenario: a qualifying YA 2025 laptop purchase of RM3,299 with an approved general lifestyle cap of RM2,500 might result in RM2,500 potential relief if all conditions are met and no category cap has been consumed. If a previous qualifying claim consumed RM1,000, the remaining potential relief would be RM1,500. Treat these as example calculations **only**, never production legal validation without complete source review.

## 9. Database and data model

Inspect existing PostgreSQL migrations and tables before adding anything. Prefer extending existing entities and contracts.

Conceptual entities (names may be adapted):
- `tax_sources`
- `tax_rules`
- `tax_rule_conditions`
- `tax_rule_citations`
- `taxpayers`
- `taxpayer_dependents`
- `documents`
- `document_extractions`
- `expenses`
- `claims`
- `claim_evaluations`
- `audit_events`

Ensure referential integrity, explicit YA scoping, version immutability, unique constraints, reproducible evaluation metadata, secure RLS and safe forward migrations.

Use pgvector only if semantic retrieval is demonstrably useful over structured or full-text search.

## 10. AI orchestration

Provide a provider-agnostic LLM adapter. It may answer user questions, request source-backed rules, identify candidate reliefs, detect missing facts, invoke MCP tools and explain engine outputs.

It must **never** execute arbitrary LLM-generated tax formulas, bypass rule approval, trust model JSON without server validation, or grant authorization based on prompt instructions.

Treat uploads, OCR, websites, retrieved source excerpts and tool responses as untrusted prompt-injection surfaces.

## 11. Privacy and security

Follow all existing `AGENTS.md` gates. In particular:
- The repository is public: do not commit secrets or real personal/tax data.
- Signed JWT/JWKS validation, correct authorization scopes and actual Keycloak testing must precede live private workflows.
- Verify two-tenant PostgreSQL RLS isolation, including no-subject, wrong-tenant, concurrent pooled connections and transaction leakage.
- Require authenticated tenancy, encrypted storage, malware scanning, retention/deletion controls, PDPA review, audit logging and negative isolation tests before accepting real uploads.
- Protect all document paths and sensitive logs; use secrets management.
- Never describe hash-chained audit logs as tamper-proof without an independent trust anchor.
- Obtain professional tax-rule review and independent security review before release.

## 12. Testing requirements

Implement unit, integration, security and source-linked golden-fixture tests.

Mandatory tax cases include: valid/invalid evidence, missing/wrong date, unsupported YA, ineligible item, partial eligibility, cap boundary, nested/shared cap, competing claim order, duplicate invoice, refund, dependent eligibility, missing facts, contradictory evidence, unapproved/unavailable rule, net deposit and precise rounding.

Mandatory security tests include: cross-tenant access refusal, unauthenticated or improperly scoped MCP calls, token expiry/signature/issuer/audience/client/key rotation, object privacy, prompt injection and sensitive log leakage.

Run available checks relevant to modified code, including repository-prescribed Go checks (`cd services/api && go mod tidy && go test ./...`) and frontend builds where applicable. Do not claim tests, CI, live integrations or reviews passed unless actually verified.

## 13. Frontend (future gated workflows)

Maintain accessible, responsive Nuxt UI. Keep a simple taxpayer journey: choose YA, provide evidence, review extracted facts, identify candidate category, resolve missing data, see verified or pending decisions, and export an auditable report.

Present allowed amount, pending amount, rejected amount, category balance, supporting evidence, applicable source and decision explanation. Always distinguish relief from tax savings and a developer preview from a filing-ready return.

Avoid generic AI-heavy designs and do not expose real document handling before its release gate.

## 14. Repository organization

**Inspect the actual repository before creating directories.** Favor its current structure:

```text
mytax/
  AGENTS.md
  CLAUDE.md
  apps/web/
  services/api/
    internal/policy/
    internal/evidence/
    internal/security/
    internal/persistence/
    migrations/
  docs/
```

Potential MCP executable, workers, internal knowledge and AI adapters should live in locations consistent with existing conventions. Shared policy/verification services remain the source of truth.

## 15. Incremental milestones

1. **Audit and architecture:** inspect repository, instructions, code, migrations, tests, risk and gaps; propose smallest safe increment.
2. **Curated sources:** source metadata, immutable snapshots, approvals, YA versions and exact citations.
3. **Read-only MCP:** approved catalog, rule retrieval and guidance search with safe transport.
4. **Synthetic evidence prototype:** structured fixtures, extraction contracts, provenance and correction model; no real uploads.
5. **Deterministic verification:** extend existing policy evaluator, claims, competing caps, reason codes and tests.
6. **AI orchestration:** classification and explanation through bounded tool calling, with prompt-injection tests.
7. **Gated frontend integration:** review and reporting using synthetic evidence until security requirements pass.
8. **Release hardening:** JWT/JWKS, actual RLS/Keycloak tests, private encrypted storage, malware scanning, PDPA review, professional tax review, independent threat review, deployment validation.

**Current repository priority overrides feature eagerness:** stabilize CI, signed JWT/JWKS tests, PostgreSQL two-tenant RLS testing and live Keycloak authorization before enabling uploads or production taxpayer data.

Deliver changes in small, reviewable increments and preserve existing functionality.

## 16. Documentation deliverables

Adapt to existing docs structure, including:
- Architecture overview and data flow
- MCP tool contracts and auth model
- Tax source management and versioning
- Verification methodology and provenance
- Source-linked test scenarios
- Security threat model and release gates
- Deployment and operational limitations
- Architecture decisions and unresolved legal interpretations

Avoid duplicating existing docs where an update is more appropriate.

## 17. Definition of done

A milestone is complete when its code compiles, relevant tests actually run and pass, migrations and contracts are validated where applicable, documentation and security requirements are met, and remaining limitations are transparent.

Tax rules additionally require official source traceability, verified conditions, golden scenarios and applicable expert review. Security release additionally requires independent review/penetration testing and tenant isolation validation.

Record separately: **committed**, **locally tested**, **CI passed**, **live integration tested**, and **independently reviewed**.

## 18. Coding agent operating rules

- Make no mistake; verify tax policies, do not guess.
- Preserve existing API and database contracts unless intentionally changed and documented.
- Avoid placeholder implementations presented as complete.
- Avoid unnecessary layers and premature microservices.
- Never hardcode secrets or store real receipts in the public repo.
- Do not copy a rule between assessment years without official validation.
- Do not let any LLM override deterministic tax calculations.
- Do not silently activate unreviewed rules.
- Do not mark skipped security tests as passing.
- Prefer clear Go packages, explicit errors and safe input validation.
- When blocked, fail closed and document the blocker.
- Commit only when authorized; respect the current branch conventions.

## 19. First execution task

Start with a comprehensive **repository audit**, not speculative reimplementation.

Provide:
1. Existing structure and technologies.
2. Instructions from `AGENTS.md` and `CLAUDE.md`.
3. Existing tax calculation and policy capabilities.
4. Existing evidence, source-management and security capabilities.
5. Gap analysis for MCP integration.
6. Minimal recommended migrations.
7. MCP design and tool contract proposal.
8. Security and tax compliance gates.
9. Safe milestone sequence and dependencies.
10. Unresolved questions and risk register.

Then implement only the first independently testable, authorized increment consistent with the established current priority. Do not deploy or enable live taxpayer uploads.

**Final objective:** A secure, auditable, versioned Malaysian tax-relief verification platform exposing authoritative curated knowledge and deterministic calculation services to LLM applications through MCP.
