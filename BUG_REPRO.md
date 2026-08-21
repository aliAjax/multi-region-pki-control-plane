# Bug Reproduction

- Bug: certificate snapshots share mutable DNS name storage across repository boundaries.
- Trigger: save a certificate, mutate its DNS slice, or mutate a returned snapshot while concurrent reads/listing run.
- Error: `TestMemoryStoreCopiesCertificateInputs` reports `input slice escaped`; the race-enabled snapshot test reports `WARNING: DATA RACE`.
