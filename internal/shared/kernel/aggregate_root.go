package kernel

import (
	"time"

	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/domain/events"
)

// AggregateRoot represents the base aggregate root
type AggregateRoot struct {
	id        uuid.UUID
	version   int
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
	events    []events.DomainEvent
}

// NewAggregateRoot creates a new aggregate root
func NewAggregateRoot(id uuid.UUID) AggregateRoot {
	now := time.Now()
	return AggregateRoot{
		id:        id,
		version:   1,
		createdAt: now,
		updatedAt: now,
		events:    make([]events.DomainEvent, 0),
	}
}

// ID returns the aggregate ID
func (ar *AggregateRoot) ID() uuid.UUID {
	return ar.id
}

// Version returns the aggregate version
func (ar *AggregateRoot) Version() int {
	return ar.version
}

// CreatedAt returns when the aggregate was created
func (ar *AggregateRoot) CreatedAt() time.Time {
	return ar.createdAt
}

// UpdatedAt returns when the aggregate was last updated
func (ar *AggregateRoot) UpdatedAt() time.Time {
	return ar.updatedAt
}

// DeletedAt returns when the aggregate was deleted (if soft deleted)
func (ar *AggregateRoot) DeletedAt() *time.Time {
	return ar.deletedAt
}

// IsDeleted checks if the aggregate is soft deleted
func (ar *AggregateRoot) IsDeleted() bool {
	return ar.deletedAt != nil
}

// IncrementVersion increments the aggregate version and updates the timestamp
func (ar *AggregateRoot) IncrementVersion() {
	ar.version++
	ar.updatedAt = time.Now()
}

// SoftDelete marks the aggregate as deleted
func (ar *AggregateRoot) SoftDelete() {
	now := time.Now()
	ar.deletedAt = &now
	ar.IncrementVersion()
}

// AddEvent adds a domain event to the aggregate
func (ar *AggregateRoot) AddEvent(event events.DomainEvent) {
	ar.events = append(ar.events, event)
}

// GetEvents returns the domain events
func (ar *AggregateRoot) GetEvents() []events.DomainEvent {
	return ar.events
}

// ClearEvents clears the domain events
func (ar *AggregateRoot) ClearEvents() {
	ar.events = make([]events.DomainEvent, 0)
}