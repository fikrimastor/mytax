# Milestone 9 — expense-level relief candidate assessment

**Update 2026-11-03:** this endpoint is now served by `internal/policy`. Response fields follow the policy shape (`claim_id`, `status`, `provisional_amount_sen`, `approved_amount_sen`, `provisional_total_sen`, `filing_ready`); rule ids `EDUCATION_UPSKILL` and `EDUCATION_OTHER` became `EDUCATION_SELF_SKILLS` and `EDUCATION_SELF_OTHER`.

Other differences from the previous eligibility response:
- Top-level `status` was removed; `package_version`, `package_status` and `notice` were added.
- Reason codes renamed: `UNKNOWN_RULE` to `UNKNOWN_RELIEF_RULE`, `FACT_NOT_VERIFIED:x` to `UNVERIFIED_FACT:x`, `DRAFT_LEGAL_RULE_NOT_APPROVED` to `DRAFT_POLICY_NOT_APPROVED`.
- Education and dental rules now use the policy fact lists (`within_year`, `qualifying_course`, `own_education`, `qualifying_beneficiary`, `no_other_claim`) instead of `incurred_in_ya`, `recognized_course`, `qualifying_education`, `qualifying_person`. Clients sending the old names get `UNVERIFIED_FACT` reasons.
- Error messages changed: duplicate claim id returns "duplicate claim ID", a bad amount returns "claim ID and positive amount required", total overflow returns "overflow".

Primary official basis: https://www.hasil.gov.my/individu/pelepasan-cukai/ (YA 2025) and https://www.hasil.gov.my/media/d1rhwb05/pelepasan-cukai-tt-2025.pdf

POST /v1/eligibility/preview accepts assessment_year=2025 and claims with immutable ID, expense_key, rule_id, amount_sen, evidence_reference and verified facts. It deterministically orders by ID and checks shared lifestyle/education/medical caps, selected nested caps, duplicate expense identifiers and first-home tier exclusivity.

This is **not a complete or officially approved legal eligibility engine**. Facts are supplied by the user, not authenticated. Cross-spouse duplicate detection, receipt identity matching, overlapping medical sublimits, exceptions and all remaining relief rules require more work. The response carries package_status "DRAFT", a notice, approved_total_sen 0 and filing_ready false, irrespective of candidate values.

Never subtract provisional amounts from real statutory computations or describe them as allowable reliefs. Production activation requires full legal source review and professionally signed rule packages.
