package custd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const dogfoodDeviceType = "server"

var dogfoodProtectedPayloadFields = map[string]struct{}{
	"sourcesystem":  {},
	"sourcecompany": {},
	"environment":   {},
	"schemaversion": {},
	"correlationid": {},
}

var dogfoodForbiddenPayloadFields = map[string]struct{}{
	"apikey":             {},
	"authorization":      {},
	"password":           {},
	"rawapiresponse":     {},
	"token":              {},
	"signedurl":          {},
	"rawprompt":          {},
	"oauthcode":          {},
	"devicesecret":       {},
	"providercredential": {},
}

// DogfoodEventInput holds the small set of fields required to build a dogfood event.
type DogfoodEventInput struct {
	EventTypeSlug string
	SchemaVersion string
	CompanySlug   string
	SourceSystem  string
	SourceCompany string
	Environment   string
	CorrelationID string
	Payload       map[string]any
}

// NewDogfoodEvent builds a valid event envelope with protected dogfood payload fields.
func NewDogfoodEvent(input *DogfoodEventInput) (*EventEnvelope, error) {
	if input == nil {
		return nil, fmt.Errorf("custd: dogfood event input is required")
	}
	if err := validateDogfoodInput(input); err != nil {
		return nil, err
	}
	encodedPayload, err := json.Marshal(dogfoodPayload(input))
	if err != nil {
		return nil, err
	}
	event := dogfoodEnvelope(input, encodedPayload)
	fillEnvelopeDefaults(event)
	return event, nil
}

func validateDogfoodInput(input *DogfoodEventInput) error {
	missing := collectMissingRequiredFields([]requiredField{
		{name: "eventTypeSlug", missing: input.EventTypeSlug == ""},
		{name: "schemaVersion", missing: input.SchemaVersion == ""},
		{name: "companySlug", missing: input.CompanySlug == ""},
		{name: "sourceSystem", missing: input.SourceSystem == ""},
		{name: "sourceCompany", missing: input.SourceCompany == ""},
		{name: "environment", missing: input.Environment == ""},
	})
	if len(missing) > 0 {
		return fmt.Errorf("custd: missing dogfood fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

func dogfoodPayload(input *DogfoodEventInput) map[string]any {
	payload := make(map[string]any, len(input.Payload)+5)
	for key, value := range input.Payload {
		if dogfoodPayloadFieldAllowed(key) {
			payload[key] = sanitizeDogfoodPayloadValue(value)
		}
	}
	payload["sourceSystem"] = input.SourceSystem
	payload["sourceCompany"] = input.SourceCompany
	payload["environment"] = input.Environment
	payload["schemaVersion"] = input.SchemaVersion
	if input.CorrelationID != "" {
		payload["correlationId"] = input.CorrelationID
	}
	return payload
}

func dogfoodPayloadFieldAllowed(key string) bool {
	if _, protected := dogfoodProtectedPayloadFields[dogfoodFieldKey(key)]; protected {
		return false
	}
	_, forbidden := dogfoodForbiddenPayloadFields[dogfoodFieldKey(key)]
	return !forbidden
}

func dogfoodFieldKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "_", ""))
}

func sanitizeDogfoodPayloadValue(value any) any {
	typed, ok := value.(map[string]any)
	if !ok {
		return value
	}
	cleaned := make(map[string]any, len(typed))
	for key, child := range typed {
		if dogfoodPayloadFieldAllowed(key) {
			cleaned[key] = sanitizeDogfoodPayloadValue(child)
		}
	}
	return cleaned
}

func dogfoodEnvelope(input *DogfoodEventInput, payload json.RawMessage) *EventEnvelope {
	return &EventEnvelope{
		EventTypeSlug: input.EventTypeSlug,
		SchemaVersion: input.SchemaVersion,
		CompanySlug:   input.CompanySlug,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Context: EventContext{
			Device: &DeviceContext{Type: dogfoodDeviceType},
		},
		Payload: payload,
	}
}
