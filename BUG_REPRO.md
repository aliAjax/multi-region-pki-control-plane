# Bug Reproduction

- Bug: typed-nil signing keys and a destroyed development HSM are treated as usable values.
- Trigger: pass a typed-nil key through signer/delete paths, then generate after destroying the HSM state.
- Error: `TestDevHSMSignerRejectsTypedNil` reports `typed nil escaped`; reinitialization panics with `assignment to entry in nil map`.
