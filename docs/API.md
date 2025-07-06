# FM-Dash API Documentation

This document provides comprehensive documentation for the FM-Dash REST API, including all endpoints, specialized features, and advanced functionality.

## Base URL

- **Development**: `http://localhost:8091/api`
- **Production**: `https://your-domain.com/api`

## Authentication

Currently, the API does not require authentication. All endpoints are publicly accessible for development and personal use.

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
  "error": "Error details (only present on errors)",
  "request_id": "unique-request-identifier",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Core Data Endpoints

### 1. Health Check

Monitor API server health and performance metrics.

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
    "memory_limit": "1GB",
    "goroutines": 12,
    "cpu_usage": "2.3%",
    "disk_usage": "15.6 GB",
    "database_connections": 5,
    "cache_hit_rate": "94.2%",
    "version": "1.3.0",
    "build_time": "2024-01-15T08:00:00Z",
    "go_version": "go1.24.3",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### 2. File Upload and Processing

Upload Football Manager HTML export files for comprehensive analysis.

```http
POST /api/upload
Content-Type: multipart/form-data
```

**Parameters:**
- `file` (required): HTML file from Football Manager export
- `processing_options` (optional): JSON string with processing preferences

**Processing Options:**
```json
{
  "enable_fifa_ratings": true,
  "calculate_percentiles": true,
  "extract_team_logos": true,
  "include_historical_data": false,
  "performance_mode": "balanced" // "fast", "balanced", "accurate"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "File uploaded and processed successfully",
  "data": {
    "upload_id": "550e8400-e29b-41d4-a716-446655440000",
    "players_count": 2847,
    "teams_count": 156,
    "leagues_count": 23,
    "processing_time": "1.23s",
    "file_size": "2.4 MB",
    "fifa_ratings_calculated": true,
    "percentiles_analyzed": true,
    "team_logos_matched": 134,
    "data_quality_score": 96.8,
    "estimated_accuracy": "high",
    "processing_metadata": {
      "parser_version": "2.1.4",
      "algorithms_used": ["fifa_v2", "percentile_enhanced"],
      "worker_count": 8,
      "memory_peak": "120MB"
    }
  }
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Invalid file format",
  "error": "File must be an HTML export from Football Manager",
  "error_code": "INVALID_FILE_FORMAT",
  "suggested_action": "Export data from Football Manager in HTML format"
}
```

### 3. Player Data Retrieval

Comprehensive player data with advanced filtering and search capabilities.

```http
GET /api/players
```

**Query Parameters:**

**Pagination:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 50, max: 1000)

**Sorting:**
- `sort` (string): Sort field (default: "name")
  - Available: `name`, `age`, `overall`, `potential`, `transfer_value`, `wage`, `fifa_overall`
- `order` (string): Sort order - "asc" or "desc" (default: "asc")

**Basic Filters:**
- `position` (string): Filter by position (e.g., "ST", "CM", "CB")
- `club` (string): Filter by club name
- `nationality` (string): Filter by nationality
- `league` (string): Filter by league name

**Advanced Filters:**
- `min_age` (int): Minimum age filter
- `max_age` (int): Maximum age filter
- `min_overall` (int): Minimum overall rating (1-100)
- `max_overall` (int): Maximum overall rating (1-100)
- `min_potential` (int): Minimum potential rating (1-100)
- `max_potential` (int): Maximum potential rating (1-100)
- `min_value` (string): Minimum transfer value (e.g., "1M", "500K")
- `max_value` (string): Maximum transfer value
- `min_wage` (string): Minimum wage (e.g., "50K", "100K")
- `max_wage` (string): Maximum wage

**Specialized Filters:**
- `wonderkids` (bool): Players aged ≤21 with potential ≥80
- `bargains` (bool): Players with high value-to-cost ratio
- `free_agents` (bool): Players without contracts
- `loan_listed` (bool): Players available on loan
- `injury_prone` (bool): Players with injury history
- `form` (string): Current form - "excellent", "good", "average", "poor"

