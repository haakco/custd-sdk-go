package custd

import (
	"encoding/json"
	"fmt"
	"time"
)

// This file mirrors the canonical reporting-pack contract owned by
// libs/go/reporting.Pack so a Go client can build, validate, preview and
// publish a valid definition. Every JSON field name, type and required/optional
// distinction is copied from the canonical source; the closed unions are named
// types whose decoder rejects a value the canonical contract does not define.

// CalculationType is the closed set of metric calculations.
type CalculationType string

const (
	CalculationCount         CalculationType = "count"
	CalculationFailureCount  CalculationType = "failure_count"
	CalculationFlow          CalculationType = "flow"
	CalculationDistinctCount CalculationType = "distinct_count"
	CalculationSum           CalculationType = "sum"
	CalculationAverage       CalculationType = "average"
	CalculationPercentile    CalculationType = "percentile"
	CalculationRatio         CalculationType = "ratio"

	CalculationLatest              CalculationType = "latest"
	CalculationMinimum             CalculationType = "minimum"
	CalculationMaximum             CalculationType = "maximum"
	CalculationFirstLastDelta      CalculationType = "firstLastDelta"
	CalculationTimeWeightedAverage CalculationType = "timeWeightedAverage"
	CalculationRate                CalculationType = "rate"
	CalculationIntegral            CalculationType = "integral"

	CalculationUsageCharge       CalculationType = "usage_charge"
	CalculationMarginalRate      CalculationType = "marginal_rate"
	CalculationProjection        CalculationType = "projection"
	CalculationPurchasedVsUsed   CalculationType = "purchased_vs_used"
	CalculationObservedBalance   CalculationType = "observed_balance"
	CalculationProjectedBalance  CalculationType = "projected_balance"
	CalculationAllocatedRecovery CalculationType = "allocated_recovery"
	CalculationCorrectionCount   CalculationType = "correction_count"
)

var calculationTypeValues = map[CalculationType]bool{
	CalculationCount: true, CalculationFailureCount: true, CalculationFlow: true, CalculationDistinctCount: true,
	CalculationSum: true, CalculationAverage: true, CalculationPercentile: true, CalculationRatio: true,
	CalculationLatest: true, CalculationMinimum: true, CalculationMaximum: true, CalculationFirstLastDelta: true,
	CalculationTimeWeightedAverage: true, CalculationRate: true, CalculationIntegral: true,
	CalculationUsageCharge: true, CalculationMarginalRate: true, CalculationProjection: true,
	CalculationPurchasedVsUsed: true, CalculationObservedBalance: true, CalculationProjectedBalance: true,
	CalculationAllocatedRecovery: true, CalculationCorrectionCount: true,
}

func (value *CalculationType) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, calculationTypeValues)
}

// AggregationMode is the closed set of template aggregation modes.
type AggregationMode string

const (
	AggregationCount         AggregationMode = "count"
	AggregationFlow          AggregationMode = "flow"
	AggregationDistinctCount AggregationMode = "distinct_count"
	AggregationSum           AggregationMode = "sum"
	AggregationAverage       AggregationMode = "average"
	AggregationPercentile    AggregationMode = "percentile"
	AggregationRatio         AggregationMode = "ratio"
)

var aggregationModeValues = map[AggregationMode]bool{
	AggregationCount: true, AggregationFlow: true, AggregationDistinctCount: true, AggregationSum: true,
	AggregationAverage: true, AggregationPercentile: true, AggregationRatio: true,
}

func (value *AggregationMode) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, aggregationModeValues)
}

// FlowResult is the closed set of flow computation outputs.
type FlowResult string

const (
	FlowResultSteps          FlowResult = "steps"
	FlowResultStarts         FlowResult = "starts"
	FlowResultCompletions    FlowResult = "completions"
	FlowResultDropOffs       FlowResult = "drop_offs"
	FlowResultCompletionRate FlowResult = "completion_rate"
	FlowResultUncorrelated   FlowResult = "uncorrelated"
)

var flowResultValues = map[FlowResult]bool{
	FlowResultSteps: true, FlowResultStarts: true, FlowResultCompletions: true,
	FlowResultDropOffs: true, FlowResultCompletionRate: true, FlowResultUncorrelated: true,
}

func (value *FlowResult) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, flowResultValues)
}

// SelectorType is the closed set of JSON value types a selector may read.
type SelectorType string

const (
	SelectorTypeString SelectorType = "string"
	NumericTypeInteger SelectorType = "integer"
	NumericTypeNumber  SelectorType = "number"
	NumericTypeString  SelectorType = "string"
)

var selectorTypeValues = map[SelectorType]bool{
	SelectorTypeString: true, NumericTypeInteger: true, NumericTypeNumber: true,
}

func (value *SelectorType) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, selectorTypeValues)
}

// MetricKind is the closed set of metric classifications.
type MetricKind string

const (
	MetricKindCount        MetricKind = "count"
	MetricKindUniqueCount  MetricKind = "unique_count"
	MetricKindRate         MetricKind = "rate"
	MetricKindLatency      MetricKind = "latency"
	MetricKindPercentile   MetricKind = "percentile"
	MetricKindFailureRatio MetricKind = "failure_ratio"
	MetricKindFreshnessAge MetricKind = "freshness_age"
	MetricKindLatestValue  MetricKind = "latest_value"

	MetricKindProbability        MetricKind = "probability"
	MetricKindQuantile           MetricKind = "quantile"
	MetricKindCredibleInterval   MetricKind = "credible_interval"
	MetricKindForecastDate       MetricKind = "forecast_date"
	MetricKindValueOfInformation MetricKind = "value_of_information"
)

