package custd

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestIngestRawLinesPostsToTheRawEndpointWithTheParserBinding(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true,"processed":2,"failed":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	result, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content:    "first line\nsecond line\n",
		ParserUUID: "11111111-1111-1111-1111-111111111111",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := doer.requests[0]
	if req.Method != http.MethodPost || !strings.HasSuffix(req.URL, "/api/v1/raw") {
		t.Fatalf("request = %s %s", req.Method, req.URL)
	}
	if req.Headers["X-Parser-UUID"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("parser header = %q", req.Headers["X-Parser-UUID"])
	}
	if string(req.Body) != "first line\nsecond line\n" {
		t.Fatalf("body = %q", req.Body)
	}
	// A 202 alone is not success: the body carries the per-record counts.
	if !result.Success || result.Processed != 2 || result.Failed != 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestIngestRawLinesAppendsThePairedSourceTarget(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true,"processed":3,"failed":1}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	result, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content:    "line\n",
		ParserUUID: "11111111-1111-1111-1111-111111111111",
		Target:     &RawSourceTarget{SubClient: "acme-store", Resource: "srv-01"},
	})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := url.Parse(doer.requests[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("subClient") != "acme-store" || query.Get("resource") != "srv-01" {
		t.Fatalf("query = %v", query)
	}
	// A 202 with refused lines is surfaced, not swallowed as success.
	if result.Processed != 3 || result.Failed != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestIngestRawLinesValidatesBeforeSending(t *testing.T) {
	cases := []struct {
		name string
		req  RawIngestRequest
	}{
		{"missing parser", RawIngestRequest{Content: "x"}},
		{"malformed parser", RawIngestRequest{Content: "x", ParserUUID: "not-a-uuid"}},
		{"partial target", RawIngestRequest{
			Content: "x", ParserUUID: "11111111-1111-1111-1111-111111111111",
			Target: &RawSourceTarget{SubClient: "acme-store"},
		}},
		{"malformed code", RawIngestRequest{
			Content: "x", ParserUUID: "11111111-1111-1111-1111-111111111111",
			Target: &RawSourceTarget{SubClient: "bad code", Resource: "srv-01"},
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusAccepted, `{"success":true}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			if _, err := client.IngestRawLines(context.Background(), testCase.req); err == nil {
				t.Fatal("expected a validation error")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("validation should not cost a request, sent %d", len(doer.requests))
			}
		})
	}
}

// rawAcceptanceUnknown asserts the error is the typed unknown-acceptance error.
func rawAcceptanceUnknown(t *testing.T, err error) {
	t.Helper()
	var unknown *RawAcceptanceUnknownError
	if !errors.As(err, &unknown) {
		t.Fatalf("err = %v, want *RawAcceptanceUnknownError", err)
	}
}

func TestIngestRawLinesRejectsMalformedAcceptanceBodies(t *testing.T) {
	cases := map[string]string{
		"empty body":           ``,
		"top-level null":       `null`,
		"empty object":         `{}`,
		"missing failed":       `{"success":true,"processed":2}`,
		"missing processed":    `{"success":true,"failed":0}`,
		"missing success":      `{"processed":2,"failed":0}`,
		"null failed":          `{"success":true,"processed":2,"failed":null}`,
		"wrong success type":   `{"success":"yes","processed":2,"failed":0}`,
		"wrong processed type": `{"success":true,"processed":"2","failed":0}`,
		"negative processed":   `{"success":true,"processed":-1,"failed":0}`,
		"negative failed":      `{"success":true,"processed":2,"failed":-1}`,
		"fractional processed": `{"success":true,"processed":2.5,"failed":0}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusAccepted, body)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			_, err := client.IngestRawLines(context.Background(), RawIngestRequest{
				Content: "x\n", ParserUUID: "11111111-1111-1111-1111-111111111111",
			})
			rawAcceptanceUnknown(t, err)
		})
	}
}

func TestIngestRawLinesPreservesLegitimatePartialCounts(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true,"processed":3,"failed":1}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	result, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content: "x\n", ParserUUID: "11111111-1111-1111-1111-111111111111",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.Processed != 3 || result.Failed != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestIngestRawLinesClassifiesTransportFailureAsUnknownAcceptance(t *testing.T) {
	client := newAdminTestClient(t, &captureDoer{}, "http://localhost:8080")
	client.config.HTTPClient = failingDoer{}
	_, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content: "x\n", ParserUUID: "11111111-1111-1111-1111-111111111111",
	})
	rawAcceptanceUnknown(t, err)
}

