package domain

// EventService defines the interface for event recording service
// It provides a generic method to record different types of events, currently mainly evaluation events
type EventService interface {
	// RecordEvaluationEvent asynchronously records an evaluation event
	RecordEvaluationEvent(event *EvaluationEvent)
	// Close gracefully shuts down the event service, draining in-flight events
	Close() error
}
