<template>
  <div class="processing-status-monitor">
    <q-card class="status-card" flat bordered>
      <q-card-section class="status-content">
        <!-- Processing Icon -->
        <div class="status-icon-container">
          <q-icon
            :name="statusIcon"
            :color="statusColor"
            size="3rem"
            class="status-icon"
          />
          <div class="processing-dots" v-if="status === 'processing'">
            <div class="dot" :class="{ active: dotIndex >= 0 }"></div>
            <div class="dot" :class="{ active: dotIndex >= 1 }"></div>
            <div class="dot" :class="{ active: dotIndex >= 2 }"></div>
          </div>
        </div>

        <!-- Status Message -->
        <div class="status-message">
          <h4 class="status-title">{{ statusTitle }}</h4>
          <p class="status-subtitle">{{ statusSubtitle }}</p>
        </div>

        <!-- Progress for processing -->
        <div v-if="status === 'processing'" class="progress-section">
          <q-linear-progress
            :value="progressValue"
            size="6px"
            color="primary"
            rounded
            class="progress-bar"
            animation-speed="300"
          />
          <div class="progress-text">
            {{ Math.round(progressValue * 100) }}% estimated
          </div>
        </div>

        <!-- Stats for completed -->
        <div v-if="status === 'completed'" class="stats-section">
          <div class="stat-item">
            <q-icon name="group" size="1.2rem" />
            <span>{{ formatNumber(playerCount) }} players processed</span>
          </div>
          <div class="stat-item">
            <q-icon name="currency_exchange" size="1.2rem" />
            <span>{{ currencySymbol }} detected</span>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="action-buttons">
          <q-btn
            v-if="status === 'processing'"
            flat
            color="primary"
            icon="refresh"
            label="Check Again"
            @click="checkStatus"
            :loading="checking"
            size="sm"
          />
          <q-btn
            v-if="status === 'completed'"
            unelevated
            color="positive"
            icon="visibility"
            label="View Dataset"
            @click="viewDataset"
            size="sm"
          />
        </div>
      </q-card-section>
    </q-card>
  </div>
</template>

<script>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import playerService from '../services/playerService.js'

