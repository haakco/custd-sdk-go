package custd

import (
	"context"
	"net/http"
	"net/url"
)

type DataLabelAdminClient struct{ admin *AdminClient }

type DataLabelValue struct {
	UUID        string `json:"uuid"`
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
type DataLabelDefinition struct {
	UUID              string           `json:"uuid"`
	Authority         string           `json:"authority"`
	Key               string           `json:"key"`
	DisplayName       string           `json:"displayName"`
	Description       string           `json:"description"`
	AllowedScopes     []string         `json:"allowedScopes"`
	Sensitivity       string           `json:"sensitivity"`
	IntendedUse       string           `json:"intendedUse"`
	Synonyms          []string         `json:"synonyms"`
	PropagationPolicy string           `json:"propagationPolicy"`
	Enabled           bool             `json:"enabled"`
	CreatedAt         string           `json:"createdAt"`
	UpdatedAt         string           `json:"updatedAt"`
	Values            []DataLabelValue `json:"values"`
}
type DataLabelDefinitionCreateRequest struct {
	Key               string   `json:"key"`
	DisplayName       string   `json:"displayName"`
	Description       string   `json:"description"`
	AllowedScopes     []string `json:"allowedScopes"`
	Sensitivity       string   `json:"sensitivity"`
	IntendedUse       string   `json:"intendedUse"`
	Synonyms          []string `json:"synonyms"`
	PropagationPolicy string   `json:"propagationPolicy"`
}
type DataLabelDefinitionUpdateRequest struct {
	DisplayName       string   `json:"displayName"`
	Description       string   `json:"description"`
	AllowedScopes     []string `json:"allowedScopes"`
	Sensitivity       string   `json:"sensitivity"`
	IntendedUse       string   `json:"intendedUse"`
	Synonyms          []string `json:"synonyms"`
	PropagationPolicy string   `json:"propagationPolicy"`
}
type DataLabelValueCreateRequest struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}
type DataLabelValueUpdateRequest struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}
type DataLabelDefinitionListResponse struct {
	Definitions []DataLabelDefinition `json:"definitions"`
}
type DataLabelUsage struct {
	DefinitionUUID         string `json:"definitionUuid"`
	ValueUUID              string `json:"valueUuid"`
	EventTypeDefaults      int64  `json:"eventTypeDefaults"`
	SchemaFieldAssignments int64  `json:"schemaFieldAssignments"`
	Total                  int64  `json:"total"`
}
type DataLabelUsageListResponse struct {
	Usage []DataLabelUsage `json:"usage"`
}
type EventTypeDataLabelDefault struct {
	EventTypeSlug  string `json:"eventTypeSlug"`
	DefinitionUUID string `json:"definitionUuid"`
	ValueUUID      string `json:"valueUuid"`
}
type EventTypeDataLabelDefaultRequest struct {
	ValueUUID string `json:"valueUuid"`
}
type SchemaFieldDataLabelAssignment struct {
	AssignmentUUID string `json:"assignmentUuid"`
	SchemaUUID     string `json:"schemaUuid"`
	FieldPath      string `json:"fieldPath"`
	DefinitionUUID string `json:"definitionUuid"`
	ValueUUID      string `json:"valueUuid"`
}
type SchemaFieldDataLabelAssignmentRequest struct {
	FieldPath      string `json:"fieldPath"`
	DefinitionUUID string `json:"definitionUuid"`
	ValueUUID      string `json:"valueUuid"`
}
type DataLabelAssignmentListResponse struct {
	EventTypeDefaults      []EventTypeDataLabelDefault      `json:"eventTypeDefaults"`
	SchemaFieldAssignments []SchemaFieldDataLabelAssignment `json:"schemaFieldAssignments"`
}
type DescriptiveDataLabel struct {
	DefinitionUUID string   `json:"definitionUuid"`
	Key            string   `json:"key"`
	DisplayName    string   `json:"displayName"`
	Description    string   `json:"description,omitempty"`
	Sensitivity    string   `json:"sensitivity,omitempty"`
	IntendedUse    string   `json:"intendedUse,omitempty"`
	Synonyms       []string `json:"synonyms,omitempty"`
	Enabled        bool     `json:"enabled"`
}
type DataLabelCatalogue struct {
	Labels []DescriptiveDataLabel `json:"labels"`
}
type DataLabelCatalogueAssignment struct {
	Scope          string `json:"scope"`
	EventTypeSlug  string `json:"eventTypeSlug,omitempty"`
	SchemaUUID     string `json:"schemaUuid,omitempty"`
	FieldPath      string `json:"fieldPath,omitempty"`
	DefinitionUUID string `json:"definitionUuid"`
	ValueUUID      string `json:"valueUuid"`
}
type DataLabelCatalogueDataset struct {
	DisplayName    string   `json:"displayName"`
	Unit           string   `json:"unit,omitempty"`
	DefinitionUUID string   `json:"definitionUuid,omitempty"`
	ValueUUIDs     []string `json:"valueUuids,omitempty"`
}
type DataLabelCataloguePack struct {
	DisplayName string                      `json:"displayName"`
	Enabled     bool                        `json:"enabled"`
	Datasets    []DataLabelCatalogueDataset `json:"datasets"`
}
type DataLabelCatalogueResponse struct {
	Catalogue      DataLabelCatalogue             `json:"catalogue"`
	Assignments    []DataLabelCatalogueAssignment `json:"assignments"`
	ReportingPacks []DataLabelCataloguePack       `json:"reportingPacks"`
	Fingerprint    string                         `json:"fingerprint"`
}

