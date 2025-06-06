package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vasu74/Call_Session_Management/internal/model"
)

func StartSessionHandler(c *gin.Context) {
	var req model.StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session model.Session
	if err := session.StartSession(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Session started successfully",
		"session": session,
	})
}

func LogSessionEventHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	var req model.LogEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event model.SessionEvent
	if err := event.LogEvent(sessionID, req); err != nil {
		if err.Error() == "session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		// Check for event_time constraint violation
		if strings.Contains(err.Error(), "valid_event_time") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "event_time must be within the last year",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event logged successfully",
		"event":   event,
	})
}

func EndSessionHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	var req model.EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var session model.Session
	if err := session.EndSession(sessionID, req); err != nil {
		switch err.Error() {
		case "session not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "session is already ended with status: completed",
			"session is already ended with status: failed",
			"session could not be ended - it may have been ended by another request":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			// Add this block:
			if strings.Contains(err.Error(), "valid_session_times") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "end_time must be after or equal to started_at",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session ended successfully",
		"session": session,
	})
}

func GetSessionDetailsHandler(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	details, err := model.GetSessionDetails(sessionID)
	if err != nil {
		if err.Error() == "session not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

func ListSessionsHandler(c *gin.Context) {
	var filter model.SessionFilter

	// Parse query parameters
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = &t
		}
	}
	if status := c.Query("status"); status != "" {
		filter.Status = model.SessionStatus(status)
	}
	if callerID := c.Query("caller_id"); callerID != "" {
		filter.CallerID = callerID
	}
	if calleeID := c.Query("callee_id"); calleeID != "" {
		filter.CalleeID = calleeID
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Validate status if provided
	if filter.Status != "" && filter.Status != model.SessionStatusOngoing && filter.Status != model.SessionStatusCompleted && filter.Status != model.SessionStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	// Get sessions
	sessions, err := model.ListSessions(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
