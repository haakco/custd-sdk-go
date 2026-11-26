package custd

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestEnvironmentAuthDomainLifecycle(t *testing.T) {
	ctx := t.Context()
	const projectID = "project/one"
	const environmentID = "environment/two"
	const path = "http://localhost:8080/api/v1/admin/auth-projects/project%2Fone/environments/environment%2Ftwo/domain"
	const host = "auth.example.test"
	const issuer = "https://" + host + "/"
	options := AuthProjectRequestOptions{}
	request := AuthProjectDomainBindRequest{Host: host, SiteDomainID: "site-domain-id"}

	for _, check := range []struct {
		method string
		suffix string
		status AuthProjectDomainStatus
		call   func(*AuthProjectAdminClient) (*AuthDomainBinding, error)
	}{
		{http.MethodGet, "", AuthProjectDomainActive, func(c *AuthProjectAdminClient) (*AuthDomainBinding, error) {
			return c.GetEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPut, "", AuthProjectDomainPending, func(c *AuthProjectAdminClient) (*AuthDomainBinding, error) {
			return c.BindEnvironmentAuthDomain(ctx, projectID, environmentID, request, options)
		}},
		{http.MethodPost, "/admit", AuthProjectDomainActive, func(c *AuthProjectAdminClient) (*AuthDomainBinding, error) {
			return c.AdmitEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPost, "/retry", AuthProjectDomainPending, func(c *AuthProjectAdminClient) (*AuthDomainBinding, error) {
			return c.RetryEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPost, "/disable", AuthProjectDomainDisabled, func(c *AuthProjectAdminClient) (*AuthDomainBinding, error) {
			return c.DisableEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPost, "/release", AuthProjectDomainReleasing, func(c *AuthProjectAdminClient) (*AuthDomainBinding, error) {
			return c.ReleaseEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
	} {
		t.Run(check.method+check.suffix, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{
				"environmentId":"environment/two","slotRef":"slot-a","siteDomainId":"site-domain-id",
				"host":"`+host+`","issuer":"`+issuer+`","status":"`+string(check.status)+`",
				"lifecycle":"admitted","intentRevision":7,
				"images":{"expected":{"gateway":"registry.example.test/gateway:1","hydra":"registry.example.test/hydra:1"},
					"running":{"gateway":"registry.example.test/gateway:0","hydra":"registry.example.test/hydra:0"},
					"convergence":"drifted"},
				"dnsTarget":"edge.example.test",
				"dnsProof":{"recordName":"_custd-auth.auth.example.test","recordValue":"custd-auth=proof"},
				"createdAt":"2026-11-19T09:00:00Z","observedAt":"2026-11-19T09:01:00Z",
				"certificateExpiresAt":"2027-02-19T09:02:00Z"
			}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			binding, err := check.call(client.Admin.AuthProjects)
			if err != nil {
				t.Fatalf("domain request: %v", err)
			}
			if len(doer.requests) != 1 {
				t.Fatalf("requests made: %d", len(doer.requests))
			}
			sent := doer.requests[0]
			assertAuthProjectRequest(t, sent, check.method, path+check.suffix, "", "")
			if check.method == http.MethodPut {
				body := decodeAuthProjectBody(t, sent)
				if len(body) != 2 || body["host"] != host || body["siteDomainId"] != request.SiteDomainID {
					t.Fatalf("bind body = %+v", body)
				}
			} else if len(sent.Body) != 0 {
				t.Fatalf("unexpected body for %s%s", check.method, check.suffix)
			}
			if binding.Status != check.status || binding.Issuer != issuer || binding.DNSTarget != "edge.example.test" ||
				binding.DNSProof.RecordName != "_custd-auth.auth.example.test" || binding.DNSProof.RecordValue != "custd-auth=proof" ||
				binding.CreatedAt.Format(time.RFC3339) != "2026-11-19T09:00:00Z" ||
				binding.ObservedAt.Format(time.RFC3339) != "2026-11-19T09:01:00Z" ||
				binding.CertificateExpiresAt.Format(time.RFC3339) != "2027-02-19T09:02:00Z" {
				t.Fatalf("binding = %+v", binding)
			}
			if binding.SlotRef != "slot-a" || binding.Lifecycle != AuthProjectDomainAdmitted || binding.IntentRevision != 7 {
				t.Fatalf("admission = slotRef %q, lifecycle %q, intentRevision %d",
					binding.SlotRef, binding.Lifecycle, binding.IntentRevision)
			}
			images := binding.Images
			if images.Expected.Gateway != "registry.example.test/gateway:1" ||
				images.Expected.Hydra != "registry.example.test/hydra:1" ||
				images.Running.Gateway != "registry.example.test/gateway:0" ||
				images.Running.Hydra != "registry.example.test/hydra:0" ||
				images.Convergence != AuthProjectDomainDrifted {
				t.Fatalf("images = %+v", images)
			}
		})
	}
}

// Admission refusals carry stable problem codes, so a caller branches on the code
// rather than the server's prose.
func TestEnvironmentAuthDomainAdmitRefusalsCarryProblemCodes(t *testing.T) {
	ctx := t.Context()
	for _, check := range []struct {
		name   string
		status int
		code   string
		want   string
	}{
		{"capacity exhausted", http.StatusConflict, "auth_domain_capacity_exhausted", AuthProjectDomainCapacityExhaustedCode},
		{"paused", http.StatusConflict, "auth_domain_paused", AuthProjectDomainPausedCode},
		{"retiring", http.StatusConflict, "auth_domain_retiring", AuthProjectDomainRetiringCode},
		{"retired", http.StatusConflict, "auth_domain_retired", AuthProjectDomainRetiredCode},
		{"slot uninstalled", http.StatusConflict, "auth_domain_slot_uninstalled", AuthProjectDomainSlotUninstalledCode},
		{"not bound", http.StatusNotFound, "auth_domain_not_bound", AuthProjectDomainNotBoundCode},
		{"conflict", http.StatusConflict, "auth_domain_conflict", AuthProjectDomainConflictCode},
		{"busy", http.StatusConflict, "auth_domain_busy", AuthProjectDomainBusyCode},
		{"unverified", http.StatusUnprocessableEntity, "auth_domain_unverified", AuthProjectDomainUnverifiedCode},
		{"unavailable", http.StatusServiceUnavailable, "auth_domain_unavailable", AuthProjectDomainUnavailableCode},
	} {
		t.Run(check.name, func(t *testing.T) {
			doer := newCaptureDoer(check.status, `{"type":"about:blank","title":"refused","status":`+
				`0,"detail":"admin: refused","code":"`+check.code+`"}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			_, err := client.Admin.AuthProjects.AdmitEnvironmentAuthDomain(ctx, "project-1", "environment-1", AuthProjectRequestOptions{})
			var requestErr *RequestError
			if !errors.As(err, &requestErr) || requestErr.StatusCode != check.status ||
				requestErr.Problem == nil || requestErr.Problem.Code != check.want {
				t.Fatalf("err = %#v", err)
			}
		})
	}
}