var metricKindValues = map[MetricKind]bool{
	MetricKindCount: true, MetricKindUniqueCount: true, MetricKindRate: true, MetricKindLatency: true,
	MetricKindPercentile: true, MetricKindFailureRatio: true, MetricKindFreshnessAge: true, MetricKindLatestValue: true,
	MetricKindProbability: true, MetricKindQuantile: true, MetricKindCredibleInterval: true,
	MetricKindForecastDate: true, MetricKindValueOfInformation: true,
}

func (value *MetricKind) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, metricKindValues)
}

// MeasurementKind is the closed set of observation measurement kinds.
type MeasurementKind string

const (
	MeasurementKindGauge             MeasurementKind = "gauge"
	MeasurementKindMonotonicCounter  MeasurementKind = "monotonicCounter"
	MeasurementKindIntervalIncrement MeasurementKind = "intervalIncrement"
	MeasurementKindState             MeasurementKind = "state"
)

var measurementKindValues = map[MeasurementKind]bool{
	MeasurementKindGauge: true, MeasurementKindMonotonicCounter: true,
	MeasurementKindIntervalIncrement: true, MeasurementKindState: true,
}

func (value *MeasurementKind) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, measurementKindValues)
}

// Unit is the closed observation unit vocabulary.
type Unit string

const (
	UnitWatt         Unit = "W"
	UnitKilowatt     Unit = "kW"
	UnitWattHour     Unit = "Wh"
	UnitKilowattHour Unit = "kWh"
	UnitVolt         Unit = "V"
	UnitAmpere       Unit = "A"
	UnitLitre        Unit = "L"
	UnitCelsius      Unit = "Cel"
	UnitByte         Unit = "By"
	UnitNone         Unit = "1"
)

var unitValues = map[Unit]bool{
	UnitWatt: true, UnitKilowatt: true, UnitWattHour: true, UnitKilowattHour: true, UnitVolt: true,
	UnitAmpere: true, UnitLitre: true, UnitCelsius: true, UnitByte: true, UnitNone: true,
}

func (value *Unit) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, unitValues)
}

// NumericUnit is the closed output-unit vocabulary for numeric measures.
type NumericUnit string

const (
	NumericUnitCount           NumericUnit = "count"
	NumericUnitMilliseconds    NumericUnit = "milliseconds"
	NumericUnitBytes           NumericUnit = "bytes"
	NumericUnitPercent         NumericUnit = "percent"
	NumericUnitCurrency        NumericUnit = "currency"
	NumericUnitWatt            NumericUnit = "W"
	NumericUnitKilowatt        NumericUnit = "kW"
	NumericUnitWattHour        NumericUnit = "Wh"
	NumericUnitKilowattHour    NumericUnit = "kWh"
	NumericUnitVolt            NumericUnit = "V"
	NumericUnitAmpere          NumericUnit = "A"
	NumericUnitLitre           NumericUnit = "L"
	NumericUnitCelsius         NumericUnit = "Cel"
	NumericUnitLitresPerSecond NumericUnit = "L/s"
	NumericUnitCurrencyPerUnit NumericUnit = "currency_per_unit"
)

var numericUnitValues = map[NumericUnit]bool{
	NumericUnitCount: true, NumericUnitMilliseconds: true, NumericUnitBytes: true, NumericUnitPercent: true,
	NumericUnitCurrency: true, NumericUnitWatt: true, NumericUnitKilowatt: true, NumericUnitWattHour: true,
	NumericUnitKilowattHour: true, NumericUnitVolt: true, NumericUnitAmpere: true, NumericUnitLitre: true,
	NumericUnitCelsius: true, NumericUnitLitresPerSecond: true, NumericUnitCurrencyPerUnit: true,
}

func (value *NumericUnit) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, numericUnitValues)
}

// CounterResetPolicy is the closed set of monotonic-counter reset behaviors.
type CounterResetPolicy string

const (
	CounterResetAsIncrement CounterResetPolicy = "as_increment"
	CounterResetReject      CounterResetPolicy = "reject"
	CounterResetRollover    CounterResetPolicy = "rollover"
)

var counterResetPolicyValues = map[CounterResetPolicy]bool{
	CounterResetAsIncrement: true, CounterResetReject: true, CounterResetRollover: true,
}

func (value *CounterResetPolicy) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, counterResetPolicyValues)
}

// FilterOperator is the closed set of operators a filter may use.
type FilterOperator string

const (
	OpEq       FilterOperator = "eq"
	OpNeq      FilterOperator = "neq"
	OpIn       FilterOperator = "in"
	OpContains FilterOperator = "contains"
)

var filterOperatorValues = map[FilterOperator]bool{
	OpEq: true, OpNeq: true, OpIn: true, OpContains: true,
}

func (value *FilterOperator) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, filterOperatorValues)
}

// SourceMode is the closed set of storage backends a template may read.
type SourceMode string

