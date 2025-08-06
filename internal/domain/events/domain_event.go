package events

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventID() uuid.UUID
	EventType() string
	AggregateID() uuid.UUID
	AggregateType() string
	EventVersion() int
	OccurredOn() time.Time
	EventData() map[string]interface{}
}

// BaseDomainEvent provides a base implementation for domain events
type BaseDomainEvent struct {
	eventID       uuid.UUID
	eventType     string
	aggregateID   uuid.UUID
	aggregateType string
	eventVersion  int
	occurredOn    time.Time
	eventData     map[string]interface{}
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(eventType string, aggregateID uuid.UUID, aggregateType string, eventData map[string]interface{}) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:       uuid.New(),
		eventType:     eventType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		eventVersion:  1,
		occurredOn:    time.Now(),
		eventData:     eventData,
	}
}

// EventID returns the event ID
func (e BaseDomainEvent) EventID() uuid.UUID {
	return e.eventID
}

// EventType returns the event type
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// AggregateID returns the aggregate ID
func (e BaseDomainEvent) AggregateID() uuid.UUID {
	return e.aggregateID
}

// AggregateType returns the aggregate type
func (e BaseDomainEvent) AggregateType() string {
	return e.aggregateType
}

// EventVersion returns the event version
func (e BaseDomainEvent) EventVersion() int {
	return e.eventVersion
}

// OccurredOn returns when the event occurred
func (e BaseDomainEvent) OccurredOn() time.Time {
	return e.occurredOn
}

// EventData returns the event data
func (e BaseDomainEvent) EventData() map[string]interface{} {
	return e.eventData
}