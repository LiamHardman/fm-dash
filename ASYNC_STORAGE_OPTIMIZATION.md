# Async Storage Optimization

## Overview

This optimization improves API response times by making dataset storage asynchronous. Previously, the upload handler would wait for the complete storage operation (including S3/disk writes) before returning a response to the user. Now, data is made available immediately in memory, and persistent storage happens in the background.

## Performance Improvement

**Before (Synchronous Storage):**
```
1. Parse uploaded file ✅
2. Process player data ✅  
3. Store to S3/disk (BLOCKING) ⛔ <- User waits here
4. Return response to user
```

**After (Asynchronous Storage):**
```
1. Parse uploaded file ✅
2. Process player data ✅
3. Store to memory (immediate) ✅
4. Return response to user ✅ <- User gets response immediately
5. Store to S3/disk (background) ✅ <- Happens asynchronously
```

## Key Changes

### New Functions Added

1. **`StoreDatasetAsync()`** in `store.go`
   - Asynchronous version of `StoreDataset()`
   - Handles persistent storage in a background goroutine
   - Includes comprehensive tracing and error handling

2. **`SetPlayerDataAsync()`** in `store.go`
   - Asynchronous version of `SetPlayerData()`
   - Immediately stores data in memory for fast retrieval
   - Queues persistent storage as a background operation

### Modified Functions

1. **Upload Handler** in `handlers.go`
   - Changed from `SetPlayerData()` to `SetPlayerDataAsync()`
   - Added tracing attributes to track async operations

2. **Percentile Processing** in `handlers.go`
   - Updated `processPercentilesAsync()` to use async storage
   - Maintains consistency with new storage pattern

## Benefits

### 1. **Faster Response Times**
- Users get immediate feedback when upload completes
- No waiting for S3/disk I/O operations
- Particularly beneficial for large datasets

### 2. **Better User Experience**
- Reduced perceived latency
- More responsive API
- Users can start working with data immediately

### 3. **Improved Scalability**
- Reduced blocking operations
- Better resource utilization
- Storage operations don't block request threads

### 4. **Maintained Reliability**
- Data is immediately available in memory
- Persistent storage still happens (just asynchronously)
- Comprehensive error handling and logging for background operations

## Technical Details

### Memory vs Persistent Storage

The optimization uses a dual-storage approach:

1. **In-Memory Storage (Immediate)**
   - Fast map-based storage for immediate data access
   - Used by legacy `GetPlayerData()` function
   - Enables instant response to API requests

2. **Persistent Storage (Asynchronous)**
   - S3 or local file system storage
   - Happens in background goroutines
   - Provides durability and persistence

### Error Handling

- Background storage failures are logged but don't affect user response
- Data remains available in memory even if persistent storage fails
- Comprehensive tracing for monitoring async operations

### Monitoring

Added tracing and metrics for:
- Async storage operation duration
- Success/failure rates of background storage
- Performance comparison between sync and async methods

## Testing

### Benchmark Tests
- `BenchmarkDatasetStorage` compares sync vs async performance
- Response time simulation shows real-world improvement
- Performance metrics demonstrate the optimization benefits

### Correctness Tests
- `TestAsyncStorageCorrectness` verifies data is immediately available
- Ensures no functionality is lost with async approach
- Validates both memory and persistent storage work correctly

## Usage

The async storage is now used automatically in:
- File upload processing (`uploadHandler`)
- Percentile calculation updates (`processPercentilesAsync`)

No changes needed in client code - the optimization is transparent to API users.

## Monitoring Recommendations

1. **Watch for Background Storage Errors**
   ```
   grep "Error storing dataset.*asynchronously" logs/
   ```

2. **Monitor Storage Performance**
   - Check `async_storage.duration_ms` metrics
   - Compare sync vs async response times

3. **Verify Data Persistence**
   - Ensure background storage operations complete successfully
   - Monitor storage system health (S3/disk space)

## Future Enhancements

1. **Storage Queue Management**
   - Implement bounded queues to prevent memory buildup
   - Add retry logic for failed background storage

2. **Batch Storage Operations**
   - Group multiple storage operations for efficiency
   - Implement smart batching based on system load

3. **Storage Health Monitoring**
   - Add health checks for persistent storage systems
   - Implement alerting for storage failures 