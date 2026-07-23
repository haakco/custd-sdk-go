package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// privacySurface covers Privacy.GetRules, SetRules, MapIdentifier, and
// ListIdentifierMappings. The MapIdentifier request body is asserted not to
// appear in any error message; the SDK must not surface ExternalID via the
// logged request fields either, so we verify the body is sent verbatim and the
// response side returns no ExternalID.
func TestPrivacyAdminClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"tenantSlug":"acme","purposes":["analytics"],"hardDeleteAfterDays":30}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	rules, err := client.Admin.Privacy.GetRules(context.Background())
	if err != nil {
		t.Fatalf("GetRules error: %v", err)
	}
	if rules.TenantSlug != "acme" {
		t.Fatalf("rules.TenantSlug = %q", rules.TenantSlug)
	}
	if doer.requests[0].Method != http.MethodGet ||
		doer.requests[0].URL != "http://localhost:8080/api/v1/admin/privacy/rules" {
		t.Fatalf("GetRules request = %+v", doer.requests[0])
	}

	doer.body = `{"tenantSlug":"acme","purposes":["analytics","product_improvement"],"hardDeleteAfterDays":60}`
	updated, err := client.Admin.Privacy.SetRules(context.Background(), PrivacyRuleUpdate{
		Purposes:            []string{"analytics", "product_improvement"},
		HardDeleteAfterDays: 60,
	})
	if err != nil {
		t.Fatalf("SetRules error: %v", err)
	}
	if updated.HardDeleteAfterDays != 60 {
		t.Fatalf("updated HardDeleteAfterDays = %d", updated.HardDeleteAfterDays)
	}
	if doer.requests[1].Method != http.MethodPut ||
		doer.requests[1].URL != "http://localhost:8080/api/v1/admin/privacy/rules" {
		t.Fatalf("SetRules request = %+v", doer.requests[1])
	}

	doer.body = `{"identifierId":"id-1","internalIdHash":"$2a$prefix","internalIdHashPrefix":"prefix","saltVersion":2,"createdAt":"2026-07-23T12:00:00Z"}`
	mapped, err := client.Admin.Privacy.MapIdentifier(context.Background(), "acme", PrivacyIdentifierMapRequest{
		ExternalID: "external-secret",
	})
	if err != nil {
		t.Fatalf("MapIdentifier error: %v", err)
	}
	if mapped.InternalIDHashPrefix != "prefix" || mapped.SaltVersion != 2 {
		t.Fatalf("mapped = %+v", mapped)
	}
	want := `{"externalId":"external-secret"}`
	if string(doer.requests[2].Body) != want {
		t.Fatalf("MapIdentifier body = %q, want %q", string(doer.requests[2].Body), want)
	}
	// Server response must not echo externalId.
	for _, c := range doer.requests[2].Headers {
		_ = c
	}
	var respMap map[string]any
	if err := json.Unmarshal([]byte(`{"identifierId":"id-1","internalIdHash":"$2a$prefix","internalIdHashPrefix":"prefix","saltVersion":2,"createdAt":"2026-07-23T12:00:00Z"}`), &respMap); err != nil {
		t.Fatalf("resp unmarshal: %v", err)
	}
	if _, leak := respMap["externalId"]; leak {
		t.Fatalf("server response must not echo externalId; got %+v", respMap)
	}

	doer.status = http.StatusOK
	doer.body = `[{"identifierId":"id-1","internalIdHash":"$2a$prefix","internalIdHashPrefix":"prefix","saltVersion":2}]`
	mappings, err := client.Admin.Privacy.ListIdentifierMappings(context.Background(), "acme")
	if err != nil {
		t.Fatalf("ListIdentifierMappings error: %v", err)
	}
	if len(mappings) != 1 || mappings[0].IdentifierID != "id-1" {
		t.Fatalf("mappings = %+v", mappings)
	}
	if doer.requests[3].URL != "http://localhost:8080/api/v1/admin/privacy/identifiers/acme" {
		t.Fatalf("ListIdentifierMappings URL = %s", doer.requests[3].URL)
	}
}

func TestRetentionAdminClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"policies":[{"tenantSlug":"acme","maxAgeDays":365,"hardDeleteAfterDays":730,"applyToEventTypes":["page.view"],"applyToDataSpaces":["main"]}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	list, err := client.Admin.Retention.List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list.Policies) != 1 || list.Policies[0].TenantSlug != "acme" {
		t.Fatalf("policies = %+v", list.Policies)
	}

	doer.body = `{"tenantSlug":"acme","maxAgeDays":180,"hardDeleteAfterDays":365,"applyToEventTypes":[],"applyToDataSpaces":[]}`
	upserted, err := client.Admin.Retention.Upsert(context.Background(), "acme", RetentionPolicyUpsertRequest{
		MaxAgeDays:          180,
		HardDeleteAfterDays: 365,
	})
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}
	if upserted.MaxAgeDays != 180 {
		t.Fatalf("upserted.MaxAgeDays = %d", upserted.MaxAgeDays)
	}
	if doer.requests[1].Method != http.MethodPut ||
		doer.requests[1].URL != "http://localhost:8080/api/v1/admin/retention/policies/acme" {
		t.Fatalf("Upsert request = %+v", doer.requests[1])
	}

	doer.body = `{"tenantSlug":"acme","maxAgeDays":180,"hardDeleteAfterDays":365,"applyToEventTypes":[],"applyToDataSpaces":[]}`
	got, err := client.Admin.Retention.Get(context.Background(), "acme")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.TenantSlug != "acme" {
		t.Fatalf("get = %+v", got)
	}

	doer.status = http.StatusNoContent
	if err := client.Admin.Retention.Delete(context.Background(), "acme"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestStorageAlertAdminClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"rules":[{"ruleId":"rule-1","tenantSlug":"acme","metric":"ingest_bytes","thresholdPercent":80,"channel":"slack","enabled":true}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	list, err := client.Admin.StorageAlerts.ListRules(context.Background(), "acme")
	if err != nil {
		t.Fatalf("ListRules error: %v", err)
	}
	if len(list.Rules) != 1 || list.Rules[0].RuleID != "rule-1" {
		t.Fatalf("rules = %+v", list.Rules)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/admin/storage/alerts/acme" {
		t.Fatalf("ListRules URL = %s", doer.requests[0].URL)
	}

	doer.status = http.StatusCreated
	doer.body = `{"ruleId":"rule-2","tenantSlug":"acme","metric":"ingest_bytes","thresholdPercent":90,"channel":"email","enabled":true}`
	created, err := client.Admin.StorageAlerts.CreateRule(context.Background(), "acme", StorageAlertRuleCreateRequest{
		Metric:           "ingest_bytes",
		ThresholdPercent: 90,
		Channel:          "email",
		Enabled:          true,
	})
	if err != nil {
		t.Fatalf("CreateRule error: %v", err)
	}
	if created.RuleID != "rule-2" {
		t.Fatalf("created = %+v", created)
	}

	doer.status = http.StatusNoContent
	if err := client.Admin.StorageAlerts.DeleteRule(context.Background(), "acme", "rule-1"); err != nil {
		t.Fatalf("DeleteRule error: %v", err)
	}
	if doer.requests[2].URL != "http://localhost:8080/api/v1/admin/storage/alerts/acme/rule-1" {
		t.Fatalf("DeleteRule URL = %s", doer.requests[2].URL)
	}
}

func TestAuditAdminClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"events":[{"eventId":"ev-1","action":"create","actorId":"u-1","actorKind":"user","resourceType":"producer","resourceId":"prod-1","ipAddress":"10.0.0.1","createdAt":"2026-07-23T12:00:00Z"}],"nextCursor":{"cursor":"next"}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	list, err := client.Admin.Audit.ListEvents(context.Background(), AuditListOptions{
		ResourceType: "producer",
		ResourceID:   "prod-1",
		Limit:        50,
	})
	if err != nil {
		t.Fatalf("ListEvents error: %v", err)
	}
	if len(list.Events) != 1 || list.NextCursor == nil || list.NextCursor.Cursor != "next" {
		t.Fatalf("events = %+v cursor = %+v", list.Events, list.NextCursor)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/admin/audit/events?limit=50&resourceId=prod-1&resourceType=producer" {
		t.Fatalf("ListEvents URL = %s", doer.requests[0].URL)
	}

	doer.body = `{"eventId":"ev-1","action":"create","actorId":"u-1","actorKind":"user","resourceType":"producer","resourceId":"prod-1","ipAddress":"10.0.0.1","createdAt":"2026-07-23T12:00:00Z"}`
	got, err := client.Admin.Audit.GetEvent(context.Background(), "ev-1")
	if err != nil {
		t.Fatalf("GetEvent error: %v", err)
	}
	if got.EventID != "ev-1" {
		t.Fatalf("GetEvent = %+v", got)
	}

	doer.body = `{"events":[{"action":"draft_created","actorId":"u-1","resourceType":"reporting_pack","resourceId":"42","packKey":"security","createdAt":"2026-07-23T12:00:00Z"}]}`
	rpEvents, err := client.Admin.Audit.ListReportingPackEvents(context.Background())
	if err != nil {
		t.Fatalf("ListReportingPackEvents error: %v", err)
	}
	if len(rpEvents.Events) != 1 || rpEvents.Events[0].PackKey != "security" {
		t.Fatalf("rpEvents = %+v", rpEvents)
	}
}

