package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Tenant measurement reads: the overview tiles and a project's recorded forecast story.

type MeasurementOverviewPoint struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type MeasurementOverviewArea struct {
	Detail     string                     `json:"detail"`
	Headline   string                     `json:"headline"`
	Key        string                     `json:"key"`
	Points     []MeasurementOverviewPoint `json:"points,omitempty"`
	Provenance string                     `json:"provenance,omitempty"`
	State      string                     `json:"state"`
	Title      string                     `json:"title"`
	Value      string                     `json:"value"`
	WidgetKind string                     `json:"widgetKind"`
}

type MeasurementOverviewTenant struct {
	CompanySlug string `json:"companySlug"`
	Label       string `json:"label"`
}

type MeasurementOverviewResponse struct {
	Areas []MeasurementOverviewArea `json:"areas"`
	// UTC generation time in RFC 3339 format.
	GeneratedAt time.Time                 `json:"generatedAt"`
	Tenant      MeasurementOverviewTenant `json:"tenant"`
}

type MeasurementForecastStoryEstimate struct {
	EstimatedAt  time.Time `json:"estimatedAt"`
	High         time.Time `json:"high"`
	Low          time.Time `json:"low"`
	MovementDays float64   `json:"movementDays,omitempty"`
	Note         string    `json:"note,omitempty"`
	WidthDays    float64   `json:"widthDays"`
}

type MeasurementForecastStoryDelivery struct {
	Estimates  []MeasurementForecastStoryEstimate `json:"estimates"`
	TargetDate *time.Time                         `json:"targetDate,omitempty"`
}

type MeasurementForecastStoryObservation struct {
	ObservedAt time.Time `json:"observedAt"`
	Value      float64   `json:"value"`
}

type MeasurementPredictionContribution struct {
	Delta        float64 `json:"delta"`
	FeatureKey   string  `json:"featureKey"`
	FeatureValue float64 `json:"featureValue,omitempty"`
	Movement     string  `json:"movement"`
	Value        float64 `json:"value"`
}

type MeasurementPredictionStoryPoint struct {
	AsOfAt       time.Time  `json:"asOfAt"`
	HorizonEndAt *time.Time `json:"horizonEndAt,omitempty"`
	InputHash    string     `json:"inputHash,omitempty"`
	Probability  float64    `json:"probability"`
	RunUUID      string     `json:"runUuid"`
}

type MeasurementPredictionCalibrationBucket struct {
	Lower         float64 `json:"lower"`
	ObservedRate  float64 `json:"observedRate,omitempty"`
	PositiveCount int64   `json:"positiveCount"`
	Predicted     float64 `json:"predicted"`
	SampleCount   int64   `json:"sampleCount"`
	Upper         float64 `json:"upper"`
}

type MeasurementPredictionStoryEvaluation struct {
	BrierScore       float64                                  `json:"brierScore,omitempty"`
	Calibration      []MeasurementPredictionCalibrationBucket `json:"calibration"`
	LogLoss          float64                                  `json:"logLoss,omitempty"`
	NegativeCount    int64                                    `json:"negativeCount"`
	PositiveCount    int64                                    `json:"positiveCount"`
	ResolvedRunCount int64                                    `json:"resolvedRunCount"`
	WindowEndAt      time.Time                                `json:"windowEndAt"`
	WindowStartAt    time.Time                                `json:"windowStartAt"`
}

