package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Hosted-app data access: an app creates an item with its public Site key,
// then reads and changes it with the item's owner link, or only reads it with
// its share link. A link token travels only in the Authorization header, never
// in a URL, a log line or an error message.

const (
	appItemRoot = "/api/v1/app"
	// AppItemSiteKeyHeader carries the hosted app's public Site key.
	AppItemSiteKeyHeader = "X-Custd-Site-Key"
	// AppItemUserTokenHeader carries the signed-in user's access token beside the link token.
	AppItemUserTokenHeader = "X-Custd-User-Token"
)

// AppItemClientConfig names the hosted app a client acts for. Origin must be one
// of the app's allowed origins; a browser sends it itself, a server must name it.
type AppItemClientConfig struct {
	SiteKey string
	Origin  string
}

// AppItemAccess is how a call reaches its item: a link token, the signed-in
// user's access token, or both. Item names a user-owned item reached with the
// user token and no link, or an entry of a collection the link opens.
type AppItemAccess struct {
	LinkToken string
	UserToken string
	Item      string
}

// AppItemClient calls the hosted-app item routes. It never sends the client's
// own API credential: the Site key, link and user tokens are the only authority.
type AppItemClient struct {
	client *CustdClient
	config AppItemClientConfig
}

// NewAppItemClient returns a hosted-app client that sends to the client's BaseURL.
func NewAppItemClient(client *CustdClient, config AppItemClientConfig) (*AppItemClient, error) {
	if strings.TrimSpace(config.SiteKey) == "" {
		return nil, fmt.Errorf("custd: app item client requires a site key")
	}
	if strings.TrimSpace(config.Origin) == "" {
		return nil, fmt.Errorf("custd: app item client requires an allowed origin")
	}
	return &AppItemClient{client: client, config: config}, nil
}

// AppAlertUnsubscribeRequest stops the email subscription an unsubscribe token names.
type AppAlertUnsubscribeRequest struct {
	Token string `json:"token"`
}

// AppAlertUnsubscribePage is the HTML page an alert email's unsubscribe link opens.
type AppAlertUnsubscribePage struct {
	HTML string
}

// GetAlertUnsubscribePage returns the page an alert email's unsubscribe link
// opens; opening it changes nothing. It needs no Site key or link.
func (c *AppItemClient) GetAlertUnsubscribePage(ctx context.Context, token string) (*AppAlertUnsubscribePage, error) {
	response, err := c.exchange(ctx, AppItemAccess{}, http.MethodGet,
		siteQueryPath("/item/alerts/unsubscribe", url.Values{"token": {token}}), nil)
	if err != nil {
		return nil, err
	}
	return &AppAlertUnsubscribePage{HTML: string(response.Body)}, nil
}

// UnsubscribeAlertEmailByToken stops the email subscription an unsubscribe token
// names. It is idempotent and needs no Site key or link.
func (c *AppItemClient) UnsubscribeAlertEmailByToken(ctx context.Context, req AppAlertUnsubscribeRequest) error {
	return c.send(ctx, AppItemAccess{}, http.MethodPost, "/item/alerts/unsubscribe", req, nil)
}

// send issues one call to a route under /api/v1/app and decodes a JSON answer into out.
func (c *AppItemClient) send(ctx context.Context, access AppItemAccess, method, path string, payload, out any) error {
	response, err := c.exchange(ctx, access, method, path, payload)
	if err != nil {
		return err
	}
	if out == nil || len(response.Body) == 0 {
		return nil
	}
	if err := unmarshalResponse(response.Body, out); err != nil {
		return validationWrap("custd: decode app item response", err)
	}
	return nil
}

func (c *AppItemClient) exchange(ctx context.Context, access AppItemAccess, method, path string, payload any) (*HTTPResponse, error) {
	var body []byte
	headers := map[string]string{
		"X-Custd-Sdk": sdkIdentity, "Origin": c.config.Origin, AppItemSiteKeyHeader: c.config.SiteKey,
	}
	if access.LinkToken != "" {
		headers["Authorization"] = "Bearer " + access.LinkToken
	}
	if access.UserToken != "" {
		headers[AppItemUserTokenHeader] = access.UserToken
	}
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("custd: encode app item request: %w", err)
		}
		body = encoded
		headers["Content-Type"] = "application/json"
	}
	response, err := c.client.requestApplication(ctx, &HTTPRequest{
		Method: method, URL: c.endpoint(path, access.Item), Headers: headers, Body: body,
	})
	if err != nil {
		return nil, err
	}
	if err := c.client.checkStatus(response.StatusCode, response.Body); err != nil {
		return nil, err
	}
	return response, nil
}

