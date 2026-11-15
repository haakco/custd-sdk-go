package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
)

// SendEventsRequest is a synchronous batch with stable caller-owned event IDs.
type SendEventsRequest struct {
	Events []EventEnvelope `json:"events"`
}

// SendEventsResponse preserves partial outcomes and identifies only confirmed accepted events.
type SendEventsResponse struct {
	Success          bool                `json:"success"`
	Results          []EventIngestResult `json:"results"`
	AcceptedEventIDs []string            `json:"acceptedEventIds"`
}

// SendEvents sends once without the local queue or automatic retry. It validates every
// returned ID and status before exposing accepted IDs. Partial rejection is a valid
// response; a transport/malformed-response error returns no acknowledgement.
func (c *CustdClient) SendEvents(ctx context.Context, request SendEventsRequest) (*SendEventsResponse, error) {
	if err := c.validateTransport(); err != nil {
		return nil, err
	}
	events, err := c.prepareConfirmedEvents(request.Events)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(eventBatchRequest{Events: events})
	if err != nil {
		return nil, fmt.Errorf("custd: marshal event batch: %w", err)
	}
	body, gzipped, err := c.maybeCompress(body)
	if err != nil {
		return nil, err
	}
	response, err := c.sendBatchRequest(ctx, body, gzipped)
	if err != nil {
		return nil, err
	}
	if err := c.checkStatus(response.StatusCode, response.Body); err != nil {
		return nil, err
	}
	fields, err := decodeJSONObject(response.Body, "batch response")
	if err != nil {
		return nil, err
	}
	if err := requireJSONFields(fields, "batch response", "success", "results"); err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, "batch response", "results", "eventUuid", "success", "status"); err != nil {
		return nil, err
	}
	var result SendEventsResponse
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return nil, validationWrap("custd: decode batch response", err)
	}
	if err := validateConfirmedEventResults(events, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *CustdClient) prepareConfirmedEvents(input []EventEnvelope) ([]EventEnvelope, error) {
	if len(input) < 1 || len(input) > 1000 {
		return nil, fmt.Errorf("custd: confirmed batch must contain 1 to 1000 events")
	}
	if err := validateEnvironmentValue(c.config.Environment); err != nil {
		return nil, err
	}
	events := append([]EventEnvelope(nil), input...)
	seen := make(map[string]bool, len(events))
	for i := range events {
		event := &events[i]
		if err := ValidateEvent(event); err != nil {
			return nil, fmt.Errorf("custd: event %d: %w", i, err)
		}
		if seen[event.EventUUID] {
			return nil, fmt.Errorf("custd: batch contains duplicate eventUuid")
		}
		seen[event.EventUUID] = true
		event.Labels = maps.Clone(event.Labels)
		applyEnvironmentLabel(event, c.config.Environment)
	}
	return events, nil
}

func validateConfirmedEventResults(events []EventEnvelope, response *SendEventsResponse) error {
	if len(response.Results) != len(events) {
		return validationErrorf("custd: batch response result count does not match submitted events")
	}
	expected := make(map[string]bool, len(events))
	for _, event := range events {
		expected[event.EventUUID] = true
	}
	accepted := make(map[string]bool, len(events))
	for _, result := range response.Results {
		if !expected[result.EventUUID] {
			return validationErrorf("custd: batch response contains an unknown or duplicate eventUuid")
		}
		delete(expected, result.EventUUID)
		if result.Success {
			if result.Status < 200 || result.Status >= 300 {
				return validationErrorf("custd: accepted batch result must have a 2xx status")
			}
			accepted[result.EventUUID] = true
		} else if result.Status < 400 || result.Status > 599 {
			return validationErrorf("custd: rejected batch result must have an error status")
		}
	}
	if response.Success != (len(accepted) == len(events)) {
		return validationErrorf("custd: batch response success disagrees with per-event results")
	}
	// Rebuild acknowledgements from validated results, never trust a collector-supplied acceptedEventIds field.
	response.AcceptedEventIDs = make([]string, 0, len(accepted))
	for _, event := range events {
		if accepted[event.EventUUID] {
			response.AcceptedEventIDs = append(response.AcceptedEventIDs, event.EventUUID)
		}
	}
	return nil
}