const (
	SourceModeAuto     SourceMode = "auto"
	SourceModePostgres SourceMode = "postgres"
	SourceModeDuckDB   SourceMode = "duckdb"
)

var sourceModeValues = map[SourceMode]bool{
	SourceModeAuto: true, SourceModePostgres: true, SourceModeDuckDB: true,
}

func (value *SourceMode) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, sourceModeValues)
}

// WidgetKind is the closed set of widget render kinds.
type WidgetKind string

const (
	WidgetStatTile        WidgetKind = "stat_tile"
	WidgetTimeSeries      WidgetKind = "time_series"
	WidgetBreakdownTable  WidgetKind = "breakdown_table"
	WidgetTopList         WidgetKind = "top_list"
	WidgetHealthList      WidgetKind = "health_list"
	WidgetFreshnessPanel  WidgetKind = "freshness_panel"
	WidgetExportTable     WidgetKind = "export_table"
	WidgetLifecycleFunnel WidgetKind = "lifecycle_funnel"
	WidgetCohortRetention WidgetKind = "cohort_retention"

	WidgetForecastFan           WidgetKind = "forecast_fan"
	WidgetUncertaintyBand       WidgetKind = "uncertainty_band"
	WidgetProbabilityTile       WidgetKind = "probability_tile"
	WidgetDistributionHistogram WidgetKind = "distribution_histogram"
	WidgetVOIRankedList         WidgetKind = "voi_ranked_list"
	WidgetCalibrationPanel      WidgetKind = "calibration_panel"
	WidgetRunwayTile            WidgetKind = "runway_tile"
	WidgetAnomalyTimeline       WidgetKind = "anomaly_timeline"
	WidgetDriverBreakdown       WidgetKind = "driver_breakdown"
)

var widgetKindValues = map[WidgetKind]bool{
	WidgetStatTile: true, WidgetTimeSeries: true, WidgetBreakdownTable: true, WidgetTopList: true,
	WidgetHealthList: true, WidgetFreshnessPanel: true, WidgetExportTable: true, WidgetLifecycleFunnel: true,
	WidgetCohortRetention: true, WidgetForecastFan: true, WidgetUncertaintyBand: true, WidgetProbabilityTile: true,
	WidgetDistributionHistogram: true, WidgetVOIRankedList: true, WidgetCalibrationPanel: true,
	WidgetRunwayTile: true, WidgetAnomalyTimeline: true, WidgetDriverBreakdown: true,
}

func (value *WidgetKind) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, widgetKindValues)
}

// ExportFormat is the closed set of report artifact formats.
type ExportFormat string

const (
	FormatCSV      ExportFormat = "csv"
	FormatXLSX     ExportFormat = "xlsx"
	FormatHTML     ExportFormat = "html"
	FormatPDF      ExportFormat = "pdf"
	FormatSVG      ExportFormat = "svg"
	FormatPNG      ExportFormat = "png"
	FormatParquet  ExportFormat = "parquet"
	FormatManifest ExportFormat = "manifest"

	ScopeReportingRead           = "reporting:read"
	ScopeReportingQuery          = "reporting:query"
	ScopeReportingSubjectInsight = "reporting:subject-insight"
	ScopeReportingExport         = "reporting:export"
)

var exportFormatValues = map[ExportFormat]bool{
	FormatCSV: true, FormatXLSX: true, FormatHTML: true, FormatPDF: true,
	FormatSVG: true, FormatPNG: true, FormatParquet: true, FormatManifest: true,
}

func (value *ExportFormat) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, exportFormatValues)
}

// DashboardInteraction is the closed set of dashboard interactions.
type DashboardInteraction string

const (
	InteractionFilter    DashboardInteraction = "filter"
	InteractionDrillDown DashboardInteraction = "drill_down"
)

var dashboardInteractionValues = map[DashboardInteraction]bool{
	InteractionFilter: true, InteractionDrillDown: true,
}

func (value *DashboardInteraction) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, dashboardInteractionValues)
}

// DashboardComparisonPreset is the closed set of server-computed comparisons.
type DashboardComparisonPreset string

const (
	ComparisonPreviousPeriod DashboardComparisonPreset = "previous_period"
	ComparisonPreviousYear   DashboardComparisonPreset = "previous_year"
)

var dashboardComparisonPresetValues = map[DashboardComparisonPreset]bool{
	ComparisonPreviousPeriod: true, ComparisonPreviousYear: true,
}

func (value *DashboardComparisonPreset) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, dashboardComparisonPresetValues)
}

// InteractionResetBehavior is the closed set of selection reset behaviors.
type InteractionResetBehavior string

const (
	ResetInteraction InteractionResetBehavior = "interaction"
	ResetDashboard   InteractionResetBehavior = "dashboard"
)

var interactionResetBehaviorValues = map[InteractionResetBehavior]bool{
	ResetInteraction: true, ResetDashboard: true,
}

func (value *InteractionResetBehavior) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, interactionResetBehaviorValues)
}

// SectionLayoutPreset is the closed set of responsive section layouts.
type SectionLayoutPreset string

const (
	SectionLayoutLead   SectionLayoutPreset = "lead"
	SectionLayoutGrid   SectionLayoutPreset = "grid"
	SectionLayoutDetail SectionLayoutPreset = "detail"
)

