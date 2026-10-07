# memoguard-rules

Versioned privacy policies for MemoGuard. This is a Go library with no network, Stellar parser, CLI, or GitHub-specific code.

## Owns

- Policy schema and validation.
- Stable rule IDs, severity levels, and confidence metadata.
- Built-in policy pack and scoped, expiring exceptions.
- Policy compatibility and release notes.

## Planned package contract

`Load`, `Validate`, and `Default` produce a validated policy. Rules describe patterns and field scopes; they never handle transaction decoding.

## First implementation slice

Define the schema, load a JSON policy, reject duplicates and invalid exceptions, and test the built-in email/phone/token fixtures. The downstream consumer is `memoguard-engine`.

Product PRD and architecture live in the parent `memguard/docs` folder in the local workspace.
