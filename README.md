<p align="center"><img src="assets/logo.svg" alt="MemoGuard logo" width="112"></p>

# memoguard-rules

Versioned privacy policies for MemoGuard. This is a Go library with no network, Stellar parser, CLI, or GitHub-specific code.

## Owns

- Policy schema and validation.
- Stable rule IDs, severity levels, and confidence metadata.
- Built-in policy pack and scoped, expiring exceptions.
- Policy compatibility and release notes.

## Package contract

`Load`, `Validate`, and `Default` produce a validated policy. Rules describe patterns and field scopes; they never handle transaction decoding.

`Default()` supplies email, phone, and credential rules. `Load(io.Reader)` parses one strict JSON policy up to 1 MiB and validates it. `Policy.Validate()` rejects duplicate IDs, invalid regexes and scopes, and incomplete exceptions. Each exception names one exact field path, a reason, and an expiry; it cannot turn off a whole class of fields. `memoguard-engine` consumes this package and owns the actual scan.

Product PRD and architecture live in the parent `memguard/docs` folder in the local workspace.
