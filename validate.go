package custd

import (
	"fmt"
	"regexp"
	"strings"
)

var eventLabelKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[._-][a-z0-9]+)*$`)

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
	if len(missing) != 0 {
		return fmt.Errorf("custd: missing required fields: %s", strings.Join(missing, ", "))
	}
	return validateEventLabels(event.Labels)
}

func validateEventLabels(labels map[string]string) error {
	if len(labels) > 16 {
		return fmt.Errorf("custd: labels may contain at most 16 entries")
	}
	for key, value := range labels {
		if len([]byte(key)) > 64 || !eventLabelKeyPattern.MatchString(key) || strings.HasPrefix(key, "custd.") {
			return fmt.Errorf("custd: labels.%s has an invalid key", key)
		}
		if value == "" || value != strings.TrimSpace(value) || len([]byte(value)) > 128 {
			return fmt.Errorf("custd: labels.%s has an invalid value", key)
		}
	}
	return nil
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
