# Bug Reproduction

- Bug: startup validator registration and lease ownership checks are not synchronized for concurrent requests.
- Trigger: register and validate challenges concurrently while lease renew/release operations overlap.
- Error: the race-enabled tests emit `WARNING: DATA RACE` and validation can report a missing validator.
