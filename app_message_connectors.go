package custd

import (
	"context"
	"net/http"
	"time"
)

const appMessageSourceGenerationHeader = "X-Custd-Source-Generation"

type AppMessageConnectorState string

const (
	AppMessageConnectorActive       AppMessageConnectorState = "active"
	AppMessageConnectorPaused       AppMessageConnectorState = "paused"
	AppMessageConnectorStale        AppMessageConnectorState = "stale"
	AppMessageConnectorDisconnected AppMessageConnectorState = "disconnected"
)

type AppMessageConnectorAction string

const (
	AppMessageConnectorPause      AppMessageConnectorAction = "pause"
	AppMessageConnectorResume     AppMessageConnectorAction = "resume"
	AppMessageConnectorDisconnect AppMessageConnectorAction = "disconnect"
)

type AppMessageConnectorConnection string

const (
	AppMessageConnectorConnecting       AppMessageConnectorConnection = "connecting"
	AppMessageConnectorImportingHistory AppMessageConnectorConnection = "importing_history"
	AppMessageConnectorSyncing          AppMessageConnectorConnection = "syncing"
	AppMessageConnectorOffline          AppMessageConnectorConnection = "offline"
	AppMessageConnectorPairAgain        AppMessageConnectorConnection = "pair_again"
)

type AppMessageConnectorControl struct {
	Action AppMessageConnectorAction `json:"action"`
}

type AppMessageConnectorReport struct {
	ConnectionState AppMessageConnectorConnection `json:"connectionState"`
	PendingMedia    int64                         `json:"pendingMedia"`
}

type AppMessageConnectorStatus struct {
	State            AppMessageConnectorState      `json:"state"`
	GrantID          string                        `json:"grantId,omitempty"`
	SourceGeneration int64                         `json:"sourceGeneration,omitempty"`
	CreatedAt        *time.Time                    `json:"createdAt,omitempty"`
	UpdatedAt        *time.Time                    `json:"updatedAt,omitempty"`
	LastReportAt     *time.Time                    `json:"lastReportAt,omitempty"`
	PendingMedia     *int64                        `json:"pendingMedia,omitempty"`
	ConnectionState  AppMessageConnectorConnection `json:"connectionState,omitempty"`
}

type AppMessageConnectorEnrollment struct {
	Status      AppMessageConnectorStatus `json:"status"`
	BearerToken string                    `json:"bearerToken"`
}

type AppMessageMutationKind string

const (
	AppMessageMutationEdit   AppMessageMutationKind = "edit"
	AppMessageMutationRevoke AppMessageMutationKind = "revoke"
)

type AppMessageMutation struct {
	ExternalID  string                 `json:"externalId"`
	OperationID string                 `json:"operationId"`
	Kind        AppMessageMutationKind `json:"kind"`
	Revision    int64                  `json:"revision"`
	Text        *string                `json:"text,omitempty"`
}

type AppMessageMutationBatch struct {
	Source           AppMessageSource     `json:"source"`
	Mutations        []AppMessageMutation `json:"mutations"`
	SourceGeneration int64                `json:"sourceGeneration"`
}

type AppMessageMutationResult struct {
	ExternalID  string `json:"externalId"`
	OperationID string `json:"operationId"`
	Status      string `json:"status"`
	ErrorCode   string `json:"errorCode,omitempty"`
}

type AppMessageMutationBatchResponse struct {
	Results []AppMessageMutationResult `json:"results"`
}

// GetConnector reads the signed-in archive owner's connector.
func (c *AppMessageClient) GetConnector(ctx context.Context) (*AppMessageConnectorStatus, error) {
	var out AppMessageConnectorStatus
	err := c.jsonRequest(ctx, http.MethodGet, "/connector", nil, &out, "state")
	return &out, err
}

// EnrollConnector creates a new grant and returns its bearer once. Use a human tenant administrator token.
func (c *AppMessageClient) EnrollConnector(ctx context.Context) (*AppMessageConnectorEnrollment, error) {
	var out AppMessageConnectorEnrollment
	err := c.jsonRequest(ctx, http.MethodPost, "/connector/enroll", nil, &out, "status", "bearerToken")
	return &out, err
}

func (c *AppMessageClient) ControlConnector(ctx context.Context, request AppMessageConnectorControl) (*AppMessageConnectorStatus, error) {
	var out AppMessageConnectorStatus
	err := c.jsonRequest(ctx, http.MethodPost, "/connector/control", request, &out, "state")
	return &out, err
}

// ConnectorSession verifies the connector bearer and returns its bound source generation.
func (c *AppMessageClient) ConnectorSession(ctx context.Context) (*AppMessageSession, error) {
	var out AppMessageSession
	err := c.jsonRequest(ctx, http.MethodGet, "/connector/session", nil, &out,
		"companySlug", "principalId", "importAllowed", "whatsappSourceGeneration")
	return &out, err
}

// ReportConnector publishes connection state using the connector bearer.
func (c *AppMessageClient) ReportConnector(ctx context.Context, request AppMessageConnectorReport) (*AppMessageConnectorStatus, error) {
	var out AppMessageConnectorStatus
	err := c.jsonRequest(ctx, http.MethodPost, "/connector/report", request, &out, "state")
	return &out, err
}

// Mutate sends a batch of native message edits or revocations using the connector bearer.
func (c *AppMessageClient) Mutate(ctx context.Context, request AppMessageMutationBatch) (*AppMessageMutationBatchResponse, error) {
	var out AppMessageMutationBatchResponse
	err := c.jsonRequest(ctx, http.MethodPost, "/mutations", request, &out, "results")
	return &out, err
}
