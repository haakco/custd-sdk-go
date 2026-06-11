package custd

type AdminSchemaRegister struct {
	EventTypeSlug string         `json:"eventTypeSlug"`
	Version       string         `json:"version"`
	JSONSchema    map[string]any `json:"jsonSchema"`
}

type AdminSchemaVersionCreate struct {
	Version    string         `json:"version"`
	JSONSchema map[string]any `json:"jsonSchema"`
}

type AdminSchema struct {
	EventTypeSlug string         `json:"eventTypeSlug"`
	Version       string         `json:"version"`
	JSONSchema    map[string]any `json:"jsonSchema,omitempty"`
}

type AdminSchemaList struct {
	Schemas []AdminSchema `json:"schemas"`
}
