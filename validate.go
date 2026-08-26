package custd

import (
	"fmt"
	"strings"
)

type requiredField struct {
	name    string
	missing bool
}

// ValidateEvent checks that all required fields are present on an event envelope.
func ValidateEvent(event *EventEnvelope) error {
	if event == nil {
		return fmt.Errorf("custd: event is required")
	}
	missing := collectMissingFields(event)
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("custd: missing required fields: %s", strings.Join(missing, ", "))
}

// collectMissingFields returns the names of required fields that are empty.
func collectMissingFields(event *EventEnvelope) []string {
	var missing []string
	missing = checkTopLevelFields(event, missing)
	missing = checkDeviceType(event, missing)
	return missing
}

// checkTopLevelFields validates the top-level required fields.
func checkTopLevelFields(event *EventEnvelope, missing []string) []string {
	return append(missing, collectMissingRequiredFields([]requiredField{
		{name: "eventUuid", missing: event.EventUUID == ""},
		{name: "eventTypeSlug", missing: event.EventTypeSlug == ""},
		{name: "schemaVersion", missing: event.SchemaVersion == ""},
		{name: "timestamp", missing: event.Timestamp == ""},
		{name: "sessionId", missing: event.SessionID == ""},
		{name: "anonymousId", missing: event.AnonymousID == ""},
		{name: "companySlug", missing: event.CompanySlug == ""},
		{name: "payload", missing: event.Payload == nil},
	})...)
}

func collectMissingRequiredFields(fields []requiredField) []string {
	var missing []string
	for _, field := range fields {
		if field.missing {
			missing = append(missing, field.name)
		}
	}
	return missing
}

// checkDeviceType validates that context.device.type is present.
func checkDeviceType(event *EventEnvelope, missing []string) []string {
	if event.Context.Device == nil || event.Context.Device.Type == "" {
		missing = append(missing, "context.device.type")
	}
	return missing
}