**Search:**
- `search` (string): Search player names, clubs, and nationalities
- `search_mode` (string): "exact", "fuzzy", "advanced" (default: "fuzzy")

**Response Fields:**
- `include_attributes` (bool): Include detailed attribute breakdown
- `include_statistics` (bool): Include performance statistics
- `include_fifa_ratings` (bool): Include FIFA-style ratings
- `include_percentiles` (bool): Include percentile rankings
- `include_market_analysis` (bool): Include transfer market analysis

**Example Request:**
```http
GET /api/players?position=ST&min_overall=80&max_age=25&limit=20&sort=potential&order=desc&include_fifa_ratings=true&include_percentiles=true
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
        "name": "Erling Haaland",
        "age": 23,
        "position": "ST",
        "parsed_positions": ["ST", "CF"],
        "position_groups": ["Forward"],
        "club": "Manchester City",
        "nationality": "Norway",
        "league": "Premier League",
        "overall": 91,
        "potential": 95,
        "fifa_overall": 89,
        "transfer_value": "€180M",
        "transfer_value_raw": 180000000,
        "wage": "€400K",
        "wage_raw": 400000,
        "contract_expires": "2027-06-30",
        "market_value_trend": "rising",
        "value_ratio": 2.1,
        "attributes": {
          "pace": 89,
          "shooting": 94,
          "passing": 65,
          "dribbling": 80,
          "defending": 45,
          "physical": 92,
          "detailed_attributes": {
            "finishing": 96,
            "shot_power": 93,
            "acceleration": 91,
            "sprint_speed": 89,
            "strength": 94,
            "jumping": 88
          }
        },
        "fifa_ratings": {
          "pace": 89,
          "shooting": 91,
          "passing": 65,
          "dribbling": 80,
          "defending": 49,
          "physical": 88,
          "overall": 89
        },
        "percentiles": {
          "overall": {
            "league": 98,
            "position": 96,
            "age_group": 99,
            "global": 92
          },
          "pace": {
            "league": 85,
            "position": 88,
            "age_group": 87
          }
        },
        "performance_stats": {
          "goals": 36,
          "assists": 8,
          "apps": 38,
          "minutes": 3240,
          "goals_per_90": 1.0,
          "xg": 32.4,
          "conversion_rate": 23.5
        },
        "injury_history": {
          "days_injured": 12,
          "injury_count": 1,
          "injury_prone": false
        },
        "market_analysis": {
          "estimated_value": "€175M",
          "value_confidence": "high",
          "market_position": "elite",
          "comparable_players": ["Mbappe", "Benzema"],
          "transfer_likelihood": "low"
        }
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 12,
      "total_players": 234,
      "players_per_page": 20,
      "has_next": true,
      "has_previous": false
    },
    "filters_applied": {
      "position": "ST",
      "min_overall": 80,
      "max_age": 25
    },
    "aggregations": {
      "average_age": 22.3,
      "average_overall": 83.7,
      "total_market_value": "€2.1B",
      "nationality_breakdown": {
        "England": 45,
        "France": 32,
        "Brazil": 28
      }
    }
  }
}
```

### 4. Individual Player Details

Retrieve comprehensive information for a specific player.

```http
GET /api/players/{id}
```

**Parameters:**
- `id` (int): Player ID

**Query Parameters:**
- `include_similar` (bool): Include similar players analysis
- `include_history` (bool): Include historical performance data
- `include_projections` (bool): Include future performance projections

