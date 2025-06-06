package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vasu74/Call_Session_Management/internal/config"
)

// SessionStatus represents the possible states of a session
type SessionStatus string

const (
	SessionStatusOngoing   SessionStatus = "ongoing"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
)

// SessionMetadata represents the flexible metadata structure for sessions
type SessionMetadata map[string]interface{}

// Value implements the driver.Valuer interface for SessionMetadata
func (m SessionMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for SessionMetadata
func (m *SessionMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(SessionMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, m)
}

// Session represents a call session in the system
type Session struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	StartedAt       time.Time       `json:"started_at" db:"started_at"`
	EndedAt         *time.Time      `json:"ended_at,omitempty" db:"ended_at"`
	CallerID        string          `json:"caller_id" db:"caller_id"`
	CalleeID        string          `json:"callee_id" db:"callee_id"`
	Status          SessionStatus   `json:"status" db:"status"`
	InitialMetadata SessionMetadata `json:"initial_metadata" db:"initial_metadata"`
	Disposition     *string         `json:"disposition,omitempty" db:"disposition"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// StartSessionRequest represents the request body for starting a new session
type StartSessionRequest struct {
	CallerID        string          `json:"caller_id" binding:"required"`
	CalleeID        string          `json:"callee_id" binding:"required"`
	InitialMetadata SessionMetadata `json:"initial_metadata"`
}

// EndSessionRequest represents the request body for ending a session
type EndSessionRequest struct {
	Status      SessionStatus `json:"status" binding:"required,oneof=completed failed"`
	Disposition string        `json:"disposition" binding:"required"`
	EndTime     time.Time     `json:"end_time" binding:"required"`
}

// SessionListResponse represents the paginated response for listing sessions
type SessionListResponse struct {
	Total    int64     `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
	Sessions []Session `json:"sessions"`
}

// SessionFilter represents the filter parameters for listing sessions
type SessionFilter struct {
	StartDate *time.Time    `form:"start_date"`
	EndDate   *time.Time    `form:"end_date"`
	Status    SessionStatus `form:"status"`
	CallerID  string        `form:"caller_id"`
	CalleeID  string        `form:"callee_id"`
	Limit     int           `form:"limit,default=50"`
	Offset    int           `form:"offset,default=0"`
	SortBy    string        `form:"sort_by,default=started_at"`
	SortOrder string        `form:"sort_order,default=desc"`
}

// SessionDetails represents the detailed view of a session with its events
type SessionDetails struct {
	Session Session        `json:"session"`
	Events  []SessionEvent `json:"events"`
}

// StartSession creates a new session with the given request data
func (s *Session) StartSession(req StartSessionRequest) error {
	now := time.Now()
	s.ID = uuid.New()
	s.StartedAt = now
	s.CallerID = req.CallerID
	s.CalleeID = req.CalleeID
	s.Status = SessionStatusOngoing
	s.InitialMetadata = req.InitialMetadata
	s.CreatedAt = now
	s.UpdatedAt = now

	// Insert into database
	query := `
		INSERT INTO sessions (id, started_at, caller_id, callee_id, status, initial_metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, started_at, caller_id, callee_id, status, initial_metadata, created_at, updated_at`

	err := config.DB.QueryRow(
		query,
		s.ID, s.StartedAt, s.CallerID, s.CalleeID, s.Status, s.InitialMetadata, s.CreatedAt, s.UpdatedAt,
	).Scan(&s.ID, &s.StartedAt, &s.CallerID, &s.CalleeID, &s.Status, &s.InitialMetadata, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// EndSession marks the session as ended with the given status and disposition
func (s *Session) EndSession(sessionID string, req EndSessionRequest) error {
	// First get the session
	query := `SELECT id, started_at, caller_id, callee_id, status, initial_metadata, created_at, updated_at 
		FROM sessions WHERE id = $1`

	err := config.DB.QueryRow(query, sessionID).Scan(
		&s.ID, &s.StartedAt, &s.CallerID, &s.CalleeID, &s.Status, &s.InitialMetadata, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("session not found")
		}
		return err
	}

	// Check if session is already ended
	if s.Status != SessionStatusOngoing {
		return fmt.Errorf("session is already ended with status: %s", s.Status)
	}

	// Update session
	updateQuery := `
		UPDATE sessions 
		SET status = $1, disposition = $2, ended_at = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND status = 'ongoing'
		RETURNING id, started_at, ended_at, caller_id, callee_id, status, initial_metadata, disposition, created_at, updated_at`

	err = config.DB.QueryRow(
		updateQuery,
		req.Status, req.Disposition, req.EndTime, sessionID,
	).Scan(
		&s.ID, &s.StartedAt, &s.EndedAt, &s.CallerID, &s.CalleeID, &s.Status, &s.InitialMetadata, &s.Disposition, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("session could not be ended - it may have been ended by another request")
		}
		return err
	}

	return nil
}

// GetSessionDetails retrieves a session and its events
func GetSessionDetails(sessionID string) (*SessionDetails, error) {
	var details SessionDetails

	// Get session
	query := `SELECT id, started_at, ended_at, caller_id, callee_id, status, initial_metadata, disposition, created_at, updated_at 
		FROM sessions WHERE id = $1`

	err := config.DB.QueryRow(query, sessionID).Scan(
		&details.Session.ID, &details.Session.StartedAt, &details.Session.EndedAt,
		&details.Session.CallerID, &details.Session.CalleeID, &details.Session.Status,
		&details.Session.InitialMetadata, &details.Session.Disposition,
		&details.Session.CreatedAt, &details.Session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	// Get events
	eventsQuery := `SELECT id, session_id, event_type, event_time, metadata, created_at 
		FROM session_events WHERE session_id = $1 ORDER BY event_time ASC`

	rows, err := config.DB.Query(eventsQuery, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event SessionEvent
		err := rows.Scan(
			&event.ID, &event.SessionID, &event.EventType, &event.EventTime,
			&event.Metadata, &event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		details.Events = append(details.Events, event)
	}

	return &details, nil
}

// ListSessions retrieves sessions based on filter criteria
func ListSessions(filter SessionFilter) (*SessionListResponse, error) {
	var response SessionListResponse
	response.Limit = filter.Limit
	response.Offset = filter.Offset

	// Build query
	query := `SELECT id, started_at, ended_at, caller_id, callee_id, status, initial_metadata, disposition, created_at, updated_at 
		FROM sessions WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND started_at >= $%d", argCount)
		args = append(args, filter.StartDate)
		argCount++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND started_at <= $%d", argCount)
		args = append(args, filter.EndDate)
		argCount++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}
	if filter.CallerID != "" {
		query += fmt.Sprintf(" AND caller_id = $%d", argCount)
		args = append(args, filter.CallerID)
		argCount++
	}
	if filter.CalleeID != "" {
		query += fmt.Sprintf(" AND callee_id = $%d", argCount)
		args = append(args, filter.CalleeID)
		argCount++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", query)
	err := config.DB.QueryRow(countQuery, args...).Scan(&response.Total)
	if err != nil {
		return nil, err
	}

	// Add sorting and pagination
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d",
		filter.SortBy, filter.SortOrder, argCount, argCount+1)
	args = append(args, filter.Limit, filter.Offset)

	// Get sessions
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID, &session.StartedAt, &session.EndedAt,
			&session.CallerID, &session.CalleeID, &session.Status,
			&session.InitialMetadata, &session.Disposition,
			&session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		response.Sessions = append(response.Sessions, session)
	}

	return &response, nil
}
