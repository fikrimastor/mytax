# Spouse and child candidates, s.6A rebates, dividend review (YA 2025, draft)

Status: **DRAFT. Not approved, not reviewed by a tax professional, not wired into any API endpoint or the assessment flow.** Nothing here reduces chargeable income or tax in a filing sense.

Sources: Income Tax Act 1967 (Act 53) updated texts as at 21 May 2024 and 1 January 2026 (not authentic texts), Finance Act 2024 (Act 862) s.8, s.11, s.12, s.16, s.17. See `docs/ACT53_RECONCILIATION_YA2025.md` and `docs/sources/act53/`.

## `internal/family` (candidates only, YA 2025 only)
Every result has `approved_total_sen` 0, `filing_ready` false, state `REVIEW_REQUIRED`, and a `DRAFT_LEGAL_RULE_NOT_APPROVED` reason when a candidate amount is offered. All facts are caller-asserted and unverified.

### Children, s.48
- Amounts: RM2,000 (under 18, full-time instruction, articles or indentures); RM8,000 for a disabled child (s.48(2)(b), Finance Act 2024); s.48(3)(a)(i): four times RM2,000 for a child over 18 in approved higher education or serving articles; s.48(3)(a)(ii): a further RM8,000 for a disabled child in the same situation. Outside Malaysia the course must be a degree or equivalent.
- s.48(4): two claimants for the same child get 50% each.
- s.48(1) proviso: a wife living with her husband and assessed separately may elect that the whole deduction is allowed to her. The election needs exactly two claimants and the separate-assessment fact; otherwise both are held for review.
- s.48(5): no deduction if the child's total income exceeds the deduction. Income equal to the deduction is allowed.
- Duplicate claim IDs, duplicate child IDs and negative income are rejected. The same claimant claiming the same child twice counts once.

### Spouse, s.47 (husband claims) and s.45A (wife claims)
- s.47(1)(a) RM4,000; s.47(1)(b) a further RM6,000 for a disabled wife (Finance Act 2024 s.11, YA 2025); s.45A(1) the same for a wife claiming a husband (Finance Act 2024 s.8).
- s.47(2) alimony adds to the s.47(1)(a) amount, and s.47(3) caps the two together at RM4,000. s.47(4): a wife assessed separately gives no s.47(1) deduction (alimony under s.47(2) is separate).
- s.47(6) / s.45A(2): spouse foreign-sourced income above RM4,000 blocks the deduction unless the spouse is disabled. Income equal to RM4,000 is allowed.
- The further disabled amount is offered only if the base conditions hold and the disability is evidenced.

### Readings and choices (not settled by the texts; fail closed)
1. **s.48(5) basis.** The Act compares child income with "the deduction that would otherwise be allowed to that individual". It does not say whether that is before or after the s.48(4) halving. When income is above the halved share but not above the whole, the result is `CHILD_INCOME_TEST_BASIS_UNRESOLVED` with 0.
2. **More than two claimants.** s.48(4) says each gets 50%. With three or more the result is `MORE_THAN_TWO_CLAIMANTS_UNRESOLVED` with 0.
3. **Foreign income test is always applied** to a non-disabled spouse. In s.47 it sits in subsection (6), which qualifies subsection (5) (a wife with no total income). The package does not know which case applies, so it requires verified foreign income (0 is fine) before it offers the base amount.
4. **Disability, institution approval and "paid maintenance"** are yes/no facts. Nothing verifies them.

## `internal/calculation/rebate.go` (provisional s.6A rebates)
- RM400 each for: an individual allowed s.46(1)(a), a spouse deduction under s.47(1) or (2), a wife allowed s.45A. Each only if chargeable income is RM35,000 or less (RM35,000 exactly qualifies, one sen more does not).
- Zakat, fitrah or other obligatory dues paid in the basis year (s.6A(3)), only with a receipt from a statutory religious authority. No income test.
- The total is capped at the tax charged. The excess is lost: **no refund and no carry-forward** (s.6A(4)); the result has `refund_sen` 0 and shows the lost amount.
- Result status `PROVISIONAL_NOT_FILING_READY`. "Was allowed a deduction" and the receipt are caller-asserted facts.
- **Not implemented:** the umrah departure-levy rebate (s.6A(2A), (2B)) needs a count of earlier rebated trips across years; the rule that the RM35,000 test uses income from all sources where Part XIV or XV of Schedule 1 applies; PCB and s.132 / s.133 credits.

## `internal/calculation/dividend.go` (review only, no tax computed)
Finance Act 2024: Schedule 1 Part XXII (YA 2025 onward) charges an individual's dividend deemed derived from Malaysia "in excess of one hundred thousand ringgit" at 2%, "notwithstanding Part I"; Schedule 6 paragraph 12B(2) exempts a dividend to an individual of RM100,000 or less where the paying company is not entitled to deduct tax. The supplied texts do not settle:
- whether only the excess or the whole dividend is charged once the total passes RM100,000;
- whether RM100,000 is per company or across all companies;
- the chargeable income "as prescribed by the Minister" where the individual has other income.

So every call returns `REVIEW_REQUIRED` and `computed` false. The only number is `illustrative_excess_tax_sen`, 2% of the excess rounded half-up to the sen, labelled illustration. YA other than 2025 is rejected.

## Tests
- `internal/family`: amounts per category, sharing, election, income-test boundaries, alimony cap, foreign-income boundary, s.45A, fail-closed facts, order independence, rejected input.
- `internal/calculation`: rebate boundaries (RM35,000, exact tax cap, zakat receipt), real gross tax of RM450 at RM30,000 income less RM400, dividend boundaries and rounding.
- Mutation checks (50% split, income test, RM4,000 cap, foreign boundary, three claimants, rebate threshold, tax cap, zakat receipt, dividend rounding) each make a test fail.

## Still not done
- Wiring these into `assessment` and the API, and any UI.
- s.46(3) and s.50(2) spouse deeming (needs the s.45(2) election context), s.49 insurance conditions, s.51 deduction must be claimed.
- Compost machine claimable once across YA 2025 to 2027 (needs data across years). Breastfeeding once per two years.
- Statutory rounding. The supplied texts have no rounding rule.
- Enabling YA 2023 and YA 2024: a decision for the repository owner.
- Tax-professional review. CI has not run on this branch. Nine PostgreSQL integration tests skip without a database (seven use the two-tenant `needDB` fixture), and CI fails on any skipped test.
