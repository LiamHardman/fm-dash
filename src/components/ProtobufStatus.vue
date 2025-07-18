<template>
  <div class="protobuf-status">
    <div class="status-header">
      <h3>Protobuf API Status</h3>
      <q-toggle 
        v-model="protobufEnabled" 
        @update:model-value="toggleProtobuf"
        label="Enable Protobuf"
        color="primary"
      />
    </div>
    
    <div class="status-grid">
      <div class="status-card">
        <div class="status-label">Client Support</div>
        <div class="status-value" :class="status.protobufSupported ? 'success' : 'error'">
          {{ status.protobufSupported ? 'Supported' : 'Not Supported' }}
        </div>
      </div>
      
      <div class="status-card">
        <div class="status-label">Client Enabled</div>
        <div class="status-value" :class="status.protobufEnabled ? 'success' : 'warning'">
          {{ status.protobufEnabled ? 'Enabled' : 'Disabled' }}
        </div>
      </div>
      
      <div class="status-card">
        <div class="status-label">Server Support</div>
        <div class="status-value" :class="getServerStatusClass()">
          {{ getServerStatusText() }}
        </div>
      </div>
      
      <div class="status-card">
        <div class="status-label">Definitions Loaded</div>
        <div class="status-value" :class="status.definitionsLoaded ? 'success' : 'error'">
          {{ status.definitionsLoaded ? 'Loaded' : 'Not Loaded' }}
        </div>
      </div>
    </div>
    
    <div class="test-section">
      <h4>Test API Requests</h4>
      <div class="test-buttons">
        <q-btn 
          @click="testRolesAPI" 
          :loading="testingRoles"
          color="primary"
          label="Test Roles API"
        />
        <q-btn 
          @click="testConfigAPI" 
          :loading="testingConfig"
          color="secondary"
          label="Test Config API"
        />
      </div>
      
      <div v-if="lastTestResult" class="test-result">
        <h5>Last Test Result:</h5>
        <div class="result-format">
          Format: <span :class="lastTestResult._protobuf?.format === 'protobuf' ? 'success' : 'warning'">
            {{ lastTestResult._protobuf?.format || 'unknown' }}
          </span>
        </div>
        <div v-if="lastTestResult._protobuf?.fallbackReason" class="result-fallback">
          Fallback Reason: {{ lastTestResult._protobuf.fallbackReason }}
        </div>
        <div class="result-timing">
          Processing Time: {{ lastTestResult._protobuf?.processingTime?.toFixed(2) || 'N/A' }}ms
        </div>
        <div class="result-size">
          Payload Size: {{ formatBytes(lastTestResult._protobuf?.payloadSize || 0) }}
        </div>
      </div>
    </div>
    
    <div class="metrics-section">
      <h4>Performance Metrics</h4>
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">Total Requests</div>
          <div class="metric-value">{{ metrics.totalRequests }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Protobuf Requests</div>
          <div class="metric-value">{{ metrics.protobufRequests }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">JSON Requests</div>
          <div class="metric-value">{{ metrics.jsonRequests }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Failed Requests</div>
          <div class="metric-value">{{ metrics.failedRequests }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Avg Request Time</div>
          <div class="metric-value">{{ metrics.averageRequestTime?.toFixed(2) || 0 }}ms</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Avg Payload Size</div>
          <div class="metric-value">{{ formatBytes(metrics.averagePayloadSize || 0) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useProtobufApi } from '../composables/useProtobufApi'

const { getRoles, getConfig, getClientStatus, setProtobufEnabled } = useProtobufApi()

const status = ref({
  protobufSupported: false,
  protobufEnabled: false,
  initialized: false,
  serverSupportsProtobuf: null,
  definitionsLoaded: false
})

const metrics = ref({
  totalRequests: 0,
  protobufRequests: 0,
  jsonRequests: 0,
  failedRequests: 0,
  averageRequestTime: 0,
  averagePayloadSize: 0
})

const protobufEnabled = ref(true)
const testingRoles = ref(false)
const testingConfig = ref(false)
const lastTestResult = ref(null)

const getServerStatusClass = () => {
  if (status.value.serverSupportsProtobuf === null) return 'warning'
  return status.value.serverSupportsProtobuf ? 'success' : 'error'
}

const getServerStatusText = () => {
  if (status.value.serverSupportsProtobuf === null) return 'Unknown'
  return status.value.serverSupportsProtobuf ? 'Supported' : 'Not Supported'
}

const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const updateStatus = () => {
  const clientStatus = getClientStatus()
  status.value = { ...clientStatus.status }
  metrics.value = { ...clientStatus.metrics }
}

const toggleProtobuf = (enabled) => {
  setProtobufEnabled(enabled)
  localStorage.setItem('protobuf_enabled', enabled.toString())
  updateStatus()
}

const testRolesAPI = async () => {
  testingRoles.value = true
  try {
    const result = await getRoles()
    lastTestResult.value = result
    console.log('Roles API test result:', result)
  } catch (error) {
    console.error('Roles API test failed:', error)
    lastTestResult.value = { error: error.message }
  } finally {
    testingRoles.value = false
    updateStatus()
  }
}

const testConfigAPI = async () => {
  testingConfig.value = true
  try {
    const result = await getConfig()
    lastTestResult.value = result
    console.log('Config API test result:', result)
  } catch (error) {
    console.error('Config API test failed:', error)
    lastTestResult.value = { error: error.message }
  } finally {
    testingConfig.value = false
    updateStatus()
  }
}

onMounted(() => {
  updateStatus()
  protobufEnabled.value = status.value.protobufEnabled
  
  // Update status every 5 seconds
  setInterval(updateStatus, 5000)
})
</script>

<style scoped>
.protobuf-status {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.status-grid, .metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 30px;
}

.status-card, .metric-card {
  background: #f5f5f5;
  border-radius: 8px;
  padding: 15px;
  text-align: center;
}

.status-label, .metric-label {
  font-size: 12px;
  color: #666;
  text-transform: uppercase;
  margin-bottom: 5px;
}

.status-value, .metric-value {
  font-size: 16px;
  font-weight: bold;
}

.success {
  color: #4caf50;
}

.warning {
  color: #ff9800;
}

.error {
  color: #f44336;
}

.test-section {
  margin-bottom: 30px;
}

.test-buttons {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.test-result {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid #2196f3;
}

.test-result h5 {
  margin: 0 0 10px 0;
}

.result-format, .result-fallback, .result-timing, .result-size {
  margin-bottom: 5px;
}
</style>