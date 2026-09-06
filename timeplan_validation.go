package custd

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var timePlanUUIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func (r TimePlanRunRequest) Validate() error {
	if strings.TrimSpace(r.PlanUUID) == "" {
		return fmt.Errorf("custd: time-plan run planUuid is required")
	}
	return validateTimePlanSchedule(r.ScheduledStartsAt, r.ScheduledEndsAt)
}

func (r TimePlanCommandRequest) Validate() error {
	if !validTimePlanUUID(r.CommandID) || strings.TrimSpace(r.IdempotencyKey) == "" || len(r.IdempotencyKey) > 128 || r.ExpectedVersion < 0 {
		return fmt.Errorf("custd: time-plan command metadata is invalid")
	}
	if !validTimePlanCommandType(r.Type) || len(r.Reason) > 500 {
		return fmt.Errorf("custd: time-plan command type or reason is invalid")
	}
	for name, value := range map[string]string{
		"clientOccurredAt": r.ClientOccurredAt, "boundaryEndsAt": r.BoundaryEndsAt,
		"scheduledStartsAt": r.ScheduledStartsAt, "scheduledEndsAt": r.ScheduledEndsAt,
	} {
		if err := validateOptionalTime(name, value); err != nil {
			return err
		}
	}
	if r.Type != "append_correction" && (r.Corrected != nil || strings.TrimSpace(r.SupersedesTransitionUUID) != "") {
		return fmt.Errorf("custd: non-correction command cannot contain correction fields")
	}
	if r.Type == "append_correction" {
		if !validTimePlanUUID(r.SupersedesTransitionUUID) || r.Corrected == nil {
			return fmt.Errorf("custd: correction requires a superseded transition and replacement")
		}
		if err := r.Corrected.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c TimePlanCorrectedCommand) Validate() error {
	if !validTimePlanCommandType(c.Type) || c.Type == "append_correction" {
		return fmt.Errorf("custd: corrected time-plan command type is invalid")
	}
	return validateRequiredTime("effectiveAt", c.EffectiveAt)
}

func (i TimePlanAnnotationInput) Validate() error {
	if len(i.Text) > 4_000 {
		return fmt.Errorf("custd: time-plan annotation text cannot exceed 4000 bytes")
	}
	if err := validateOptionalTime("dueDate", i.DueDate); err != nil {
		return err
	}
	switch i.Type {
	case "note":
		if strings.TrimSpace(i.Text) == "" || i.MarkerLabel != "" || i.DecisionStatus != "" {
			return fmt.Errorf("custd: note annotation fields are invalid")
		}
	case "decision":
		if strings.TrimSpace(i.Text) == "" || i.MarkerLabel != "" || (i.DecisionStatus != "" && !validTimePlanDecisionStatus(i.DecisionStatus)) {
			return fmt.Errorf("custd: decision annotation fields are invalid")
		}
	case "marker":
		if strings.TrimSpace(i.MarkerLabel) == "" || len(i.MarkerLabel) > 100 || i.Text != "" || i.DecisionStatus != "" {
			return fmt.Errorf("custd: marker annotation fields are invalid")
		}
	case "action":
		if strings.TrimSpace(i.Text) == "" || !validTimePlanActionStatus(i.ActionStatus) || i.MarkerLabel != "" || i.DecisionStatus != "" || len(i.AssigneeRef) > 200 {
			return fmt.Errorf("custd: action annotation fields are invalid")
		}
	default:
		return fmt.Errorf("custd: time-plan annotation type is invalid")
	}
	if i.Type != "action" && (i.AssigneeRef != "" || i.DueDate != "" || i.ActionStatus != "") {
		return fmt.Errorf("custd: action annotation fields require action type")
	}
	return nil
}

func (r TimePlanRedactionRequest) Validate() error {
	if length := len(strings.TrimSpace(r.Reason)); length == 0 || length > 200 {
		return fmt.Errorf("custd: time-plan redaction reason must contain 1 to 200 bytes")
	}
	return nil
}

func validateTimePlanSchedule(startsAt, endsAt string) error {
	if startsAt == "" && endsAt == "" {
		return nil
	}
	if startsAt == "" || endsAt == "" {
		return fmt.Errorf("custd: time-plan schedule requires an ordered start and end")
	}
	start, startErr := time.Parse(time.RFC3339, startsAt)
	end, endErr := time.Parse(time.RFC3339, endsAt)
	if startErr != nil || endErr != nil || !end.After(start) {
		return fmt.Errorf("custd: time-plan schedule requires an ordered start and end")
	}
	return nil
}

func validateOptionalTime(name, value string) error {
	if value != "" {
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("custd: time-plan %s is invalid", name)
		}
	}
	return nil
}

func validateRequiredTime(name, value string) error {
	if value == "" {
		return fmt.Errorf("custd: time-plan %s is required", name)
	}
	return validateOptionalTime(name, value)
}

func validTimePlanUUID(value string) bool {
	return timePlanUUIDPattern.MatchString(strings.TrimSpace(value))
}

func validTimePlanCommandType(value string) bool {
	switch value {
	case "schedule_run", "start_run", "start_block", "complete_block", "skip_block", "change_boundary", "complete_run", "cancel_run", "append_correction":
		return true
	default:
		return false
	}
}

func validTimePlanDecisionStatus(value string) bool {
	return value == "proposed" || value == "accepted" || value == "rejected" || value == "superseded"
}

func validTimePlanActionStatus(value string) bool {
	return value == "open" || value == "completed" || value == "cancelled"
}
