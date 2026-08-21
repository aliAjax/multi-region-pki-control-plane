# Bug Reproduction

- Bug: retry backoff ignores an already-cancelled or subsequently-cancelled context.
- Trigger: cancel the request before a retry, or cancel it while the backoff timer is waiting in both retry paths.
- Error: `TestBreakerStopsWhenContextEnds` reports `retry continued after cancellation: 3 calls`; pre-cancel tests report `pre-cancel ignored`.