var sectionLayoutPresetValues = map[SectionLayoutPreset]bool{
	SectionLayoutLead: true, SectionLayoutGrid: true, SectionLayoutDetail: true,
}

func (value *SectionLayoutPreset) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, sectionLayoutPresetValues)
}

// NarrativeValue is the closed set of canonical values available to narrative.
type NarrativeValue string

const (
	NarrativeValuePrimary NarrativeValue = "primary"
	NarrativeValueDelta   NarrativeValue = "delta"
)

var narrativeValueValues = map[NarrativeValue]bool{
	NarrativeValuePrimary: true, NarrativeValueDelta: true,
}

func (value *NarrativeValue) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, narrativeValueValues)
}

// IdentityRole selects one pseudonymous identity declared by the pack.
type IdentityRole string

const (
	IdentityRoleSubject     IdentityRole = "subject"
	IdentityRoleSession     IdentityRole = "session"
	IdentityRoleCorrelation IdentityRole = "correlation"
)

var identityRoleValues = map[IdentityRole]bool{
	IdentityRoleSubject: true, IdentityRoleSession: true, IdentityRoleCorrelation: true,
}

func (value *IdentityRole) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, identityRoleValues)
}

// AssetEdgeMode is the closed set of asset-hierarchy edge modes.
type AssetEdgeMode string

const (
	AssetEdgeInclusive     AssetEdgeMode = "inclusive"
	AssetEdgeExclusive     AssetEdgeMode = "exclusive"
	AssetEdgeInformational AssetEdgeMode = "informational"
)

var assetEdgeModeValues = map[AssetEdgeMode]bool{
	AssetEdgeInclusive: true, AssetEdgeExclusive: true, AssetEdgeInformational: true,
}

func (value *AssetEdgeMode) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, assetEdgeModeValues)
}

// SourceSelectionMode is the closed set of reconciliation selection modes.
type SourceSelectionMode string

const (
	SourceSelectionPriority     SourceSelectionMode = "priority"
	SourceSelectionFreshness    SourceSelectionMode = "freshness"
	SourceSelectionAvailability SourceSelectionMode = "availability"
	SourceSelectionAgreement    SourceSelectionMode = "agreement"
)

var sourceSelectionModeValues = map[SourceSelectionMode]bool{
	SourceSelectionPriority: true, SourceSelectionFreshness: true,
	SourceSelectionAvailability: true, SourceSelectionAgreement: true,
}

func (value *SourceSelectionMode) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, sourceSelectionModeValues)
}

// BillingBoundary is the closed set of accounting billing boundaries.
type BillingBoundary string

const (
	BillingBoundaryCalendarMonth BillingBoundary = "calendar_month"
	BillingBoundaryCalendarDay   BillingBoundary = "calendar_day"
)

var billingBoundaryValues = map[BillingBoundary]bool{
	BillingBoundaryCalendarMonth: true, BillingBoundaryCalendarDay: true,
}

func (value *BillingBoundary) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, billingBoundaryValues)
}

// RoundingMode is the closed set of accounting rounding modes.
type RoundingMode string

const (
	RoundingHalfUp   RoundingMode = "half_up"
	RoundingHalfEven RoundingMode = "half_even"
	RoundingDown     RoundingMode = "down"
	RoundingUp       RoundingMode = "up"
)

var roundingModeValues = map[RoundingMode]bool{
	RoundingHalfUp: true, RoundingHalfEven: true, RoundingDown: true, RoundingUp: true,
}

func (value *RoundingMode) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, roundingModeValues)
}

// ExecutiveRecipe is the closed set of executive report recipes.
type ExecutiveRecipe string

const (
	ExecutiveRecipeBlank             ExecutiveRecipe = "blank"
	ExecutiveRecipeWeeklyPerformance ExecutiveRecipe = "weekly-performance"
	ExecutiveRecipeOperationsHealth  ExecutiveRecipe = "operations-health"
	ExecutiveRecipeForecastDecision  ExecutiveRecipe = "forecast-decision"
)

var executiveRecipeValues = map[ExecutiveRecipe]bool{
	ExecutiveRecipeBlank: true, ExecutiveRecipeWeeklyPerformance: true,
	ExecutiveRecipeOperationsHealth: true, ExecutiveRecipeForecastDecision: true,
}

func (value *ExecutiveRecipe) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, executiveRecipeValues)
}

// ExecutiveComparisonBasis is the closed set of executive comparison bases.
type ExecutiveComparisonBasis string

const (
	ExecutiveComparisonNone           ExecutiveComparisonBasis = "none"
	ExecutiveComparisonPreviousPeriod ExecutiveComparisonBasis = "previous-period"
	ExecutiveComparisonTarget         ExecutiveComparisonBasis = "target"
)

var executiveComparisonBasisValues = map[ExecutiveComparisonBasis]bool{
	ExecutiveComparisonNone: true, ExecutiveComparisonPreviousPeriod: true, ExecutiveComparisonTarget: true,
}

func (value *ExecutiveComparisonBasis) UnmarshalJSON(raw []byte) error {
	return unmarshalClosedUnion(raw, value, executiveComparisonBasisValues)
}

