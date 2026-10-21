package custd

import (
	"crypto/rand"
	"fmt"
	"time"
)

func fillEnvelopeDefaults(event *EventEnvelope) {
	if event == nil {
		return
	}
	if event.EventUUID == "" {
		event.EventUUID = randomUUID()
	}
	if event.SessionID == "" {
		event.SessionID = randomUUID()
	}
	if event.AnonymousID == "" {
		event.AnonymousID = randomUUID()
	}
}

func randomUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("custd: generate uuid: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// NewIdempotencyKey returns a canonical UUIDv7, the retry identity Custd's mutating admin
// routes require.
//
// Version 7 rather than 4 because the server validates the version: it keeps keys ordered
// by creation, so a replay is matched against a bounded window rather than a whole table.
func NewIdempotencyKey() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("custd: generate idempotency key: %w", err)
	}
	millis := uint64(time.Now().UnixMilli())
	b[0], b[1], b[2] = byte(millis>>40), byte(millis>>32), byte(millis>>24)
	b[3], b[4], b[5] = byte(millis>>16), byte(millis>>8), byte(millis)
	b[6] = (b[6] & 0x0f) | 0x70 // version 7
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