func dataLabelIncludeDisabled(include bool) string {
	if include {
		return "?includeDisabled=true"
	}
	return ""
}
func (c *DataLabelAdminClient) List(ctx context.Context, include bool) (*DataLabelDefinitionListResponse, error) {
	var out DataLabelDefinitionListResponse
	err := c.admin.request(ctx, http.MethodGet, "/data-labels"+dataLabelIncludeDisabled(include), nil, &out)
	return &out, err
}
func (c *DataLabelAdminClient) Catalogue(ctx context.Context, include bool) (*DataLabelCatalogueResponse, error) {
	var out DataLabelCatalogueResponse
	err := c.admin.request(ctx, http.MethodGet, "/data-labels/catalogue"+dataLabelIncludeDisabled(include), nil, &out)
	return &out, err
}
func (c *DataLabelAdminClient) Get(ctx context.Context, uuid string, include bool) (*DataLabelDefinition, error) {
	var out DataLabelDefinition
	err := c.admin.request(ctx, http.MethodGet, "/data-labels/"+url.PathEscape(uuid)+dataLabelIncludeDisabled(include), nil, &out)
	return &out, err
}
func (c *DataLabelAdminClient) Create(ctx context.Context, body DataLabelDefinitionCreateRequest) (*DataLabelDefinition, error) {
	var out DataLabelDefinition
	err := c.admin.request(ctx, http.MethodPost, "/data-labels", body, &out)
	return &out, err
}
func (c *DataLabelAdminClient) Update(ctx context.Context, uuid string, body DataLabelDefinitionUpdateRequest) (*DataLabelDefinition, error) {
	var out DataLabelDefinition
	err := c.admin.request(ctx, http.MethodPatch, "/data-labels/"+url.PathEscape(uuid), body, &out)
	return &out, err
}
func (c *DataLabelAdminClient) Disable(ctx context.Context, uuid string) error {
	return c.admin.request(ctx, http.MethodPost, "/data-labels/"+url.PathEscape(uuid)+"/disable", nil, nil)
}
func (c *DataLabelAdminClient) CreateValue(ctx context.Context, uuid string, body DataLabelValueCreateRequest) (*DataLabelValue, error) {
	var out DataLabelValue
	err := c.admin.request(ctx, http.MethodPost, "/data-labels/"+url.PathEscape(uuid)+"/values", body, &out)
	return &out, err
}
func (c *DataLabelAdminClient) UpdateValue(ctx context.Context, uuid string, body DataLabelValueUpdateRequest) (*DataLabelValue, error) {
	var out DataLabelValue
	err := c.admin.request(ctx, http.MethodPatch, "/data-label-values/"+url.PathEscape(uuid), body, &out)
	return &out, err
}
func (c *DataLabelAdminClient) DisableValue(ctx context.Context, uuid string) error {
	return c.admin.request(ctx, http.MethodPost, "/data-label-values/"+url.PathEscape(uuid)+"/disable", nil, nil)
}
func (c *DataLabelAdminClient) ListUsage(ctx context.Context) (*DataLabelUsageListResponse, error) {
	var out DataLabelUsageListResponse
	err := c.admin.request(ctx, http.MethodGet, "/data-labels/usage", nil, &out)
	return &out, err
}
func (c *DataLabelAdminClient) ListAssignments(ctx context.Context) (*DataLabelAssignmentListResponse, error) {
	var out DataLabelAssignmentListResponse
	err := c.admin.request(ctx, http.MethodGet, "/data-label-assignments", nil, &out)
	return &out, err
}
func (c *DataLabelAdminClient) SetEventTypeDefault(ctx context.Context, slug, definitionUUID string, body EventTypeDataLabelDefaultRequest) error {
	return c.admin.request(ctx, http.MethodPut, "/event-types/"+url.PathEscape(slug)+"/data-label-defaults/"+url.PathEscape(definitionUUID), body, nil)
}
func (c *DataLabelAdminClient) RemoveEventTypeDefault(ctx context.Context, slug, definitionUUID string) error {
	return c.admin.request(ctx, http.MethodDelete, "/event-types/"+url.PathEscape(slug)+"/data-label-defaults/"+url.PathEscape(definitionUUID), nil, nil)
}
func (c *DataLabelAdminClient) SetSchemaFieldAssignment(ctx context.Context, schemaUUID string, body SchemaFieldDataLabelAssignmentRequest) error {
	return c.admin.request(ctx, http.MethodPut, "/event-schemas/"+url.PathEscape(schemaUUID)+"/field-data-labels", body, nil)
}
func (c *DataLabelAdminClient) RemoveSchemaFieldAssignment(ctx context.Context, assignmentUUID string) error {
	return c.admin.request(ctx, http.MethodDelete, "/data-label-assignments/schema-fields/"+url.PathEscape(assignmentUUID), nil, nil)
}
