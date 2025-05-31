# API Documentation

This document provides comprehensive documentation for the Football Manager Data Browser (FMDB) REST API.

## Base URL

- **Development**: `http://localhost:8091/api`
- **Production**: `https://your-domain.com/api`

## Authentication

Currently, the API does not require authentication. All endpoints are publicly accessible.

## Content Types

- **Request**: `application/json` or `multipart/form-data` (for file uploads)
- **Response**: `application/json`

## Common Response Format

All API responses follow a consistent structure:

```json
{
  "status": "success|error",
  "message": "Human readable message",
  "data": {}, // Response data (varies by endpoint)
  "error": "Error details (only present on errors)"
}
```

## Endpoints

### 1. Health Check

Check the health status of the API server.

```http
GET /api/health
```

**Response:**
```json
{
  "status": "success",
  "message": "Server is healthy",
  "data": {
    "uptime": "2h34m12s",
    "memory_usage": "45.2 MB",
    "goroutines": 12,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### 2. File Upload

Upload a Football Manager HTML export file for processing.

```http
POST /api/upload
Content-Type: multipart/form-data
```

**Parameters:**
- `file` (required): HTML file from Football Manager export

**Response:**
```json
{
  "status": "success",
  "message": "File uploaded and processed successfully",
  "data": {
    "upload_id": "uuid-string",
    "players_count": 2847,
    "processing_time": "1.23s",
    "file_size": "2.4 MB"
  }
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Invalid file format",
  "error": "File must be an HTML export from Football Manager"
}
```

### 3. Get All Players

Retrieve all players with optional filtering and pagination.

```http
GET /api/players
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 50, max: 1000)
- `sort` (string): Sort field (default: "name")
- `order` (string): Sort order - "asc" or "desc" (default: "asc")
- `position` (string): Filter by position (e.g., "ST", "CM", "CB")
- `club` (string): Filter by club name
- `nationality` (string): Filter by nationality
- `min_age` (int): Minimum age filter
- `max_age` (int): Maximum age filter
- `min_overall` (int): Minimum overall rating
- `max_overall` (int): Maximum overall rating
- `min_value` (string): Minimum transfer value (e.g., "1M", "500K")
- `max_value` (string): Maximum transfer value
- `search` (string): Search player names

**Example:**
```http
GET /api/players?position=ST&min_overall=80&limit=20&sort=overall&order=desc
```

**Response:**
```json
{
  "status": "success",
  "message": "Players retrieved successfully",
  "data": {
    "players": [
      {
        "id": 1,
        "name": "Lionel Messi",
        "age": 36,
        "position": "RW, CAM, CF",
        "club": "Inter Miami",
        "nationality": "Argentina",
        "overall": 91,
        "potential": 91,
        "transfer_value": "€30M",
        "wage": "€500K",
        "attributes": {
          "pace": 85,
          "shooting": 92,
          "passing": 95,
          "dribbling": 96,
          "defending": 35,
          "physical": 68
        },
        "performance_stats": {
          "goals": 25,
          "assists": 15,
          "apps": 30
        }
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 143,
      "total_players": 2847,
      "players_per_page": 20
    },
    "filters_applied": {
      "position": "ST",
      "min_overall": 80
    }
  }
}
```

### 4. Get Player by ID

Retrieve detailed information for a specific player.

```http
GET /api/players/{id}
```

**Parameters:**
- `id` (int): Player ID

**Response:**
```json
{
  "status": "success",
  "message": "Player retrieved successfully",
  "data": {
    "player": {
      "id": 1,
      "name": "Lionel Messi",
      "age": 36,
      "position": "RW, CAM, CF",
      "parsed_positions": ["RW", "CAM", "CF"],
      "position_groups": ["Forward", "Midfielder"],
      "club": "Inter Miami",
      "nationality": "Argentina",
      "overall": 91,
      "potential": 91,
      "fifa_overall": 89,
      "transfer_value": "€30M",
      "transfer_value_raw": 30000000,
      "wage": "€500K",
      "wage_raw": 500000,
      "contract_expires": "2025-06-30",
      "attributes": {
        "pace": 85,
        "shooting": 92,
        "passing": 95,
        "dribbling": 96,
        "defending": 35,
        "physical": 68,
        "technical": {
          "corners": 90,
          "crossing": 85,
          "dribbling": 96,
          "finishing": 92,
          "first_touch": 95,
          "free_kick_taking": 95,
          "heading": 70,
          "long_shots": 90,
          "long_throws": 8,
          "marking": 20,
          "passing": 95,
          "penalty_taking": 85,
          "tackling": 25,
          "technique": 96
        },
        "mental": {
          "aggression": 65,
          "anticipation": 92,
          "bravery": 75,
          "composure": 96,
          "concentration": 85,
          "decisions": 95,
          "determination": 85,
          "flair": 98,
          "leadership": 90,
          "off_the_ball": 92,
          "positioning": 85,
          "teamwork": 90,
          "vision": 95,
          "work_rate": 75
        },
        "physical": {
          "acceleration": 85,
          "agility": 92,
          "balance": 95,
          "jumping_reach": 68,
          "natural_fitness": 85,
          "pace": 85,
          "stamina": 75,
          "strength": 65
        }
      },
      "performance_stats": {
        "apps": 30,
        "goals": 25,
        "assists": 15,
        "yellow_cards": 2,
        "red_cards": 0,
        "minutes": 2700,
        "rating": 7.8
      }
    }
  }
}
```

