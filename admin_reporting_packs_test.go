package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestReportingPacksAdminClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"drafts":[{"id":42,"revision":1,"definition":{"key":"security","displayName":"Security","enabled":true,"eventTypes":["login.success"]}}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	list, err := client.Admin.ReportingPacks.ListDrafts(context.Background())
	if err != nil {
		t.Fatalf("ListDrafts error: %v", err)
	}
	if len(list.Drafts) != 1 || list.Drafts[0].ID != 42 {
		t.Fatalf("drafts = %+v", list.Drafts)
	}

	doer.body = `{"id":42,"revision":1,"definition":{"key":"security","displayName":"Security","enabled":true,"eventTypes":["login.success"]}}`
	got, err := client.Admin.ReportingPacks.GetDraft(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetDraft error: %v", err)
	}
	if got.ID != 42 {
		t.Fatalf("got = %+v", got)
	}

	doer.status = http.StatusCreated
	doer.body = `{"id":43,"revision":1,"definition":{"key":"security","displayName":"Security","enabled":true,"eventTypes":["login.success"]}}`
	created, err := client.Admin.ReportingPacks.CreateDraft(context.Background(), ReportingPackDraftCreate{
		Definition: Pack{Key: "security", DisplayName: "Security", Enabled: true, EventTypes: []string{"login.success"}},
	})
	if err != nil {
		t.Fatalf("CreateDraft error: %v", err)
	}
	if created.ID != 43 {
		t.Fatalf("created = %+v", created)
	}

	doer.status = http.StatusOK
	doer.body = `{"id":43,"revision":2,"definition":{"key":"security","displayName":"Security v2","enabled":true,"eventTypes":["login.success"]}}`
	updated, err := client.Admin.ReportingPacks.UpdateDraft(context.Background(), "43", ReportingPackDraftUpdate{
		Definition:       Pack{Key: "security", DisplayName: "Security v2", Enabled: true, EventTypes: []string{"login.success"}},
		ExpectedRevision: 1,
	})
	if err != nil {
		t.Fatalf("UpdateDraft error: %v", err)
	}
	if updated.Revision != 2 {
		t.Fatalf("updated = %+v", updated)
	}

	doer.body = `{"valid":true}`
	val, err := client.Admin.ReportingPacks.Validate(context.Background(), ReportingPackDraftCreate{
		Definition: Pack{Key: "security", DisplayName: "Security"},
	})
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if !val.Valid {
		t.Fatalf("Validate = %+v", val)
	}

	doer.body = `{"buckets":[{"date":"2026-07-23","value":{"value":1.0,"unit":"count","sampleCount":1,"dataSufficiency":"complete","complete":true},"source":"preview","queryDurationMs":10}],"value":{"value":1.0,"unit":"count","sampleCount":1,"dataSufficiency":"complete","complete":true}}`
	prev, err := client.Admin.ReportingPacks.Preview(context.Background(), ReportingPackPreviewRequest{
		Definition: Pack{Key: "security", DisplayName: "Security"},
		TenantSlug: "acme",
		Query:      QueryHint{Template: "count", Metrics: []string{"events"}},
	})
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}
	if len(prev.Buckets) != 1 {
		t.Fatalf("prev = %+v", prev)
	}

	doer.body = `{"id":99,"generationNumber":7,"sourceDraftId":43,"definition":{"key":"security","displayName":"Security"},"state":"publishing","createdAt":"2026-07-23T12:00:00Z"}`
	gen, err := client.Admin.ReportingPacks.Publish(context.Background(), "43")
	if err != nil {
		t.Fatalf("Publish error: %v", err)
	}
	if gen.ID != 99 || gen.State != "publishing" {
		t.Fatalf("gen = %+v", gen)
	}

	doer.status = http.StatusAccepted
	doer.body = `{"id":100,"generationNumber":8,"sourceDraftId":43,"definition":{"key":"security","displayName":"Security"},"state":"restarting"}`
	restart, err := client.Admin.ReportingPacks.Restart(context.Background(), "43")
	if err != nil {
		t.Fatalf("Restart error: %v", err)
	}
	if restart.GenerationNumber != 8 {
		t.Fatalf("restart = %+v", restart)
	}

	doer.status = http.StatusOK
	doer.body = `{"id":99,"generationNumber":7,"sourceDraftId":43,"definition":{"key":"security","displayName":"Security"},"state":"published"}`
	getGen, err := client.Admin.ReportingPacks.GetGeneration(context.Background(), "99")
	if err != nil {
		t.Fatalf("GetGeneration error: %v", err)
	}
	if getGen.State != "published" {
		t.Fatalf("getGen = %+v", getGen)
	}

	doer.body = `{"generation":{"id":99,"generationNumber":7,"packKey":"security","state":"published"},"acknowledgements":[{"accepted":true,"consumer":"tracklab","observedAt":"2026-07-23T12:00:00Z","observedGenerationId":99}]}`
	status, err := client.Admin.ReportingPacks.GetGenerationStatus(context.Background(), "99")
	if err != nil {
		t.Fatalf("GetGenerationStatus error: %v", err)
	}
	if status.Generation.PackKey != "security" || len(status.Acknowledgements) != 1 {
		t.Fatalf("status = %+v", status)
	}

	doer.status = http.StatusNoContent
	if err := client.Admin.ReportingPacks.RollbackGeneration(context.Background(), "99"); err != nil {
		t.Fatalf("RollbackGeneration error: %v", err)
	}

	doer.status = http.StatusOK
	doer.body = `{"generationId":99,"definitionFingerprint":"abc","tenantSlug":"acme","materializations":[{"status":"ready","definitionFingerprint":"abc","sourceCoverageCount":12}]}`
	prov, err := client.Admin.ReportingPacks.GetRollupProvenance(context.Background(), "99")
	if err != nil {
		t.Fatalf("GetRollupProvenance error: %v", err)
	}
	if prov.DefinitionFingerprint != "abc" || len(prov.Materializations) != 1 {
		t.Fatalf("prov = %+v", prov)
	}
}

func TestProducerReservationsClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"producerSlug":"webhook","parentCompanySlug":"acme","status":"reserved","reservedAt":"2026-07-23T12:00:00Z","expiresAt":"2026-07-23T12:05:00Z","maxTtlSeconds":300}`)
	client := newProvisioningTestClient(t, doer)

	reserved, err := client.Provisioning.Reservations.Reserve(context.Background(), "acme", ProducerReservationCreateRequest{
		ProducerSlug: "webhook",
		TTLSeconds:   300,
	})
	if err != nil {
		t.Fatalf("Reserve error: %v", err)
	}
	if reserved.ProducerSlug != "webhook" || reserved.Status != "reserved" {
		t.Fatalf("reserved = %+v", reserved)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/data-spaces/acme/producer-reservations" {
		t.Fatalf("Reserve URL = %s", doer.requests[0].URL)
	}

	doer.status = http.StatusOK
	doer.body = `{"reservations":[{"producerSlug":"webhook","parentCompanySlug":"acme","status":"reserved"}]}`
	list, err := client.Provisioning.Reservations.List(context.Background(), "acme")
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list.Reservations) != 1 {
		t.Fatalf("list = %+v", list.Reservations)
	}

	doer.body = `{"producerSlug":"webhook","parentCompanySlug":"acme","status":"claimed","claimedByClientId":"custd-acme-webhook"}`
	claimed, err := client.Provisioning.Reservations.Claim(context.Background(), "acme", "webhook", ProducerReservationClaimRequest{
		ClaimedByClientID: "custd-acme-webhook",
	})
	if err != nil {
		t.Fatalf("Claim error: %v", err)
	}
	if claimed.Status != "claimed" || claimed.ClaimedByClientID != "custd-acme-webhook" {
		t.Fatalf("claimed = %+v", claimed)
	}
	if doer.requests[2].URL != "http://localhost:8080/api/v1/data-spaces/acme/producer-reservations/webhook/claim" {
		t.Fatalf("Claim URL = %s", doer.requests[2].URL)
	}

	doer.status = http.StatusNoContent
	if err := client.Provisioning.Reservations.Release(context.Background(), "acme", "webhook"); err != nil {
		t.Fatalf("Release error: %v", err)
	}
	if doer.requests[3].URL != "http://localhost:8080/api/v1/data-spaces/acme/producer-reservations/webhook" {
		t.Fatalf("Release URL = %s", doer.requests[3].URL)
	}
}