// unmarshalClosedUnion decodes a closed-union value and rejects any member the
// canonical contract does not define. An empty string is the zero value and is
// accepted so an unfilled definition can still round-trip through JSON.
func unmarshalClosedUnion[T ~string](raw []byte, value *T, allowed map[T]bool) error {
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return err
	}
	if text != "" && !allowed[T(text)] {
		return fmt.Errorf("custd: %q is not a valid closed-union value", text)
	}
	*value = T(text)
	return nil
}

// MetricDefinition describes a server-owned metric a widget may request.
type MetricDefinition struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Kind        MetricKind `json:"kind"`
	Unit        string     `json:"unit,omitempty"`
	Description string     `json:"description,omitempty"`
}

// DimensionDefinition describes a group-by / filter dimension.
type DimensionDefinition struct {
	Key                 string   `json:"key"`
	Label               string   `json:"label"`
	Description         string   `json:"description,omitempty"`
	LabelDefinitionUUID string   `json:"labelDefinitionUuid,omitempty"`
	LabelDefinitionKey  string   `json:"labelDefinitionKey,omitempty"`
	LabelValueUUIDs     []string `json:"labelValueUuids,omitempty"`
}

// FilterDefinition declares a dimension and the operators permitted against it.
type FilterDefinition struct {
	Dimension string           `json:"dimension"`
	Operators []FilterOperator `json:"operators"`
}

// QueryTemplate is a server-owned, named query shape.
type QueryTemplate struct {
	Name              string             `json:"name"`
	AllowedMetrics    []string           `json:"allowedMetrics"`
	AllowedDimensions []string           `json:"allowedDimensions,omitempty"`
	AllowedFilters    []FilterDefinition `json:"allowedFilters,omitempty"`
	SourceModes       []SourceMode       `json:"sourceModes"`
	MaxRows           int                `json:"maxRows"`
	DefaultRange      string             `json:"defaultRange,omitempty"`
	SubjectScope      *SubjectScopeRule  `json:"subjectScope,omitempty"`
	CompositionRules  []string           `json:"compositionRules,omitempty"`
}

// SubjectScopeRule marks a template for the trusted subject-insight endpoint.
type SubjectScopeRule struct {
	Required  bool   `json:"required"`
	Dimension string `json:"dimension"`
}

// SelectorDefinition declares the JSON path and value type a selector reads.
type SelectorDefinition struct {
	Selector string       `json:"selector"`
	Type     SelectorType `json:"type"`
}

// NumericSelector binds a numeric measure to a selector and output unit.
type NumericSelector struct {
	Selector string       `json:"selector"`
	Type     SelectorType `json:"type"`
	Unit     NumericUnit  `json:"unit"`
}