**Response:**
```json
{
  "status": "success",
  "message": "Player retrieved successfully",
  "data": {
    "player": {
      "id": 1,
      "name": "Lionel Messi",
      "full_name": "Lionel Andrés Messi",
      "age": 36,
      "date_of_birth": "1987-06-24",
      "position": "RW, CAM, CF",
      "parsed_positions": ["RW", "CAM", "CF"],
      "position_groups": ["Forward", "Midfielder"],
      "preferred_position": "RW",
      "club": "Inter Miami",
      "nationality": "Argentina",
      "second_nationality": null,
      "league": "MLS",
      "height": "170cm",
      "weight": "72kg",
      "preferred_foot": "Left",
      "weak_foot": 4,
      "skill_moves": 4,
      "work_rates": "Medium/Low",
      "overall": 91,
      "potential": 91,
      "fifa_overall": 89,
      "form": "Excellent",
      "morale": "Very Happy",
      "transfer_value": "€30M",
      "transfer_value_raw": 30000000,
      "wage": "€500K",
      "wage_raw": 500000,
      "contract_expires": "2025-12-31",
      "release_clause": null,
      "agent": "Jorge Messi",
      "value_trend": "declining",
      "similar_players": [
        {
          "id": 123,
          "name": "Kevin De Bruyne",
          "similarity_score": 87.3,
          "common_attributes": ["passing", "vision", "technique"]
        }
      ],
      "playing_style": {
        "primary": "Playmaker",
        "secondary": "Inside Forward",
        "traits": ["Tries Killer Balls Often", "Places Shots", "Curls Ball"]
      },
      "detailed_attributes": {
        "technical": {
          "corners": 93,
          "crossing": 85,
          "dribbling": 95,
          "finishing": 89,
          "first_touch": 96,
          "free_kicks": 94,
          "heading": 70,
          "long_shots": 90,
          "long_throws": 55,
          "marking": 22,
          "passing": 93,
          "penalty_taking": 85,
          "tackling": 35,
          "technique": 96
        },
        "mental": {
          "aggression": 48,
          "anticipation": 94,
          "bravery": 68,
          "composure": 96,
          "concentration": 93,
          "decisions": 95,
          "determination": 90,
          "flair": 97,
          "leadership": 85,
          "off_the_ball": 91,
          "positioning": 93,
          "teamwork": 92,
          "vision": 96,
          "work_rate": 85
        },
        "physical": {
          "acceleration": 85,
          "agility": 91,
          "balance": 95,
          "jumping": 68,
          "natural_fitness": 95,
          "pace": 80,
          "stamina": 88,
          "strength": 68
        }
      },
      "performance_projections": {
        "next_season": {
          "estimated_overall": 89,
          "decline_factors": ["age", "physical_attributes"],
          "projected_performance": "excellent"
        },
        "career_trajectory": {
          "peak_years_remaining": 2,
          "decline_rate": "gradual",
          "retirement_estimate": "2026-2027"
        }
      }
    }
  }
}
```

## Team and League Endpoints

### 5. Team Information

Retrieve team data with squad analysis and statistics.

```http
GET /api/teams
```

**Query Parameters:**
- `league` (string): Filter by league
- `country` (string): Filter by country
- `tier` (int): Filter by league tier (1, 2, 3, etc.)
- `include_squad` (bool): Include squad details
- `include_statistics` (bool): Include team statistics

**Response:**
```json
{
  "status": "success",
  "data": {
    "teams": [
      {
        "id": 1,
        "name": "Manchester City",
        "full_name": "Manchester City Football Club",
        "short_name": "Man City",
        "league": "Premier League",
        "country": "England",
        "tier": 1,
        "founded": 1880,
        "stadium": "Etihad Stadium",
        "capacity": 55000,
        "manager": "Pep Guardiola",
        "logo_url": "/logos/manchester_city.png",
        "squad_size": 25,
        "average_age": 26.8,
        "total_market_value": "€1.2B",
        "average_overall": 82.3,
        "team_statistics": {
          "goals_scored": 89,
          "goals_conceded": 31,
          "clean_sheets": 18,
          "possession_avg": 68.5,
          "pass_accuracy": 89.2
        },
        "formation": "4-3-3",
        "playing_style": "Possession-based",
        "reputation": "World Class"
      }
    ]
  }
}
```

### 6. League Information

Comprehensive league data and statistics.

```http
GET /api/leagues
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "leagues": [
      {
        "id": 1,
        "name": "Premier League",
        "country": "England",
        "tier": 1,
        "teams_count": 20,
        "total_players": 500,
        "average_overall": 78.5,
        "total_market_value": "€15.2B",
        "reputation": "World Class",
        "competitiveness": 95,
        "financial_power": 98
      }
    ]
  }
}
```

