# Bug Reproduction

- Bug: the revoke HTTP handler and public error helpers discard not-found classification and wrapped causes.
- Trigger: revoke a certificate ID that is absent and inspect the response/error chain.
- Error: `TestRevokeEndpointReturnsNotFoundForMissingCertificate` reports `missing certificate mapped to 400` with code `invalid_request`.
