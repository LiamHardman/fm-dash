# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Football Manager Data Browser (FM-Dash).

## Quick Diagnostics

### Health Check Commands

```bash
# Check overall system health
npm run check

# Individual component checks
npm run test:api          # Test API connectivity
npm run test:frontend     # Test frontend build
npm run test:go          # Test Go backend
npm run test:s3          # Test S3 connectivity

# Service status
curl http://localhost:8091/api/health
```

### Log Analysis

```bash
# View application logs
docker logs fm-dash-app -f

# View specific service logs
kubectl logs -f deployment/fm-dash-app

# Check error logs only
docker logs fm-dash-app 2>&1 | grep -i error

# Monitor real-time logs with filtering
docker logs fm-dash-app -f --since=30m | grep -E "(ERROR|WARN|FATAL)"
```

## Common Issues

### 1. Application Won't Start

#### Symptoms
- Server fails to start
- Port binding errors
- Configuration errors

#### Diagnosis
```bash
# Check if port is already in use
lsof -i :8091
netstat -tulpn | grep 8091

# Verify configuration
npm run config:validate

# Check environment variables
env | grep -E "(SERVER|S3|CORS)"
```

#### Solutions

**Port Already in Use**
```bash
# Kill process using the port
kill -9 $(lsof -t -i:8091)

# Or use a different port
export SERVER_PORT=8092
```

**Configuration Issues**
```bash
# Reset to default configuration
cp config.yaml.example config.yaml

# Validate environment variables
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8091
export S3_ENDPOINT=localhost:9000
```

**Permission Issues**
```bash
# Fix file permissions
chmod +x entrypoint.sh
chown -R $USER:$USER /tmp/uploads
```

### 2. File Upload Issues

#### Symptoms
- Upload fails with timeout
- "File too large" errors
- "Invalid file format" errors

#### Diagnosis
```bash
# Check upload configuration
grep -E "(MAX_UPLOAD|UPLOAD_TIMEOUT)" .env

# Test upload endpoint
curl -X POST http://localhost:8091/api/upload \
  -F "file=@test.html" \
  -H "Content-Type: multipart/form-data"

# Check available disk space
df -h /tmp
```

#### Solutions

**File Size Issues**
```bash
# Increase upload limits
export MAX_UPLOAD_SIZE=100MB
export UPLOAD_TIMEOUT=600s

# Nginx configuration
client_max_body_size 100M;
client_body_timeout 300s;
```

**Format Issues**
```bash
# Verify file is valid HTML
file your-export.html
head -n 10 your-export.html

# Check for Football Manager markers
grep -i "football manager" your-export.html
```

**Timeout Issues**
```bash
# Increase timeouts
export UPLOAD_TIMEOUT=900s
export PROCESSING_TIMEOUT=1200s

# Monitor processing logs
tail -f debug_server.log | grep -i upload
```

### 3. S3/Storage Issues

#### Symptoms
- "Cannot connect to S3" errors
- File not found errors
- Permission denied errors

#### Diagnosis
```bash
# Test S3 connectivity
aws s3 ls s3://fm-dash-data --endpoint-url http://localhost:9000

# Check S3 configuration
echo $S3_ENDPOINT
echo $S3_BUCKET
echo $S3_ACCESS_KEY

# Test MinIO connection
mc alias set local http://localhost:9000 minioadmin minioadmin
mc ls local/fm-dash-data
```

#### Solutions

**Connection Issues**
```bash
# Verify MinIO is running
docker ps | grep minio

# Check network connectivity
curl http://localhost:9000/minio/health/live

# Restart MinIO
docker restart minio_container
```

**Authentication Issues**
```bash
# Verify credentials
export S3_ACCESS_KEY=your-access-key
export S3_SECRET_KEY=your-secret-key

# Test credentials
aws configure set aws_access_key_id $S3_ACCESS_KEY
aws configure set aws_secret_access_key $S3_SECRET_KEY
```

**Bucket Issues**
```bash
# Create bucket if missing
mc mb local/fm-dash-data

# Check bucket policy
mc policy get local/fm-dash-data

# Set public read policy
mc policy set public local/fm-dash-data
```

### 4. Performance Issues

#### Symptoms
- Slow file processing
- High memory usage
- Timeouts during large file uploads

#### Diagnosis
```bash
# Monitor system resources
top -p $(pgrep -f fm-dash)
htop

# Check memory usage
cat /proc/meminfo | grep -E "(MemTotal|MemAvailable)"

# Monitor Go memory stats
curl http://localhost:8091/api/health | jq '.memory_usage'

# Check processing time
time curl -X POST http://localhost:8091/api/upload -F "file=@large-file.html"
```

#### Solutions

