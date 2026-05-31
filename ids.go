package custd

import (
	"crypto/rand"
	"fmt"
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