### 5. Search Players

Advanced search endpoint with multiple criteria.

```http
POST /api/players/search
Content-Type: application/json
```

**Request Body:**
```json
{
  "search_term": "messi",
  "filters": {
    "positions": ["RW", "CAM", "CF"],
    "clubs": ["Inter Miami", "PSG"],
    "nationalities": ["Argentina"],
    "age_range": {
      "min": 30,
      "max": 40
    },
    "overall_range": {
      "min": 85,
      "max": 95
    },
    "value_range": {
      "min": "10M",
      "max": "100M"
    },
    "attributes": {
      "dribbling": {
        "min": 90
      },
      "passing": {
        "min": 85
      }
    }
  },
  "sort": {
    "field": "overall",
    "order": "desc"
  },
  "pagination": {
    "page": 1,
    "limit": 50
  }
}
```

**Response:**
Same format as GET /api/players but with search results.

### 6. Player Comparison

Compare multiple players side by side.

```http
POST /api/players/compare
Content-Type: application/json
```

**Request Body:**
```json
{
  "player_ids": [1, 2, 3]
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Players comparison retrieved successfully",
  "data": {
    "players": [
      // Array of player objects (same format as single player)
    ],
    "comparison_metrics": {
      "strongest_attributes": {
        "1": ["dribbling", "passing", "vision"],
        "2": ["pace", "shooting", "finishing"],
        "3": ["defending", "heading", "strength"]
      },
      "position_compatibility": {
        "common_positions": ["CAM"],
        "unique_positions": {
          "1": ["RW", "CF"],
          "2": ["ST"],
          "3": ["CB", "CDM"]
        }
      }
    }
  }
}
```

### 7. Statistics Endpoint

Get aggregated statistics about the dataset.

```http
GET /api/stats
```

**Response:**
```json
{
  "status": "success",
  "message": "Statistics retrieved successfully",
  "data": {
    "overview": {
      "total_players": 2847,
      "total_clubs": 156,
      "total_nationalities": 89,
      "average_age": 24.3,
      "average_overall": 67.8
    },
    "positions": {
      "GK": 142,
      "CB": 425,
      "LB": 178,
      "RB": 165,
      "CDM": 234,
      "CM": 398,
      "CAM": 187,
      "LW": 156,
      "RW": 143,
      "ST": 289
    },
    "age_distribution": {
      "16-20": 587,
      "21-25": 1234,
      "26-30": 756,
      "31-35": 234,
      "36+": 36
    },
    "overall_distribution": {
      "50-59": 423,
      "60-69": 1567,
      "70-79": 756,
      "80-89": 98,
      "90+": 3
    },
    "top_clubs": [
      {
        "name": "Manchester City",
        "player_count": 45,
        "average_overall": 78.2
      }
    ],
    "top_nationalities": [
      {
        "name": "England",
        "player_count": 234,
        "average_overall": 69.1
      }
    ]
  }
}
```

### 8. Export Data

Export filtered player data in various formats.

```http
POST /api/export
Content-Type: application/json
```

**Request Body:**
```json
{
  "format": "csv|json|xlsx",
  "filters": {
    // Same filter structure as search endpoint
  },
  "fields": [
    "name",
    "age",
    "position",
    "club",
    "overall",
    "transfer_value"
  ]
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Export prepared successfully",
  "data": {
    "download_url": "/api/downloads/export-uuid.csv",
    "expires_at": "2024-01-15T11:30:00Z",
    "record_count": 156
  }
}
```

## Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Malformed request body or parameters |
| 404 | `PLAYER_NOT_FOUND` | Player with specified ID not found |
| 413 | `FILE_TOO_LARGE` | Uploaded file exceeds size limit |
| 415 | `UNSUPPORTED_FORMAT` | File format not supported |
| 422 | `VALIDATION_ERROR` | Request validation failed |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Server temporarily unavailable |

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Default**: 100 requests per minute per IP
- **Upload**: 5 uploads per hour per IP
- **Export**: 10 exports per hour per IP

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

## Examples

### Basic Player Search

```javascript
// Search for strikers over 80 overall
const response = await fetch('/api/players?position=ST&min_overall=80');
const data = await response.json();
console.log(data.data.players);
```

### Upload and Process File

```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const response = await fetch('/api/upload', {
  method: 'POST',
  body: formData
});

const result = await response.json();
if (result.status === 'success') {
  console.log(`Processed ${result.data.players_count} players`);
}
```

### Advanced Search

```javascript
const searchQuery = {
  search_term: "ronaldo",
  filters: {
    positions: ["ST", "RW"],
    age_range: { min: 25, max: 40 },
    overall_range: { min: 85 }
  },
  sort: { field: "overall", order: "desc" }
};

const response = await fetch('/api/players/search', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(searchQuery)
});
```

## WebSocket Events (Future)

The API is designed to support real-time features via WebSocket connections:

- `upload_progress` - File upload and processing progress
- `data_updated` - When new data is available
- `export_ready` - When export file is ready for download

---

For questions or issues with the API, please refer to the [troubleshooting guide](TROUBLESHOOTING.md) or open an issue on the project repository. 