# API Documentation

## Base URL

```
http://localhost:8080
```

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. To access protected endpoints:

1. Register a new user or login to get a JWT token
2. Include the token in the Authorization header: `Authorization: Bearer <token>`

### Authentication Endpoints

#### Register a New User

```http
POST /auth/register
```

Creates a new user account.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (201 Created):**

```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user",
    "created_at": "2024-03-20T10:00:00Z"
  }
}
```

**Error Responses:**

- `400 Bad Request`: Invalid request body
- `409 Conflict`: User already exists
- `500 Internal Server Error`: Server error

#### Login

```http
POST /auth/login
```

Authenticates a user and returns a JWT token.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user",
    "created_at": "2024-03-20T10:00:00Z"
  }
}
```

**Error Responses:**

- `400 Bad Request`: Invalid request body
- `401 Unauthorized`: Invalid credentials
- `500 Internal Server Error`: Server error

#### Get User Profile

```http
GET /api/profile
```

Retrieves the current user's profile information.

**Headers:**

- `Authorization: Bearer <token>`

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "user",
  "created_at": "2024-03-20T10:00:00Z",
  "updated_at": "2024-03-20T10:00:00Z"
}
```

**Error Responses:**

- `401 Unauthorized`: Missing or invalid token
- `500 Internal Server Error`: Server error

## API Endpoints

### Sessions

#### Start a New Session

```http
POST api/sessions/start
```

Creates a new call session.

**Request Body:**

```json
{
  "caller_id": "user123",
  "callee_id": "user456",
  "initial_metadata": {
    "call_type": "voice",
    "priority": "high",
    "notes": "Initial customer support call"
  }
}
```

**Response (201 Created):**

```json
{
  "message": "Session started successfully",
  "session": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "started_at": "2024-03-20T10:00:00Z",
    "caller_id": "user123",
    "callee_id": "user456",
    "status": "ongoing",
    "initial_metadata": {
      "call_type": "voice",
      "priority": "high",
      "notes": "Initial customer support call"
    },
    "created_at": "2024-03-20T10:00:00Z",
    "updated_at": "2024-03-20T10:00:00Z"
  }
}
```

**Error Responses:**

- `400 Bad Request`: Invalid request body
- `500 Internal Server Error`: Server error

#### Log Session Event

```http
POST api/sessions/{sessionId}/events
```

Logs an event for an existing session.

**Path Parameters:**

- `sessionId` (UUID): ID of the session

**Request Body:**

```json
{
  "event_type": "assistant_response",
  "event_time": "2025-03-20T10:01:00Z",
  "metadata": {
    "response_type": "text",
    "content": "Hello, how can I help you today?",
    "confidence": 0.95
  }
}
```

**Response (201 Created):**

```json
{
  "message": "Event logged successfully",
  "event": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "event_type": "assistant_response",
    "event_time": "2024-03-20T10:01:00Z",
    "metadata": {
      "response_type": "text",
      "content": "Hello, how can I help you today?",
      "confidence": 0.95
    },
    "created_at": "2024-03-20T10:01:00Z"
  }
}
```

**Error Responses:**

- `400 Bad Request`: Invalid request body or event time
- `404 Not Found`: Session not found
- `500 Internal Server Error`: Server error

#### End Session

```http
POST api/sessions/{sessionId}/end
```

Ends an ongoing session.

**Path Parameters:**

- `sessionId` (UUID): ID of the session

**Request Body:**

```json
{
  "status": "completed",
  "disposition": "successful_resolution",
  "end_time": "2025-06-27T10:30:00Z"
}
```

**Response (200 OK):**

```json
{
  "message": "Session ended successfully",
  "session": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "started_at": "2025-06-06 08:40:43.749867",
    "ended_at": "2025-06-27 10:30:00",
    "caller_id": "user123",
    "callee_id": "user456",
    "status": "completed",
    "disposition": "successful_resolution",
    "initial_metadata": {
      "call_type": "voice",
      "priority": "high",
      "notes": "Initial customer support call"
    },
    "created_at": "2025-06-06 14:07:02.277883",
    "updated_at": "2025-06-06 08:40:43.749867"
  }
}
```

**Error Responses:**

