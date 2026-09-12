package custd

import (
	"fmt"
	"regexp"
	"strings"
)

// EnvironmentLabelKey is the reserved envelope label that carries a declared
// environment. Clients declare it through EventEnvelope.Environment or
// ClientConfig.Environment rather than through Labels, because every other
// reserved key is rejected.
const EnvironmentLabelKey = "custd.environment"

// UnclassifiedEnvironment is the report-side value for events without a usable
// environment. It is reserved: a client cannot declare it.
const UnclassifiedEnvironment = "unclassified"

var environmentValuePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

// validateEnvironmentValue checks a declared environment against the same
// contract ingest applies, so a rejected declaration fails locally instead of
// failing the event at the API.
func validateEnvironmentValue(value string) error {
	if value == "" {
		return nil
	}
	if value != strings.TrimSpace(value) || !environmentValuePattern.MatchString(value) {
		return fmt.Errorf("custd: environment %q must be lowercase letters, digits and hyphens, at most 32 characters", value)
	}
	if value == UnclassifiedEnvironment {
		return fmt.Errorf("custd: environment %q is reserved", value)
	}
	return nil
}

// applyEnvironmentLabel stamps the effective environment onto the envelope
// labels. The event's own declaration wins over the client default, and the
// function never overwrites a label it did not set. It runs after label
// validation because the label key is reserved: callers declare the environment
// through the field, not through Labels.
func applyEnvironmentLabel(event *EventEnvelope, clientEnvironment string) {
	if event == nil {
		return
	}
	effective := event.Environment
	if effective == "" {
		effective = clientEnvironment
	}
	if effective == "" {
		return
	}
	if event.Labels == nil {
		event.Labels = make(map[string]string, 1)
	}
	if _, exists := event.Labels[EnvironmentLabelKey]; exists {
		return
	}
	event.Labels[EnvironmentLabelKey] = effective
}
