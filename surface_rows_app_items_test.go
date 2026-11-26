package custd

import "context"

var appItemSurfaceRows = []surfaceRow{
	{
		id:   "admin-api POST /api/v1/app/items",
		path: "/api/v1/app/items",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Create(context.Background(), AppItemAccess{}, AppItemCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/items",
		path: "/api/v1/app/items",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().ListMine(context.Background(), AppItemAccess{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/claim",
		path: "/api/v1/app/item/claim",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Claim(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/item",
		path: "/api/v1/app/item",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Get(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/app/item",
		path: "/api/v1/app/item",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Update(context.Background(), surfaceItemAccess, AppItemSettingsRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/item/plan/forecast",
		path: "/api/v1/app/item/plan/forecast",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetPlanForecast(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/app/item/plan",
		path: "/api/v1/app/item/plan",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().UpdatePlan(context.Background(), surfaceItemAccess, AppPlanUpdate{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/item/measurements",
		path: "/api/v1/app/item/measurements",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().ListMeasurements(context.Background(), surfaceItemAccess, AppItemPageOptions{After: "a2", Limit: 5, Run: "a3"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/measurements",
		path: "/api/v1/app/item/measurements",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AddMeasurements(context.Background(), surfaceItemAccess, MeasurementObservationBulkRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/measurements/csv",
		path: "/api/v1/app/item/measurements/csv",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().ImportMeasurementsCSV(context.Background(), surfaceItemAccess, MeasurementObservationCSVImportRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/estimates",
		path: "/api/v1/app/item/estimates",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().CreateEstimate(context.Background(), surfaceItemAccess, MeasurementEstimateCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/complete",
		path: "/api/v1/app/item/complete",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Complete(context.Background(), surfaceItemAccess, MeasurementProjectComplete{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/cancel",
		path: "/api/v1/app/item/cancel",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Cancel(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/reopen",
		path: "/api/v1/app/item/reopen",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().Reopen(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/item/runs",
		path: "/api/v1/app/item/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().ListRuns(context.Background(), surfaceItemAccess, AppItemPageOptions{After: "a2", Limit: 5, Run: "a3"})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/runs",
		path: "/api/v1/app/item/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().StartRun(context.Background(), surfaceItemAccess, MeasurementRunCreate{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/entries",
		path: "/api/v1/app/item/entries",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AddEntry(context.Background(), surfaceItemAccess, AppLinkCollectionEntryAddRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/app/item/entries/{entryUuid}",
		path: "/api/v1/app/item/entries/a1",
		call: func(e *surfaceEnv) error {
			return e.appItems().RemoveEntry(context.Background(), surfaceItemAccess, "a1")
		},
	},
	{
		id:   "admin-api DELETE /api/v1/app/item/links/{linkUuid}",
		path: "/api/v1/app/item/links/a1",
		call: func(e *surfaceEnv) error {
			return e.appItems().RevokeLink(context.Background(), surfaceItemAccess, "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/links/{linkUuid}/rotate",
		path: "/api/v1/app/item/links/a1/rotate",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().RotateLink(context.Background(), surfaceItemAccess, "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/model/versions",
		path: "/api/v1/app/item/model/versions",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AddModelVersion(context.Background(), surfaceItemAccess, AppModelVersionCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/model/copy",
		path: "/api/v1/app/item/model/copy",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().CopyModel(context.Background(), surfaceItemAccess, UncertaintyCopyModelRequest{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/item/alerts",
		path: "/api/v1/app/item/alerts",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetAlertSettings(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/app/item/alerts/deliveries",
		path: "/api/v1/app/item/alerts/deliveries",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().ListAlertDeliveries(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/alerts/rules",
		path: "/api/v1/app/item/alerts/rules",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AddAlertRule(context.Background(), surfaceItemAccess, AppItemAlertRuleCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/app/item/alerts/rules/{ruleUuid}",
		path: "/api/v1/app/item/alerts/rules/a1",
		call: func(e *surfaceEnv) error {
			return e.appItems().RemoveAlertRule(context.Background(), surfaceItemAccess, "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/alerts/webhooks",
		path: "/api/v1/app/item/alerts/webhooks",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AddAlertWebhook(context.Background(), surfaceItemAccess, AppItemAlertWebhookCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/app/item/alerts/webhooks/{webhookUuid}",
		path: "/api/v1/app/item/alerts/webhooks/a1",
		call: func(e *surfaceEnv) error {
			return e.appItems().RemoveAlertWebhook(context.Background(), surfaceItemAccess, "a1")
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/alerts/slack",
		path: "/api/v1/app/item/alerts/slack",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AddAlertSlack(context.Background(), surfaceItemAccess, AppItemAlertSlackCreateRequest{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/app/item/alerts/subscription",
		path: "/api/v1/app/item/alerts/subscription",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().SubscribeAlertEmail(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/app/item/alerts/subscription",
		path: "/api/v1/app/item/alerts/subscription",
		call: func(e *surfaceEnv) error {
			return e.appItems().UnsubscribeAlertEmail(context.Background(), surfaceItemAccess)
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/meeting/start",
		path: "/api/v1/app/item/meeting/start",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().StartMeeting(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/meeting/advance",
		path: "/api/v1/app/item/meeting/advance",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().AdvanceMeeting(context.Background(), surfaceItemAccess, AppMeetingAdvanceRequest{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/meeting/end",
		path: "/api/v1/app/item/meeting/end",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().EndMeeting(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "admin-api PATCH /api/v1/app/item/meeting/minutes",
		path: "/api/v1/app/item/meeting/minutes",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().UpdateMeetingMinutes(context.Background(), surfaceItemAccess, AppMeetingMinutesUpdate{})
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/app/forecast/runs",
		path: "/api/v1/app/forecast/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().RunForecast(context.Background(), surfaceItemAccess, MeasurementForecastRunRequest{})
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/app/forecast/runs/{runUuid}",
		path: "/api/v1/app/forecast/runs/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetForecastRun(context.Background(), surfaceItemAccess, "a1")
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/app/forecast/latest",
		path: "/api/v1/app/forecast/latest",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetLatestForecast(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/app/forecast/story",
		path: "/api/v1/app/forecast/story",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetForecastStory(context.Background(), surfaceItemAccess)
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/app/models/runs",
		path: "/api/v1/app/models/runs",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().RunModel(context.Background(), surfaceItemAccess, UncertaintyRunRequest{})
			return err
		},
	},
	{
		id:   "analytics-api POST /api/v1/app/models/calculate",
		path: "/api/v1/app/models/calculate",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().CalculateModel(context.Background(), surfaceItemAccess, AppModelCalculateRequest{})
			return err
		},
	},
	{
		id:   "analytics-api GET /api/v1/app/models/latest",
		path: "/api/v1/app/models/latest",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetLatestModelRun(context.Background(), surfaceItemAccess)
			return err
		},
	},

	{
		id:   "admin-api GET /api/v1/app/item/alerts/unsubscribe",
		path: "/api/v1/app/item/alerts/unsubscribe",
		call: func(e *surfaceEnv) error {
			_, err := e.appItems().GetAlertUnsubscribePage(context.Background(), "a1")
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/app/item/alerts/unsubscribe",
		path: "/api/v1/app/item/alerts/unsubscribe",
		call: func(e *surfaceEnv) error {
			return e.appItems().UnsubscribeAlertEmailByToken(context.Background(), AppAlertUnsubscribeRequest{})
		},
	},
}

// appItems is the hosted-app client the app rows call through.
func (e *surfaceEnv) appItems() *AppItemClient {
	client, err := NewAppItemClient(e.client, AppItemClientConfig{SiteKey: "site_pk_test", Origin: "https://app.example"})
	if err != nil {
		e.t.Fatalf("app item client: %v", err)
	}
	return client
}

// surfaceItemAccess names an item, so a row sends the item query the routes declare.
var surfaceItemAccess = AppItemAccess{Item: "i1"}
