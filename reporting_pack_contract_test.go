package custd

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestReportingPackContractRoundTripsCanonicalDefinition(t *testing.T) {
	original := canonicalPackFixture()
	first, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal canonical pack: %v", err)
	}

	keys := packContractKeys(t, first)
	for _, required := range []string{
		"key", "displayName", "owner", "version", "enabled", "eventTypes",
		"metrics", "dimensions", "templates", "trust", "proof",
	} {
		if _, ok := keys[required]; !ok {
			t.Fatalf("required field %q missing from %s", required, first)
		}
	}

	var decoded Pack
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("unmarshal canonical pack: %v", err)
	}
	second, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("re-marshal canonical pack: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("round-trip changed the definition:\nfirst:  %s\nsecond: %s", first, second)
	}
	assertPackContractNestedFields(t, first)
}

func TestReportingPackContractOmitsUnsetOptionalFields(t *testing.T) {
	minimal := Pack{
		Key: "security", DisplayName: "Security", Owner: "platform", Version: 1, Enabled: true,
		EventTypes: []string{"login.success"},
		Metrics:    []PackMetric{}, Dimensions: []PackDimension{}, Templates: []PackTemplate{},
		Trust: TrustDiagnostics{SafeFields: []string{"coverage.status"}},
		Proof: ProofProfile{Key: "security_proof", OutputLayout: "summary"},
	}
	raw, err := json.Marshal(minimal)
	if err != nil {
		t.Fatalf("marshal minimal pack: %v", err)
	}
	keys := packContractKeys(t, raw)
	for _, optional := range []string{
		"flowRules", "rollupRules", "dashboards", "exportFormats", "redactionRules", "identity",
		"transitionRules", "activationRules", "retentionRules", "activityRules", "observation",
		"reconciliations", "assetHierarchy", "accounting",
	} {
		if _, ok := keys[optional]; ok {
			t.Fatalf("optional field %q should be omitted from %s", optional, raw)
		}
	}
}

func TestReportingPackContractRejectsInvalidClosedUnion(t *testing.T) {
	var metric PackMetric
	if err := json.Unmarshal([]byte(`{"calculation":"count"}`), &metric); err != nil {
		t.Fatalf("valid closed-union value rejected: %v", err)
	}
	if metric.Calculation != CalculationCount {
		t.Fatalf("calculation = %q", metric.Calculation)
	}
	if err := json.Unmarshal([]byte(`{"calculation":"not_a_calculation"}`), &metric); err == nil {
		t.Fatal("expected an invalid closed-union value to be rejected")
	}
}

func packContractKeys(t *testing.T, data []byte) map[string]json.RawMessage {
	t.Helper()
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatalf("decode object keys: %v", err)
	}
	return keys
}

