# Versioned statutory rules — milestone 2

## Status
**YA 2025 draft sample catalog only. No rules legally approved. No YA 2026 numbers silently copied.**

Source for YA 2025 relief limits: https://www.hasil.gov.my/individu/pelepasan-cukai/

Includes demonstration of:
- Shared group limit (lifestyle RM2,500, across multiple expense types)
- Nested sublimit (eligible vaccination RM1,000 inside medical group RM10,000)
- Nested sublimit (self skills education RM2,000 inside education group RM7,000)
- Required facts, document presence and duplicate flags
- Stable allocation of capped amounts by claim ID

**Caution:** These are purposefully incomplete rule conditions. A successful illustrative check is never LHDN approval and cannot be used for an actual tax return. The medical, education and lifestyle categories have additional legal requirements and sublimits. The 'APPROVED' state in unit tests is a fixture to verify arithmetic, not a shipped signed policy.

## Acceptance checklist for activation
- Primary sources (Act, applicable orders, Public Rulings, explanatory notes) checked for exact YA
- Statutory requirements modeled in full, not inferred from simple yes/no fields
- Rule source snapshot, review date, approver and review trail
- End-to-end tax computation reconciled to LHDN examples
- Spouse/child, duplicates across categories, reimbursement and record retention tests
- Nonresident, refunds, PCB and tax rebates handled with correct statutory order
- Professional review completed

No automatic tax return preparation or upload of sensitive documents is supported in this milestone.
