# Milestone 7 — intake classification and policy release gates

Source of truth: https://www.hasil.gov.my/individu/pelepasan-cukai/ and official YA2025 Form BE explanatory notes.

The module classifies candidate income sources (employment, pension, interest, rental, royalty, dividend, business and other). Classification is intentionally preliminary. No raw amount is converted into statutory income automatically. Business income triggers Form B review and dividends trigger special tax review.

The release-record validator requires named reviewer, review date, snapshot SHA-256, approval reference and complete official source metadata. This structural validator is **not sufficient to authorize activation**: an independently reviewed, trusted release registry and statutory golden tests are still required. ActivePackage currently always fails closed.

Existing draft relief caps are useful examples, not complete official conditions. Filing readiness remains false for all milestone 7 workflows.

Next gate: complete Form BE crosswalk from primary sources, rule-level legal conditions, independent Malaysian tax professional sign-off and official calculation fixtures.
