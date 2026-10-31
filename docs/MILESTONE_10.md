# Milestone 10 — evidence fingerprints and audit integrity prototype

Endpoint: POST /v1/evidence/inspect with `items` consisting of synthetic, non-sensitive metadata only.

Features:
- SHA-256 format validation for document digests (the caller computes digest; server does NOT receive documents).
- Exact-file duplicate detection by digest.
- Candidate invoice duplicate detection by merchant+invoice key.
- Deterministic, hash-chained audit events and tampering verification tests.
- No records persisted; no upload, OCR or user identity integration.

Limitations: hash chains alone do not make a tamper-proof audit system. Production needs immutable append-only storage, trusted timestamping or keyed signing, access control, tenant isolation, encryption, malware scanning, retention/deletion policies, PDPA review, and an authenticated digest computation flow. SHA-256 of predictable personal metadata is not sufficient privacy protection: avoid raw identifiers in this endpoint.

The endpoint is disabled unless environment variable `MYTAX_ENABLE_EVIDENCE_PREVIEW=true`; even then, **send synthetic data only**. No claim is considered verified or filing-ready.
