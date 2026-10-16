package custd

import (
	"context"
	"net/http"
	"net/url"
)

// ReportingPacksAdminClient owns the admin-owned reporting-pack lifecycle. The
// authority model is company_id; the SDK sends no TenantSlug on these
// surfaces, matching the Stage 1 frozen contract.
type ReportingPacksAdminClient struct {
	admin *AdminClient
}

type Pack struct {
	Key         string          `json:"key"`
	DisplayName string          `json:"displayName"`
	Owner       string          `json:"owner,omitempty"`
	Enabled     bool            `json:"enabled"`
	EventTypes  []string        `json:"eventTypes"`
	Metrics     []PackMetric    `json:"metrics"`
	Dimensions  []PackDimension `json:"dimensions"`
	Dashboards  []PackDashboard `json:"dashboards"`
}

type PackMetric struct {
	Key         string   `json:"key"`
	DisplayName string   `json:"displayName"`
	Template    string   `json:"template"`
	Metrics     []string `json:"metrics"`
	Dimensions  []string `json:"dimensions"`
}

type PackDimension struct {
	Key         string `json:"key"`
	DisplayName string `json:"displayName"`
}

type PackDashboard struct {
	Key            string            `json:"key"`
	Title          string            `json:"title"`
	Hidden         bool              `json:"hidden"`
	DefaultRange   string            `json:"defaultRange"`
	RefreshSeconds int               `json:"refreshSeconds"`
	RequiredScopes []string          `json:"requiredScopes"`
	Widgets        []ReportingWidget `json:"widgets"`
}

type ReportingPackDraft struct {
	ID         int64  `json:"id"`
	Revision   int64  `json:"revision"`
	Definition Pack   `json:"definition"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
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

type QueryHint struct {
	Template   string   `json:"template"`
	Metrics    []string `json:"metrics"`
	Dimensions []string `json:"dimensions"`
	RangeDays  int      `json:"rangeDays,omitempty"`
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
