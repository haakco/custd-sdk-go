package custd

import (
	"net/http"
	"testing"
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
		call   func(*AuthProjectAdminClient) (*AuthProjectDomainBinding, error)
	}{
		{http.MethodGet, "", AuthProjectDomainActive, func(c *AuthProjectAdminClient) (*AuthProjectDomainBinding, error) {
			return c.GetEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPut, "", AuthProjectDomainPending, func(c *AuthProjectAdminClient) (*AuthProjectDomainBinding, error) {
			return c.BindEnvironmentAuthDomain(ctx, projectID, environmentID, request, options)
		}},
		{http.MethodPost, "/retry", AuthProjectDomainPending, func(c *AuthProjectAdminClient) (*AuthProjectDomainBinding, error) {
			return c.RetryEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPost, "/disable", AuthProjectDomainDisabled, func(c *AuthProjectAdminClient) (*AuthProjectDomainBinding, error) {
			return c.DisableEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
		{http.MethodPost, "/release", AuthProjectDomainReleasing, func(c *AuthProjectAdminClient) (*AuthProjectDomainBinding, error) {
			return c.ReleaseEnvironmentAuthDomain(ctx, projectID, environmentID, options)
		}},
	} {
		t.Run(check.method+check.suffix, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{
				"environmentId":"environment/two","siteDomainId":"site-domain-id","host":"`+host+`",
				"issuer":"`+issuer+`","status":"`+string(check.status)+`","dnsTarget":"edge.example.test",
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
				binding.CreatedAt != "2026-11-19T09:00:00Z" || binding.ObservedAt != "2026-11-19T09:01:00Z" ||
				binding.CertificateExpiresAt != "2027-02-19T09:02:00Z" {
				t.Fatalf("binding = %+v", binding)
			}
		})
	}
}
