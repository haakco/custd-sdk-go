package custd

import "sync"

// queue is a thread-safe in-memory event queue.
type queue struct {
	mu       sync.Mutex
	events   []EventEnvelope
	capacity int
}

// newQueue creates a queue with the given maximum capacity.
func newQueue(capacity int) *queue {
	return &queue{
		events:   make([]EventEnvelope, 0, capacity),
		capacity: capacity,
	}
}

// enqueue adds an event to the queue, dropping the oldest if at capacity.
// It returns the new queue length so callers can make atomic threshold decisions.
func (q *queue) enqueue(event *EventEnvelope) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.events) >= q.capacity {
		q.events = q.events[1:]
	}
	q.events = append(q.events, *event)
	return len(q.events)
}

// dequeue removes and returns up to n events from the front of the queue.
func (q *queue) dequeue(n int) []EventEnvelope {
	q.mu.Lock()
	defer q.mu.Unlock()

	if n > len(q.events) {
		n = len(q.events)
	}
	batch := make([]EventEnvelope, n)
	copy(batch, q.events[:n])
	q.events = q.events[n:]
	return batch
}

// requeue prepends events back to the front of the queue.
// If the result exceeds capacity, the oldest events (front) are dropped
// to stay consistent with enqueue's "drop oldest" policy.
func (q *queue) requeue(events []EventEnvelope) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.events = append(events, q.events...)
	if len(q.events) > q.capacity {
		q.events = q.events[len(q.events)-q.capacity:]
	}
}

// len returns the current queue length.
func (q *queue) len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.events)
}