## Advanced Analysis Endpoints

### 7. Player Analysis Tools

#### Bargain Hunter Analysis

Find undervalued players based on various criteria.

```http
POST /api/analysis/bargains
```

**Request Body:**
```json
{
  "budget": 50000000,
  "currency": "EUR",
  "positions": ["ST", "CAM"],
  "max_age": 28,
  "min_overall": 75,
  "analysis_type": "value_ratio", // "value_ratio", "potential_vs_cost", "wage_efficiency"
  "market_factors": {
    "position_inflation": true,
    "age_premium": true,
    "league_adjustment": true
  }
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "bargain_players": [
      {
        "player_id": 456,
        "name": "Viktor Osimhen",
        "value_ratio": 2.3,
        "market_value": "€120M",
        "estimated_fair_value": "€276M",
        "bargain_score": 87.5,
        "analysis": {
          "value_factors": ["young_age", "high_potential", "proven_performance"],
          "risk_factors": ["injury_history"],
          "recommendation": "Strong Buy"
        }
      }
    ],
    "analysis_summary": {
      "total_analyzed": 1247,
      "bargains_found": 23,
      "average_value_ratio": 1.8,
      "best_bargain_score": 94.2
    }
  }
}
```

#### Wonderkids Discovery

Identify young talents with high potential.

```http
POST /api/analysis/wonderkids
```

**Request Body:**
```json
{
  "max_age": 21,
  "min_potential": 80,
  "positions": ["all"],
  "budget_limit": 100000000,
  "include_growth_analysis": true,
  "scout_requirements": {
    "min_current_ability": 60,
    "personality_traits": ["professional", "determined"],
    "injury_resistance": "medium"
  }
}
```

#### Squad Upgrade Finder

Analyze squad weaknesses and suggest improvements.

```http
POST /api/analysis/upgrades
```

**Request Body:**
```json
{
  "current_squad": [1, 2, 3, 4], // Player IDs
  "formation": "4-3-3",
  "budget": 200000000,
  "priority_positions": ["CB", "CM"],
  "upgrade_criteria": {
    "min_improvement": 5, // Overall rating improvement
    "age_preference": "young", // "young", "experienced", "mixed"
    "style_compatibility": true
  }
}
```

### 8. Wishlist Management

#### Retrieve User Wishlist

```http
GET /api/wishlist
```

**Query Parameters:**
- `user_id` (string): User identifier (for future multi-user support)
- `include_analysis` (bool): Include analysis for wishlist players

**Response:**
```json
{
  "status": "success",
  "data": {
    "wishlist": {
      "id": "wishlist_123",
      "user_id": "user_456",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "players": [
        {
          "player_id": 789,
          "added_at": "2024-01-10T15:20:00Z",
          "notes": "Potential summer target",
          "priority": "high",
          "budget_allocation": 50000000
        }
      ],
      "total_estimated_cost": "€280M",
      "analysis": {
        "position_coverage": {
          "ST": 2,
          "CM": 1,
          "CB": 1
        },
        "age_distribution": {
          "under_23": 2,
          "23_to_28": 1,
          "over_28": 1
        },
        "potential_synergies": ["pace_attack", "creative_midfield"]
      }
    }
  }
}
```

#### Add Player to Wishlist

```http
POST /api/wishlist
```

**Request Body:**
```json
{
  "player_id": 789,
  "notes": "Excellent young striker with high potential",
  "priority": "high", // "low", "medium", "high"
  "budget_allocation": 50000000,
  "target_date": "2024-07-01"
}
```

#### Remove from Wishlist

```http
DELETE /api/wishlist/{player_id}
```

#### Export Wishlist

```http
GET /api/wishlist/export
```

**Query Parameters:**
- `format` (string): "json", "csv", "pdf"
- `include_analysis` (bool): Include detailed analysis

### 9. Team Logo Management

#### Search Team Logos

```http
GET /api/logos/search
```