type MeasurementPredictionStoryOutcome struct {
	IsLateEvidence bool       `json:"isLateEvidence"`
	ObservedValue  int64      `json:"observedValue"`
	OutcomeKind    string     `json:"outcomeKind,omitempty"`
	Resolution     string     `json:"resolution"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	RunUUID        string     `json:"runUuid"`
}

type MeasurementPredictionSourceHealth struct {
	ConsecutiveFailures int64      `json:"consecutiveFailures"`
	DisplayName         string     `json:"displayName"`
	LastAttemptAt       *time.Time `json:"lastAttemptAt,omitempty"`
	LastSucceededAt     *time.Time `json:"lastSucceededAt,omitempty"`
	NextEligibleAt      *time.Time `json:"nextEligibleAt,omitempty"`
	SourceKey           string     `json:"sourceKey"`
	Status              string     `json:"status"`
}

type MeasurementPredictionStory struct {
	Contributions  []MeasurementPredictionContribution   `json:"contributions"`
	Current        *MeasurementPredictionStoryPoint      `json:"current,omitempty"`
	DefinitionKey  string                                `json:"definitionKey"`
	DefinitionUUID string                                `json:"definitionUuid"`
	DisplayName    string                                `json:"displayName"`
	Evaluation     *MeasurementPredictionStoryEvaluation `json:"evaluation,omitempty"`
	History        []MeasurementPredictionStoryPoint     `json:"history"`
	Outcomes       []MeasurementPredictionStoryOutcome   `json:"outcomes"`
	QualityState   string                                `json:"qualityState"`
	Sources        []MeasurementPredictionSourceHealth   `json:"sources"`
	State          string                                `json:"state"`
}

type PresentationPoint struct {
	Date         string  `json:"date,omitempty"`
	LowerDate    string  `json:"lowerDate,omitempty"`
	LowerValue   float64 `json:"lowerValue,omitempty"`
	Selected     bool    `json:"selected,omitempty"`
	SnapshotUUID string  `json:"snapshotUuid,omitempty"`
	UpperDate    string  `json:"upperDate,omitempty"`
	UpperValue   float64 `json:"upperValue,omitempty"`
	Value        float64 `json:"value,omitempty"`
	X            string  `json:"x"`
}

type PresentationLayer struct {
	Label  string              `json:"label"`
	Points []PresentationPoint `json:"points"`
	Role   string              `json:"role"`
	YKind  string              `json:"yKind"`
}

type PresentationReference struct {
	Date  string  `json:"date,omitempty"`
	Label string  `json:"label"`
	Role  string  `json:"role"`
	Value float64 `json:"value,omitempty"`
}

type PresentationTableColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type PresentationSupportingTableRow struct {
	Cells []string `json:"cells"`
	Key   string   `json:"key"`
}

type PresentationSupportingTable struct {
	Columns []PresentationTableColumn        `json:"columns"`
	Key     string                           `json:"key"`
	Label   string                           `json:"label"`
	Rows    []PresentationSupportingTableRow `json:"rows"`
}

type PresentationTableRow struct {
	Cells        []string `json:"cells"`
	SnapshotUUID string   `json:"snapshotUuid"`
}

type CompiledPresentation struct {
	Caption          string                        `json:"caption"`
	Height           int64                         `json:"height"`
	Layers           []PresentationLayer           `json:"layers"`
	NonColorEncoding string                        `json:"nonColorEncoding"`
	PresetID         string                        `json:"presetId"`
	References       []PresentationReference       `json:"references"`
	Renderer         string                        `json:"renderer"`
	Summary          string                        `json:"summary"`
	SupportingTables []PresentationSupportingTable `json:"supportingTables,omitempty"`
	TableColumns     []PresentationTableColumn     `json:"tableColumns"`
	TableRows        []PresentationTableRow        `json:"tableRows"`
	Version          int64                         `json:"version"`
	Width            int64                         `json:"width"`
}

type MeasurementForecastStoryProvenance struct {
	GeneratedAt     time.Time `json:"generatedAt"`
	LatestEngine    string    `json:"latestEngine,omitempty"`
	LatestInputHash string    `json:"latestInputHash,omitempty"`
	Owner           string    `json:"owner"`
}

type MeasurementForecastStoryPercentiles struct {
	P10 *time.Time `json:"p10,omitempty"`
	P25 *time.Time `json:"p25,omitempty"`
	P50 *time.Time `json:"p50,omitempty"`
	P75 *time.Time `json:"p75,omitempty"`
	P90 *time.Time `json:"p90,omitempty"`
}

type MeasurementForecastStorySnapshot struct {
	AsOfAt            time.Time                           `json:"asOfAt"`
	GeneratedAt       time.Time                           `json:"generatedAt"`
	HorizonEndAt      *time.Time                          `json:"horizonEndAt,omitempty"`
	InputCount        int64                               `json:"inputCount"`
	InputHash         string                              `json:"inputHash,omitempty"`
	IntervalState     string                              `json:"intervalState"`
	Method            string                              `json:"method"`
	Movement          string                              `json:"movement"`
	ObservedValue     float64                             `json:"observedValue,omitempty"`
	OutputProbability float64                             `json:"outputProbability,omitempty"`
	Percentiles       MeasurementForecastStoryPercentiles `json:"percentiles"`
	RunUUID           string                              `json:"runUuid"`
	SampleCount       int64                               `json:"sampleCount"`
	Status            string                              `json:"status"`
	Warnings          []MeasurementForecastWarning        `json:"warnings,omitempty"`
}

type MeasurementForecastStoryTarget struct {
	Available bool    `json:"available"`
	Deadline  string  `json:"deadline,omitempty"`
	Label     string  `json:"label"`
	Value     float64 `json:"value"`
}

type MeasurementForecastStoryTrust struct {
	Evidence         string `json:"evidence"`
	ObservationCount int64  `json:"observationCount"`
	SnapshotCount    int64  `json:"snapshotCount"`
	State            string `json:"state"`
}

type MeasurementForecastStoryResponse struct {
	CurrentSnapshotUUID string                                `json:"currentSnapshotUuid,omitempty"`
	Delivery            MeasurementForecastStoryDelivery      `json:"delivery"`
	Observations        []MeasurementForecastStoryObservation `json:"observations"`
	Prediction          *MeasurementPredictionStory           `json:"prediction,omitempty"`
	Presentation        CompiledPresentation                  `json:"presentation"`
	ProjectKind         string                                `json:"projectKind"`
	ProjectName         string                                `json:"projectName"`
	ProjectUUID         string                                `json:"projectUuid"`
	Provenance          MeasurementForecastStoryProvenance    `json:"provenance"`
	Snapshots           []MeasurementForecastStorySnapshot    `json:"snapshots"`
	State               string                                `json:"state"`
	Target              MeasurementForecastStoryTarget        `json:"target"`
	Truncated           bool                                  `json:"truncated"`
	Trust               MeasurementForecastStoryTrust         `json:"trust"`
}

// Overview gets measurement overview.
func (c *MeasurementClient) Overview(ctx context.Context) (*MeasurementOverviewResponse, error) {
	var out MeasurementOverviewResponse
	err := c.request(ctx, http.MethodGet, "/overview", nil, &out)
	return &out, err
}

// GetForecastStory gets recorded forecast story.
func (c *MeasurementProjectClient) GetForecastStory(ctx context.Context, projectUUID string) (*MeasurementForecastStoryResponse, error) {
	var out MeasurementForecastStoryResponse
	err := c.measurement.request(ctx, http.MethodGet, "/projects/"+url.PathEscape(projectUUID)+"/forecast-story", nil, &out)
	return &out, err
}