export default {
  name: 'ProcessingStatusMonitor',
  props: {
    datasetId: {
      type: String,
      required: true
    },
    autoRefresh: {
      type: Boolean,
      default: true
    },
    refreshInterval: {
      type: Number,
      default: 5000 // 5 seconds
    }
  },
  emits: ['status-changed', 'completed'],
  setup(props, { emit }) {
    const router = useRouter()
    const status = ref('processing')
    const message = ref('Checking processing status...')
    const playerCount = ref(0)
    const currencySymbol = ref('£')
    const checking = ref(false)
    const dotIndex = ref(0)
    const progressValue = ref(0.3) // Start at 30%
    const checkCount = ref(0)
    const maxChecks = 60 // Maximum 5 minutes (60 * 5 seconds)

    let statusInterval = null
    let dotInterval = null
    let progressInterval = null

    // Computed properties
    const statusIcon = computed(() => {
      switch (status.value) {
        case 'completed':
          return 'check_circle'
        case 'processing':
          return 'sync'
        default:
          return 'help'
      }
    })

    const statusColor = computed(() => {
      switch (status.value) {
        case 'completed':
          return 'positive'
        case 'processing':
          return 'primary'
        default:
          return 'grey'
      }
    })

    const statusTitle = computed(() => {
      switch (status.value) {
        case 'completed':
          return 'Processing Complete!'
        case 'processing':
          return 'Processing Your Dataset'
        default:
          return 'Checking Status'
      }
    })

    const statusSubtitle = computed(() => {
      return message.value
    })

    // Methods
    const checkStatus = async () => {
      if (checking.value) return
      
      checking.value = true
      checkCount.value++

      try {
        const response = await playerService.checkProcessingStatus(props.datasetId)
        
        status.value = response.status
        message.value = response.message
        playerCount.value = response.playerCount || 0
        currencySymbol.value = response.currencySymbol || '£'

        if (status.value === 'completed') {
          handleCompleted()
        } else if (checkCount.value >= maxChecks) {
          handleTimeout()
        }
      } catch (error) {
        console.error('Error checking processing status:', error)
        message.value = 'Error checking status. Please try again.'
      } finally {
        checking.value = false
      }
    }

    const handleCompleted = () => {
      // Stop all intervals
      stopIntervals()
      
      // Emit events
      emit('status-changed', { status: 'completed', playerCount: playerCount.value })
      emit('completed', { datasetId: props.datasetId, playerCount: playerCount.value })
      
      // Update progress to 100%
      progressValue.value = 1
    }

    const handleTimeout = () => {
      message.value = 'Processing is taking longer than expected. You can check back later.'
      stopIntervals()
    }

    const viewDataset = () => {
      router.push(`/dataset/${props.datasetId}`)
    }

    const startIntervals = () => {
      // Status check interval
      if (props.autoRefresh) {
        statusInterval = setInterval(checkStatus, props.refreshInterval)
      }

      // Dot animation interval
      dotInterval = setInterval(() => {
        dotIndex.value = (dotIndex.value + 1) % 4
      }, 500)

      // Progress animation interval
      progressInterval = setInterval(() => {
        if (status.value === 'processing' && progressValue.value < 0.9) {
          progressValue.value += 0.01 // Slowly increase progress
        }
      }, 1000)
    }

    const stopIntervals = () => {
      if (statusInterval) {
        clearInterval(statusInterval)
        statusInterval = null
      }
      if (dotInterval) {
        clearInterval(dotInterval)
        dotInterval = null
      }
      if (progressInterval) {
        clearInterval(progressInterval)
        progressInterval = null
      }
    }

    const formatNumber = (num) => {
      return num.toLocaleString()
    }

    // Lifecycle
    onMounted(() => {
      // Initial check
      checkStatus()
      
      // Start intervals
      startIntervals()
    })

    onUnmounted(() => {
      stopIntervals()
    })

    return {
      status,
      message,
      playerCount,
      currencySymbol,
      checking,
      dotIndex,
      progressValue,
      statusIcon,
      statusColor,
      statusTitle,
      statusSubtitle,
      checkStatus,
      viewDataset,
      formatNumber
    }
  }
}
</script>

<style lang="scss" scoped>
.processing-status-monitor {
  .status-card {
    width: 100%;
    max-width: 500px;
    margin: 0 auto;
    border-radius: 16px;
    background: linear-gradient(135deg, #f8f9fc 0%, #ffffff 100%);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
    border: 1px solid rgba(46, 116, 181, 0.1);

    .body--dark & {
      background: linear-gradient(135deg, #2a2a2a 0%, #1e1e1e 100%);
      border-color: rgba(255, 255, 255, 0.1);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    }
  }

  .status-content {
    padding: 2rem;
    text-align: center;
  }

  .status-icon-container {
    position: relative;
    margin-bottom: 2rem;
    height: 80px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    .status-icon {
      animation: pulse 2s infinite ease-in-out;
    }

    .processing-dots {
      display: flex;
      justify-content: center;
      gap: 0.5rem;
      margin-top: 1rem;

      .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: rgba(46, 116, 181, 0.3);
        transition: all 0.3s ease;

        &.active {
          background: $primary;
          transform: scale(1.2);
        }

        .body--dark & {
          background: rgba(46, 116, 181, 0.4);

          &.active {
            background: $primary;
          }
        }
      }
    }
  }

  .status-message {
    margin-bottom: 2rem;

    .status-title {
      font-size: 1.4rem;
      font-weight: 600;
      color: $primary;
      margin: 0 0 0.5rem 0;
      line-height: 1.2;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }

    .status-subtitle {
      font-size: 1rem;
      color: $secondary;
      margin: 0;
      line-height: 1.3;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }
  }

  .progress-section {
    margin-bottom: 2rem;

    .progress-bar {
      margin-bottom: 0.5rem;
      transition: all 0.2s ease;
    }

    .progress-text {
      font-size: 0.9rem;
      color: $secondary;
      font-weight: 500;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }
  }

  .stats-section {
    margin-bottom: 2rem;

    .stat-item {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
      margin-bottom: 0.5rem;
      font-size: 1rem;
      color: $secondary;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }
  }

  .action-buttons {
    display: flex;
    justify-content: center;
    gap: 1rem;
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}
</style> 