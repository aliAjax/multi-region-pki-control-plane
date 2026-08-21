# Bug Reproduction

- Bug: issuance failure leaves the certificate in `issuing`, ignores cancellation in CA creation, and drops certificate/parent error sentinels.
- Trigger: fail CA or PEM parsing after issuance starts, cancel CA creation, and inspect the returned errors.
- Error: `TestIssuanceFailureRollsBackState` reports `state issuing err ca certificate unavailable`; cancellation and PEM tests report ignored/lost errors.