**Memory Optimization**
```bash
# Increase memory limits
export MEMORY_LIMIT=2GB
export WORKER_COUNT=2  # Reduce workers to save memory

# Enable garbage collection tuning
export GOGC=50  # More aggressive GC
```

**Processing Optimization**
```bash
# Increase worker count for CPU-bound tasks
export WORKER_COUNT=8
export BATCH_SIZE=50

# Optimize for large files
export STREAM_PROCESSING=true
export CHUNK_SIZE=1MB
```

**Timeout Adjustments**
```bash
# Increase timeouts for large files
export PROCESSING_TIMEOUT=1800s  # 30 minutes
export HTTP_READ_TIMEOUT=300s
export HTTP_WRITE_TIMEOUT=300s
```

### 5. Frontend Issues

#### Symptoms
- "Network Error" in browser
- CORS errors
- Blank page or loading indefinitely

#### Diagnosis
```bash
# Check frontend build
npm run build
ls -la dist/

# Test API connectivity from frontend
curl -i http://localhost:8091/api/health

# Check browser console for errors
# Open Developer Tools > Console

# Verify CORS configuration
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     http://localhost:8091/api/upload
```

#### Solutions

**CORS Issues**
```bash
# Update CORS configuration
export CORS_ALLOWED_ORIGINS="http://localhost:3000,https://yourdomain.com"

# Verify CORS headers in response
curl -H "Origin: http://localhost:3000" \
     http://localhost:8091/api/health -I
```

**API Connection Issues**
```bash
# Check API URL configuration
echo $VITE_API_URL

# Update API URL
export VITE_API_URL=http://localhost:8091

# Restart development server
npm run dev
```

**Build Issues**
```bash
# Clear cache and rebuild
rm -rf node_modules dist
npm install
npm run build

# Check for dependency issues
npm audit
npm audit fix
```

### 6. Database/Data Issues

#### Symptoms
- "No players found" after upload
- Inconsistent search results
- Data appears corrupted

#### Diagnosis
```bash
# Check processed data
ls -la /path/to/processed/data/

# Verify JSON structure
cat processed/players.json | jq '.[] | select(.name == null)'

# Count processed players
cat processed/players.json | jq '. | length'

# Check for parsing errors
grep -i "error\|failed" debug_server.log
```

#### Solutions

**Data Processing Issues**
```bash
# Re-process uploaded file
curl -X POST http://localhost:8091/api/reprocess \
  -d '{"upload_id": "your-upload-id"}'

# Check HTML file structure
head -n 50 your-export.html | grep -E "<table|<th|<tr"

# Validate player data
npm run validate:data
```

**Search/Filter Issues**
```bash
# Clear search cache
curl -X DELETE http://localhost:8091/api/cache/search

# Rebuild search index
curl -X POST http://localhost:8091/api/search/rebuild

# Test search directly
curl "http://localhost:8091/api/players?search=test"
```

## Error Messages

### Backend Error Messages

#### `failed to connect to S3: connection refused`
**Cause**: S3/MinIO service is not running or not accessible

**Solutions**:
```bash
# Start MinIO
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ACCESS_KEY=minioadmin \
  -e MINIO_SECRET_KEY=minioadmin \
  minio/minio server /data --console-address ":9001"

# Check S3 endpoint
export S3_ENDPOINT=localhost:9000
```

#### `file size exceeds maximum allowed size`
**Cause**: Uploaded file is larger than configured limit

**Solutions**:
```bash
# Increase file size limit
export MAX_UPLOAD_SIZE=100MB

# Check file size
ls -lh your-file.html

# Compress file if possible
gzip your-file.html
```

#### `invalid HTML structure: no player table found`
**Cause**: HTML file doesn't contain expected Football Manager export structure

**Solutions**:
```bash
# Verify file is FM export
grep -i "football manager" your-file.html

# Check for table structure
grep -E "<table|thead|tbody" your-file.html

# Re-export from Football Manager with correct settings
```

#### `context deadline exceeded`
**Cause**: Operation timeout due to large file or slow processing

**Solutions**:
```bash
# Increase timeout
export PROCESSING_TIMEOUT=1800s

# Reduce batch size
export BATCH_SIZE=50

# Monitor processing progress
tail -f debug_server.log
```

### Frontend Error Messages

#### `Network Error: Request failed with status code 0`
**Cause**: API server is not running or CORS issues

**Solutions**:
```bash
# Check API server status
curl http://localhost:8091/api/health

# Update API URL
export VITE_API_URL=http://localhost:8091

# Check CORS configuration
```

#### `Cannot read property 'data' of undefined`
**Cause**: API response format has changed or request failed

**Solutions**:
```javascript
// Add error handling to API calls
try {
  const response = await api.get('/players');
  if (response && response.data) {
    return response.data;
  }
} catch (error) {
  console.error('API request failed:', error);
  throw error;
}
```

