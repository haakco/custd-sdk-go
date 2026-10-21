package custd

import (
	"context"
	"net/http"
	"net/url"
)

// ReportingPacksAdminClient owns the admin-owned reporting-pack lifecycle and
// the reviewed setup templates. The authority model is company_id; the SDK
// sends no TenantSlug on these surfaces, matching the frozen contract.
type ReportingPacksAdminClient struct {
	admin *AdminClient
}

type ReportingPackDraft struct {
	ID         int64 `json:"id"`
	Revision   int64 `json:"revision"`
	Definition Pack  `json:"definition"`
	// ValidationError names why a stored draft cannot be published, and is empty when it can.
	ValidationError string `json:"validationError,omitempty"`
	CreatedAt       string `json:"createdAt,omitempty"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}

type ReportingPackDraftCreate struct {
	Definition Pack `json:"definition"`
}

type ReportingPackDraftUpdate struct {
	Definition       Pack  `json:"definition"`
	ExpectedRevision int64 `json:"expectedRevision"`
}

type ReportingPackDraftListResponse struct {
	Drafts []ReportingPackDraft `json:"drafts"`
}

type ReportingPackValidateResponse struct {
	Valid bool `json:"valid"`
}

type ReportingPackPreviewRequest struct {
	Definition Pack      `json:"definition"`
	TenantSlug string    `json:"tenantSlug"`
	Query      QueryHint `json:"query"`
}

type ReportingPackPreviewResponse struct {
	Buckets []RenderedWidgetBucket `json:"buckets"`
	Value   RenderedMetricValue    `json:"value"`
}

// QueryHint mirrors the reporting contract's browser-supplied query. The SDK cannot import
// the reporting library, so the shape is restated here field for field.
//
// Every field matters on the wire: the server validates the query before any SQL runs and
// rejects a hint without a positive maxRows, so an omitted field is a refused preview rather
// than a defaulted one.
type QueryHint struct {
	DashboardKey string            `json:"dashboardKey,omitempty"`
	WidgetKey    string            `json:"widgetKey,omitempty"`
	Template     string            `json:"template"`
	Metrics      []string          `json:"metrics"`
	Dimensions   []string          `json:"dimensions,omitempty"`
	Filters      []RequestedFilter `json:"filters,omitempty"`
	From         string            `json:"from,omitempty"`
	To           string            `json:"to,omitempty"`
	RangeDays    int               `json:"rangeDays,omitempty"`
	MaxRows      int               `json:"maxRows,omitempty"`
	CountOnly    bool              `json:"countOnly,omitempty"`
	Comparison   string            `json:"comparison,omitempty"`
}

// RequestedFilter is one exact predicate on a query dimension.
type RequestedFilter struct {
	Dimension string `json:"dimension"`
	Operator  string `json:"operator"`
	Value     string `json:"value,omitempty"`
}

type ReportingPackGeneration struct {
	ID               int64  `json:"id"`
	GenerationNumber int64  `json:"generationNumber"`
	SourceDraftID    int64  `json:"sourceDraftId"`
	Definition       Pack   `json:"definition"`
	State            string `json:"state"`
	CreatedAt        string `json:"createdAt,omitempty"`
}

type ReportingPackGenerationStatusResponse struct {
	Generation       ReportingPackGenerationStatusDetail      `json:"generation"`
	Acknowledgements []ReportingPackGenerationAcknowledgement `json:"acknowledgements"`
}

type ReportingPackGenerationStatusDetail struct {
	ID               int64  `json:"id"`
	GenerationNumber int64  `json:"generationNumber"`
	PackKey          string `json:"packKey"`
	State            string `json:"state"`
}

type ReportingPackGenerationAcknowledgement struct {
	Accepted             bool   `json:"accepted"`
	Consumer             string `json:"consumer"`
	ErrorDetail          string `json:"errorDetail,omitempty"`
	ObservedAt           string `json:"observedAt,omitempty"`
	ObservedGenerationID int64  `json:"observedGenerationId"`
}

type ReportingPackRollupProvenance struct {
	GenerationID          int64                                    `json:"generationId"`
	DefinitionFingerprint string                                   `json:"definitionFingerprint"`
	TenantSlug            string                                   `json:"tenantSlug,omitempty"`
	Materializations      []ReportingPackMaterializationProvenance `json:"materializations"`
}

type ReportingPackMaterializationProvenance struct {
	Status                string `json:"status"`
	DefinitionFingerprint string `json:"definitionFingerprint"`
	SourceCoverageCount   int64  `json:"sourceCoverageCount"`
}

// ReportingPackSetupBindingField describes one client-selectable binding a
// setup template renders. The value and its validation belong to the backend.
type ReportingPackSetupBindingField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
}

// ReportingPackSetupTemplate is one reviewed setup template: its canonical
// definition plus the client-selectable bindings the client renders.
type ReportingPackSetupTemplate struct {
	Kind        string                           `json:"kind"`
	DisplayName string                           `json:"displayName"`
	Definition  Pack                             `json:"definition"`
	Bindings    []ReportingPackSetupBindingField `json:"bindings"`
}

// ReportingPackSetupTemplateListResponse is the public list contract for the
// reviewed setup templates.
type ReportingPackSetupTemplateListResponse struct {
	Templates []ReportingPackSetupTemplate `json:"templates"`
}

// LogObservabilityBindings is the typed log-observability binding a client
// selects. The backend owns the metric, dimension, template and widget keys and
// the validation; it is never an untyped dictionary.
type LogObservabilityBindings struct {
	EventType     string `json:"eventType"`
	SeverityField string `json:"severityField"`
	ResourceField string `json:"resourceField"`
}

type ReportingPackSetupTemplateBindRequest struct {
	Bindings LogObservabilityBindings `json:"bindings"`
}

type ReportingPackSetupTemplateBindResponse struct {
	Kind       string `json:"kind"`
	Definition Pack   `json:"definition"`
}

func (c *ReportingPacksAdminClient) ListDrafts(ctx context.Context) (*ReportingPackDraftListResponse, error) {
	var out ReportingPackDraftListResponse
	err := c.admin.request(ctx, http.MethodGet, "/reporting-packs/drafts", nil, &out)
	return &out, err
}

func (c *ReportingPacksAdminClient) GetDraft(ctx context.Context, draftID string) (*ReportingPackDraft, error) {
	var out ReportingPackDraft
	err := c.admin.request(ctx, http.MethodGet, "/reporting-packs/drafts/"+url.PathEscape(draftID), nil, &out)
	return &out, err
}

func (c *ReportingPacksAdminClient) CreateDraft(ctx context.Context, req ReportingPackDraftCreate) (*ReportingPackDraft, error) {
	var out ReportingPackDraft
	err := c.admin.request(ctx, http.MethodPost, "/reporting-packs/drafts", req, &out)
	return &out, err
}

func (c *ReportingPacksAdminClient) UpdateDraft(
	ctx context.Context,
	draftID string,
	req ReportingPackDraftUpdate,
) (*ReportingPackDraft, error) {
	var out ReportingPackDraft
	err := c.admin.request(
		ctx,
		http.MethodPut,
		"/reporting-packs/drafts/"+url.PathEscape(draftID),
		req,
		&out,
	)
	return &out, err
}

func (c *ReportingPacksAdminClient) Validate(ctx context.Context, req ReportingPackDraftCreate) (*ReportingPackValidateResponse, error) {
	var out ReportingPackValidateResponse
	err := c.admin.request(ctx, http.MethodPost, "/reporting-packs/validate", req, &out)
	return &out, err
}

func (c *ReportingPacksAdminClient) Preview(
	ctx context.Context,
	req ReportingPackPreviewRequest,
) (*ReportingPackPreviewResponse, error) {
	var out ReportingPackPreviewResponse
	err := c.admin.request(ctx, http.MethodPost, "/reporting-packs/preview", req, &out)
	return &out, err
}

func (c *ReportingPacksAdminClient) Publish(ctx context.Context, draftID string) (*ReportingPackGeneration, error) {
	var out ReportingPackGeneration
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/reporting-packs/drafts/"+url.PathEscape(draftID)+"/publish",
		nil,
		&out,
	)
	return &out, err
}

func (c *ReportingPacksAdminClient) Restart(ctx context.Context, draftID string) (*ReportingPackGeneration, error) {
	var out ReportingPackGeneration
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/reporting-packs/drafts/"+url.PathEscape(draftID)+"/restart",
		nil,
		&out,
	)
	return &out, err
}

func (c *ReportingPacksAdminClient) GetGeneration(ctx context.Context, generationID string) (*ReportingPackGeneration, error) {
	var out ReportingPackGeneration
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/reporting-packs/generations/"+url.PathEscape(generationID),
		nil,
		&out,
	)
	return &out, err
}

func (c *ReportingPacksAdminClient) GetGenerationStatus(ctx context.Context, generationID string) (*ReportingPackGenerationStatusResponse, error) {
	var out ReportingPackGenerationStatusResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/reporting-packs/generations/"+url.PathEscape(generationID)+"/status",
		nil,
		&out,
	)
	return &out, err
}

func (c *ReportingPacksAdminClient) RollbackGeneration(ctx context.Context, generationID string) error {
	return c.admin.request(
		ctx,
		http.MethodPost,
		"/reporting-packs/generations/"+url.PathEscape(generationID)+"/rollback",
		nil,
		nil,
	)
}

func (c *ReportingPacksAdminClient) GetRollupProvenance(
	ctx context.Context,
	generationID string,
) (*ReportingPackRollupProvenance, error) {
	var out ReportingPackRollupProvenance
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/reporting-packs/generations/"+url.PathEscape(generationID)+"/rollup-provenance",
		nil,
		&out,
	)
	return &out, err
}

// ListSetupTemplates lists the reviewed setup templates, each with its canonical
// definition and the client-selectable bindings it renders.
func (c *ReportingPacksAdminClient) ListSetupTemplates(ctx context.Context) (*ReportingPackSetupTemplateListResponse, error) {
	var out ReportingPackSetupTemplateListResponse
	err := c.admin.request(ctx, http.MethodGet, "/reporting-packs/setup-templates", nil, &out)
	return &out, err
}

// BindSetupTemplate applies typed bindings to one reviewed setup template and
// returns the server-built, validated definition. It is a pure build: it creates
// no data and grants no scope, so its result must be posted through CreateDraft
// or Preview only when the caller intends to.
func (c *ReportingPacksAdminClient) BindSetupTemplate(
	ctx context.Context,
	kind string,
	req ReportingPackSetupTemplateBindRequest,
) (*ReportingPackSetupTemplateBindResponse, error) {
	var out ReportingPackSetupTemplateBindResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/reporting-packs/setup-templates/"+url.PathEscape(kind)+"/bind",
		req,
		&out,
	)
	return &out, err
}

// ReportingPackGenerationValidity states whether one published generation can still serve
// reads, and names why not when it cannot.
type ReportingPackGenerationValidity struct {
	ID               int64  `json:"id"`
	PackKey          string `json:"packKey"`
	GenerationNumber int    `json:"generationNumber"`
	State            string `json:"state"`
	ValidationError  string `json:"validationError,omitempty"`
}

// ReportingPackGenerationValidityListResponse is the named collection the route answers with.
type ReportingPackGenerationValidityListResponse struct {
	Generations []ReportingPackGenerationValidity `json:"generations"`
}

// ListGenerationValidity reports every generation's usability in one read, so a doctor can
// find a pack whose published generation no longer serves without checking them one by one.
func (c *ReportingPacksAdminClient) ListGenerationValidity(
	ctx context.Context,
) (*ReportingPackGenerationValidityListResponse, error) {
	var out ReportingPackGenerationValidityListResponse
	err := c.admin.request(ctx, http.MethodGet, "/reporting-packs/generation-validity", nil, &out)
	return &out, err
}