// assertPackContractNestedFields pins the canonical field names that only appear
// below the top level, including the accounting rounding object and the embedded
// executive recipe selection.
func assertPackContractNestedFields(t *testing.T, raw []byte) {
	t.Helper()
	var payload struct {
		Metrics    []map[string]json.RawMessage `json:"metrics"`
		Templates  []map[string]json.RawMessage `json:"templates"`
		Dashboards []struct {
			Executive map[string]json.RawMessage `json:"executive"`
		} `json:"dashboards"`
		Accounting struct {
			Schedules []struct {
				Rounding map[string]json.RawMessage `json:"rounding"`
			} `json:"schedules"`
		} `json:"accounting"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode nested fields: %v", err)
	}
	assertPackContractFields(t, payload.Metrics[0], "calculation", "numeric", "maximumGapSeconds", "percentile", "seriesKey")
	assertPackContractFields(t, payload.Templates[0], "name", "aggregation", "allowedMetrics", "sourceModes")
	assertPackContractFields(t, payload.Dashboards[0].Executive, "recipe", "comparisonBasis")
	assertPackContractFields(t, payload.Accounting.Schedules[0].Rounding, "mode", "decimalPlaces")
}

func assertPackContractFields(t *testing.T, fields map[string]json.RawMessage, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			t.Fatalf("field %q missing from %v", name, fields)
		}
	}
}

func canonicalPackFixture() Pack {
	return Pack{
		Key: "managed_audit_security", DisplayName: "Managed Audit Security", Owner: "platform",
		Version: 1, Enabled: true,
		EventTypes: []string{"audit.event"},
		Metrics: []PackMetric{{
			MetricDefinition: MetricDefinition{Key: "audit_events", Label: "Events", Kind: MetricKindCount, Unit: "count", Description: "audit events"},
			Calculation:      CalculationCount,
			Match:            []MatchPredicate{{Selector: "payload.status", Exact: "ok", AnyOf: []string{"a", "b"}, NonEmpty: true, Pattern: "^a"}},
			FlowResult:       FlowResultSteps,
			Numeric:          &NumericSelector{Selector: "payload.duration", Type: NumericTypeNumber, Unit: NumericUnitMilliseconds},
			Distinct:         &SelectorDefinition{Selector: "payload.actor", Type: SelectorTypeString},
			MeasurementKind:  MeasurementKindGauge,
			PackMetricCounter: PackMetricCounter{
				MaximumGapSeconds: 300, ResetPolicy: CounterResetRollover, RolloverAt: 1.5, CanonicalUnit: UnitKilowatt,
			},
			PackMetricRatio:     PackMetricRatio{Percentile: 95, NumeratorMetric: "a", DenominatorMetric: "b"},
			PackMetricHierarchy: PackMetricHierarchy{SeriesKey: "series", HierarchyRootKey: "root"},
		}},
		Dimensions: []PackDimension{{
			DimensionDefinition: DimensionDefinition{
				Key: "audit_status", Label: "Status", Description: "status",
				LabelDefinitionUUID: "uuid", LabelDefinitionKey: "def", LabelValueUUIDs: []string{"v1"},
			},
			Selector: "payload.status", AllowedValuePattern: "^[a-z]+$", Fallback: "unknown", MaxCardinality: 100,
		}},
		Templates: []PackTemplate{{
			QueryTemplate: QueryTemplate{
				Name: "audit_overview", AllowedMetrics: []string{"audit_events"}, AllowedDimensions: []string{"audit_status"},
				AllowedFilters: []FilterDefinition{{Dimension: "audit_status", Operators: []FilterOperator{OpEq, OpIn}}},
				SourceModes:    []SourceMode{SourceModeAuto}, MaxRows: 250, DefaultRange: "7d",
				SubjectScope: &SubjectScopeRule{Required: true, Dimension: "audit_status"}, CompositionRules: []string{"rule"},
			},
			EventTypes: []string{"audit.event"}, Aggregation: AggregationCount, FlowRule: "audit_flow",
		}},
		FlowRules: []FlowRule{{
			Key: "audit_flow", CorrelationSelector: "payload.trace", FamilySelector: "payload.family",
			FamilyValue: "value", StepSelector: "payload.step", StartMarkers: []string{"start"},
			CompletionMarkers: []string{"done"}, IncludeUncorrelated: true,
		}},
		RollupRules: []RollupRule{{
			Family: "family", Templates: []string{"audit_overview"}, Metric: "audit_events",
			SourceEventType: "audit.event", Match: []MatchPredicate{{Selector: "payload.status"}},
			Dimensions: []RollupDimension{{
				Slot: 0, Dimension: "audit_status", Selector: "payload.status", AllowedValuePattern: "^[a-z]+$",
			}},
			CompletenessRequired: true,
		}},
		Dashboards: []DashboardDefinition{{
			Key: "audit", PackKey: "managed_audit_security", ExportFormats: []ExportFormat{FormatCSV},
			Title: "Audit", Description: "audit dashboard", Group: "security", DefaultRange: "7d",
			RefreshSeconds: 60, RequiredScopes: []string{ScopeReportingRead},
			Widgets: []WidgetDefinition{{
				Key: "overview", Title: "Overview", Kind: WidgetTimeSeries, Template: "audit_overview",
				Metrics: []string{"audit_events"}, Dimensions: []string{"audit_status"},
				Presentation: &PresentationPresetSelection{
					PresetID: "forecast_convergence_v1", Version: 1,
					NonColorEncoding: "line_symbol_label", SummaryTemplateID: "forecast_window_evidence",
				},
			}},
			Interactions: []DashboardInteractionRule{{
				Key: "drill", SourceWidget: "overview", TargetWidget: "overview", Dimension: "audit_status",
				Action: InteractionFilter, LinkedFilterGroup: "filters", ComparisonPreset: ComparisonPreviousPeriod,
				ResetBehavior: ResetInteraction,
			}},
			SharePolicy: &DashboardSharePolicy{
				RelativeRanges: []string{"7d"}, ComparisonPresets: []DashboardComparisonPreset{ComparisonPreviousPeriod},
			},
			Sections: []ReportSectionDefinition{{
				Key: "lead", Title: "Lead", Layout: SectionLayoutLead, ColumnSpan: 1, WidgetKeys: []string{"overview"},
				Narrative: []NarrativeBlock{{
					Text: "narrative", References: []NarrativeValueReference{{WidgetKey: "overview", Value: NarrativeValuePrimary}},
				}},
			}},
			Executive: &ExecutiveReportDefinition{
				ExecutiveRecipeSelection: ExecutiveRecipeSelection{
					Recipe: ExecutiveRecipeWeeklyPerformance, PrimaryWidgetKey: "overview",
					PrimaryMetricKey: "audit_events", PrimaryLabel: "Events", PrimaryUnit: "count", ScopeLabel: "all",
				},
				ComparisonBasis: ExecutiveComparisonPreviousPeriod, ComparisonLabel: "previous",
				TargetValue: packContractFloat64(10), TargetUnit: "count",
				Findings: []string{"finding"}, Actions: []string{"action"}, SupportingWidgetKeys: []string{"overview"},
			},
		}},
		ExportFormats: []ExportFormat{FormatCSV},
		Trust:         TrustDiagnostics{SafeFields: []string{"coverage.status"}, RedactionGuard: []string{"email"}},
		RedactionRules: []RedactionRule{{
			SourceSystem: "sys", SourceEventType: "audit.event", SelectorFields: []string{"payload.email"},
			AllowedFields: []string{"payload.status"}, ReasonClasses: []string{"pii"}, ResultShape: "summary",
		}},
		Proof: ProofProfile{
			Key: "audit_proof", Templates: []string{"audit_overview"}, SafeMetadataFields: []string{"coverage.status"},
			ForbiddenFields: []string{"email"}, OutputLayout: "summary",
		},
		Identity: &IdentitySelectors{
			Subject:     SelectorDefinition{Selector: "anonymous_id", Type: SelectorTypeString},
			Correlation: SelectorDefinition{Selector: "payload.trace", Type: SelectorTypeString},
		},
		TransitionRules: []TransitionRule{{
			Key: "transition", Identity: IdentityRoleSubject, StepSelector: "payload.step",
			Steps: []string{"a", "b"}, MaxDurationMs: 1000, MaxEvents: 10,
		}},
		ActivationRules: []ActivationRule{{
			Key: "activation", Identity: IdentityRoleSubject,
			Predicate: MatchPredicate{Selector: "payload.status", Exact: "first"}, CohortWindowDays: 7,
		}},
		RetentionRules: []RetentionRule{{
			Key: "retention", ActivationRule: "activation", ReturnWindowDays: []int{1, 7},
			ReturnPredicate: MatchPredicate{Selector: "payload.status"},
		}},
		ActivityRules: []ActivityRule{{Key: "activity", Identity: IdentityRoleSession, BucketDays: 1}},
		Observation: &ObservationContract{
			Version: 1, Sources: []ObservationSource{{Key: "source", Label: "Source"}},
			Assets: []ObservationAsset{{Key: "asset", Label: "Asset"}},
			Series: []SemanticSeries{{
				Key: "series", AssetKey: "asset", MeasurementKind: MeasurementKindGauge, CanonicalUnit: UnitKilowatt,
				AllowedSourceUnits: []Unit{UnitWatt, UnitKilowatt}, ExpectedCadenceSeconds: 60, MaximumGapSeconds: 300,
				Sources: []SourceSeries{{SourceKey: "source", SourceSeriesKey: "source_series", Priority: 1}},
			}},
		},
		Reconciliations: []SourceReconciliation{{
			Version: 1, TenantKey: "acme", SeriesKey: "series", CanonicalUnit: UnitKilowatt,
			Mode: SourceSelectionPriority, MaxAgeSeconds: 3600, AbsoluteTolerance: 0.5, RelativeTolerance: 0.1,
			Sources: []SourceSeries{{SourceKey: "source", SourceSeriesKey: "source_series", Priority: 1}},
		}},
		AssetHierarchy: &AssetHierarchy{
			Version: 1, TenantKey: "acme", MaxDepth: 3,
			Nodes: []AssetNode{{TenantKey: "acme", Key: "root", Label: "Root"}},
			Edges: []AssetHierarchyEdge{{TenantKey: "acme", ParentKey: "root", ChildKey: "child", Mode: AssetEdgeInclusive}},
		},
		Accounting: &AccountingContract{
			Version: 1, UsageMetric: "usage",
			Schedules: []SteppedRateSchedule{{
				Key: "schedule", UsageMetric: "usage", UsageUnit: UnitKilowattHour, Timezone: "UTC",
				BillingBoundary: BillingBoundaryCalendarMonth, EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Currency: "USD", Tiers: []RateTier{
					{LowerBound: 0, UpperBound: packContractFloat64(100), PricePerUnit: 1.5},
					{LowerBound: 100, PricePerUnit: 2},
				},
				Rounding: RoundingPolicy{Mode: RoundingHalfUp, DecimalPlaces: 2},
			}},
			Allocations: []AllocationRule{{
				Key: "allocation", SourceSeriesKey: "series", Targets: []AllocationTarget{{SeriesKey: "target", Weight: 1}},
				Rounding: RoundingPolicy{Mode: RoundingHalfEven, DecimalPlaces: 3},
			}},
		},
	}
}

func packContractFloat64(value float64) *float64 {
	return &value
}