## Docker Issues

### Container Won't Start

#### Diagnosis
```bash
# Check container status
docker ps -a | grep fm-dash

# View container logs
docker logs fm-dash-app

# Check resource usage
docker stats fm-dash-app

# Inspect container configuration
docker inspect fm-dash-app
```

#### Solutions
```bash
# Remove and recreate container
docker rm -f fm-dash-app
docker run -d --name fm-dash-app -p 8080:8080 fm-dash:latest

# Check available resources
docker system df
docker system prune

# Update image
docker pull fm-dash:latest
```

### Volume Mount Issues

#### Symptoms
- Configuration files not loading
- Upload directory not accessible
- Permission denied errors

#### Solutions
```bash
# Fix file permissions
sudo chown -R 1000:1000 /path/to/volume

# Check mount points
docker inspect fm-dash-app | jq '.[0].Mounts'

# Test volume access
docker exec fm-dash-app ls -la /etc/config
```

## Kubernetes Issues

### Pod Issues

#### Diagnosis
```bash
# Check pod status
kubectl get pods -l app=fm-dash

# View pod logs
kubectl logs -f deployment/fm-dash-app

# Describe pod for events
kubectl describe pod fm-dash-app-xxx

# Check resource usage
kubectl top pods
```

#### Solutions
```bash
# Scale deployment
kubectl scale deployment fm-dash-app --replicas=3

# Update deployment
kubectl set image deployment/fm-dash-app fm-dash=fm-dash:latest

# Check config and secrets
kubectl get configmap fm-dash-config -o yaml
kubectl get secret fm-dash-secrets -o yaml
```

### Service Issues

#### Diagnosis
```bash
# Check service
kubectl get svc fm-dash-service

# Test service connectivity
kubectl port-forward svc/fm-dash-service 8080:80

# Check endpoints
kubectl get endpoints fm-dash-service
```

## Performance Debugging

### Memory Leaks

#### Monitoring
```bash
# Monitor Go memory
watch -n 5 'curl -s http://localhost:8091/api/health | jq .memory_usage'

# Profile memory usage
go tool pprof http://localhost:8091/debug/pprof/heap

# Check for goroutine leaks
curl http://localhost:8091/debug/pprof/goroutine?debug=1
```

#### Solutions
```bash
# Enable memory profiling
export ENABLE_PPROF=true

# Adjust GC settings
export GOGC=50
export GOMEMLIMIT=1GB

# Monitor and restart if needed
```

### High CPU Usage

#### Diagnosis
```bash
# Profile CPU usage
go tool pprof http://localhost:8091/debug/pprof/profile?seconds=30

# Check system load
uptime
top -p $(pgrep -f fm-dash)
```

#### Solutions
```bash
# Reduce worker count
export WORKER_COUNT=2

# Optimize processing
export BATCH_SIZE=25
export ENABLE_CACHING=true
```

## Debugging Tools

### Enable Debug Mode

```bash
# Backend debug mode
export LOG_LEVEL=debug
export DEBUG=true

# Frontend debug mode
export VITE_DEBUG=true
export NODE_ENV=development
```

### Logging Configuration

```yaml
# Enhanced logging configuration
logging:
  level: debug
  format: json
  outputs:
    - stdout
    - file: /var/log/fm-dash.log
  components:
    api: debug
    processing: info
    storage: warn
```

### Profiling Endpoints

```bash
# CPU profiling
curl http://localhost:8091/debug/pprof/profile?seconds=30 > cpu.prof

# Memory profiling
curl http://localhost:8091/debug/pprof/heap > heap.prof

# Goroutine profiling
curl http://localhost:8091/debug/pprof/goroutine > goroutine.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof heap.prof
```

## Getting Help

### Information to Gather

When reporting issues, please include:

1. **Environment Information**
   ```bash
   # System information
   uname -a
   docker --version
   kubectl version --client
   
   # Application version
   curl http://localhost:8091/api/health
   ```

2. **Configuration**
   ```bash
   # Environment variables (sanitized)
   env | grep -E "(SERVER|S3|CORS|UPLOAD)" | sed 's/=.*/=***/'
   
   # Configuration files
   cat config.yaml
   ```

3. **Logs**
   ```bash
   # Recent application logs
   docker logs fm-dash-app --tail=100
   
   # Error logs
   grep -i error /var/log/fm-dash.log | tail -20
   ```

4. **Steps to Reproduce**
   - Exact steps that led to the issue
   - Expected vs actual behavior
   - File sizes and types being processed

### Support Channels

- **GitHub Issues**: For bugs and feature requests
- **Documentation**: Check existing docs for solutions
- **Community Forum**: For usage questions and discussions

---

If you can't find a solution here, please create an issue with the information gathered above. 