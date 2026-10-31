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
