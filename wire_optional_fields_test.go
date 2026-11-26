package custd

import (
	"encoding/json"
	"testing"
)

// The server omits these fields when they carry no value, so the SDK must keep both shapes: a set
// value, including a real zero, is sent, and an unset one is left out rather than sent empty.
func TestOptionalWireFieldsMatchServerOmission(t *testing.T) {
	zero := 0.0
	cases := []struct {
		name  string
		value any
		key   string
		want  string
	}{
		{"resolved zero bound input", UncertaintyBoundInput{State: UncertaintyBoundInputStateResolved, Value: &zero}, "value", "0"},
		{"missing bound input", UncertaintyBoundInput{State: UncertaintyBoundInputStateValueMissing}, "value", ""},
		{"remedy without shape", AdminSiteRemedy{Element: "x"}, "kind", ""},
		{"last audit page", AuditListCursor{}, "cursor", ""},
		{"site create default identity", AdminSiteCreate{CompanySlug: "acme"}, "identityMode", ""},
		{"zero progress point", PresentationPoint{X: "d1", Value: &zero}, "value", "0"},
		{"zero interval bound", PresentationPoint{X: "d1", LowerValue: &zero, UpperValue: &zero}, "lowerValue", "0"},
		{"temporal point", PresentationPoint{X: "d1", Date: "2026-11-26"}, "value", ""},
		{"zero reference", PresentationReference{Label: "target", Value: &zero}, "value", "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			if got := string(fields[tc.key]); got != tc.want {
				t.Fatalf("%s = %q in %s, want %q", tc.key, got, encoded, tc.want)
			}
		})
	}
}

// A platform-minted hostname needs no TXT record, so the server leaves dnsProof out.
func TestMintedDomainBindingHasNoDNSProof(t *testing.T) {
	var binding AuthDomainBinding
	body := `{"environmentId":"e","host":"h","issuer":"i","status":"active","lifecycle":"admitted",` +
		`"intentRevision":1,"images":{"expected":{},"running":{},"source":"release-default","convergence":"converged"},` +
		`"createdAt":"2026-11-19T09:00:00Z"}`
	if err := json.Unmarshal([]byte(body), &binding); err != nil || binding.DNSProof != nil {
		t.Fatalf("binding = %+v, %v", binding, err)
	}
}
