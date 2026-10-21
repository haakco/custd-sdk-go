package custd

import (
	"context"
	"net/http"
	"time"
)

// The definition-math proof is produced by the server from its own allocator
// implementation. The expected vector values and the fixture vocabulary stay
// on the server: only the observed results cross this boundary.
const (
	TimePlanDefinitionMathProofSchemaVersion            = "time-plan.definition-math-proof.v1"
	TimePlanDefinitionMathProofSource                   = "admin-api.timeplans"
	TimePlanDefinitionMathProofConformanceSchemaVersion = "time-plan.conformance.v1"
	TimePlanDefinitionMathProofAllocatorVersion         = "largest-remainder.v1"
)

// TimePlanDefinitionMathProofResponse is the server-owned observation of the
// allocator conformance run for the authenticated tenant scope.
type TimePlanDefinitionMathProofResponse struct {
	SchemaVersion            string                             `json:"schemaVersion"`
	CompanySlug              string                             `json:"companySlug"`
	Source                   string                             `json:"source"`
	ObservedAt               time.Time                          `json:"observedAt"`
	ConformanceSchemaVersion string                             `json:"conformanceSchemaVersion"`
	AllocatorVersion         string                             `json:"allocatorVersion"`
	Definitions              []TimePlanDefinitionMathDefinition `json:"definitions"`
	Allocations              []TimePlanDefinitionMathAllocation `json:"allocations"`
	Adjustments              []TimePlanDefinitionMathAdjustment `json:"adjustments"`
	Summary                  TimePlanDefinitionMathProofSummary `json:"summary"`
}

// TimePlanDefinitionMathDefinition records one initial allocation vector.
type TimePlanDefinitionMathDefinition struct {
	VectorID               string `json:"vectorId"`
	Index                  int    `json:"index"`
	HorizonMS              int64  `json:"horizonMs"`
	BlockCount             int    `json:"blockCount"`
	AllocatedMS            int64  `json:"allocatedMs"`
	ExactConservation      bool   `json:"exactConservation"`
	RuntimeVocabularyClean bool   `json:"runtimeVocabularyClean"`
	Passed                 bool   `json:"passed"`
}

// TimePlanDefinitionMathAllocationValue is one block allocation result.
type TimePlanDefinitionMathAllocationValue struct {
	BlockID    string `json:"blockId"`
	Sequence   int    `json:"sequence"`
	DurationMS int64  `json:"durationMs"`
}

// TimePlanDefinitionMathAllocation records one complete allocation vector.
type TimePlanDefinitionMathAllocation struct {
	VectorID              string                                  `json:"vectorId"`
	Index                 int                                     `json:"index"`
	HorizonMS             int64                                   `json:"horizonMs"`
	Allocations           []TimePlanDefinitionMathAllocationValue `json:"allocations"`
	ExactConservation     bool                                    `json:"exactConservation"`
	DeterministicTieBreak bool                                    `json:"deterministicTieBreak"`
	NegativeAllocations   int                                     `json:"negativeAllocations"`
	Passed                bool                                    `json:"passed"`
}

// TimePlanDefinitionMathAdjustmentTiming contains the measured and remaining
// durations of one adjustment vector. It is embedded so the wire shape stays
// flat while the adjustment record remains easy to scan.
type TimePlanDefinitionMathAdjustmentTiming struct {
	HorizonMS   int64 `json:"horizonMs"`
	ElapsedMS   int64 `json:"elapsedMs"`
	ActualMS    int64 `json:"actualMs"`
	RemainingMS int64 `json:"remainingMs"`
	UnusedMS    int64 `json:"unusedMs"`
	OverrunMS   int64 `json:"overrunMs"`
}

// TimePlanDefinitionMathAdjustment records one adjustment allocation vector.
type TimePlanDefinitionMathAdjustment struct {
	VectorID string `json:"vectorId"`
	Index    int    `json:"index"`
	Reason   string `json:"reason"`
	TimePlanDefinitionMathAdjustmentTiming
	MarkFutureNotRun    bool                                    `json:"markFutureNotRun"`
	Allocations         []TimePlanDefinitionMathAllocationValue `json:"allocations"`
	ExactConservation   bool                                    `json:"exactConservation"`
	NegativeAllocations int                                     `json:"negativeAllocations"`
	Passed              bool                                    `json:"passed"`
}

// TimePlanDefinitionMathProofSummary aggregates all conformance vectors.
type TimePlanDefinitionMathProofSummary struct {
	InitialVectorCount    int  `json:"initialVectorCount"`
	AdjustmentVectorCount int  `json:"adjustmentVectorCount"`
	VectorCount           int  `json:"vectorCount"`
	Passed                bool `json:"passed"`
	ExactConservation     bool `json:"exactConservation"`
	DeterministicTieBreak bool `json:"deterministicTieBreak"`
	NegativeAllocations   int  `json:"negativeAllocations"`
}

// GetTimePlanDefinitionMathProof reads the source-owned allocator and
// adjustment conformance results for the authenticated tenant scope.
func (c *TimePlanAdminClient) GetTimePlanDefinitionMathProof(
	ctx context.Context, companySlug string,
) (*TimePlanDefinitionMathProofResponse, error) {
	var out TimePlanDefinitionMathProofResponse
	err := c.admin.request(
		ctx, http.MethodGet, timePlanSDKPath(timePlanOperationsPath+"/definition-math", companySlug, 0), nil, &out,
	)
	return &out, err
}
