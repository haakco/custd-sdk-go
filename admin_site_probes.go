package custd

import (
	"context"
	"net/http"
	"time"
)

// AdminSiteProbeKind identifies which check runs against the Site domain.
type AdminSiteProbeKind string

const (
	AdminSiteProbePing         AdminSiteProbeKind = "ping"
	AdminSiteProbeTCP          AdminSiteProbeKind = "tcp"
	AdminSiteProbeDNS          AdminSiteProbeKind = "dns"
	AdminSiteProbeHTTPDelivery AdminSiteProbeKind = "http-delivery"
)

// AdminSiteProbeSpec configures one network check.
type AdminSiteProbeSpec struct {
	Kind              AdminSiteProbeKind `json:"kind"`
	Port              int                `json:"port,omitempty"`
	ExpectedAddresses []string           `json:"expectedAddresses,omitempty"`
	Protocol          string             `json:"protocol,omitempty"`
}

// AdminSiteProbe is a configured check and its latest result.
type AdminSiteProbe struct {
	AdminSiteProbeSpec
	Status        AdminSiteUptimeStatus `json:"status"`
	LastCheckedAt *time.Time            `json:"lastCheckedAt,omitempty"`
	ResponseMs    int                   `json:"responseMs,omitempty"`
	LastFailure   string                `json:"lastFailure,omitempty"`
	Detail        *AdminSiteProbeDetail `json:"detail,omitempty"`
}

// AdminSiteProbeDetail contains observations relevant to the configured check.
type AdminSiteProbeDetail struct {
	ResolvedIP          string                    `json:"resolvedIp,omitempty"`
	Addresses           []string                  `json:"addresses,omitempty"`
	MissingAddresses    []string                  `json:"missingAddresses,omitempty"`
	UnexpectedAddresses []string                  `json:"unexpectedAddresses,omitempty"`
	StatusCode          int                       `json:"statusCode,omitempty"`
	NegotiatedProtocol  string                    `json:"negotiatedProtocol,omitempty"`
	TLSVersion          string                    `json:"tlsVersion,omitempty"`
	BytesRead           int64                     `json:"bytesRead,omitempty"`
	Timings             *AdminSiteDeliveryTimings `json:"timings,omitempty"`
	Cache               *AdminSiteDeliveryCache   `json:"cache,omitempty"`
}

// AdminSiteDeliveryTimings records delivery phases reached, in milliseconds.
type AdminSiteDeliveryTimings struct {
	DNSMs       *int `json:"dnsMs,omitempty"`
	ConnectMs   *int `json:"connectMs,omitempty"`
	TLSMs       *int `json:"tlsMs,omitempty"`
	FirstByteMs *int `json:"firstByteMs,omitempty"`
	TotalMs     int  `json:"totalMs"`
}

// AdminSiteDeliveryCache holds selected cache and delivery headers.
type AdminSiteDeliveryCache struct {
	CacheControl string `json:"cacheControl,omitempty"`
	Age          string `json:"age,omitempty"`
	CacheStatus  string `json:"cacheStatus,omitempty"`
	Server       string `json:"server,omitempty"`
	Via          string `json:"via,omitempty"`
}

// AdminSiteProbeList contains a domain's configured checks and results.
type AdminSiteProbeList struct {
	Probes []AdminSiteProbe `json:"probes"`
}

type adminSiteProbesRequest struct {
	Probes []AdminSiteProbeSpec `json:"probes"`
}

// ListProbes reads the domain's network probe settings and results.
func (c *SiteDomainAdminClient) ListProbes(ctx context.Context, domainUUID string) (*AdminSiteProbeList, error) {
	var result AdminSiteProbeList
	err := c.request(ctx, http.MethodGet, siteDomainPath(domainUUID)+"/probes", nil, &result)
	return &result, err
}

// SetProbes replaces the domain's network probes. An empty slice disables them.
func (c *SiteDomainAdminClient) SetProbes(ctx context.Context, domainUUID string, probes []AdminSiteProbeSpec) (*AdminSiteProbeList, error) {
	var result AdminSiteProbeList
	err := c.request(ctx, http.MethodPut, siteDomainPath(domainUUID)+"/probes",
		adminSiteProbesRequest{Probes: nonNilSlice(probes)}, &result)
	return &result, err
}
