package custd

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// SnapshotError freezes safe code locations before an asynchronous handoff. Error messages are never read.
func SnapshotError(request CaptureExceptionRequest) ErrorCaptureEvent {
	mode := request.Mode
	if mode == "" {
		mode = ErrorCaptureHandled
	}
	name := "Error"
	kind := reflect.TypeOf(request.Error)
	if kind != nil {
		if kind.Kind() == reflect.Pointer {
			kind = kind.Elem()
		}
		if captureTypePattern.MatchString(kind.Name()) {
			name = kind.Name()
		}
	}
	pcs := make([]uintptr, 24)
	count := runtime.Callers(2, pcs)
	iterator := runtime.CallersFrames(pcs[:count])
	frames := make([]ErrorCaptureFrame, 0, count)
	for {
		frame, more := iterator.Next()
		name := filepath.Base(frame.File)
		internal := strings.HasPrefix(frame.Function, "github.com/haakco/custd-sdk-go/v3.") && (name == "error_capture.go" || name == "error_capture_snapshot.go" || name == "error_capture_reporter.go")
		if !internal && captureFilenamePattern.MatchString(name) && frame.Line > 0 && frame.Line <= 1000000 {
			frames = append(frames, ErrorCaptureFrame{Filename: name, Lineno: frame.Line})
		}
		if !more {
			break
		}
	}
	// Collector frame order is oldest first, with the raising frame last.
	for left, right := 0, len(frames)-1; left < right; left, right = left+1, right-1 {
		frames[left], frames[right] = frames[right], frames[left]
	}
	return ErrorCaptureEvent{Platform: ErrorPlatformGo, ExceptionType: name, Frames: frames, Mode: mode, CorrelationID: request.CorrelationID}
}

type errorEnvelopeException struct {
	Type      string `json:"type"`
	Mechanism struct {
		Handled bool `json:"handled"`
	} `json:"mechanism"`
	Stacktrace struct {
		Frames []ErrorCaptureFrame `json:"frames"`
	} `json:"stacktrace"`
}
type errorEnvelopeEvent struct {
	ID          string        `json:"event_id"`
	Timestamp   string        `json:"timestamp"`
	Platform    ErrorPlatform `json:"platform"`
	Level       string        `json:"level"`
	Environment string        `json:"environment"`
	Release     string        `json:"release,omitempty"`
	SDK         struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"sdk"`
	Tags      *errorEnvelopeTags `json:"tags,omitempty"`
	Exception struct {
		Values []errorEnvelopeException `json:"values"`
	} `json:"exception"`
}
type errorEnvelopeTags struct {
	CorrelationID string `json:"correlation_id"`
}

func newErrorEnvelopeEvent(id string, input ErrorCaptureEvent) errorEnvelopeEvent {
	event := errorEnvelopeEvent{ID: id, Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Platform: input.Platform, Level: "error", Environment: input.Environment, Release: input.Release}
	event.SDK.Name = "custd.go"
	event.SDK.Version = Version
	if input.CorrelationID != "" {
		event.Tags = &errorEnvelopeTags{CorrelationID: input.CorrelationID}
	}
	exception := errorEnvelopeException{Type: input.ExceptionType}
	exception.Mechanism.Handled = input.Mode != ErrorCaptureUnhandled
	exception.Stacktrace.Frames = append([]ErrorCaptureFrame{}, input.Frames...)
	event.Exception.Values = []errorEnvelopeException{exception}
	return event
}
