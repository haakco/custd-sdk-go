package custd

// TemporalTrust is the coverage and provenance the server attaches to one
// rendered value.
type TemporalTrust struct {
	CoverageStart      string                `json:"coverageStart,omitempty"`
	CoverageEnd        string                `json:"coverageEnd,omitempty"`
	WindowCoverage     float64               `json:"windowCoverage,omitempty"`
	FreshnessMs        int64                 `json:"freshnessMs,omitempty"`
	Warnings           []string              `json:"warnings,omitempty"`
	SourceRange        *SourceRange          `json:"sourceRange,omitempty"`
	CalculationVersion string                `json:"calculationVersion,omitempty"`
	Reconciliation     *ReconciliationResult `json:"reconciliation,omitempty"`
}

// SourceRange is the source time range a rendered value covers.
type SourceRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ReconciliationResult is the server's selected source for a series and the
// evidence behind it. A stale or unavailable result carries no value.
type ReconciliationResult struct {
	TenantKey            string               `json:"tenantKey"`
	SeriesKey            string               `json:"seriesKey"`
	SelectedSourceKey    string               `json:"selectedSourceKey,omitempty"`
	SelectedSourceSeries string               `json:"selectedSourceSeries,omitempty"`
	ObservedAt           string               `json:"observedAt,omitempty"`
	Value                *float64             `json:"value,omitempty"`
	Unit                 string               `json:"unit,omitempty"`
	Availability         string               `json:"availability"`
	Status               string               `json:"status"`
	ContributingSources  []SourceContribution `json:"contributingSources,omitempty"`
	Divergence           SourceDivergence     `json:"divergence"`
}

// SourceContribution is one source considered during reconciliation.
type SourceContribution struct {
	SourceKey        string   `json:"sourceKey"`
	SourceSeriesKey  string   `json:"sourceSeriesKey"`
	ObservedAt       string   `json:"observedAt,omitempty"`
	Value            *float64 `json:"value,omitempty"`
	Unit             string   `json:"unit,omitempty"`
	Availability     string   `json:"availability"`
	FreshnessSeconds int64    `json:"freshnessSeconds,omitempty"`
	Stale            bool     `json:"stale,omitempty"`
	Selected         bool     `json:"selected,omitempty"`
}

// SourceDivergence reports whether reconciled sources disagreed.
type SourceDivergence struct {
	Detected           bool    `json:"detected"`
	AbsoluteDifference float64 `json:"absoluteDifference,omitempty"`
	RelativeDifference float64 `json:"relativeDifference,omitempty"`
}

// EmailDeliveryRecipients is the delivery outcome per recipient of a
// reporting email widget.
type EmailDeliveryRecipients struct {
	Complete   bool                     `json:"complete"`
	Recipients []EmailDeliveryRecipient `json:"recipients"`
}

// EmailDeliveryRecipient is one recipient's latest delivery outcome.
type EmailDeliveryRecipient struct {
	RecipientID string `json:"recipientId"`
	Outcome     string `json:"outcome"`
	OccurredAt  string `json:"occurredAt"`
}

// SnapshotLabelMetadata names a data label and its values used by a result.
type SnapshotLabelMetadata struct {
	DefinitionUUID string               `json:"definitionUuid"`
	Key            string               `json:"key"`
	DisplayName    string               `json:"displayName"`
	Sensitivity    string               `json:"sensitivity"`
	Values         []SnapshotLabelValue `json:"values"`
}

// SnapshotLabelValue is one data label value.
type SnapshotLabelValue struct {
	UUID        string `json:"uuid"`
	DisplayName string `json:"displayName"`
}

// RenderedVisualResult is the server-computed shape behind a specialised
// widget kind; only the list matching Kind is populated.
type RenderedVisualResult struct {
	Kind         string                      `json:"kind"`
	Summary      string                      `json:"summary"`
	Method       string                      `json:"method,omitempty"`
	Lifecycle    []RenderedLifecycleStage    `json:"lifecycle,omitempty"`
	Cohorts      []RenderedCohortCell        `json:"cohorts,omitempty"`
	Distribution []RenderedDistributionBin   `json:"distribution,omitempty"`
	Quantiles    []RenderedQuantile          `json:"quantiles,omitempty"`
	Intervals    []RenderedInterval          `json:"intervals,omitempty"`
	Calibration  []RenderedCalibrationBucket `json:"calibration,omitempty"`
	Anomalies    []RenderedAnomaly           `json:"anomalies,omitempty"`
	Drivers      []RenderedDriver            `json:"drivers,omitempty"`
}

type RenderedLifecycleStage struct {
	Key       string              `json:"key"`
	Label     string              `json:"label"`
	Order     int                 `json:"order"`
	Entered   RenderedMetricValue `json:"entered"`
	Converted RenderedMetricValue `json:"converted"`
	DropOff   RenderedMetricValue `json:"dropOff"`
}

type RenderedCohortCell struct {
	Cohort     string              `json:"cohort"`
	Window     string              `json:"window"`
	Population int                 `json:"population"`
	Suppressed bool                `json:"suppressed,omitempty"`
	Value      RenderedMetricValue `json:"value"`
}

type RenderedDistributionBin struct {
	Lower float64             `json:"lower"`
	Upper float64             `json:"upper"`
	Value RenderedMetricValue `json:"value"`
}

type RenderedQuantile struct {
	Quantile float64             `json:"quantile"`
	Value    RenderedMetricValue `json:"value"`
}

type RenderedInterval struct {
	Label    string              `json:"label"`
	Lower    RenderedMetricValue `json:"lower"`
	Estimate RenderedMetricValue `json:"estimate"`
	Upper    RenderedMetricValue `json:"upper"`
}

type RenderedCalibrationBucket struct {
	PredictedLower float64             `json:"predictedLower"`
	PredictedUpper float64             `json:"predictedUpper"`
	Observed       RenderedMetricValue `json:"observed"`
}

type RenderedAnomaly struct {
	At    string              `json:"at"`
	Value RenderedMetricValue `json:"value"`
	Lower float64             `json:"lower"`
	Upper float64             `json:"upper"`
}

type RenderedDriver struct {
	Key          string              `json:"key"`
	Label        string              `json:"label"`
	Method       string              `json:"method"`
	Contribution RenderedMetricValue `json:"contribution"`
	Rank         int                 `json:"rank"`
}
