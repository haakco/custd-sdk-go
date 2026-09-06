package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type TimePlanAdminClient struct{ admin *AdminClient }

type TimePlanAnnotationType string

const (
	TimePlanAnnotationNote     TimePlanAnnotationType = "note"
	TimePlanAnnotationMarker   TimePlanAnnotationType = "marker"
	TimePlanAnnotationDecision TimePlanAnnotationType = "decision"
	TimePlanAnnotationAction   TimePlanAnnotationType = "action"
)

type TimePlanAnnotationField string

const (
	TimePlanAnnotationText           TimePlanAnnotationField = "text"
	TimePlanAnnotationMarkerLabel    TimePlanAnnotationField = "markerLabel"
	TimePlanAnnotationDecisionStatus TimePlanAnnotationField = "decisionStatus"
	TimePlanAnnotationAssigneeRef    TimePlanAnnotationField = "assigneeRef"
	TimePlanAnnotationDueDate        TimePlanAnnotationField = "dueDate"
	TimePlanAnnotationActionStatus   TimePlanAnnotationField = "actionStatus"
)

type TimePlanThresholdCueSeverity string

const (
	TimePlanThresholdInfo     TimePlanThresholdCueSeverity = "info"
	TimePlanThresholdWarning  TimePlanThresholdCueSeverity = "warning"
	TimePlanThresholdCritical TimePlanThresholdCueSeverity = "critical"
)

type TimePlanDefinition struct {
	HorizonMS          int64                     `json:"horizonMs"`
	DefaultStartsAt    string                    `json:"defaultStartsAt,omitempty"`
	DefaultEndsAt      string                    `json:"defaultEndsAt,omitempty"`
	RedistributionMode string                    `json:"redistributionMode,omitempty"`
	AutoAdvance        bool                      `json:"autoAdvance,omitempty"`
	AnnotationSchema   *TimePlanAnnotationSchema `json:"annotationSchema,omitempty"`
	ThresholdCues      []TimePlanThresholdCue    `json:"thresholdCues,omitempty"`
	Blocks             []TimePlanBlock           `json:"blocks"`
}

type TimePlanAnnotationSchema struct {
	AllowedTypes []TimePlanAnnotationType  `json:"allowedTypes,omitempty"`
	Fields       []TimePlanAnnotationField `json:"fields,omitempty"`
}

type TimePlanThresholdCue struct {
	RemainingFractionBPS *int64                       `json:"remainingFractionBps,omitempty"`
	RemainingMS          *int64                       `json:"remainingMs,omitempty"`
	Severity             TimePlanThresholdCueSeverity `json:"severity"`
}

// Validate checks the closed metadata contract before it crosses the HTTP boundary.
func (s *TimePlanAnnotationSchema) Validate() error {
	if s == nil {
		return nil
	}
	if len(s.AllowedTypes) > 4 || len(s.Fields) > 6 {
		return fmt.Errorf("custd: time-plan annotation schema has too many values")
	}
	seenTypes := make(map[TimePlanAnnotationType]struct{}, len(s.AllowedTypes))
	for _, value := range s.AllowedTypes {
		if value != TimePlanAnnotationNote && value != TimePlanAnnotationMarker &&
			value != TimePlanAnnotationDecision && value != TimePlanAnnotationAction {
			return fmt.Errorf("custd: time-plan annotation schema contains an unsupported type")
		}
		if _, exists := seenTypes[value]; exists {
			return fmt.Errorf("custd: time-plan annotation schema types must be unique")
		}
		seenTypes[value] = struct{}{}
	}
	seenFields := make(map[TimePlanAnnotationField]struct{}, len(s.Fields))
	for _, value := range s.Fields {
		if value != TimePlanAnnotationText && value != TimePlanAnnotationMarkerLabel &&
			value != TimePlanAnnotationDecisionStatus && value != TimePlanAnnotationAssigneeRef &&
			value != TimePlanAnnotationDueDate && value != TimePlanAnnotationActionStatus {
			return fmt.Errorf("custd: time-plan annotation schema contains an unsupported field")
		}
		if _, exists := seenFields[value]; exists {
			return fmt.Errorf("custd: time-plan annotation schema fields must be unique")
		}
		seenFields[value] = struct{}{}
	}
	return nil
}