// MatchPredicate is one bounded selector predicate.
type MatchPredicate struct {
	Selector string   `json:"selector"`
	Exact    string   `json:"exact,omitempty"`
	AnyOf    []string `json:"anyOf,omitempty"`
	NonEmpty bool     `json:"nonEmpty,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
}

// PackMetricCounter groups the bounded counter measurement knobs.
type PackMetricCounter struct {
	MaximumGapSeconds int                `json:"maximumGapSeconds,omitempty"`
	ResetPolicy       CounterResetPolicy `json:"resetPolicy,omitempty"`
	RolloverAt        float64            `json:"rolloverAt,omitempty"`
	CanonicalUnit     Unit               `json:"canonicalUnit,omitempty"`
}

// PackMetricRatio groups the ratio-input bindings plus the percentile knob.
type PackMetricRatio struct {
	Percentile        int    `json:"percentile,omitempty"`
	NumeratorMetric   string `json:"numeratorMetric,omitempty"`
	DenominatorMetric string `json:"denominatorMetric,omitempty"`
}

// PackMetricHierarchy groups the hierarchy-binding fields.
type PackMetricHierarchy struct {
	SeriesKey        string `json:"seriesKey,omitempty"`
	HierarchyRootKey string `json:"hierarchyRootKey,omitempty"`
}

// PackMetric is the canonical pack metric (MetricDefinition plus its calculation).
type PackMetric struct {
	MetricDefinition
	Calculation     CalculationType     `json:"calculation"`
	Match           []MatchPredicate    `json:"match,omitempty"`
	FlowResult      FlowResult          `json:"flowResult,omitempty"`
	Numeric         *NumericSelector    `json:"numeric,omitempty"`
	Distinct        *SelectorDefinition `json:"distinct,omitempty"`
	MeasurementKind MeasurementKind     `json:"measurementKind,omitempty"`
	PackMetricCounter
	PackMetricRatio
	PackMetricHierarchy
}

// PackDimension is the canonical pack dimension (DimensionDefinition plus its selector).
type PackDimension struct {
	DimensionDefinition
	Selector            string `json:"selector"`
	AllowedValuePattern string `json:"allowedValuePattern,omitempty"`
	Fallback            string `json:"fallback,omitempty"`
	MaxCardinality      int    `json:"maxCardinality,omitempty"`
}

// PackTemplate is the canonical query template filled with its flow rule.
type PackTemplate struct {
	QueryTemplate
	EventTypes  []string        `json:"eventTypes"`
	Aggregation AggregationMode `json:"aggregation"`
	FlowRule    string          `json:"flowRule,omitempty"`
}

// FlowRule declares an unordered funnel correlation rule.
type FlowRule struct {
	Key                 string   `json:"key"`
	CorrelationSelector string   `json:"correlationSelector"`
	FamilySelector      string   `json:"familySelector"`
	FamilyValue         string   `json:"familyValue,omitempty"`
	StepSelector        string   `json:"stepSelector"`
	StartMarkers        []string `json:"startMarkers"`
	CompletionMarkers   []string `json:"completionMarkers"`
	IncludeUncorrelated bool     `json:"includeUncorrelated"`
}

// RollupDimension is one slot a rollup rule derives.
type RollupDimension struct {
	Slot                int    `json:"slot"`
	Dimension           string `json:"dimension"`
	Selector            string `json:"selector"`
	AllowedValuePattern string `json:"allowedValuePattern,omitempty"`
}

// RollupRule maps source templates to a derived rollup metric.
type RollupRule struct {
	Family               string            `json:"family"`
	Templates            []string          `json:"templates,omitempty"`
	Metric               string            `json:"metric"`
	SourceEventType      string            `json:"sourceEventType"`
	Match                []MatchPredicate  `json:"match,omitempty"`
	Dimensions           []RollupDimension `json:"dimensions,omitempty"`
	CompletenessRequired bool              `json:"completenessRequired"`
}

// TrustDiagnostics lists the safe output fields and the redaction guard.
type TrustDiagnostics struct {
	SafeFields     []string `json:"safeFields"`
	RedactionGuard []string `json:"redactionGuard"`
}

// RedactionRule declares one redaction bound to a source event shape.
type RedactionRule struct {
	SourceSystem    string   `json:"sourceSystem"`
	SourceEventType string   `json:"sourceEventType"`
	SelectorFields  []string `json:"selectorFields"`
	AllowedFields   []string `json:"allowedFields"`
	ReasonClasses   []string `json:"reasonClasses"`
	ResultShape     string   `json:"resultShape"`
}

// ProofProfile names the proof template and its safe/forbidden fields.
type ProofProfile struct {
	Key                string   `json:"key"`
	Templates          []string `json:"templates"`
	SafeMetadataFields []string `json:"safeMetadataFields"`
	ForbiddenFields    []string `json:"forbiddenFields"`
	OutputLayout       string   `json:"outputLayout"`
}

// IdentitySelectors names the optional identity selectors a pack declares.
type IdentitySelectors struct {
	Subject     SelectorDefinition `json:"subject,omitempty"`
	Session     SelectorDefinition `json:"session,omitempty"`
	Entity      SelectorDefinition `json:"entity,omitempty"`
	Cohort      SelectorDefinition `json:"cohort,omitempty"`
	Correlation SelectorDefinition `json:"correlation,omitempty"`
}

// TransitionRule declares an ordered, bounded identity sequence.
type TransitionRule struct {
	Key           string       `json:"key"`
	Identity      IdentityRole `json:"identity"`
	StepSelector  string       `json:"stepSelector"`
	Steps         []string     `json:"steps"`
	MaxDurationMs int64        `json:"maxDurationMs"`
	MaxEvents     int          `json:"maxEvents"`
}

// ActivationRule declares a bounded first qualifying event for an identity.
type ActivationRule struct {
	Key              string         `json:"key"`
	Identity         IdentityRole   `json:"identity"`
	Predicate        MatchPredicate `json:"predicate"`
	CohortWindowDays int            `json:"cohortWindowDays"`
}

// RetentionRule declares return windows from a referenced activation rule.
type RetentionRule struct {
	Key              string         `json:"key"`
	ActivationRule   string         `json:"activationRule"`
	ReturnWindowDays []int          `json:"returnWindowDays"`
	ReturnPredicate  MatchPredicate `json:"returnPredicate"`
}

// ActivityRule declares a bounded UTC activity bucket for a pack identity.
type ActivityRule struct {
	Key        string       `json:"key"`
	Identity   IdentityRole `json:"identity"`
	BucketDays int          `json:"bucketDays"`
}

// DashboardInteractionRule connects two declared widgets through one dimension.
type DashboardInteractionRule struct {
	Key               string                    `json:"key"`
	SourceWidget      string                    `json:"sourceWidget"`
	TargetWidget      string                    `json:"targetWidget"`
	Dimension         string                    `json:"dimension"`
	Action            DashboardInteraction      `json:"action"`
	LinkedFilterGroup string                    `json:"linkedFilterGroup"`
	ComparisonPreset  DashboardComparisonPreset `json:"comparisonPreset,omitempty"`
	ResetBehavior     InteractionResetBehavior  `json:"resetBehavior"`
}

// DashboardSharePolicy declares the bounded state a share URL may encode.
type DashboardSharePolicy struct {
	RelativeRanges    []string                    `json:"relativeRanges,omitempty"`
	ComparisonPresets []DashboardComparisonPreset `json:"comparisonPresets,omitempty"`
}

// NarrativeValueReference identifies one rendered value without an expression.
type NarrativeValueReference struct {
	WidgetKey string         `json:"widgetKey"`
	Value     NarrativeValue `json:"value"`
}

// NarrativeBlock is bounded plain text with references to canonical values.
type NarrativeBlock struct {
	Text       string                    `json:"text"`
	References []NarrativeValueReference `json:"references,omitempty"`
}

// ReportSectionDefinition is one ordered, bounded report composition block.
type ReportSectionDefinition struct {
	Key        string              `json:"key"`
	Title      string              `json:"title"`
	Layout     SectionLayoutPreset `json:"layout"`
	ColumnSpan int                 `json:"columnSpan"`
	WidgetKeys []string            `json:"widgetKeys,omitempty"`
	Narrative  []NarrativeBlock    `json:"narrative,omitempty"`
}

// PresentationPresetSelection is the only visualization choice a pack may persist.
type PresentationPresetSelection struct {
	PresetID          string `json:"presetId"`
	Version           int    `json:"version"`
	NonColorEncoding  string `json:"nonColorEncoding"`
	SummaryTemplateID string `json:"summaryTemplateId"`
}

// WidgetDefinition binds a template to a render kind and a metric subset.
type WidgetDefinition struct {
	Key          string                       `json:"key"`
	Title        string                       `json:"title"`
	Kind         WidgetKind                   `json:"kind"`
	Template     string                       `json:"template"`
	Metrics      []string                     `json:"metrics"`
	Dimensions   []string                     `json:"dimensions,omitempty"`
	Presentation *PresentationPresetSelection `json:"presentation,omitempty"`
}

// ExecutiveRecipeSelection groups the recipe choice and its primary selections.
type ExecutiveRecipeSelection struct {
	Recipe           ExecutiveRecipe `json:"recipe"`
	PrimaryWidgetKey string          `json:"primaryWidgetKey"`
	PrimaryMetricKey string          `json:"primaryMetricKey"`
	PrimaryLabel     string          `json:"primaryLabel"`
	PrimaryUnit      string          `json:"primaryUnit"`
	ScopeLabel       string          `json:"scopeLabel"`
}

// ExecutiveReportDefinition stores the tenant-owned meaning of a dashboard.
type ExecutiveReportDefinition struct {
	ExecutiveRecipeSelection
	ComparisonBasis      ExecutiveComparisonBasis `json:"comparisonBasis"`
	ComparisonLabel      string                   `json:"comparisonLabel,omitempty"`
	TargetValue          *float64                 `json:"targetValue,omitempty"`
	TargetUnit           string                   `json:"targetUnit,omitempty"`
	Findings             []string                 `json:"findings,omitempty"`
	Actions              []string                 `json:"actions,omitempty"`
	SupportingWidgetKeys []string                 `json:"supportingWidgetKeys,omitempty"`
}

// DashboardDefinition is an ordered set of widgets with refresh metadata.
type DashboardDefinition struct {
	Key            string                     `json:"key"`
	PackKey        string                     `json:"packKey,omitempty"`
	ExportFormats  []ExportFormat             `json:"exportFormats,omitempty"`
	Title          string                     `json:"title"`
	Description    string                     `json:"description,omitempty"`
	Group          string                     `json:"group,omitempty"`
	DefaultRange   string                     `json:"defaultRange"`
	RefreshSeconds int                        `json:"refreshSeconds"`
	RequiredScopes []string                   `json:"requiredScopes"`
	Widgets        []WidgetDefinition         `json:"widgets"`
	Interactions   []DashboardInteractionRule `json:"interactions,omitempty"`
	SharePolicy    *DashboardSharePolicy      `json:"sharePolicy,omitempty"`
	Sections       []ReportSectionDefinition  `json:"sections,omitempty"`
	Executive      *ExecutiveReportDefinition `json:"executive,omitempty"`
}

// ObservationSource is one configured observation source.
type ObservationSource struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ObservationAsset is one configured logical asset.
type ObservationAsset struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// SourceSeries identifies one source-series within a source.
type SourceSeries struct {
	SourceKey       string `json:"sourceKey"`
	SourceSeriesKey string `json:"sourceSeriesKey"`
	Priority        int    `json:"priority"`
}

// SemanticSeries maps configured source series to one logical series.
type SemanticSeries struct {
	Key                    string          `json:"key"`
	AssetKey               string          `json:"assetKey,omitempty"`
	AssetKeys              []string        `json:"assetKeys,omitempty"`
	MeasurementKind        MeasurementKind `json:"measurementKind"`
	CanonicalUnit          Unit            `json:"canonicalUnit"`
	AllowedSourceUnits     []Unit          `json:"allowedSourceUnits"`
	ExpectedCadenceSeconds int             `json:"expectedCadenceSeconds"`
	MaximumGapSeconds      int             `json:"maximumGapSeconds"`
	Sources                []SourceSeries  `json:"sources"`
}

// ObservationContract is the bounded observation vocabulary a pack declares.
type ObservationContract struct {
	Version int                 `json:"version"`
	Sources []ObservationSource `json:"sources"`
	Assets  []ObservationAsset  `json:"assets"`
	Series  []SemanticSeries    `json:"series"`
}

// SourceReconciliation maps configured source series to one logical series.
type SourceReconciliation struct {
	Version           int                 `json:"version"`
	TenantKey         string              `json:"tenantKey"`
	SeriesKey         string              `json:"seriesKey"`
	CanonicalUnit     Unit                `json:"canonicalUnit,omitempty"`
	Mode              SourceSelectionMode `json:"mode"`
	MaxAgeSeconds     int                 `json:"maxAgeSeconds"`
	AbsoluteTolerance float64             `json:"absoluteTolerance,omitempty"`
	RelativeTolerance float64             `json:"relativeTolerance,omitempty"`
	Sources           []SourceSeries      `json:"sources"`
}

// AssetNode is one node in a tenant-scoped asset hierarchy.
type AssetNode struct {
	TenantKey string `json:"tenantKey"`
	Key       string `json:"key"`
	Label     string `json:"label"`
}

// AssetHierarchyEdge is one directed edge in a tenant-scoped asset hierarchy.
type AssetHierarchyEdge struct {
	TenantKey string        `json:"tenantKey"`
	ParentKey string        `json:"parentKey"`
	ChildKey  string        `json:"childKey"`
	Mode      AssetEdgeMode `json:"mode"`
}

// AssetHierarchy is a bounded, tenant-scoped DAG of logical assets.
type AssetHierarchy struct {
	Version   int                  `json:"version"`
	TenantKey string               `json:"tenantKey"`
	MaxDepth  int                  `json:"maxDepth,omitempty"`
	Nodes     []AssetNode          `json:"nodes"`
	Edges     []AssetHierarchyEdge `json:"edges,omitempty"`
}

// RateTier is one contiguous tier in a stepped rate schedule.
type RateTier struct {
	LowerBound   float64  `json:"lowerBound"`
	UpperBound   *float64 `json:"upperBound,omitempty"`
	PricePerUnit float64  `json:"pricePerUnit"`
}

// RoundingPolicy is the closed rounding contract for an accounting rule.
type RoundingPolicy struct {
	Mode          RoundingMode `json:"mode"`
	DecimalPlaces int          `json:"decimalPlaces"`
}

// SteppedRateSchedule is one bounded, contiguous usage rate schedule.
type SteppedRateSchedule struct {
	Key             string          `json:"key"`
	UsageMetric     string          `json:"usageMetric"`
	UsageUnit       Unit            `json:"usageUnit"`
	Timezone        string          `json:"timezone"`
	BillingBoundary BillingBoundary `json:"billingBoundary"`
	EffectiveFrom   time.Time       `json:"effectiveFrom"`
	EffectiveTo     *time.Time      `json:"effectiveTo,omitempty"`
	Currency        string          `json:"currency"`
	Tiers           []RateTier      `json:"tiers"`
	Rounding        RoundingPolicy  `json:"rounding"`
}

// AllocationTarget is one weighted target of an allocation rule.
type AllocationTarget struct {
	SeriesKey string  `json:"seriesKey"`
	Weight    float64 `json:"weight"`
}

// AllocationRule distributes a source series across weighted targets.
type AllocationRule struct {
	Key             string             `json:"key"`
	SourceSeriesKey string             `json:"sourceSeriesKey"`
	Targets         []AllocationTarget `json:"targets"`
	Rounding        RoundingPolicy     `json:"rounding"`
}

// AccountingContract is the bounded accounting vocabulary a pack declares.
type AccountingContract struct {
	Version     int                   `json:"version"`
	UsageMetric string                `json:"usageMetric"`
	Schedules   []SteppedRateSchedule `json:"schedules"`
	Allocations []AllocationRule      `json:"allocations,omitempty"`
}

// Pack is the canonical reporting-pack definition (libs/go/reporting.Pack). Every
// contracted field is carried so a definition round-trips losslessly through
// create/validate/preview/publish; unknown extension fields are the only thing a
// decode may drop.
type Pack struct {
	Key            string                `json:"key"`
	DisplayName    string                `json:"displayName"`
	Owner          string                `json:"owner"`
	Version        int                   `json:"version"`
	Enabled        bool                  `json:"enabled"`
	EventTypes     []string              `json:"eventTypes"`
	Metrics        []PackMetric          `json:"metrics"`
	Dimensions     []PackDimension       `json:"dimensions"`
	Templates      []PackTemplate        `json:"templates"`
	FlowRules      []FlowRule            `json:"flowRules,omitempty"`
	RollupRules    []RollupRule          `json:"rollupRules,omitempty"`
	Dashboards     []DashboardDefinition `json:"dashboards,omitempty"`
	ExportFormats  []ExportFormat        `json:"exportFormats,omitempty"`
	Trust          TrustDiagnostics      `json:"trust"`
	RedactionRules []RedactionRule       `json:"redactionRules,omitempty"`
	Proof          ProofProfile          `json:"proof"`
	// Identity is a pointer so an unset identity is genuinely omitted from a
	// request. Canonical declares it as a value with `omitempty`, which is a
	// no-op on a struct, so canonical always emits `"identity": {}`. Both forms
	// decode to the same zero value, so no canonical value is dropped.
	Identity        *IdentitySelectors     `json:"identity,omitempty"`
	TransitionRules []TransitionRule       `json:"transitionRules,omitempty"`
	ActivationRules []ActivationRule       `json:"activationRules,omitempty"`
	RetentionRules  []RetentionRule        `json:"retentionRules,omitempty"`
	ActivityRules   []ActivityRule         `json:"activityRules,omitempty"`
	Observation     *ObservationContract   `json:"observation,omitempty"`
	Reconciliations []SourceReconciliation `json:"reconciliations,omitempty"`
	AssetHierarchy  *AssetHierarchy        `json:"assetHierarchy,omitempty"`
	Accounting      *AccountingContract    `json:"accounting,omitempty"`
}
