# Competitive landscape

Status: informal research, not a market study. Prepared 2026-11-01 from **public pages only** (no login, no app testing).

## How to read this

- Competitor facts below are **their own claims** from their marketing, store and legal pages. We did not test the apps, their OCR, or their tax rules, so accuracy is **unverified**.
- Pricing and features change. Re-check the sources before relying on any row.
- mytax facts come from this repository (`README.md`, `AGENTS.md`, `docs/`) and describe the current developer preview.
- mytax is not affiliated with LHDN, and neither are the products below (each says so on its own pages).

## Products looked at

| | mytax (this repo) | MyTaxMate | Fiskal |
|---|---|---|---|
| Model | Open source, developer preview | Paid consumer app (web, iOS, Android) | Consumer app (iOS listing reviewed) with free and paid tiers |
| Audience | Individuals; business/Form B later | Employees (Form BE), freelancers and business owners (Form B), landlords | Individuals, business and rental in Pro |
| Receipt OCR | No | Yes ("SmartScan"), suggestions with confidence labels | Yes, suggests a relief category |
| Document storage | No (blocked until security gates pass) | Receipt vault, 7-year retention claim, audit pack | Receipt vault, 7-year retention claim, audit pack (paid) |
| Accounts / login | No | Yes | Yes |
| Tax estimate | Gross-tax preview only; no final tax | Estimated tax balance; YA 2026 referenced | Tax and refund estimates; YA 2026 rules claimed, YA 2025 unchanged |
| Filing help | Form BE worksheet preview only | Form BE/B co-pilot showing what to enter in MyTax | Filing figures in MyTax order |
| Rules transparency | Rules are versioned, source-linked and gated on review (see `docs/COMPLIANCE.md`); YA 2025 policy not yet approved | Claims mapping to LHDN's schedule; no verification method shown | Per-relief explanations; no verification method seen |
| Source code | Public | Not stated | Not stated |
| Price (as listed) | Free | RM6.66/month or RM79.90/year (Standard); RM169.90/year (Pro); lifetime from RM599.90 | Free tier; RM59.90/year (Standard); RM99.90/year (Pro) |
| Data location claim | N/A (no user data) | Singapore | Singapore |

Sources (accessed 2026-11-01):
- MyTaxMate: <https://mytaxmate.my/> and <https://mytaxmate.my/terms>
- Fiskal: <https://apps.apple.com/app/id6798135605>

## What the competitors have in common

- Both are paid or freemium consumer apps built around receipt scanning and a filing assistant.
- Both disclaim LHDN affiliation, tax-advice status and filing on the user's behalf.
- Both ask users to keep personal documents in their cloud storage, which mytax has deliberately not enabled yet.
- Neither shows how a relief rule is sourced or reviewed, at least not on the pages we read.

## Where mytax is different today

- **Verifiable rules:** the project's gate (official source, assessment year, effective date, professional review, golden fixtures) is aimed at making rule correctness checkable. That is a goal, not a delivered capability: no YA is approved yet.
- **Fail closed:** unapproved claims never reduce chargeable income and unknown cases are `REVIEW_REQUIRED`.
- **Open source:** the logic can be inspected and challenged.

## Where mytax is behind

- No OCR, accounts, vault or business and rental tools. Several are blocked on purpose until the security, PDPA and review gates in `AGENTS.md` pass.
- No approved rule package for any year, while competitors already advertise YA 2026 support.
- No mobile app, billing or support channel.

## Risks to watch

- Coverage race: users compare features first, and mytax cannot match rule coverage until its review gate is cleared.
- A free open-source tool competes on trust, not features. That only works if the trust claim is demonstrable (published sources, tests and review status).
- Competitor privacy claims are not always consistent with their own disclosures. The Fiskal listing says it advertises nothing, yet its App Privacy section lists contact info under advertising or marketing use. Do not copy privacy or legal wording from competitors; get it reviewed.

## Open questions

- How accurate are competitors' rules and OCR? Needs a legitimate test with synthetic data, not real taxpayer data.
- Who else serves this need (payroll and HR portals, accounting suites, other apps)? Only two web searches were run.
- Should "provable, source-linked rules per year" be mytax's stated differentiator? That is a product decision, not a finding.
