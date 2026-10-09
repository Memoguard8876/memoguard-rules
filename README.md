<p align="center"><img src="assets/logo.svg" alt="MemoGuard logo" width="112"></p>

<h1 align="center">memoguard-rules</h1>

<p align="center"><b>Validated, versioned privacy policies for Stellar transaction data.</b></p>

<p align="center">
  <a href="https://github.com/Memoguard8876/memoguard-rules/actions/workflows/ci.yml"><img src="https://github.com/Memoguard8876/memoguard-rules/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/Memoguard8876/memoguard-rules?color=blue" alt="License: MIT"></a>
  <a href="https://github.com/Memoguard8876/memoguard-rules/tags"><img src="https://img.shields.io/github/v/tag/Memoguard8876/memoguard-rules?label=release&color=brightgreen" alt="Latest release"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/Memoguard8876/memoguard-rules?color=00ADD8" alt="Go version">
  <a href="https://github.com/Memoguard8876/memoguard-rules/issues"><img src="https://img.shields.io/github/issues/Memoguard8876/memoguard-rules?color=orange" alt="Open issues"></a>
  <a href="https://github.com/Memoguard8876/memoguard-rules/issues?q=is%3Aopen+label%3A%22help+wanted%22"><img src="https://img.shields.io/badge/help%20wanted-welcome-8A2BE2" alt="Help wanted"></a>
  <img src="https://img.shields.io/badge/built%20for-Stellar-black" alt="Built for Stellar">
</p>

<p align="center">
  <a href="https://cjay-1.gitbook.io/memoguard-docs/">Documentation</a> ·
  <a href="https://github.com/Memoguard8876/memoguard-rules/releases">Releases</a> ·
  <a href="https://github.com/Memoguard8876/memoguard-rules/issues">Issues</a> ·
  <a href="CONTRIBUTING.md">Contributing</a> ·
  <a href="SECURITY.md">Security</a>
</p>

---

## What it is

