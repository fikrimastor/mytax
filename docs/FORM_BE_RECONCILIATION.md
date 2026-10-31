# YA2025 indicative Form BE arithmetic — milestone 5

Official form notes: https://www.hasil.gov.my/media/pshpbomm/explanatorynotes_be2025_2.pdf

POST /v1/reconciliation/preview with integer sen amounts:
```json
{"assessment_year":2025,"declared_total_income_sen":7000000,"draft_reliefs_sen":200000,"verified_reliefs_sen":0,"entered_rebates_sen":0,"pcb_sen":100000}
```

This is a deliberately limited arithmetic preview and **not** a completed Form BE:
- The declared income is used as hypothetical chargeable income because statutory income aggregation, allowable deductions and verified reliefs are not complete.
- Draft reliefs are never subtracted.
- The server refuses nonzero client-supplied verified reliefs or rebates.
- PCB is provisionally subtracted from gross tax, for informational arithmetic only.
- Dividend-related Part XXII computation, credits, joint assessment, rounding and all relevant special cases remain out of scope.
- Never represent the balance as an LHDN-certified liability or refund.
- No production filing should use this API result.

The complete statutory integration requires an approved policy package, Form BE crosswalk by field, documented user input provenance, audit trails and reviewed fixture tests.