// Validate checks the exactly-one trigger, bounds, and closed severity enum.
func (c *TimePlanThresholdCue) Validate() error {
	if c == nil {
		return fmt.Errorf("custd: time-plan threshold cue is required")
	}
	if (c.RemainingMS == nil) == (c.RemainingFractionBPS == nil) {
		return fmt.Errorf("custd: time-plan threshold cue must have one remaining threshold")
	}
	if c.RemainingMS != nil && (*c.RemainingMS < 0 || *c.RemainingMS > 2_419_200_000) {
		return fmt.Errorf("custd: time-plan remainingMs is out of range")
	}
	if c.RemainingFractionBPS != nil && (*c.RemainingFractionBPS < 0 || *c.RemainingFractionBPS > 10_000) {
		return fmt.Errorf("custd: time-plan remainingFractionBps is out of range")
	}
	if c.Severity != TimePlanThresholdInfo && c.Severity != TimePlanThresholdWarning &&
		c.Severity != TimePlanThresholdCritical {
		return fmt.Errorf("custd: time-plan threshold cue severity is invalid")
	}
	return nil
}

// Validate checks the typed definition metadata and every nested DTO.
func (d *TimePlanDefinition) Validate() error {
	if d == nil {
		return fmt.Errorf("custd: time-plan definition is required")
	}
	if d.AnnotationSchema != nil {
		if err := d.AnnotationSchema.Validate(); err != nil {
			return err
		}
	}
	if len(d.ThresholdCues) > 16 {
		return fmt.Errorf("custd: time-plan thresholdCues cannot contain more than 16 values")
	}
	seenTriggers := make(map[string]struct{}, len(d.ThresholdCues))
	for index := range d.ThresholdCues {
		cue := &d.ThresholdCues[index]
		if err := cue.Validate(); err != nil {
			return fmt.Errorf("custd: time-plan threshold cue %d: %w", index, err)
		}
		var trigger string
		switch {
		case cue.RemainingMS != nil:
			trigger = fmt.Sprintf("ms:%d", *cue.RemainingMS)
		case cue.RemainingFractionBPS != nil:
			trigger = fmt.Sprintf("bps:%d", *cue.RemainingFractionBPS)
		default:
			return fmt.Errorf("custd: time-plan threshold cue must have one remaining threshold")
		}
		if _, exists := seenTriggers[trigger]; exists {
			return fmt.Errorf("custd: time-plan thresholdCues must not contain duplicate triggers")
		}
		seenTriggers[trigger] = struct{}{}
	}
	for _, block := range d.Blocks {
		if block.Basis != "absolute" && block.Basis != "horizon_fraction" && block.Basis != "remainder_weight" {
			return fmt.Errorf("custd: time-plan block basis is invalid")
		}
	}
	return nil
}

type TimePlanBlock struct {
	UUID        string   `json:"uuid"`
	SemanticKey string   `json:"semanticKey"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Basis       string   `json:"basis"`
	DurationMS  int64    `json:"durationMs,omitempty"`
	Numerator   int64    `json:"numerator,omitempty"`
	Denominator int64    `json:"denominator,omitempty"`
	Weight      int64    `json:"weight,omitempty"`
}

type TimePlanDraftRequest struct {
	PlanKey     string             `json:"planKey"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Definition  TimePlanDefinition `json:"definition"`
}

type TimePlanRevisionRequest struct {
	ExpectedRevision int64 `json:"expectedRevision"`
}

type TimePlanDraftRevisionRequest struct {
	ExpectedRevision int64              `json:"expectedRevision"`
	PlanKey          string             `json:"planKey"`
	Name             string             `json:"name"`
	Description      string             `json:"description,omitempty"`
	Definition       TimePlanDefinition `json:"definition"`
}

