package custd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func readResponseBody(body io.ReadCloser) ([]byte, error) {
	data, readErr := io.ReadAll(body)
	closeErr := body.Close()
	return data, errors.Join(readErr, closeErr)
}

// ResponseValidationError reports a success response body that does not match
// its named DTO: an empty body, an object that omits an owner-required field,
// or a field with the wrong shape. It is exported so a caller can identify a
// malformed body with errors.As instead of matching on message text; the
// request and transport errors use RequestError and Problem instead.
type ResponseValidationError struct {
	// Message is the full human-readable reason.
	Message string
	// Err is the nested JSON decode error the failure wrapped, if any.
	Err error
}

// Error renders the validation failure. When the failure wrapped a JSON decode
// error the cause is appended so a log still shows the underlying parse detail.
func (e *ResponseValidationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap exposes the wrapped JSON decode error, if any.
func (e *ResponseValidationError) Unwrap() error { return e.Err }

// validationErrorf builds a ResponseValidationError from a format string.
func validationErrorf(format string, args ...any) *ResponseValidationError {
	return &ResponseValidationError{Message: fmt.Sprintf(format, args...)}
}

// validationWrap builds a ResponseValidationError that wraps a nested decode
// error so errors.Is/As can still reach the cause.
func validationWrap(message string, err error) *ResponseValidationError {
	return &ResponseValidationError{Message: message, Err: err}
}

// decodeJSONObject decodes data as a JSON object, rejecting an empty body or a
// body whose top level is not an object. The named response DTOs go through it
// so a malformed success body fails at the boundary instead of decoding to a
// zero value.
func decodeJSONObject(data []byte, context string) (map[string]json.RawMessage, error) {
	if len(data) == 0 {
		return nil, validationErrorf("custd: %s body is empty", context)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, validationWrap(fmt.Sprintf("custd: %s must be a JSON object", context), err)
	}
	if fields == nil {
		return nil, validationErrorf("custd: %s must be a JSON object", context)
	}
	return fields, nil
}

// requireJSONFields fails when any named field is absent or explicitly null. A
// field the owner documents as optional must be checked separately rather than
// listed here, so a genuinely optional omission is not mistaken for a missing
// required one.
func requireJSONFields(fields map[string]json.RawMessage, context string, names ...string) error {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			return validationErrorf("custd: %s field %s is required", context, name)
		}
	}
	return nil
}

// requireJSONObject validates that an object field is present and carries every
// required nested field. A present null is rejected here: the owner never emits
// a null for these DTO fields.
func requireJSONObject(fields map[string]json.RawMessage, context, name string, required ...string) error {
	raw, ok := fields[name]
	if !ok || string(raw) == "null" {
		return validationErrorf("custd: %s field %s is required", context, name)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return validationWrap(fmt.Sprintf("custd: %s field %s must be an object", context, name), err)
	}
	if object == nil {
		return validationErrorf("custd: %s field %s must be an object", context, name)
	}
	return requireJSONFields(object, fmt.Sprintf("%s %s", context, name), required...)
}

// requireJSONObjectList validates that a field is a JSON list of objects and
// that every element carries the named required fields. A non-list container or
// a null element is rejected; an empty list is accepted because the owner
// emits [] for an empty collection.
func requireJSONObjectList(fields map[string]json.RawMessage, context, name string, required ...string) error {
	raw, ok := fields[name]
	if !ok || string(raw) == "null" {
		return validationErrorf("custd: %s field %s is required", context, name)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return validationWrap(fmt.Sprintf("custd: %s field %s must be a list", context, name), err)
	}
	for index, item := range items {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(item, &object); err != nil {
			return validationWrap(fmt.Sprintf("custd: %s %s[%d] must be an object", context, name, index), err)
		}
		if object == nil {
			return validationErrorf("custd: %s %s[%d] must be an object", context, name, index)
		}
		if err := requireJSONFields(object, fmt.Sprintf("%s %s[%d]", context, name, index), required...); err != nil {
			return err
		}
	}
	return nil
}