`memoguard-rules` is a Go library that defines **what counts as a leak**. It holds the policy format, validates policy files, ships the built-in rules, and handles scoped, expiring exceptions. It has no network code, no Stellar parser, no CLI and no GitHub code. The [engine](https://github.com/Memoguard8876/memoguard-engine) imports it to do the scanning.

## Features

- Strict JSON policy loading. Unknown fields and trailing data are rejected.
- Rule validation: unique IDs, valid Go (RE2) patterns, valid scopes, severity and confidence values.
- Built-in policy `v1` with email, phone and credential rules.
- Field-path scopes (`*`, exact path, or trailing `.*`).
- Exceptions that name one rule, one exact field, a reason and an expiry.
- Hard limits on policy size, rule count, pattern length and scopes.
- Fuzz-tested policy parser.

## Install

```bash
go get github.com/memoguard8876/memoguard-rules@v0.1.1
```

Needs Go 1.24 or newer. No token is needed.

## Quick start

```go
package main

import (
	"fmt"
	"os"

	rules "github.com/memoguard8876/memoguard-rules"
)

func main() {
	// Built-in policy
	policy := rules.Default()
	fmt.Println(policy.Version, len(policy.Rules)) // v1 3

	// Or load and validate your own
	file, err := os.Open("policy.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	custom, err := rules.Load(file) // strict JSON, up to 1 MiB, validated
	if err != nil {
		panic(err)
	}
	_ = custom
}
```

## Built-in rules

| Rule ID | Severity | Confidence | Detects |
| --- | --- | --- | --- |
| `personal.email` | block | high | Email addresses |
| `personal.phone` | warning | medium | International numbers starting with `+` |
| `secret.token` | block | high | `api_key`, `secret` or `token`, then `:` or `=`, then 16 or more characters |

All three apply to every field (`*`). Government and customer ID formats vary, so no ID rule is on by default.

## Policy file

```json
{
  "version": "acme-v1",
  "rules": [
    {
      "id": "custom.customer_id",
      "description": "Customer ID in public transaction data",
      "remediation": "Replace the customer ID with an opaque reference",
      "pattern": "\\bCUST-[0-9]{8}\\b",
      "scopes": ["transaction.*"],
      "severity": "block",
      "confidence": "high"
    }
  ],
  "exceptions": [
    {
      "rule_id": "custom.customer_id",
      "field_path": "transaction.memo.text",
      "reason": "Public demo account",
      "expires_at": "2027-01-31T00:00:00Z"
    }
  ]
}
```

| Limit | Value |
| --- | --- |
| Policy file | 1 MiB, one JSON object |
| Rules | 1 to 256 |
| Rule ID | up to 80 characters |
| Pattern | 1 to 4096 characters |
| Scopes per rule | 1 to 32, each up to 256 characters |

> **Replace, not extend.** A policy file passed to the CLI replaces the built-in policy. Copy the built-in rules into your file if you want them.
>
> **Watch the trailing dot.** `transaction.memo.*` matches `transaction.memo.text` but not `transaction.memo`. Flat JSON such as `{"memo": "..."}` has the path `transaction.memo`, so use `transaction.*` or the exact path.

A synthetic example lives in [`examples/government-id-policy.json`](examples/government-id-policy.json). Replace its fictional `GOV-` format with a reviewed local one.

## API

```go
func Default() Policy
func Load(r io.Reader) (Policy, error)
func (p Policy) Validate() error
func (p Policy) IsExcepted(ruleID, fieldPath string, now time.Time) bool
func MatchesScope(scope, fieldPath string) bool
```

Types: `Policy`, `Rule`, `Exception`, `Severity` (`block`, `warning`), `Confidence` (`low`, `medium`, `high`).

## Develop

```bash
go test ./...
go vet ./...
go test -run '^$' -fuzz FuzzPolicyLoadDoesNotPanic -fuzztime 30s
```

No network service or credentials are needed.

## Limits of this repository

This library does not decode Stellar data and never sees a transaction. Rules are patterns: they can miss real leaks and flag harmless text. Tune them against synthetic copies of your own data.

## The MemoGuard family

MemoGuard is four independent Go repositories. Each builds from tagged releases of the one before it.

```text
memoguard-rules ──► memoguard-engine ──► memoguard-cli ──► memoguard-action
```

| Repository | Role | Latest |
| --- | --- | --- |
| [memoguard-rules](https://github.com/Memoguard8876/memoguard-rules) | Policy schema, validation, built-in rules, expiring exceptions | v0.1.1 |
| [memoguard-engine](https://github.com/Memoguard8876/memoguard-engine) | Stellar XDR decoding, field extraction, scanning, redacted findings | v0.2.1 |
| [memoguard-cli](https://github.com/Memoguard8876/memoguard-cli) | `memoguard scan` command, output formats, exit codes, release binaries | v0.2.2 |
| [memoguard-action](https://github.com/Memoguard8876/memoguard-action) | GitHub Action: pinned CLI, annotations, failure threshold | v0.2.1 |

Full guides, the field-path reference and walkthroughs are in **[MemoGuard Docs](https://cjay-1.gitbook.io/memoguard-docs/)**. Product requirements and architecture are versioned in [memoguard-cli/product/docs](https://github.com/Memoguard8876/memoguard-cli/tree/main/product/docs).

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md).

- Browse [open issues](https://github.com/Memoguard8876/memoguard-rules/issues). Labels show the type (`enhancement`, `documentation`, `testing`), `help wanted`, and a `complexity` level.
- Comment on an issue and wait to be assigned before you start.
- Use **synthetic data only** in tests and fixtures. Never commit real customer data, keys, tokens or real transactions.
- Add tests for every behaviour change, including malformed input, and a line in [CHANGELOG.md](CHANGELOG.md).

## Security

A clean scan is not a guarantee, and there has been no formal security audit. Report vulnerabilities privately through GitHub's private vulnerability reporting for this repository. See [SECURITY.md](SECURITY.md). Do not post exploit details or real private data in a public issue.

## Maintainers

| Maintainer | Role | Contact |
| --- | --- | --- |
| [Memoguard8876](https://github.com/Memoguard8876) | Organization owner, releases | [GitHub issues](https://github.com/Memoguard8876/memoguard-rules/issues) |

## Community

Ask questions and propose changes in [GitHub issues](https://github.com/Memoguard8876/memoguard-rules/issues). Read the [documentation](https://cjay-1.gitbook.io/memoguard-docs/) first; the [FAQ](https://cjay-1.gitbook.io/memoguard-docs/faq) answers the common questions.

## Contributors

<a href="https://github.com/Memoguard8876/memoguard-rules/graphs/contributors"><img src="https://contrib.rocks/image?repo=Memoguard8876/memoguard-rules" alt="Contributors"></a>

## License

[MIT](LICENSE) © MemoGuard contributors.
