package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestSiteProbesDNSExpectationsAndClear(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"probes":[{"kind":"dns","expectedAddresses":["1.1.1.1"],"status":"down","detail":{"addresses":["8.8.8.8"],"missingAddresses":["1.1.1.1"],"unexpectedAddresses":["8.8.8.8"]}}]}`)
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	_, err := domains.SetProbes(ctx, "d-1", []AdminSiteProbeSpec{
		{Kind: AdminSiteProbeDNS, ExpectedAddresses: []string{"1.1.1.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := domains.ListProbes(ctx, "d-1")
	if err != nil || len(list.Probes) != 1 || list.Probes[0].Status != AdminSiteUptimeDown || list.Probes[0].Detail == nil {
		t.Fatalf("probe list = %+v, %v", list, err)
	}
	detail := list.Probes[0].Detail
	if len(detail.MissingAddresses) != 1 || detail.MissingAddresses[0] != "1.1.1.1" ||
		len(detail.UnexpectedAddresses) != 1 || detail.UnexpectedAddresses[0] != "8.8.8.8" {
		t.Fatalf("DNS differences = %+v", detail)
	}
	doer.body = `{"probes":[]}`
	cleared, err := domains.SetProbes(ctx, "d-1", nil)
	if err != nil || len(cleared.Probes) != 0 {
		t.Fatalf("cleared probes = %+v, %v", cleared, err)
	}
	path := siteTestBase + "/probes"
	assertSiteRequests(t, doer.requests, []string{"PUT " + path, "GET " + path, "PUT " + path})
	assertSiteBody(t, doer.requests[0], `{"probes":[{"kind":"dns","expectedAddresses":["1.1.1.1"]}]}`)
	assertSiteBody(t, doer.requests[2], `{"probes":[]}`)
}
