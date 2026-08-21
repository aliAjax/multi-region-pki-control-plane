# Bug Reproduction

- Bug: replication vectors and conflict snapshots expose shared map storage.
- Trigger: merge a remote vector, mutate the merged result or cursor, and then inspect the local/fencing state.
- Error: `TestVectorMergeDoesNotMutateInputs` reports `left vector was contaminated` and snapshot tests report `alias`.
