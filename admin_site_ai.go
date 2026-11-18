package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type SiteAIMode string

const (
	SiteAIOff      SiteAIMode = "off"
	SiteAIPlatform SiteAIMode = "platform"
	SiteAIOwnKey   SiteAIMode = "own-key"
)

type SiteAIConsentAction string

const (
	SiteAIGrantConsent  SiteAIConsentAction = "grant"
	SiteAIRevokeConsent SiteAIConsentAction = "revoke"
)

type SiteAIUsagePeriod string

const (
	SiteAIUsageDay           SiteAIUsagePeriod = "day"
	SiteAIUsageMonth         SiteAIUsagePeriod = "month"
	SiteAIUsagePreviousMonth SiteAIUsagePeriod = "previous-month"
)

// SiteAIProduct identifies an approved product with attributed answer-sampling usage.
type SiteAIProduct string

const (
	SiteAIProductHosting SiteAIProduct = "hosting"
	SiteAIProductCustd   SiteAIProduct = "custd"
	SiteAIProductCB      SiteAIProduct = "cb"
	SiteAIProductAwthy   SiteAIProduct = "awthy"
)

var siteAIProductValues = map[SiteAIProduct]bool{
	SiteAIProductHosting: true, SiteAIProductCustd: true, SiteAIProductCB: true, SiteAIProductAwthy: true,
}

func (value *SiteAIProduct) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, siteAIProductValues)
}

// SiteAISettings contains the current tenant policy and terms versions.
type SiteAISettings struct {
	Mode                 SiteAIMode `json:"aiMode"`
	MonthlyCapUSDMicros  int64      `json:"monthlyCapUsdMicros"`
	KeyConfigured        bool       `json:"keyConfigured"`
	KeyRefused           bool       `json:"keyRefused,omitempty"`
	CurrentTermsVersion  string     `json:"currentTermsVersion,omitempty"`
	AcceptedTermsVersion string     `json:"acceptedTermsVersion,omitempty"`
	ConsentAcceptedAt    *time.Time `json:"consentAcceptedAt,omitempty"`
}

// SiteAISettingsRequest replaces the tenant policy. Omitting the key can reuse an existing own key.
type SiteAISettingsRequest struct {
	Mode                SiteAIMode `json:"aiMode"`
	MonthlyCapUSDMicros int64      `json:"monthlyCapUsdMicros"`
	APIKey              string     `json:"apiKey,omitempty"`
}

type SiteAIConsentRequest struct {
	Action       SiteAIConsentAction `json:"action"`
	TermsVersion string              `json:"termsVersion,omitempty"`
}

// SiteAIUsageRow groups attempt counts, provider units and costs in USD micros.
type SiteAIUsageRow struct {
	TenantSlug             string        `json:"tenantSlug,omitempty"`
	Feature                string        `json:"feature"`
	SamplingProduct        SiteAIProduct `json:"samplingProduct,omitempty"`
	Provider               string        `json:"provider"`
	Method                 string        `json:"method"`
	Model                  string        `json:"model"`
	Payer                  string        `json:"payer"`
	PriceRevision          string        `json:"priceRevision"`
	Attempts               int64         `json:"attempts"`
	UnsettledAttempts      int64         `json:"unsettledAttempts"`
	UnknownCostAttempts    int64         `json:"unknownCostAttempts"`
	FailedAttempts         int64         `json:"failedAttempts"`
	NotSentAttempts        int64         `json:"notSentAttempts"`
	InputTokens            int64         `json:"inputTokens"`
	CachedInputTokens      int64         `json:"cachedInputTokens"`
	OutputTokens           int64         `json:"outputTokens"`
	Searches               int64         `json:"searches"`
	Pages                  int64         `json:"pages"`
	AudioMilliseconds      int64         `json:"audioMilliseconds"`
	ConfirmedCostUSDMicros int64         `json:"confirmedCostUsdMicros"`
	EstimatedCostUSDMicros int64         `json:"estimatedCostUsdMicros"`
	CapAccountedUSDMicros  int64         `json:"capAccountedUsdMicros"`
}

type SiteAIUsage struct {
	Period SiteAIUsagePeriod `json:"period"`
	From   time.Time         `json:"from"`
	Until  time.Time         `json:"until"`
	Rows   []SiteAIUsageRow  `json:"rows"`
}

type SiteAITenantUsage struct {
	SiteAIUsage
	Mode                SiteAIMode `json:"aiMode"`
	MonthlyCapUSDMicros int64      `json:"monthlyCapUsdMicros"`
}

type SiteAIPlatformUsage struct {
	SiteAIUsage
	SamplingMonthlyCapUSDMicros int64 `json:"samplingMonthlyCapUsdMicros"`
}

const siteAISettingsPath = siteDomainsPath + "/ai-settings"
const siteAITenantUsagePath = siteDomainsPath + "/ai-usage"
const siteAIPlatformUsagePath = "/site-insights/ai-usage"

// GetAISettings reads the tenant's current AI policy and provider terms version.
func (c *SiteDomainAdminClient) GetAISettings(ctx context.Context) (*SiteAISettings, error) {
	var out SiteAISettings
	err := c.request(ctx, http.MethodGet, siteAISettingsPath, nil, &out)
	return &out, err
}

// SetAISettings replaces the tenant's AI mode and cap. The API key is write-only.
func (c *SiteDomainAdminClient) SetAISettings(ctx context.Context, body SiteAISettingsRequest) (*SiteAISettings, error) {
	var out SiteAISettings
	err := c.request(ctx, http.MethodPut, siteAISettingsPath, body, &out)
	return &out, err
}

// SetAIConsent records explicit acceptance of the supplied terms version or revokes consent.
func (c *SiteDomainAdminClient) SetAIConsent(ctx context.Context, body SiteAIConsentRequest) (*SiteAISettings, error) {
	var out SiteAISettings
	err := c.request(ctx, http.MethodPut, siteAISettingsPath+"/consent", body, &out)
	return &out, err
}

// GetAIUsage reads the tenant's current UTC month's AI usage.
func (c *SiteDomainAdminClient) GetAIUsage(ctx context.Context) (*SiteAITenantUsage, error) {
	var out SiteAITenantUsage
	err := c.request(ctx, http.MethodGet, siteAITenantUsagePath, nil, &out)
	return &out, err
}

// GetSiteAIUsage reads AI usage across tenants. It requires a system administrator.
// An empty period uses the server default.
func (c *SystemAdminClient) GetSiteAIUsage(ctx context.Context, period SiteAIUsagePeriod) (*SiteAIPlatformUsage, ResponseMetadata, error) {
	path := siteAIPlatformUsagePath
	if period != "" {
		path += "?" + url.Values{"period": {string(period)}}.Encode()
	}
	var out SiteAIPlatformUsage
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, path, nil, &out)
	return &out, metadata, err
}
