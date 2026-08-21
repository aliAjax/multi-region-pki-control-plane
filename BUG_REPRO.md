# Bug Reproduction

- Bug: fixed-clock and certificate lifecycle checks disagree at validity and renewal boundaries, so OCSP can use stale status.
- Trigger: evaluate issued/published/renewing certificates before and after their validity window with an injected fixed clock.
- Error: `TestLifecycleTransitionAndActivity` reports `expired issued certificate should not be active`; the fixed-clock boundary tests fail as well.