**Query Parameters:**
- `team_name` (string): Team name for fuzzy search
- `league` (string): Filter by league
- `exact_match` (bool): Require exact match

**Response:**
```json
{
  "status": "success",
  "data": {
    "matches": [
      {
        "team_name": "Manchester City",
        "logo_url": "/logos/manchester_city.png",
        "match_confidence": 95.2,
        "logo_quality": "high",
        "format": "png",
        "size": "512x512"
      }
    ]
  }
}
```

#### Upload Custom Logo

```http
POST /api/logos/upload
Content-Type: multipart/form-data
```

**Parameters:**
- `logo` (file): Logo image file
- `team_name` (string): Associated team name
- `league` (string): Team's league

## Performance and Analytics

### 10. System Metrics

```http
GET /api/metrics
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "performance_metrics": {
      "request_count": 1547,
      "average_response_time": "120ms",
      "error_rate": "0.02%",
      "cache_hit_rate": "94.2%",
      "database_query_time": "15ms"
    },
    "system_resources": {
      "cpu_usage": "15.3%",
      "memory_usage": "512MB",
      "disk_usage": "23.4GB",
      "active_connections": 8
    },
    "business_metrics": {
      "total_players_processed": 125847,
      "total_uploads": 1247,
      "unique_teams": 2847,
      "average_processing_time": "2.3s"
    }
  }
}
```

### 11. Analytics and Statistics

#### Player Statistics Aggregation

```http
GET /api/analytics/players
```

**Query Parameters:**
- `group_by` (string): "position", "league", "nationality", "age_group"
- `metric` (string): "overall", "potential", "market_value", "age"
- `filters` (object): Same filters as player search

**Response:**
```json
{
  "status": "success",
  "data": {
    "aggregations": {
      "position": {
        "ST": {
          "count": 234,
          "average_overall": 76.8,
          "average_value": "€12.5M",
          "top_player": "Erling Haaland"
        },
        "CM": {
          "count": 456,
          "average_overall": 74.2,
          "average_value": "€8.7M",
          "top_player": "Kevin De Bruyne"
        }
      }
    },
    "trends": {
      "value_inflation": {
        "year_over_year": "12.3%",
        "position_leaders": ["ST", "CAM", "CB"]
      },
      "age_demographics": {
        "average_age": 26.4,
        "youth_percentage": 23.7
      }
    }
  }
}
```

## WebSocket Endpoints

### 12. Real-time Processing Updates

For large file uploads, connect to WebSocket for real-time progress updates.

```
WebSocket: ws://localhost:8091/ws/processing/{upload_id}
```

**Messages:**
```json
{
  "type": "progress",
  "data": {
    "stage": "parsing_html",
    "progress": 65,
    "players_processed": 1500,
    "estimated_remaining": "30s",
    "current_operation": "Calculating FIFA ratings"
  }
}
```

## Error Handling

### Error Response Format

```json
{
  "status": "error",
  "message": "Human readable error message",
  "error": "Detailed technical error",
  "error_code": "VALIDATION_ERROR",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "age",
    "value": "invalid",
    "constraint": "must be between 15 and 50"
  }
}
```

### Common Error Codes

- `VALIDATION_ERROR`: Input validation failed
- `FILE_TOO_LARGE`: Upload file exceeds size limit
- `INVALID_FILE_FORMAT`: File format not supported
- `PROCESSING_FAILED`: HTML parsing or processing failed
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server internal error
- `NOT_FOUND`: Resource not found
- `TIMEOUT`: Request timeout

## Rate Limiting

- **General API**: 1000 requests per hour per IP
- **File Upload**: 10 uploads per hour per IP
- **Search/Filter**: 500 requests per hour per IP
- **Analytics**: 100 requests per hour per IP

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:
```
GET /api/openapi.json
GET /api/swagger-ui
```

## SDK and Client Libraries

Official client libraries are available for:
- **JavaScript/TypeScript**: `@fm-dash/client`
- **Python**: `fm-dash-py`
- **Go**: `github.com/fm-dash/go-client`

---

*For additional API support or feature requests, please open an issue on our GitHub repository.* 