# Bug Reproduction

- Bug: revocation audit, outbox, CRL repository, and issuer failures lose their original error chain.
- Trigger: make each downstream persistence or certificate parsing operation return an error during revoke/CRL generation.
- Error: `TestRevokePropagatesAuditAndOutboxProblem` reports `expected outbox failure, got outbox unavailable`; audit and CRL tests report lost errors.