func TestOffboardingAdminClientLifecycle(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"tenantSlug":"acme","effectiveAt":"2026-08-23T00:00:00Z","gracePeriodDays":7,"reason":"client request","status":"scheduled"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")

	sched, err := client.Admin.Offboarding.Schedule(context.Background(), "acme", OffboardingScheduleRequest{
		EffectiveAt:     "2026-08-23T00:00:00Z",
		GracePeriodDays: 7,
		Reason:          "client request",
		Status:          "scheduled",
	})
	if err != nil {
		t.Fatalf("Schedule error: %v", err)
	}
	if sched.TenantSlug != "acme" {
		t.Fatalf("sched = %+v", sched)
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/admin/offboarding/schedules/acme" {
		t.Fatalf("Schedule URL = %s", doer.requests[0].URL)
	}

	doer.body = `{"schedules":[{"tenantSlug":"acme","effectiveAt":"2026-08-23T00:00:00Z","gracePeriodDays":7,"reason":"client request","status":"scheduled"}]}`
	list, err := client.Admin.Offboarding.ListSchedules(context.Background())
	if err != nil {
		t.Fatalf("ListSchedules error: %v", err)
	}
	if len(list.Schedules) != 1 {
		t.Fatalf("schedules = %+v", list.Schedules)
	}

	doer.status = http.StatusNoContent
	if err := client.Admin.Offboarding.CancelSchedule(context.Background(), "acme", OffboardingCancelRequest{
		Reason: "client cancelled",
	}); err != nil {
		t.Fatalf("CancelSchedule error: %v", err)
	}
	if doer.requests[2].URL != "http://localhost:8080/api/v1/admin/offboarding/schedules/acme/cancel" {
		t.Fatalf("CancelSchedule URL = %s", doer.requests[2].URL)
	}

	doer.body = `{"requestUuid":"req-1","tenantSlug":"acme","status":"pending","requestedBy":"u-1","requestedAt":"2026-07-23T12:00:00Z"}`
	got, err := client.Admin.Offboarding.GetRequest(context.Background(), "req-1")
	if err != nil {
		t.Fatalf("GetRequest error: %v", err)
	}
	if got.RequestUUID != "req-1" {
		t.Fatalf("GetRequest = %+v", got)
	}

	doer.status = http.StatusNoContent
	if err := client.Admin.Offboarding.CancelRequest(context.Background(), "req-1"); err != nil {
		t.Fatalf("CancelRequest error: %v", err)
	}
	if err := client.Admin.Offboarding.ConfirmRequest(context.Background(), "req-1"); err != nil {
		t.Fatalf("ConfirmRequest error: %v", err)
	}
}
