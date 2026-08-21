# Bug Reproduction

- Bug: notification idempotency is recorded before a send succeeds, and provider/template failure paths do not preserve request behavior.
- Trigger: make the first provider call fail, retry the same message, and exercise provider cancellation or a missing template variable.
- Error: `TestDispatcherRetriesAfterProviderIssue` reports `retry did not reach provider`; duplicate suppression and provider cancellation tests also fail.
