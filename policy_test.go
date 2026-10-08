package rules

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestGovernmentIDExampleLoads(t *testing.T) {
	file, err := os.Open("examples/government-id-policy.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	policy, err := Load(file)
	if err != nil || len(policy.Rules) != 1 || policy.Rules[0].ID != "custom.government_id" {
		t.Fatalf("example policy: %#v, %v", policy, err)
	}
}

func TestDefaultValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	input := `{"version":"v1","rules":[],"unexpected":true}`
	if _, err := Load(strings.NewReader(input)); err == nil {
		t.Fatal("expected strict JSON decoding error")
	}
}

func TestExceptionIsExactAndExpires(t *testing.T) {
	p := Default()
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	p.Exceptions = []Exception{{RuleID: "personal.email", FieldPath: "transaction.memo.text", Reason: "public support inbox", ExpiresAt: now.Add(time.Hour)}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if !p.IsExcepted("personal.email", "transaction.memo.text", now) {
		t.Fatal("expected scoped exception")
	}
	if p.IsExcepted("personal.email", "transaction.operations.0.destination", now) || p.IsExcepted("personal.email", "transaction.memo.text", now.Add(2*time.Hour)) {
		t.Fatal("exception escaped its scope or expiry")
	}
}

func FuzzPolicyLoadDoesNotPanic(f *testing.F) {
	f.Add([]byte(`{"version":"v1","rules":[]}`))
	f.Add([]byte(`{"version":"v1","rules":[{"id":"x","description":"x","remediation":"x","pattern":"x","scopes":["*"],"severity":"warning","confidence":"low"}]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > MaxPolicyBytes {
			return
		}
		_, _ = Load(strings.NewReader(string(data)))
	})
}
