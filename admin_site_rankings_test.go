package custd

import (
	"context"
	"net/http"
	"testing"
)

func siteRankingsFixture(t *testing.T, key string) string {
	t.Helper()
	return recordedFixture(t, "site-rankings.json", key)
}

func TestSiteRankingsDecodeStatesAndSources(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, siteRankingsFixture(t, "rankings"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()

	rankings, err := domains.GetRankings(ctx, "d-1", 3)
	if err != nil || rankings.Source != AdminSiteRankingsSourceActive || len(rankings.Keywords) != 2 {
		t.Fatalf("rankings = %+v, %v", rankings, err)
	}
	active, paused := rankings.Keywords[0], rankings.Keywords[1]
	notRanking, ranking := active.Observations[1], active.Observations[2]
	if notRanking.State != AdminSiteRankStateNotRanking || notRanking.Position != nil || notRanking.RankingURL != "" ||
		ranking.State != AdminSiteRankStateRanking || ranking.Position == nil || *ranking.Position != 7.4 ||
		active.Device != AdminSiteRankDeviceMobile || paused.Status != AdminSiteTrackedKeywordPaused ||
		paused.Country != AdminSiteRankAll || paused.Latest.State != AdminSiteRankStateNoData {
		t.Fatalf("keywords = %+v", rankings.Keywords)
	}
	if rankings.Pool != AdminSiteSearchPoolJoined || active.Pool == nil ||
		active.Pool.DemandBand != AdminSiteSearchPoolDemand10KTo100K || active.Pool.Positions.Top10 != 40 || paused.Pool != nil {
		t.Fatalf("pool = %q, %+v, %+v", rankings.Pool, active.Pool, paused.Pool)
	}
	doer.body = siteRankingsFixture(t, "rankingsNotConnected")
	none, err := domains.GetRankings(ctx, "d-1", 0)
	if err != nil || none.Source != AdminSiteRankSourceNotConnected || none.Pool != AdminSiteSearchPoolNotJoined ||
		len(none.Keywords) != 0 {
		t.Fatalf("not connected = %+v, %v", none, err)
	}
	doer.body = siteRankingsFixture(t, "rankingsFailing")
	failing, err := domains.GetRankings(ctx, "d-1", 0)
	if err != nil || failing.Source != AdminSiteRankingsSourceFailing || failing.SourceFailure == "" {
		t.Fatalf("failing = %+v, %v", failing, err)
	}
	assertSiteRequests(t, doer.requests[:2], []string{
		"GET " + siteTestBase + "/rankings?days=3", "GET " + siteTestBase + "/rankings",
	})
}

func TestSiteTrackedKeywordRoutesAndProblems(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, siteRankingsFixture(t, "trackedKeyword"))
	domains := newAdminTestClient(t, doer, "http://localhost:8080").Admin.SiteDomains
	ctx := context.Background()
	track := AdminSiteTrackKeywordRequest{Keyword: "Pricing Tool", Country: "zaf", Device: AdminSiteRankDeviceMobile}

	tracked, err := domains.TrackKeyword(ctx, "d-1", track)
	if err != nil || tracked.NotifyPosition != AdminSiteDefaultRankNotifyPosition || tracked.Status != AdminSiteTrackedKeywordActive {
		t.Fatalf("tracked = %+v, %v", tracked, err)
	}
	doer.status, doer.body = http.StatusOK, siteRankingsFixture(t, "updatedTrackedKeyword")
	updated, err := domains.UpdateTrackedKeyword(ctx, "d-1", "k-1", AdminSiteTrackedKeywordUpdate{NotifyPosition: 5})
	if err != nil || updated.NotifyPosition != 5 {
		t.Fatalf("updated = %+v, %v", updated, err)
	}
	if _, err := domains.UpdateTrackedKeyword(ctx, "d-1", "k-1", AdminSiteTrackedKeywordUpdate{Paused: true}); err != nil {
		t.Fatal(err)
	}
	doer.status, doer.body = http.StatusNoContent, ""
	if err := domains.UntrackKeyword(ctx, "d-1", "k-1"); err != nil {
		t.Fatal(err)
	}
	keywords := siteTestBase + "/rankings/keywords"
	assertSiteRequests(t, doer.requests, []string{
		"POST " + keywords, "PUT " + keywords + "/k-1", "PUT " + keywords + "/k-1", "DELETE " + keywords + "/k-1",
	})
	assertSiteBody(t, doer.requests[0], siteRankingsFixture(t, "trackKeywordRequest"))
	assertSiteBody(t, doer.requests[1], siteRankingsFixture(t, "updateTrackedKeywordRequest"))
	assertSiteBody(t, doer.requests[2], siteRankingsFixture(t, "updateTrackedKeywordPauseRequest"))
	for _, tc := range []struct {
		key, code string
		status    int
		call      func() error
	}{
		{"trackedKeywordInvalidProblem", SiteTrackedKeywordInvalidCode, http.StatusBadRequest, func() error {
			_, err := domains.TrackKeyword(ctx, "d-1", AdminSiteTrackKeywordRequest{})
			return err
		}},
		{"trackedKeywordLimitProblem", SiteTrackedKeywordLimitCode, http.StatusConflict, func() error {
			_, err := domains.TrackKeyword(ctx, "d-1", track)
			return err
		}},
		{"trackedKeywordDuplicateProblem", SiteTrackedKeywordDuplicateCode, http.StatusConflict, func() error {
			_, err := domains.TrackKeyword(ctx, "d-1", track)
			return err
		}},
		{"trackedKeywordNotFoundProblem", SiteTrackedKeywordNotFoundCode, http.StatusNotFound, func() error {
			return domains.UntrackKeyword(ctx, "d-1", "gone")
		}},
		{"rankingsDaysInvalidProblem", SiteRankingsDaysInvalidCode, http.StatusBadRequest, func() error {
			_, err := domains.GetRankings(ctx, "d-1", 91)
			return err
		}},
	} {
		doer.status, doer.body = tc.status, siteRankingsFixture(t, tc.key)
		assertSiteProblemCode(t, tc.call(), tc.status, tc.code)
	}
}