type TimePlan struct {
	UUID          string             `json:"uuid"`
	PlanKey       string             `json:"planKey"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	Status        string             `json:"status"`
	DraftRevision int64              `json:"draftRevision"`
	Definition    TimePlanDefinition `json:"definition"`
	UpdatedAt     string             `json:"updatedAt"`
}

type TimePlanListResponse struct {
	Plans []TimePlan `json:"plans"`
}

type TimePlanVersion struct {
	UUID           string `json:"uuid"`
	PlanUUID       string `json:"planUuid"`
	VersionNumber  int    `json:"versionNumber"`
	DefinitionHash string `json:"definitionHash"`
	PublishedAt    string `json:"publishedAt"`
}

type TimePlanRunRequest struct {
	PlanUUID          string `json:"planUuid"`
	VersionUUID       string `json:"versionUuid,omitempty"`
	ScheduledStartsAt string `json:"scheduledStartsAt,omitempty"`
	ScheduledEndsAt   string `json:"scheduledEndsAt,omitempty"`
}

type TimePlanCreatedRun struct {
	UUID              string               `json:"uuid"`
	PlanUUID          string               `json:"planUuid"`
	VersionUUID       string               `json:"versionUuid"`
	Status            string               `json:"status"`
	BaselineHorizonMS int64                `json:"baselineHorizonMs"`
	BlockAllocations  []TimePlanAllocation `json:"blockAllocations"`
	CreatedAt         string               `json:"createdAt"`
}

type TimePlanAllocation struct {
	BlockID    string `json:"blockId"`
	Sequence   int    `json:"sequence"`
	DurationMS int64  `json:"durationMs"`
}

type TimePlanRunBlock struct {
	UUID               string `json:"uuid"`
	Sequence           int    `json:"sequence"`
	Status             string `json:"status"`
	BaselineMS         int64  `json:"baselineMs"`
	CurrentMS          int64  `json:"currentMs"`
	AllocatedAtStartMS *int64 `json:"allocatedAtStartMs,omitempty"`
	ActualActiveMS     int64  `json:"actualActiveMs"`
	WallStartedAt      string `json:"wallStartedAt,omitempty"`
	WallEndedAt        string `json:"wallEndedAt,omitempty"`
	OutcomeCensored    bool   `json:"outcomeCensored"`
}

type TimePlanRun struct {
	UUID                string             `json:"uuid"`
	PlanUUID            string             `json:"planUuid"`
	Status              string             `json:"status"`
	StreamVersion       int64              `json:"streamVersion"`
	ScheduledStartsAt   string             `json:"scheduledStartsAt,omitempty"`
	ScheduledEndsAt     string             `json:"scheduledEndsAt,omitempty"`
	EffectiveStartsAt   string             `json:"effectiveStartsAt,omitempty"`
	EffectiveEndsAt     string             `json:"effectiveEndsAt,omitempty"`
	StartPolicy         string             `json:"startPolicy,omitempty"`
	BaselineHorizonMS   int64              `json:"baselineHorizonMs"`
	ExecutableHorizonMS *int64             `json:"executableHorizonMs,omitempty"`
	LostMS              int64              `json:"lostMs"`
	UnusedMS            int64              `json:"unusedMs"`
	OverrunMS           int64              `json:"overrunMs"`
	CurrentBlockUUID    string             `json:"currentBlockUuid,omitempty"`
	Blocks              []TimePlanRunBlock `json:"blocks"`
}

type TimePlanCommandRequest struct {
	CommandID                string                    `json:"commandId"`
	IdempotencyKey           string                    `json:"idempotencyKey"`
	ExpectedVersion          int64                     `json:"expectedVersion"`
	Type                     string                    `json:"type"`
	BlockID                  string                    `json:"blockId,omitempty"`
	ClientOccurredAt         string                    `json:"clientOccurredAt,omitempty"`
	BoundaryEndsAt           string                    `json:"boundaryEndsAt,omitempty"`
	ScheduledStartsAt        string                    `json:"scheduledStartsAt,omitempty"`
	ScheduledEndsAt          string                    `json:"scheduledEndsAt,omitempty"`
	StartPolicy              string                    `json:"startPolicy,omitempty"`
	Reason                   string                    `json:"reason,omitempty"`
	SupersedesTransitionUUID string                    `json:"supersedesTransitionUuid,omitempty"`
	Corrected                *TimePlanCorrectedCommand `json:"corrected,omitempty"`
}

type TimePlanCorrectedCommand struct {
	Type           string `json:"type"`
	BlockID        string `json:"blockId,omitempty"`
	EffectiveAt    string `json:"effectiveAt"`
	BoundaryEndsAt string `json:"boundaryEndsAt,omitempty"`
	StartPolicy    string `json:"startPolicy,omitempty"`
}

type TimePlanCalculationChange struct {
	BlockID string `json:"blockId"`
	FromMS  int64  `json:"fromMs"`
	ToMS    int64  `json:"toMs"`
}

type TimePlanCalculationReceipt struct {
	AllocatorVersion string                      `json:"allocatorVersion"`
	Reason           string                      `json:"reason"`
	Summary          string                      `json:"summary"`
	Source           []TimePlanAllocation        `json:"source"`
	Result           []TimePlanAllocation        `json:"result"`
	Changes          []TimePlanCalculationChange `json:"changes"`
}

type TimePlanCommandResult struct {
	TransitionUUID string                     `json:"transitionUuid"`
	Projection     TimePlanRun                `json:"projection"`
	Receipt        TimePlanCalculationReceipt `json:"receipt"`
	Duplicate      bool                       `json:"duplicate"`
}

type TimePlanTransition struct {
	UUID                     string                     `json:"uuid"`
	RunUUID                  string                     `json:"runUuid"`
	StreamVersion            int64                      `json:"streamVersion"`
	CommandID                string                     `json:"commandId"`
	Type                     string                     `json:"type"`
	ActorKind                string                     `json:"actorKind"`
	ActorRef                 string                     `json:"actorRef"`
	ServerReceivedAt         string                     `json:"serverReceivedAt"`
	ClientOccurredAt         string                     `json:"clientOccurredAt,omitempty"`
	Reason                   string                     `json:"reason,omitempty"`
	PreviousStatus           string                     `json:"previousStatus,omitempty"`
	CurrentStatus            string                     `json:"currentStatus"`
	AllocatorVersion         string                     `json:"allocatorVersion"`
	SchemaVersion            string                     `json:"schemaVersion"`
	SupersedesTransitionUUID string                     `json:"supersedesTransitionUuid,omitempty"`
	Receipt                  TimePlanCalculationReceipt `json:"receipt"`
}

type TimePlanHistoryResponse struct {
	Transitions []TimePlanTransition `json:"transitions"`
}

type TimePlanAnnotationInput struct {
	Type           string `json:"type"`
	RunBlockUUID   string `json:"runBlockUuid,omitempty"`
	Text           string `json:"text,omitempty"`
	MarkerLabel    string `json:"markerLabel,omitempty"`
	DecisionStatus string `json:"decisionStatus,omitempty"`
	AssigneeRef    string `json:"assigneeRef,omitempty"`
	DueDate        string `json:"dueDate,omitempty"`
	ActionStatus   string `json:"actionStatus,omitempty"`
}

type TimePlanAnnotation struct {
	UUID            string `json:"uuid"`
	RunUUID         string `json:"runUuid"`
	RunBlockUUID    string `json:"runBlockUuid,omitempty"`
	Type            string `json:"type"`
	Text            string `json:"text,omitempty"`
	MarkerLabel     string `json:"markerLabel,omitempty"`
	DecisionStatus  string `json:"decisionStatus,omitempty"`
	AssigneeRef     string `json:"assigneeRef,omitempty"`
	DueDate         string `json:"dueDate,omitempty"`
	ActionStatus    string `json:"actionStatus,omitempty"`
	SupersedesUUID  string `json:"supersedesUuid,omitempty"`
	RecordedAt      string `json:"recordedAt"`
	ActorKind       string `json:"actorKind"`
	ActorRef        string `json:"actorRef"`
	RedactedAt      string `json:"redactedAt,omitempty"`
	RedactionReason string `json:"redactionReason,omitempty"`
}

type TimePlanAnnotationListResponse struct {
	Annotations []TimePlanAnnotation `json:"annotations"`
}

type TimePlanRedactionRequest struct {
	Reason string `json:"reason"`
}

func (c *TimePlanAdminClient) List(ctx context.Context, companySlug string, limit int) (TimePlanListResponse, error) {
	var out TimePlanListResponse
	err := c.admin.request(ctx, http.MethodGet, timePlanSDKPath("/time-plans", companySlug, limit), nil, &out)
	return out, err
}

func (c *TimePlanAdminClient) Get(ctx context.Context, companySlug, planUUID string) (*TimePlan, error) {
	var out TimePlan
	err := c.admin.request(ctx, http.MethodGet, timePlanSDKPath("/time-plans/"+url.PathEscape(planUUID), companySlug, 0), nil, &out)
	return &out, err
}

func (c *TimePlanAdminClient) Create(ctx context.Context, companySlug string, req TimePlanDraftRequest) (*TimePlan, error) {
	if err := req.Definition.Validate(); err != nil {
		return nil, err
	}
	var out TimePlan
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans", companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) Preview(ctx context.Context, companySlug string, definition TimePlanDefinition) (*TimePlanAllocationPreview, error) {
	if err := definition.Validate(); err != nil {
		return nil, err
	}
	var out TimePlanAllocationPreview
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/preview", companySlug, 0), definition, &out)
	return &out, err
}

type TimePlanAllocationPreview struct {
	AllocatorVersion string               `json:"allocatorVersion"`
	HorizonMS        int64                `json:"horizonMs"`
	Allocations      []TimePlanAllocation `json:"allocations"`
}

func (c *TimePlanAdminClient) Revise(ctx context.Context, companySlug, planUUID string, req TimePlanDraftRevisionRequest) (*TimePlan, error) {
	if err := req.Definition.Validate(); err != nil {
		return nil, err
	}
	var out TimePlan
	err := c.admin.request(ctx, http.MethodPatch, timePlanSDKPath("/time-plans/"+url.PathEscape(planUUID), companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) Publish(ctx context.Context, companySlug, planUUID string, req TimePlanRevisionRequest) (*TimePlanVersion, error) {
	var out TimePlanVersion
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/"+url.PathEscape(planUUID)+"/publish", companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) Retire(ctx context.Context, companySlug, planUUID string) (*TimePlan, error) {
	var out TimePlan
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/"+url.PathEscape(planUUID)+"/retire", companySlug, 0), nil, &out)
	return &out, err
}

func (c *TimePlanAdminClient) CreateRun(ctx context.Context, companySlug string, req TimePlanRunRequest) (*TimePlanCreatedRun, error) {
	var out TimePlanCreatedRun
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/runs", companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) GetRun(ctx context.Context, companySlug, runUUID string) (*TimePlanRun, error) {
	var out TimePlanRun
	err := c.admin.request(ctx, http.MethodGet, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID), companySlug, 0), nil, &out)
	return &out, err
}

func (c *TimePlanAdminClient) History(ctx context.Context, companySlug, runUUID string, limit int) (TimePlanHistoryResponse, error) {
	var out TimePlanHistoryResponse
	err := c.admin.request(ctx, http.MethodGet, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/history", companySlug, limit), nil, &out)
	return out, err
}

func (c *TimePlanAdminClient) Execute(ctx context.Context, companySlug, runUUID string, req TimePlanCommandRequest) (*TimePlanCommandResult, error) {
	var out TimePlanCommandResult
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/commands", companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) CreateAnnotation(ctx context.Context, companySlug, runUUID string, req TimePlanAnnotationInput) (*TimePlanAnnotation, error) {
	var out TimePlanAnnotation
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/annotations", companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) ListAnnotations(ctx context.Context, companySlug, runUUID string, limit int) (TimePlanAnnotationListResponse, error) {
	var out TimePlanAnnotationListResponse
	err := c.admin.request(ctx, http.MethodGet, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/annotations", companySlug, limit), nil, &out)
	return out, err
}

func (c *TimePlanAdminClient) CorrectAnnotation(ctx context.Context, companySlug, runUUID, annotationUUID string, req TimePlanAnnotationInput) (*TimePlanAnnotation, error) {
	var out TimePlanAnnotation
	err := c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/annotations/"+url.PathEscape(annotationUUID)+"/corrections", companySlug, 0), req, &out)
	return &out, err
}

func (c *TimePlanAdminClient) RedactAnnotation(
	ctx context.Context, companySlug, runUUID, annotationUUID string, req TimePlanRedactionRequest,
) error {
	return c.admin.request(ctx, http.MethodPost, timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/annotations/"+url.PathEscape(annotationUUID)+"/redact", companySlug, 0), req, nil)
}

func timePlanSDKPath(path, companySlug string, limit int) string {
	values := url.Values{}
	if companySlug != "" {
		values.Set("companySlug", companySlug)
	}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if encoded := values.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
