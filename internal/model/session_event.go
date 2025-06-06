package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vasu74/Call_Session_Management/internal/config"
)

// EventMetadata represents the flexible metadata structure for session events
type EventMetadata map[string]interface{}

// Value implements the driver.Valuer interface for EventMetadata
func (m EventMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for EventMetadata
func (m *EventMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(EventMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, m)
}

// SessionEvent represents an event that occurred during a session
type SessionEvent struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	SessionID uuid.UUID     `json:"session_id" db:"session_id"`
	EventType string        `json:"event_type" db:"event_type"`
	EventTime time.Time     `json:"event_time" db:"event_time"`
	Metadata  EventMetadata `json:"metadata" db:"metadata"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
}

// LogEventRequest represents the request body for logging a session event
type LogEventRequest struct {
	EventType string        `json:"event_type" binding:"required"`
	EventTime time.Time     `json:"event_time" binding:"required"`
	Metadata  EventMetadata `json:"metadata"`
}

// LogEvent creates a new session event with the given request data
func (e *SessionEvent) LogEvent(sessionID string, req LogEventRequest) error {
	// First verify the session exists and is not ended
	var status SessionStatus
	err := config.DB.QueryRow("SELECT status FROM sessions WHERE id = $1", sessionID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("session not found")
		}
		return err
	}

	if status != SessionStatusOngoing {
		return errors.New("cannot log events for ended session")
	}

	// Create event
	e.ID = uuid.New()
	e.SessionID = uuid.MustParse(sessionID)
	e.EventType = req.EventType
	e.EventTime = req.EventTime
	e.Metadata = req.Metadata
	e.CreatedAt = time.Now()

	// Insert into database
	query := `
		INSERT INTO session_events (id, session_id, event_type, event_time, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, session_id, event_type, event_time, metadata, created_at`

	err = config.DB.QueryRow(
		query,
		e.ID, e.SessionID, e.EventType, e.EventTime, e.Metadata, e.CreatedAt,
	).Scan(&e.ID, &e.SessionID, &e.EventType, &e.EventTime, &e.Metadata, &e.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}