- `400 Bad Request`: Invalid request body or end time
- `404 Not Found`: Session not found
- `409 Conflict`: Session already ended
- `500 Internal Server Error`: Server error

#### Get Session Details

```http
GET /sessions/{sessionId}
```

Retrieves detailed information about a session, including all events.

**Path Parameters:**

- `sessionId` (UUID): ID of the session

**Response (200 OK):**

```json
{
  "session": {
    "id": "5d5f318c-27e5-4004-a8a2-5bb685e7de17",
    "started_at": "2025-06-06T14:07:02.277883Z",
    "ended_at": "2025-06-27T10:30:00Z",
    "caller_id": "user123",
    "callee_id": "user456",
    "status": "completed",
    "initial_metadata": {
      "notes": "Initial customer support call",
      "priority": "high",
      "call_type": "voice"
    },
    "disposition": "successful_resolution",
    "created_at": "2025-06-06T14:07:02.277883Z",
    "updated_at": "2025-06-06T08:40:43.749867Z"
  },
  "events": [
    {
      "id": "e2c98260-6363-444b-bcc3-26cc7f02f47e",
      "session_id": "5d5f318c-27e5-4004-a8a2-5bb685e7de17",
      "event_type": "assistant_response",
      "event_time": "2025-03-20T10:01:00Z",
      "metadata": {
        "content": "Hello, how can I help you today?",
        "confidence": 0.95,
        "response_type": "text"
      },
      "created_at": "2025-06-06T14:08:41.997798Z"
    }
  ]
}
```

**Error Responses:**

- `404 Not Found`: Session not found
- `500 Internal Server Error`: Server error

#### List Sessions

```http
GET /api/sessions
```

Lists sessions with optional filtering and pagination.

**Query Parameters:**

- `start_date` (ISO8601): Filter sessions started after this date
- `end_date` (ISO8601): Filter sessions started before this date
- `status` (enum): Filter by status (ongoing, completed, failed)
- `caller_id` (string): Filter by caller ID
- `callee_id` (string): Filter by callee ID
- `limit` (integer, default: 50): Number of results per page
- `offset` (integer, default: 0): Pagination offset
- `sort_by` (string, default: started_at): Sort field
- `sort_order` (string, default: desc): Sort order (asc/desc)

**Response (200 OK):**

```json
{
  "total": 100,
  "limit": 50,
  "offset": 0,
  "sessions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "started_at": "2024-03-20T10:00:00Z",
      "ended_at": "2024-03-20T10:30:00Z",
      "caller_id": "user123",
      "callee_id": "user456",
      "status": "completed",
      "disposition": "successful_resolution",
      "initial_metadata": {
        "call_type": "voice",
        "priority": "high"
      },
      "created_at": "2024-03-20T10:00:00Z",
      "updated_at": "2024-03-20T10:30:00Z"
    }
  ]
}
```

**Error Responses:**

- `400 Bad Request`: Invalid query parameters
- `500 Internal Server Error`: Server error

## Data Types

### Session Status

Enum values:

- `ongoing`: Session is currently active
- `completed`: Session ended successfully
- `failed`: Session ended with an error

### Timestamps

All timestamps are in ISO8601 format with UTC timezone.

### Metadata

Both `initial_metadata` and event `metadata` are stored as JSONB, allowing flexible schema:

```json
{
  "key1": "value1",
  "key2": 123,
  "key3": {
    "nested": "value"
  },
  "key4": ["array", "of", "values"]
}
```

## Rate Limiting

Currently, no rate limiting is implemented. Future versions will include rate limiting based on IP and API key.

## Error Handling

All errors follow this format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:

- `200 OK`: Successful operation
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict
- `500 Internal Server Error`: Server error

## Versioning

The current API version is v1. Future versions will be available under `/v2`, `/v3`, etc.

## Support

For API support, please contact the development team or create an issue in the repository.

## Protected Endpoints

All endpoints under `/api/*` require authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

### Error Responses for Protected Endpoints

- `401 Unauthorized`:
  - Missing Authorization header
  - Invalid token format
  - Expired token
  - Invalid token
- `403 Forbidden`: Insufficient permissions (for role-based access)
