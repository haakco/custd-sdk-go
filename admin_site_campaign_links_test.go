package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func siteCampaignLinksFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-campaign-links.json", key)
}

func TestSiteCampaignLinkRoutesAndConflicts(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteCampaignLinksFixture(t, "campaignLinkListWithArchived"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	var req AdminSiteCampaignLinkRequest
	if err := json.Unmarshal([]byte(siteCampaignLinksFixture(t, "campaignLinkRequest")), &req); err != nil {
		t.Fatal(err)
	}

	list, err := domains.ListCampaignLinks(ctx, "d-1", true)
	if err != nil || len(list.Links) != 2 || list.MaxLinks != AdminSiteMaxCampaignLinks ||
		list.Links[0].ArchivedAt != nil || list.Links[1].ArchivedAt == nil || list.Links[1].Term != "" ||
		len(list.SuggestedSources) == 0 || len(list.SuggestedMediums) == 0 {
		t.Fatalf("list = %+v, %v", list, err)
	}
	doer.status, doer.body = http.StatusCreated, siteCampaignLinksFixture(t, "campaignLink")
	link, err := domains.CreateCampaignLink(ctx, "d-1", req)
	if err != nil || link.Term != "hosting_plans" || link.ArchivedAt != nil ||
		link.TaggedURL != "https://example.com/plans?plan=team&utm_campaign=spring_sale&utm_content=header_link"+
			"&utm_medium=email&utm_source=newsletter&utm_term=hosting_plans" {
		t.Fatalf("link = %+v, %v", link, err)
	}
	assertSiteBody(t, doer.requests[1], siteCampaignLinksFixture(t, "campaignLinkRequest"))
	doer.status, doer.body = http.StatusOK, siteCampaignLinksFixture(t, "campaignLinkArchived")
	archived, err := domains.ArchiveCampaignLink(ctx, "d-1", "0198a3b2-0000-7000-8000-000000000061")
	if err != nil || archived.ArchivedAt == nil {
		t.Fatalf("archived = %+v, %v", archived, err)
	}
	doer.body = siteCampaignLinksFixture(t, "campaignLinkListWithArchived")
	if _, err := domains.ListCampaignLinks(ctx, "d-1", false); err != nil {
		t.Fatal(err)
	}
	if _, err := domains.CreateCampaignLink(ctx, "d-1", AdminSiteCampaignLinkRequest{
		URL: "https://example.com/", Source: "google", Medium: "cpc", Campaign: "sale",
	}); err != nil {
		t.Fatal(err)
	}
	assertSiteBody(t, doer.requests[4], `{"url":"https://example.com/","source":"google","medium":"cpc","campaign":"sale"}`)
	base := siteTestBase + "/campaign-links"
	assertSiteRequests(t, doer.requests, []string{
		"GET " + base + "?archived=true", "POST " + base, "POST " + base + "/0198a3b2-0000-7000-8000-000000000061/archive",
		"GET " + base, "POST " + base,
	})
	for key, code := range map[string]string{
		"campaignLinkTakenProblem": SiteCampaignLinkTakenCode, "campaignLinkLimitProblem": SiteCampaignLinkLimitCode,
	} {
		doer.status, doer.body = http.StatusConflict, siteCampaignLinksFixture(t, key)
		_, err := domains.CreateCampaignLink(ctx, "d-1", req)
		assertSiteProblemCode(t, err, http.StatusConflict, code)
	}
}
