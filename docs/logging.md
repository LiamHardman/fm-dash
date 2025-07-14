# Logging System

## Overview

The application now supports configurable log levels to control the verbosity of logging output. This allows you to filter logs based on their importance and reduce noise in production environments. **All logs now properly integrate with SignOz via OpenTelemetry OTLP handler.**

## Log Levels

The following log levels are available (in order of severity):

1. **DEBUG** (0) - Detailed debugging information (e.g., cache hits, internal state changes)
2. **INFO** (1) - General informational messages (e.g., successful operations, configuration loaded)
3. **WARN** (2) - Warning messages that don't stop execution (e.g., fallback to defaults, recoverable errors)
4. **CRITICAL** (3) - Critical errors and fatal conditions

## Configuration

### Environment Variables

- **`LOG_LEVEL`**: Sets the minimum log level to display
  - Valid values: `DEBUG`, `INFO`, `WARN`, `CRITICAL` (case-insensitive)
  - Default: `INFO`
  - Example: `LOG_LEVEL=DEBUG` will show all log messages
  - Example: `LOG_LEVEL=WARN` will only show WARN and CRITICAL messages

- **`LOG_ALL_REQUESTS`**: Controls whether all HTTP requests are logged
  - Valid values: `true`, `false`
  - Default: `false` (only logs non-200 responses)

- **`OTEL_ENABLED`**: Controls whether OpenTelemetry (and SignOz integration) is enabled
  - Valid values: `true`, `false`
  - Default: `false`
  - When `true`, logs are sent to SignOz via OTLP

## Usage in Code

Use the appropriate logging function based on the message severity:

```go
// Debug messages - only shown when LOG_LEVEL=DEBUG
LogDebug("Retrieved config data from memory cache")
LogDebug("Processing player data for %s", playerName)

// Info messages - default level, shown for normal operations
LogInfo("Configuration initialization completed successfully")
LogInfo("Successfully loaded weights from %s with %d entries", filename, count)

// Warning messages - shown for recoverable issues
LogWarn("Could not read %s: %v. Using default weights.", filePath, err)
LogWarn("S3 connection failed, using local fallback storage")

// Critical messages - shown for serious errors
LogCritical("Database connection lost: %v", err)
LogCritical("Failed to initialize required service: %v", err)
```

## Examples

### Production Setup (Minimal Logging)
```bash
export LOG_LEVEL=WARN
export OTEL_ENABLED=true
# Only warnings and critical errors will be shown, all logs sent to SignOz
```

### Development Setup (Verbose Logging)
```bash
export LOG_LEVEL=DEBUG
export OTEL_ENABLED=true
# All log messages will be shown, including cache hits and debug info, all sent to SignOz
```

### Standard Setup (Balanced)
```bash
export LOG_LEVEL=INFO  # This is the default
export OTEL_ENABLED=true
# Shows info, warnings, and critical messages, all sent to SignOz
```

## Log Format

Each log message is prefixed with its level:

```
[DEBUG] Retrieved config data from memory cache
[INFO] Configuration initialization completed successfully
[WARN] Could not read attribute_weights.json: file not found. Using default weights.
[CRITICAL] Failed to initialize database connection
```

## SignOz Integration

When `OTEL_ENABLED=true`, all logs from the leveled logging functions (`LogDebug`, `LogInfo`, `LogWarn`, `LogCritical`) are automatically sent to SignOz via the OpenTelemetry OTLP handler. This includes:

- Structured logging with trace correlation
- Automatic context enrichment (trace IDs, span IDs)
- Integration with distributed tracing
- Centralized log aggregation in SignOz

## Migration Notes

The system is backward compatible:
- Existing `log.Printf()` calls continue to work unchanged
- New leveled functions (`LogDebug`, `LogInfo`, etc.) respect the minimum log level **and now properly send logs to SignOz**
- Cache-related debug messages have been moved to DEBUG level to reduce noise at INFO level

## Performance

- Log level checking is very fast (simple integer comparison)
- Messages below the minimum level are not formatted or processed
- No performance impact on production systems when using higher log levels
- SignOz integration adds minimal overhead via efficient OTLP batching

## Troubleshooting

### Logs not appearing in SignOz

1. **Check OTEL_ENABLED**: Ensure `OTEL_ENABLED=true` in your environment
2. **Verify LOG_LEVEL**: Make sure your log level allows the messages you expect (e.g., `LOG_LEVEL=DEBUG` for debug messages)
3. **Check SignOz configuration**: Verify `OTEL_EXPORTER_OTLP_ENDPOINT` points to your SignOz collector
4. **Use slog functions**: Only `slog.*` and the leveled functions (`LogDebug`, `LogInfo`, etc.) send logs to SignOz. Direct `log.Printf()` calls only go to stdout/stderr.

### Recent Fix (2024)

**Issue**: After implementing leveled logging functions, logs stopped appearing in SignOz.

**Root Cause**: The leveled logging functions (`LogDebug`, `LogInfo`, etc.) were using `log.Printf()` which bypasses the OpenTelemetry slog handler.

**Fix**: Updated all leveled logging functions to use `slog` instead of `log.Printf()`, ensuring proper SignOz integration while maintaining LOG_LEVEL filtering behavior. 