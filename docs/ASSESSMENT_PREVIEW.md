# Milestone 4: illustrative tax assessment and draft claim reconciliation

Endpoint: POST /v1/assessment/preview

Example request (all monetary amounts are integer sen):
```json
{"assessment_year":2025,"declared_chargeable_income_sen":7000000,"claims":[]}
```

Returns:
- marginal YA2025 **gross** tax breakdown;
- deterministic draft relief screening and its provenance/version;
- reconciliation that proves **no draft relief was subtracted**;
- warnings and filing_ready=false.

**Not an official Form BE reconciliation.** A proper Form BE field mapping, statutory rounding, confirmed income computations, approved relief amounts, rebates, credits and PCB will require reviewed official sources and more tests.

Security: this unauthenticated endpoint is a calculation prototype; do not send identity or personal documents.

## Claim shape (changed 2026-11-03)

`claims` use the same shape as `POST /v1/eligibility/preview`:

```json
{"id":"c1","rule_id":"MEDICAL_VACCINATION","amount_sen":50000,"expense_key":"inv-123","evidence_reference":"doc-9","facts":{"resident":true}}
```

`evidence_present` and `duplicate` no longer exist: the server derives duplicates from `expense_key` and checks that `evidence_reference` is set. Decisions use `claim_id`, `status`, `provisional_amount_sen`, `approved_amount_sen` and `reasons`; the result adds `provisional_total_sen` and `filing_ready` (always false).

Further differences:
- New reason codes `EXPENSE_IDENTIFIER_MISSING` and `MUTUALLY_EXCLUSIVE_RELIEF` can appear.
- The two first-home tiers share an exclusive family, so only one first-home claim (including a second claim in the same tier) is screened per request.
- The error message "claim amount overflow" is now "overflow".
- `MEDICAL_DENTAL` is a new rule.