func TestIngestRawLinesClassifiesRetryableStatusAsUnknownAcceptance(t *testing.T) {
	doer := newCaptureDoer(http.StatusServiceUnavailable, `{"title":"unavailable","status":503}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content: "x\n", ParserUUID: "11111111-1111-1111-1111-111111111111",
	})
	rawAcceptanceUnknown(t, err)
}

func TestIngestRawLinesKeepsNonRetryableStatusAsADefiniteRejection(t *testing.T) {
	doer := newCaptureDoer(http.StatusBadRequest, `{"title":"invalid","status":400}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content: "x\n", ParserUUID: "11111111-1111-1111-1111-111111111111",
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	var unknown *RawAcceptanceUnknownError
	if errors.As(err, &unknown) {
		t.Fatalf("400 must stay a definite rejection, got %v", err)
	}
}

const rawKeyTestUUID = "11111111-1111-1111-1111-111111111111"

func TestNewRawIngestRequestMintsOneRetainedKey(t *testing.T) {
	req := NewRawIngestRequest(RawIngestInput{Content: "x\n", ParserUUID: rawKeyTestUUID})
	if err := req.IdempotencyKey.Validate(); err != nil {
		t.Fatalf("minted key invalid: %v", err)
	}
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true,"processed":1,"failed":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	// The same request value resent retains the key; the SDK never mints per call.
	for i := 0; i < 2; i++ {
		if _, err := client.IngestRawLines(context.Background(), req); err != nil {
			t.Fatal(err)
		}
	}
	first := doer.requests[0].Headers[IdempotencyKeyHeader]
	second := doer.requests[1].Headers[IdempotencyKeyHeader]
	if first == "" || first != second || first != string(req.IdempotencyKey) {
		t.Fatalf("keys = %q / %q, want retained %q", first, second, req.IdempotencyKey)
	}
}

func TestNewRawIngestRequestMintsFreshKeysPerObject(t *testing.T) {
	first := NewRawIngestRequest(RawIngestInput{Content: "x\n", ParserUUID: rawKeyTestUUID})
	second := NewRawIngestRequest(RawIngestInput{Content: "x\n", ParserUUID: rawKeyTestUUID})
	if first.IdempotencyKey == second.IdempotencyKey {
		t.Fatalf("identical text must still get a new key for a genuine occurrence: %q", first.IdempotencyKey)
	}
}

func TestRawRequestIdempotencyKeyRejectsMalformed(t *testing.T) {
	cases := []RawRequestIdempotencyKey{
		"", "1234", "1234:", ":11111111-1111-4111-8111-111111111111",
		"0:11111111-1111-4111-8111-111111111111",
		"1234:11111111-1111-1111-1111-111111111111",
		"1234:11111111-1111-4111-1111-11111111111",
	}
	for _, key := range cases {
		if err := key.Validate(); err == nil {
			t.Fatalf("key %q should be rejected", key)
		}
	}
}

func TestIngestRawLinesSurfacesTheOperationKeyOnUnknownAcceptance(t *testing.T) {
	key := NewRawRequestIdempotencyKey()
	client := newAdminTestClient(t, &captureDoer{}, "http://localhost:8080")
	client.config.HTTPClient = failingDoer{}
	_, err := client.IngestRawLines(context.Background(), RawIngestRequest{
		Content: "x\n", ParserUUID: rawKeyTestUUID, IdempotencyKey: key,
	})
	var unknown *RawAcceptanceUnknownError
	if !errors.As(err, &unknown) {
		t.Fatalf("err = %v, want *RawAcceptanceUnknownError", err)
	}
	if unknown.IdempotencyKey != key {
		t.Fatalf("operation key = %q, want %q", unknown.IdempotencyKey, key)
	}
}

func TestIngestRawLinesKeepsConflictAndExpiryDefinite(t *testing.T) {
	for _, status := range []int{http.StatusConflict, http.StatusBadRequest} {
		doer := newCaptureDoer(status, `{"title":"idempotency conflict","status":409}`)
		client := newAdminTestClient(t, doer, "http://localhost:8080")
		_, err := client.IngestRawLines(context.Background(), NewRawIngestRequest(RawIngestInput{
			Content: "x\n", ParserUUID: rawKeyTestUUID,
		}))
		if err == nil {
			t.Fatalf("status %d: expected an error", status)
		}
		var unknown *RawAcceptanceUnknownError
		if errors.As(err, &unknown) {
			t.Fatalf("status %d must stay a definite rejection, got %v", status, err)
		}
	}
}
