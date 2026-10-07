package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

const MaxPolicyBytes = 1 << 20

type Severity string

const (
	Warning Severity = "warning"
	Block   Severity = "block"
)

type Confidence string

const (
	Low    Confidence = "low"
	Medium Confidence = "medium"
	High   Confidence = "high"
)

type Rule struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Remediation string     `json:"remediation"`
	Pattern     string     `json:"pattern"`
	Scopes      []string   `json:"scopes"`
	Severity    Severity   `json:"severity"`
	Confidence  Confidence `json:"confidence"`
}

type Exception struct {
	RuleID    string    `json:"rule_id"`
	FieldPath string    `json:"field_path"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Policy struct {
	Version    string      `json:"version"`
	Rules      []Rule      `json:"rules"`
	Exceptions []Exception `json:"exceptions,omitempty"`
}

func Load(reader io.Reader) (Policy, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxPolicyBytes+1))
	if err != nil {
		return Policy{}, fmt.Errorf("read policy: %w", err)
	}
	if len(data) > MaxPolicyBytes {
		return Policy{}, errors.New("policy exceeds size limit")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return Policy{}, fmt.Errorf("decode policy: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Policy{}, errors.New("policy must contain one JSON object")
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (p Policy) Validate() error {
	if strings.TrimSpace(p.Version) == "" {
		return errors.New("policy version is required")
	}
	if len(p.Rules) == 0 || len(p.Rules) > 256 {
		return errors.New("policy must contain 1 to 256 rules")
	}
	seen := make(map[string]struct{}, len(p.Rules))
	for _, rule := range p.Rules {
		if rule.ID == "" || len(rule.ID) > 80 || strings.TrimSpace(rule.Description) == "" || strings.TrimSpace(rule.Remediation) == "" {
			return fmt.Errorf("rule id, description, and remediation are required")
		}
		if _, exists := seen[rule.ID]; exists {
			return fmt.Errorf("duplicate rule %q", rule.ID)
		}
		seen[rule.ID] = struct{}{}
		if rule.Severity != Warning && rule.Severity != Block {
			return fmt.Errorf("rule %q has invalid severity", rule.ID)
		}
		if rule.Confidence != Low && rule.Confidence != Medium && rule.Confidence != High {
			return fmt.Errorf("rule %q has invalid confidence", rule.ID)
		}
		if len(rule.Pattern) == 0 || len(rule.Pattern) > 4096 {
			return fmt.Errorf("rule %q has invalid pattern length", rule.ID)
		}
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("rule %q has invalid pattern: %w", rule.ID, err)
		}
		if len(rule.Scopes) == 0 || len(rule.Scopes) > 32 {
			return fmt.Errorf("rule %q needs 1 to 32 scopes", rule.ID)
		}
		for _, scope := range rule.Scopes {
			if !validScope(scope) {
				return fmt.Errorf("rule %q has invalid scope %q", rule.ID, scope)
			}
		}
	}
	for _, exception := range p.Exceptions {
		if _, exists := seen[exception.RuleID]; !exists {
			return fmt.Errorf("exception references unknown rule %q", exception.RuleID)
		}
		if !validScope(exception.FieldPath) || exception.FieldPath == "*" || strings.HasSuffix(exception.FieldPath, ".*") {
			return errors.New("exception must target one exact field path")
		}
		if strings.TrimSpace(exception.Reason) == "" || exception.ExpiresAt.IsZero() {
			return errors.New("exception needs a reason and expiry")
		}
	}
	return nil
}

func validScope(scope string) bool {
	if scope == "*" {
		return true
	}
	if scope == "" || len(scope) > 256 || strings.ContainsAny(scope, " \t\r\n") {
		return false
	}
	if strings.Contains(scope, "*") {
		return strings.HasSuffix(scope, ".*") && strings.Count(scope, "*") == 1
	}
	return true
}

func MatchesScope(scope, fieldPath string) bool {
	if scope == "*" {
		return true
	}
	if strings.HasSuffix(scope, ".*") {
		return strings.HasPrefix(fieldPath, strings.TrimSuffix(scope, "*"))
	}
	return scope == fieldPath
}

func (p Policy) IsExcepted(ruleID, fieldPath string, now time.Time) bool {
	for _, exception := range p.Exceptions {
		if exception.RuleID == ruleID && exception.FieldPath == fieldPath && now.Before(exception.ExpiresAt) {
			return true
		}
	}
	return false
}

func Default() Policy {
	return Policy{Version: "v1", Rules: []Rule{
		{
			ID: "personal.email", Description: "Email address in public transaction data",
			Remediation: "Use an opaque reference instead of an email address",
			Pattern:     `(?i)\b[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}\b`,
			Scopes:      []string{"*"}, Severity: Block, Confidence: High,
		},
		{
			ID: "personal.phone", Description: "Possible international phone number in public transaction data",
			Remediation: "Use an opaque reference and confirm the number is not public data",
			Pattern:     `\+[1-9][0-9 ()-]{7,}[0-9]`,
			Scopes:      []string{"*"}, Severity: Warning, Confidence: Medium,
		},
		{
			ID: "secret.token", Description: "Possible credential in public transaction data",
			Remediation: "Remove the credential and rotate it if it may have been exposed",
			Pattern:     `(?i)\b(?:api[_-]?key|secret|token)\s*[:=]\s*[A-Za-z0-9_\-]{16,}\b`,
			Scopes:      []string{"*"}, Severity: Block, Confidence: High,
		},
	}}
}