// endpoint resolves a path under /api/v1/app, adding the item a call names.
func (c *AppItemClient) endpoint(path, item string) string {
	endpoint := strings.TrimRight(c.client.config.BaseURL, "/") + appItemRoot + path
	if item == "" {
		return endpoint
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return endpoint + separator + url.Values{"item": {item}}.Encode()
}

type AppMeasurementProjectCreate struct {
	Description string                    `json:"description,omitempty"`
	Kind        string                    `json:"kind"`
	Name        string                    `json:"name"`
	Series      []MeasurementSeriesCreate `json:"series"`
	Target      MeasurementTargetCreate   `json:"target"`
}

type AppMeetingBlockCreate struct {
	PredictedMs int64  `json:"predictedMs"`
	Title       string `json:"title"`
}

type AppMeetingCreate struct {
	Blocks []AppMeetingBlockCreate `json:"blocks"`
	Name   string                  `json:"name"`
}

type AppItemCreateRequest struct {
	IdleExpirySeconds  *int64                       `json:"idleExpirySeconds,omitempty"`
	Kind               string                       `json:"kind"`
	MeasurementProject *AppMeasurementProjectCreate `json:"measurementProject,omitempty"`
	Meeting            *AppMeetingCreate            `json:"meeting,omitempty"`
	// Display name of a link collection or plan.
	Name             string                         `json:"name,omitempty"`
	UncertaintyModel *UncertaintyCreateModelRequest `json:"uncertaintyModel,omitempty"`
}

type AppItemCounterProgress struct {
	Current     float64    `json:"current,omitempty"`
	ForecastP50 *time.Time `json:"forecastP50,omitempty"`
	Percent     int64      `json:"percent,omitempty"`
}

type AppItem struct {
	// When the item is deleted unless it changes first.
	ExpiresAt         time.Time `json:"expiresAt"`
	IdleExpirySeconds int64     `json:"idleExpirySeconds"`
	ItemUUID          string    `json:"itemUuid"`
	Kind              string    `json:"kind"`
	// Whether the signed-in caller owns it.
	Ownership string `json:"ownership"`
	// What the presented link, or the signed-in owner, may do.
	Role      string    `json:"role"`
	UpdatedAt time.Time `json:"updatedAt"`
	// Moves on with every change; the read's ETag carries it.
	Version int64 `json:"version"`
}

type AppLinkCollectionEntry struct {
	CreatedAt time.Time `json:"createdAt"`
	EntryUUID string    `json:"entryUuid"`
	ItemUUID  string    `json:"itemUuid"`
	Kind      string    `json:"kind"`
	Position  int64     `json:"position"`
	// Available when the stored entry has a recoverable share link.
	ShareToken string `json:"shareToken,omitempty"`
}

type AppLinkCollection struct {
	Entries []AppLinkCollectionEntry `json:"entries"`
	Name    string                   `json:"name,omitempty"`
}

type AppItemLink struct {
	CreatedAt time.Time `json:"createdAt"`
	LinkUUID  string    `json:"linkUuid"`
	Role      string    `json:"role"`
	// Available for share links created or rotated after recovery was enabled.
	ShareToken string `json:"shareToken,omitempty"`
}

type AppItemRun struct {
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	RunUUID   string    `json:"runUuid"`
	Status    string    `json:"status"`
}

type AppItemRunPage struct {
	NextCursor string       `json:"nextCursor,omitempty"`
	Runs       []AppItemRun `json:"runs"`
}

type AppMeetingBlock struct {
	// The time it took, once it is done.
	ActualMs int64 `json:"actualMs"`
	// The time it has now: an overrun earlier block shrinks the blocks after it.
	AllocatedMs int64      `json:"allocatedMs"`
	BlockUUID   string     `json:"blockUuid"`
	EndedAt     *time.Time `json:"endedAt,omitempty"`
	Position    int64      `json:"position"`
	// The time the agenda predicted.
	PredictedMs int64      `json:"predictedMs"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	Status      string     `json:"status"`
	Title       string     `json:"title"`
}

type AppMeetingSummary struct {
	// The time the blocks that have finished took.
	FinishedActualMs int64 `json:"finishedActualMs"`
	FinishedBlocks   int64 `json:"finishedBlocks"`
	// The predicted time of the blocks that have finished.
	FinishedPredictedMs int64 `json:"finishedPredictedMs"`
	// How far the meeting ran past its planned end.
	OverrunMs int64 `json:"overrunMs"`
	// The agenda's predicted length.
	PredictedMs int64 `json:"predictedMs"`
}

type AppMeeting struct {
	Blocks []AppMeetingBlock `json:"blocks"`
	// The block that is on now, if any.
	CurrentBlockUUID string     `json:"currentBlockUuid,omitempty"`
	EndsAt           *time.Time `json:"endsAt,omitempty"`
	Name             string     `json:"name"`
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	// ready until started; completed once ended.
	Status  string            `json:"status"`
	Summary AppMeetingSummary `json:"summary"`
}

type AppMeetingAction struct {
	Description string `json:"description"`
	PartUUID    string `json:"partUuid,omitempty"`
}

type AppMeetingMinutes struct {
	Actions       []AppMeetingAction `json:"actions"`
	Decisions     string             `json:"decisions"`
	Minutes       string             `json:"minutes"`
	PlanUUID      string             `json:"planUuid,omitempty"`
	SharedSummary string             `json:"sharedSummary,omitempty"`
}

type AppModelPreview struct {
	P5       float64 `json:"p5"`
	P50      float64 `json:"p50"`
	P95      float64 `json:"p95"`
	Quantity string  `json:"quantity"`
}

type AppPlanPart struct {
	ItemUUID      string   `json:"itemUuid"`
	Prerequisites []string `json:"prerequisites"`
	Required      bool     `json:"required"`
}

type AppPlan struct {
	Deadline *time.Time    `json:"deadline,omitempty"`
	Name     string        `json:"name"`
	Parts    []AppPlanPart `json:"parts"`
	Revision int64         `json:"revision"`
}

type AppPlanMeetingSummary struct {
	MeetingUUID string `json:"meetingUuid"`
	Name        string `json:"name"`
	Summary     string `json:"summary"`
}

type AppItemRead struct {
	CounterProgress      *AppItemCounterProgress  `json:"counterProgress,omitempty"`
	Estimates            *MeasurementEstimateList `json:"estimates,omitempty"`
	Item                 AppItem                  `json:"item"`
	LinkCollection       *AppLinkCollection       `json:"linkCollection,omitempty"`
	Links                []AppItemLink            `json:"links,omitempty"`
	MeasurementProject   *MeasurementProject      `json:"measurementProject,omitempty"`
	MeasurementRuns      *AppItemRunPage          `json:"measurementRuns,omitempty"`
	Meeting              *AppMeeting              `json:"meeting,omitempty"`
	MeetingMinutes       *AppMeetingMinutes       `json:"meetingMinutes,omitempty"`
	ModelPreview         *AppModelPreview         `json:"modelPreview,omitempty"`
	Plan                 *AppPlan                 `json:"plan,omitempty"`
	PlanMeetingSummaries []AppPlanMeetingSummary  `json:"planMeetingSummaries,omitempty"`
	UncertaintyModel     *UncertaintyModelDetail  `json:"uncertaintyModel,omitempty"`
}

type AppItemCreated struct {
	Item AppItemRead `json:"item"`
	// Reads and changes this item. Shown once.
	OwnerToken string `json:"ownerToken"`
	// Only reads this item. Recoverable through an owner read.
	ShareToken string `json:"shareToken"`
}

type AppItemSummary struct {
	// When the item is deleted unless it changes first.
	ExpiresAt         time.Time `json:"expiresAt"`
	IdleExpirySeconds int64     `json:"idleExpirySeconds"`
	ItemUUID          string    `json:"itemUuid"`
	Kind              string    `json:"kind"`
	// The item's display name, when set.
	Name string `json:"name"`
	// Whether the signed-in caller owns it.
	Ownership string `json:"ownership"`
	// What the presented link, or the signed-in owner, may do.
	Role string `json:"role"`
	// Code of the first series in a measurement project.
	SeriesCode string `json:"seriesCode,omitempty"`
	// The measurement project's or meeting run's status; empty for a link collection or a model.
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
	// Moves on with every change; the read's ETag carries it.
	Version int64 `json:"version"`
}

type AppItemList struct {
	Items []AppItemSummary `json:"items"`
}

type AppItemSettingsRequest struct {
	IdleExpirySeconds *int64 `json:"idleExpirySeconds,omitempty"`
	// Collection name. Empty clears it.
	Name    string                    `json:"name,omitempty"`
	Project *MeasurementProjectUpdate `json:"project,omitempty"`
}

type AppPlanProgressPoint struct {
	At        time.Time `json:"at"`
	Completed int64     `json:"completed"`
}

type AppPlanProgressBand struct {
	At     time.Time `json:"at"`
	Lower  int64     `json:"lower"`
	Median int64     `json:"median"`
	Upper  int64     `json:"upper"`
}

type AppPlanForecast struct {
	AsOf                time.Time              `json:"asOf"`
	CompletedParts      int64                  `json:"completedParts"`
	DeadlineProbability float64                `json:"deadlineProbability,omitempty"`
	Horizon             time.Time              `json:"horizon"`
	HorizonMissFraction float64                `json:"horizonMissFraction,omitempty"`
	Lower               *time.Time             `json:"lower,omitempty"`
	Median              *time.Time             `json:"median,omitempty"`
	MissingParts        []string               `json:"missingParts"`
	Model               string                 `json:"model"`
	ObservedProgress    []AppPlanProgressPoint `json:"observedProgress"`
	PreviousMedian      *time.Time             `json:"previousMedian,omitempty"`
	ProjectedProgress   []AppPlanProgressBand  `json:"projectedProgress"`
	RequiredParts       int64                  `json:"requiredParts"`
	Revision            int64                  `json:"revision"`
	Upper               *time.Time             `json:"upper,omitempty"`
}

type AppPlanPartUpdate struct {
	ItemUUID      string   `json:"itemUuid"`
	LinkToken     string   `json:"linkToken,omitempty"`
	Prerequisites []string `json:"prerequisites"`
	Required      bool     `json:"required"`
}

type AppPlanUpdate struct {
	Deadline         *time.Time          `json:"deadline,omitempty"`
	ExpectedRevision int64               `json:"expectedRevision"`
	Parts            []AppPlanPartUpdate `json:"parts"`
}

type AppItemMeasurement struct {
	ObservationUUID string    `json:"observationUuid"`
	ObservedAt      time.Time `json:"observedAt"`
	SeriesUUID      string    `json:"seriesUuid"`
	Value           float64   `json:"value"`
}

type AppItemMeasurementPage struct {
	NextCursor string               `json:"nextCursor,omitempty"`
	Rows       []AppItemMeasurement `json:"rows"`
	RunUUID    string               `json:"runUuid"`
}

type MeasurementObservationCSVImportRequest struct {
	CSV string `json:"csv"`
}

type AppLinkCollectionEntryAddRequest struct {
	// A share (or owner) link token of another item of this app.
	Token string `json:"token"`
}

type AppItemLinkRotated struct {
	Link  AppItemLink `json:"link"`
	Token string      `json:"token"`
}

type AppModelVersionCreateRequest struct {
	CounterLinks []string              `json:"counterLinks,omitempty"`
	Note         string                `json:"note,omitempty"`
	Quantities   []UncertaintyQuantity `json:"quantities"`
}

type AppItemAlertRule struct {
	// Whether the condition held at the last evaluation.
	ConditionMet bool      `json:"conditionMet"`
	CreatedAt    time.Time `json:"createdAt"`
	// How many times the condition has become true.
	Crossings int64 `json:"crossings"`
	// Quantity name for model_range_above.
	ModelOutput string `json:"modelOutput,omitempty"`
	RuleUUID    string `json:"ruleUuid"`
	// Days without a new estimate, 1 to 365; estimate_stale only.
	StaleDays int64 `json:"staleDays,omitempty"`
	// Deadline probability or model output threshold.
	Threshold float64 `json:"threshold,omitempty"`
	Type      string  `json:"type"`
}

type AppItemAlertSubscription struct {
	CreatedAt        time.Time `json:"createdAt"`
	SubscriptionUUID string    `json:"subscriptionUuid"`
}

type AppItemAlertWebhook struct {
	CreatedAt time.Time `json:"createdAt"`
	Endpoint  string    `json:"endpoint"`
	// webhook posts signed Custd JSON; slack posts a Slack message.
	Kind        string `json:"kind"`
	WebhookUUID string `json:"webhookUuid"`
}

type AppItemAlertSettings struct {
	DeliveryAvailable bool                      `json:"deliveryAvailable"`
	Rules             []AppItemAlertRule        `json:"rules"`
	Subscription      *AppItemAlertSubscription `json:"subscription,omitempty"`
	Webhooks          []AppItemAlertWebhook     `json:"webhooks"`
}

type AppItemAlertDelivery struct {
	Attempts  int64     `json:"attempts"`
	Channel   string    `json:"channel"`
	CreatedAt time.Time `json:"createdAt"`
	Crossing  int64     `json:"crossing"`
	// Matches the webhook's X-Custd-Delivery-Id.
	DeliveryUUID   string    `json:"deliveryUuid"`
	LastErrorClass string    `json:"lastErrorClass,omitempty"`
	LastStatusCode int64     `json:"lastStatusCode,omitempty"`
	Outcome        string    `json:"outcome"`
	RuleType       string    `json:"ruleType"`
	RuleUUID       string    `json:"ruleUuid"`
	UpdatedAt      time.Time `json:"updatedAt"`
	WebhookUUID    string    `json:"webhookUuid,omitempty"`
}

type AppItemAlertDeliveryList struct {
	Deliveries []AppItemAlertDelivery `json:"deliveries"`
}

type AppItemAlertRuleCreateRequest struct {
	// Quantity name for model_range_above.
	ModelOutput string `json:"modelOutput,omitempty"`
	// Days without a new estimate, 1 to 365; estimate_stale only.
	StaleDays *int64 `json:"staleDays,omitempty"`
	// Deadline probability or model output threshold.
	Threshold *float64 `json:"threshold,omitempty"`
	Type      string   `json:"type"`
}

type AppItemAlertWebhookCreateRequest struct {
	// HTTPS URL on port 443 that resolves only to public addresses.
	Endpoint string `json:"endpoint"`
}

type AppItemAlertWebhookCreated struct {
	// Verifies X-Custd-Signature. Shown once.
	SigningSecret string              `json:"signingSecret"`
	Webhook       AppItemAlertWebhook `json:"webhook"`
}

type AppItemAlertSlackCreateRequest struct {
	// Slack incoming-webhook URL, https://hooks.slack.com/services/...
	WebhookURL string `json:"webhookUrl"`
}

type AppMeetingAdvanceRequest struct {
	// The block to finish (currentBlockUuid); a repeat is a no-op, a stale one 409.
	BlockUUID string `json:"blockUuid"`
}

type AppMeetingMinutesUpdate struct {
	Actions          []AppMeetingAction `json:"actions"`
	Decisions        string             `json:"decisions"`
	ExpectedRevision int64              `json:"expectedRevision"`
	Minutes          string             `json:"minutes"`
	PlanLinkToken    string             `json:"planLinkToken,omitempty"`
	PlanUUID         string             `json:"planUuid,omitempty"`
	SharedSummary    string             `json:"sharedSummary,omitempty"`
}

type MeasurementForecastInterval struct {
	Confidence float64   `json:"confidence"`
	Lower      time.Time `json:"lower"`
	Upper      time.Time `json:"upper"`
}

type MeasurementForecastHistogramBin struct {
	Count int64     `json:"count"`
	End   time.Time `json:"end"`
	Start time.Time `json:"start"`
}

type MeasurementForecastHistogram struct {
	Bins       []MeasurementForecastHistogramBin `json:"bins"`
	NeverCount int64                             `json:"neverCount"`
}

type MeasurementForecastPoint struct {
	At    time.Time `json:"at"`
	Value float64   `json:"value"`
}

type MeasurementForecastLine struct {
	Points []MeasurementForecastPoint `json:"points"`
}

type MeasurementForecastDetail struct {
	ArrivalHistogram MeasurementForecastHistogram `json:"arrivalHistogram"`
	PosteriorLines   []MeasurementForecastLine    `json:"posteriorLines"`
}

type MeasurementForecastRate struct {
	PerDay    float64 `json:"perDay"`
	PerHour   float64 `json:"perHour"`
	PerMinute float64 `json:"perMinute"`
	PerSecond float64 `json:"perSecond"`
}

type MeasurementForecastCounterResult struct {
	CompletionInterval  *MeasurementForecastInterval `json:"completionInterval,omitempty"`
	DeadlineProbability float64                      `json:"deadlineProbability,omitempty"`
	Detail              MeasurementForecastDetail    `json:"detail"`
	PooledRunCount      int64                        `json:"pooledRunCount"`
	Rate                *MeasurementForecastRate     `json:"rate,omitempty"`
}

type MeasurementForecastRunResponse struct {
	Counter         *MeasurementForecastCounterResult `json:"counter,omitempty"`
	ForecastFrom    *time.Time                        `json:"forecastFrom,omitempty"`
	ForecastRunUUID string                            `json:"forecastRunUuid"`
	GeneratedAt     string                            `json:"generatedAt"`
	InputCount      int64                             `json:"inputCount"`
	Method          string                            `json:"method"`
	Percentiles     map[string]string                 `json:"percentiles"`
	ProjectUUID     string                            `json:"projectUuid"`
	Provenance      map[string]string                 `json:"provenance"`
	SampleCount     int64                             `json:"sampleCount"`
	Seed            int64                             `json:"seed"`
	Status          string                            `json:"status"`
	Warnings        []MeasurementForecastWarning      `json:"warnings,omitempty"`
}

type AppModelInputOverride struct {
	Input UncertaintyInput `json:"input"`
	Name  string           `json:"name"`
}

type AppModelCalculateRequest struct {
	Inputs        []AppModelInputOverride `json:"inputs,omitempty"`
	Outputs       []string                `json:"outputs"`
	SampleCount   *int64                  `json:"sampleCount,omitempty"`
	Seed          *int64                  `json:"seed,omitempty"`
	VersionNumber *int64                  `json:"versionNumber,omitempty"`
}

type AppModelCalculateResult struct {
	ModelUUID     string                      `json:"modelUuid"`
	Quantities    []UncertaintyQuantityResult `json:"quantities"`
	SampleCount   int64                       `json:"sampleCount"`
	Seed          int64                       `json:"seed"`
	VersionNumber int64                       `json:"versionNumber"`
}

// REUSED existing: UncertaintyCreateModelRequest MeasurementEstimateList MeasurementProject UncertaintyModelDetail MeasurementProjectUpdate MeasurementObservationBulkRequest MeasurementObservationBulkResponse MeasurementEstimateCreate MeasurementEstimate MeasurementProjectComplete MeasurementRunCreate UncertaintyQuantity UncertaintyCopyModelRequest MeasurementForecastRunRequest MeasurementForecastWarning MeasurementForecastStoryResponse UncertaintyRunRequest UncertaintyRunDetail UncertaintyInput UncertaintyQuantityResult

// Create creates an item and its owner and share links.
func (c *AppItemClient) Create(ctx context.Context, access AppItemAccess, req AppItemCreateRequest) (*AppItemCreated, error) {
	var out AppItemCreated
	err := c.send(ctx, access, http.MethodPost, "/items", req, &out)
	return &out, err
}

// ListMine lists the signed-in person's items.
func (c *AppItemClient) ListMine(ctx context.Context, access AppItemAccess) (*AppItemList, error) {
	var out AppItemList
	err := c.send(ctx, access, http.MethodGet, "/items", nil, &out)
	return &out, err
}

// Claim claims an unowned item with its owner link.
func (c *AppItemClient) Claim(ctx context.Context, access AppItemAccess) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/claim", nil, &out)
	return &out, err
}

// Get reads the item a link opens.
func (c *AppItemClient) Get(ctx context.Context, access AppItemAccess) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodGet, "/item", nil, &out)
	return &out, err
}

// Update changes an item's idle window and project settings.
func (c *AppItemClient) Update(ctx context.Context, access AppItemAccess, req AppItemSettingsRequest) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPatch, "/item", req, &out)
	return &out, err
}

// GetPlanForecast reads the overall completion forecast for a plan.
func (c *AppItemClient) GetPlanForecast(ctx context.Context, access AppItemAccess) (*AppPlanForecast, error) {
	var out AppPlanForecast
	err := c.send(ctx, access, http.MethodGet, "/item/plan/forecast", nil, &out)
	return &out, err
}

// UpdatePlan replaces a plan's parts and dependencies at an expected revision.
func (c *AppItemClient) UpdatePlan(ctx context.Context, access AppItemAccess, req AppPlanUpdate) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPatch, "/item/plan", req, &out)
	return &out, err
}

// ListMeasurements pages through a run's measurements in observed time order.
func (c *AppItemClient) ListMeasurements(ctx context.Context, access AppItemAccess, page AppItemPageOptions) (*AppItemMeasurementPage, error) {
	var out AppItemMeasurementPage
	err := c.send(ctx, access, http.MethodGet, siteQueryPath("/item/measurements", page.values()), nil, &out)
	return &out, err
}

// AddMeasurements adds measurements, one or in bulk.
func (c *AppItemClient) AddMeasurements(ctx context.Context, access AppItemAccess, req MeasurementObservationBulkRequest) (*MeasurementObservationBulkResponse, error) {
	var out MeasurementObservationBulkResponse
	err := c.send(ctx, access, http.MethodPost, "/item/measurements", req, &out)
	return &out, err
}

// ImportMeasurementsCSV adds measurements from CSV.
func (c *AppItemClient) ImportMeasurementsCSV(ctx context.Context, access AppItemAccess, req MeasurementObservationCSVImportRequest) (*MeasurementObservationBulkResponse, error) {
	var out MeasurementObservationBulkResponse
	err := c.send(ctx, access, http.MethodPost, "/item/measurements/csv", req, &out)
	return &out, err
}

// CreateEstimate records a dated low/high delivery estimate.
func (c *AppItemClient) CreateEstimate(ctx context.Context, access AppItemAccess, req MeasurementEstimateCreate) (*MeasurementEstimate, error) {
	var out MeasurementEstimate
	err := c.send(ctx, access, http.MethodPost, "/item/estimates", req, &out)
	return &out, err
}

// Complete completes a project on its actual date.
func (c *AppItemClient) Complete(ctx context.Context, access AppItemAccess, req MeasurementProjectComplete) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/complete", req, &out)
	return &out, err
}

// Cancel cancels a project.
func (c *AppItemClient) Cancel(ctx context.Context, access AppItemAccess) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/cancel", nil, &out)
	return &out, err
}

// Reopen reopens a counter whose target was met.
func (c *AppItemClient) Reopen(ctx context.Context, access AppItemAccess) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/reopen", nil, &out)
	return &out, err
}

// ListRuns pages through measurement runs.
func (c *AppItemClient) ListRuns(ctx context.Context, access AppItemAccess, page AppItemPageOptions) (*AppItemRunPage, error) {
	var out AppItemRunPage
	err := c.send(ctx, access, http.MethodGet, siteQueryPath("/item/runs", page.values()), nil, &out)
	return &out, err
}

// StartRun starts a new measurement run.
func (c *AppItemClient) StartRun(ctx context.Context, access AppItemAccess, req MeasurementRunCreate) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/runs", req, &out)
	return &out, err
}

// AddEntry adds the item a pasted link opens to a collection.
func (c *AppItemClient) AddEntry(ctx context.Context, access AppItemAccess, req AppLinkCollectionEntryAddRequest) (*AppLinkCollectionEntry, error) {
	var out AppLinkCollectionEntry
	err := c.send(ctx, access, http.MethodPost, "/item/entries", req, &out)
	return &out, err
}

// RemoveEntry removes an entry from a collection.
func (c *AppItemClient) RemoveEntry(ctx context.Context, access AppItemAccess, entryUUID string) error {
	return c.send(ctx, access, http.MethodDelete, "/item/entries/"+url.PathEscape(entryUUID), nil, nil)
}

// RevokeLink revokes one of the item's links.
func (c *AppItemClient) RevokeLink(ctx context.Context, access AppItemAccess, linkUUID string) error {
	return c.send(ctx, access, http.MethodDelete, "/item/links/"+url.PathEscape(linkUUID), nil, nil)
}

// RotateLink gives a link a new token.
func (c *AppItemClient) RotateLink(ctx context.Context, access AppItemAccess, linkUUID string) (*AppItemLinkRotated, error) {
	var out AppItemLinkRotated
	err := c.send(ctx, access, http.MethodPost, "/item/links/"+url.PathEscape(linkUUID)+"/rotate", nil, &out)
	return &out, err
}

// AddModelVersion saves a new version of a model's quantities.
func (c *AppItemClient) AddModelVersion(ctx context.Context, access AppItemAccess, req AppModelVersionCreateRequest) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/model/versions", req, &out)
	return &out, err
}

// CopyModel copies a model version into a new item with its own links.
func (c *AppItemClient) CopyModel(ctx context.Context, access AppItemAccess, req UncertaintyCopyModelRequest) (*AppItemCreated, error) {
	var out AppItemCreated
	err := c.send(ctx, access, http.MethodPost, "/item/model/copy", req, &out)
	return &out, err
}

// GetAlertSettings reads the item's alert rules, webhooks and your email subscription.
func (c *AppItemClient) GetAlertSettings(ctx context.Context, access AppItemAccess) (*AppItemAlertSettings, error) {
	var out AppItemAlertSettings
	err := c.send(ctx, access, http.MethodGet, "/item/alerts", nil, &out)
	return &out, err
}

// ListAlertDeliveries reads the newest alert deliveries and their outcome.
func (c *AppItemClient) ListAlertDeliveries(ctx context.Context, access AppItemAccess) (*AppItemAlertDeliveryList, error) {
	var out AppItemAlertDeliveryList
	err := c.send(ctx, access, http.MethodGet, "/item/alerts/deliveries", nil, &out)
	return &out, err
}

// AddAlertRule adds an alert rule.
func (c *AppItemClient) AddAlertRule(ctx context.Context, access AppItemAccess, req AppItemAlertRuleCreateRequest) (*AppItemAlertRule, error) {
	var out AppItemAlertRule
	err := c.send(ctx, access, http.MethodPost, "/item/alerts/rules", req, &out)
	return &out, err
}

// RemoveAlertRule removes an alert rule.
func (c *AppItemClient) RemoveAlertRule(ctx context.Context, access AppItemAccess, ruleUUID string) error {
	return c.send(ctx, access, http.MethodDelete, "/item/alerts/rules/"+url.PathEscape(ruleUUID), nil, nil)
}

// AddAlertWebhook registers a signed webhook; its secret is shown once.
func (c *AppItemClient) AddAlertWebhook(ctx context.Context, access AppItemAccess, req AppItemAlertWebhookCreateRequest) (*AppItemAlertWebhookCreated, error) {
	var out AppItemAlertWebhookCreated
	err := c.send(ctx, access, http.MethodPost, "/item/alerts/webhooks", req, &out)
	return &out, err
}

// RemoveAlertWebhook removes a webhook or Slack destination.
func (c *AppItemClient) RemoveAlertWebhook(ctx context.Context, access AppItemAccess, webhookUUID string) error {
	return c.send(ctx, access, http.MethodDelete, "/item/alerts/webhooks/"+url.PathEscape(webhookUUID), nil, nil)
}

// AddAlertSlack posts the item's alerts to a Slack incoming webhook; remove it like a webhook.
func (c *AppItemClient) AddAlertSlack(ctx context.Context, access AppItemAccess, req AppItemAlertSlackCreateRequest) (*AppItemAlertWebhook, error) {
	var out AppItemAlertWebhook
	err := c.send(ctx, access, http.MethodPost, "/item/alerts/slack", req, &out)
	return &out, err
}

// SubscribeAlertEmail gets the item's alerts by email (signed in).
func (c *AppItemClient) SubscribeAlertEmail(ctx context.Context, access AppItemAccess) (*AppItemAlertSubscription, error) {
	var out AppItemAlertSubscription
	err := c.send(ctx, access, http.MethodPut, "/item/alerts/subscription", nil, &out)
	return &out, err
}

// UnsubscribeAlertEmail stops your alert emails for the item (signed in).
func (c *AppItemClient) UnsubscribeAlertEmail(ctx context.Context, access AppItemAccess) error {
	return c.send(ctx, access, http.MethodDelete, "/item/alerts/subscription", nil, nil)
}

// StartMeeting starts a meeting: its first block goes on.
func (c *AppItemClient) StartMeeting(ctx context.Context, access AppItemAccess) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/meeting/start", nil, &out)
	return &out, err
}

// AdvanceMeeting finishes the named current block and start the next.
func (c *AppItemClient) AdvanceMeeting(ctx context.Context, access AppItemAccess, req AppMeetingAdvanceRequest) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/meeting/advance", req, &out)
	return &out, err
}

// EndMeeting ends a meeting now.
func (c *AppItemClient) EndMeeting(ctx context.Context, access AppItemAccess) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPost, "/item/meeting/end", nil, &out)
	return &out, err
}

// UpdateMeetingMinutes saves owner-only meeting minutes and linked actions.
func (c *AppItemClient) UpdateMeetingMinutes(ctx context.Context, access AppItemAccess, req AppMeetingMinutesUpdate) (*AppItemRead, error) {
	var out AppItemRead
	err := c.send(ctx, access, http.MethodPatch, "/item/meeting/minutes", req, &out)
	return &out, err
}

// RunForecast runs the item's forecast.
func (c *AppItemClient) RunForecast(ctx context.Context, access AppItemAccess, req MeasurementForecastRunRequest) (*MeasurementForecastRunResponse, error) {
	var out MeasurementForecastRunResponse
	err := c.send(ctx, access, http.MethodPost, "/forecast/runs", req, &out)
	return &out, err
}

// GetForecastRun reads one forecast run of the item.
func (c *AppItemClient) GetForecastRun(ctx context.Context, access AppItemAccess, runUUID string) (*MeasurementForecastRunResponse, error) {
	var out MeasurementForecastRunResponse
	err := c.send(ctx, access, http.MethodGet, "/forecast/runs/"+url.PathEscape(runUUID), nil, &out)
	return &out, err
}

// GetLatestForecast reads the item's latest forecast.
func (c *AppItemClient) GetLatestForecast(ctx context.Context, access AppItemAccess) (*MeasurementForecastRunResponse, error) {
	var out MeasurementForecastRunResponse
	err := c.send(ctx, access, http.MethodGet, "/forecast/latest", nil, &out)
	return &out, err
}

// GetForecastStory reads the item's measurements and forecast story.
func (c *AppItemClient) GetForecastStory(ctx context.Context, access AppItemAccess) (*MeasurementForecastStoryResponse, error) {
	var out MeasurementForecastStoryResponse
	err := c.send(ctx, access, http.MethodGet, "/forecast/story", nil, &out)
	return &out, err
}

// RunModel runs the item's uncertainty model.
func (c *AppItemClient) RunModel(ctx context.Context, access AppItemAccess, req UncertaintyRunRequest) (*UncertaintyRunDetail, error) {
	var out UncertaintyRunDetail
	err := c.send(ctx, access, http.MethodPost, "/models/runs", req, &out)
	return &out, err
}

// CalculateModel calculates the item's uncertainty model with input overrides, without storing a run.
func (c *AppItemClient) CalculateModel(ctx context.Context, access AppItemAccess, req AppModelCalculateRequest) (*AppModelCalculateResult, error) {
	var out AppModelCalculateResult
	err := c.send(ctx, access, http.MethodPost, "/models/calculate", req, &out)
	return &out, err
}

// GetLatestModelRun reads the latest run of the item's uncertainty model.
func (c *AppItemClient) GetLatestModelRun(ctx context.Context, access AppItemAccess) (*UncertaintyRunDetail, error) {
	var out UncertaintyRunDetail
	err := c.send(ctx, access, http.MethodGet, "/models/latest", nil, &out)
	return &out, err
}

// AppItemPageOptions pages a run or measurement list. After is the previous
// page's cursor, Limit caps the page (zero uses the server default) and Run
// names the measurement run to read; runs ignore it.
type AppItemPageOptions struct {
	After string
	Limit int
	Run   string
}

func (o AppItemPageOptions) values() url.Values {
	values := url.Values{"after": {o.After}, "run": {o.Run}}
	if o.Limit > 0 {
		values.Set("limit", strconv.Itoa(o.Limit))
	}
	return values
}
