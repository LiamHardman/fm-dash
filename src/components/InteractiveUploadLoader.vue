<template>
  <q-dialog
    v-model="showDialog"
    persistent
    no-escape-key
    position="standard"
    class="upload-loader-dialog"
  >
    <q-card class="upload-loader-card" flat>
      <q-card-section class="text-center loader-content">
        <!-- Animated Football Manager themed icon -->
        <div class="loader-icon-container">
          <q-icon
            :name="currentIcon"
            size="4rem"
            :color="iconColor"
            class="loader-icon"
          />
          <div class="loading-dots">
            <div class="dot" :class="{ active: dotIndex >= 0 }"></div>
            <div class="dot" :class="{ active: dotIndex >= 1 }"></div>
            <div class="dot" :class="{ active: dotIndex >= 2 }"></div>
          </div>
        </div>

        <!-- Progress bar -->
        <div class="progress-container">
          <q-linear-progress
            :value="progressValue"
            size="8px"
            color="primary"
            rounded
            class="progress-bar"
          />
          <div class="progress-text">
            {{ Math.round(progressValue * 100) }}%
          </div>
        </div>

        <!-- Fun loading message -->
        <div class="loading-message">
          <h4 class="message-title">{{ currentMessage.title }}</h4>
          <p class="message-subtitle">{{ currentMessage.subtitle }}</p>
        </div>

        <!-- Stats display -->
        <div class="upload-stats" v-if="uploadStats.filename">
          <div class="stat-item">
            <q-icon name="description" size="1.2rem" />
            <span>{{ uploadStats.filename }}</span>
          </div>
          <div class="stat-item" v-if="uploadStats.fileSize">
            <q-icon name="storage" size="1.2rem" />
            <span>{{ uploadStats.fileSize }}</span>
          </div>
          <div class="stat-item" v-if="uploadStats.playersFound > 0">
            <q-icon name="group" size="1.2rem" />
            <span>{{ uploadStats.playersFound.toLocaleString() }} players detected</span>
          </div>
        </div>

        <!-- Cancel button (only shown for the first few seconds) -->
        <div class="actions" v-if="showCancelButton">
          <q-btn
            flat
            color="negative"
            icon="close"
            label="Cancel Upload"
            @click="handleCancel"
            size="sm"
          />
        </div>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script>
import { ref, computed, watch, onUnmounted } from 'vue'

export default {
  name: 'InteractiveUploadLoader',
  emits: ['cancel'],
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    filename: {
      type: String,
      default: ''
    },
    fileSize: {
      type: String,
      default: ''
    },
    playersFound: {
      type: Number,
      default: 0
    },
    progress: {
      type: Number,
      default: 0
    }
  },
  setup(props, { emit }) {
    const showDialog = ref(false)
    const currentMessageIndex = ref(0)
    const dotIndex = ref(0)
    const showCancelButton = ref(true)
    const currentIconIndex = ref(0)
    const startTime = ref(null)

    // Fun loading messages inspired by games
    const loadingMessages = [
      {
        title: "Reticulating Splines",
        subtitle: "The classic... organizing tactical formations"
      },
      {
        title: "Calculating Player DNA",
        subtitle: "Analyzing genetic predisposition for late winners"
      },
      {
        title: "Consulting Team Talks Database",
        subtitle: "Loading passionate speeches and motivational clichés"
      },
      {
        title: "Simulating Transfer Negotiations",
        subtitle: "Adding unrealistic agent demands"
      },
      {
        title: "Calibrating Wonderkid Sensors",
        subtitle: "Detecting future Ballon d'Or winners"
      },
      {
        title: "Loading Match Engine Physics",
        subtitle: "Implementing impossible last-minute goals"
      },
      {
        title: "Generating Media Questions",
        subtitle: "Preparing repetitive press conference topics"
      },
      {
        title: "Optimizing Squad Harmony",
        subtitle: "Balancing egos and personality clashes"
      },
      {
        title: "Indexing Injury Probability Matrix",
        subtitle: "Your best player will definitely get injured"
      },
      {
        title: "Compiling Scout Reports",
        subtitle: "Discovering gems in the Faroe Islands"
      },
      {
        title: "Processing Contract Demands",
        subtitle: "Adding unreasonable wage expectations"
      },
      {
        title: "Synchronizing Board Expectations",
        subtitle: "Setting impossible targets with zero budget"
      },
      {
        title: "Initializing Referee AI",
        subtitle: "Programming controversial VAR decisions"
      },
      {
        title: "Caching Weather Patterns",
        subtitle: "Ensuring rain during important matches"
      },
      {
        title: "Loading Stadium Atmospherics",
        subtitle: "Preparing crowd reactions to substitute decisions"
      }
    ]

    // Icons that rotate during loading
    const loadingIcons = [
      'sports_soccer',
      'emoji_events',
      'timeline',
      'analytics',
      'groups',
      'stadium'
    ]

    // Icon colors that cycle
    const iconColors = ['primary', 'secondary', 'accent', 'positive', 'info']

    const currentMessage = computed(() => loadingMessages[currentMessageIndex.value])
    const currentIcon = computed(() => loadingIcons[currentIconIndex.value])
    const iconColor = computed(() => iconColors[currentIconIndex.value % iconColors.length])
    const progressValue = computed(() => Math.min(props.progress / 100, 1))

    const uploadStats = computed(() => ({
      filename: props.filename,
      fileSize: props.fileSize,
      playersFound: props.playersFound
    }))

    let messageInterval = null
    let dotInterval = null
    let iconInterval = null
    let cancelTimeout = null

    // Start animations and timers
    const startLoader = () => {
      startTime.value = Date.now()
      showCancelButton.value = true
      
      // Change message every 3 seconds
      messageInterval = setInterval(() => {
        currentMessageIndex.value = (currentMessageIndex.value + 1) % loadingMessages.length
      }, 3000)

      // Animate dots every 500ms
      dotInterval = setInterval(() => {
        dotIndex.value = (dotIndex.value + 1) % 4
      }, 500)

      // Change icon every 2 seconds
      iconInterval = setInterval(() => {
        currentIconIndex.value = (currentIconIndex.value + 1) % loadingIcons.length
      }, 2000)

      // Hide cancel button after 10 seconds
      cancelTimeout = setTimeout(() => {
        showCancelButton.value = false
      }, 10000)
    }

    // Stop all animations and timers
    const stopLoader = () => {
      if (messageInterval) {
        clearInterval(messageInterval)
        messageInterval = null
      }
      if (dotInterval) {
        clearInterval(dotInterval)
        dotInterval = null
      }
      if (iconInterval) {
        clearInterval(iconInterval)
        iconInterval = null
      }
      if (cancelTimeout) {
        clearTimeout(cancelTimeout)
        cancelTimeout = null
      }
    }

    const handleCancel = () => {
      emit('cancel')
      showDialog.value = false
    }

    // Watch for visibility changes
    watch(() => props.visible, (newVal) => {
      showDialog.value = newVal
      if (newVal) {
        startLoader()
      } else {
        stopLoader()
      }
    }, { immediate: true })

    // Cleanup on unmount
    onUnmounted(() => {
      stopLoader()
    })

    return {
      showDialog,
      currentMessage,
      currentIcon,
      iconColor,
      dotIndex,
      progressValue,
      uploadStats,
      showCancelButton,
      handleCancel
    }
  }
}
</script>

<style lang="scss" scoped>
.upload-loader-dialog {
  .upload-loader-card {
    min-width: 400px;
    max-width: 500px;
    border-radius: 16px;
    background: linear-gradient(135deg, #f8f9fc 0%, #ffffff 100%);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
    border: 1px solid rgba(26, 35, 126, 0.1);

    .body--dark & {
      background: linear-gradient(135deg, #2a2a2a 0%, #1e1e1e 100%);
      border-color: rgba(255, 255, 255, 0.1);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    }
  }

  .loader-content {
    padding: 2rem;
  }

  .loader-icon-container {
    position: relative;
    margin-bottom: 2rem;

    .loader-icon {
      animation: bounce 2s infinite ease-in-out;
    }

    .loading-dots {
      display: flex;
      justify-content: center;
      gap: 0.5rem;
      margin-top: 1rem;

      .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: rgba(26, 35, 126, 0.3);
        transition: all 0.3s ease;

        &.active {
          background: #1a237e;
          transform: scale(1.2);
        }

        .body--dark & {
          background: rgba(255, 255, 255, 0.3);

          &.active {
            background: rgba(255, 255, 255, 0.9);
          }
        }
      }
    }
  }

  .progress-container {
    margin-bottom: 2rem;

    .progress-bar {
      margin-bottom: 0.5rem;
    }

    .progress-text {
      font-size: 0.9rem;
      color: #666;
      font-weight: 500;

      .body--dark & {
        color: rgba(255, 255, 255, 0.7);
      }
    }
  }

  .loading-message {
    margin-bottom: 2rem;

    .message-title {
      font-size: 1.4rem;
      font-weight: 600;
      color: #1a237e;
      margin: 0 0 0.5rem 0;
      animation: fadeInOut 3s ease-in-out;

      .body--dark & {
        color: rgba(255, 255, 255, 0.9);
      }
    }

    .message-subtitle {
      font-size: 1rem;
      color: #666;
      margin: 0;
      animation: fadeInOut 3s ease-in-out;

      .body--dark & {
        color: rgba(255, 255, 255, 0.7);
      }
    }
  }

  .upload-stats {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 2rem;
    padding: 1rem;
    background: rgba(26, 35, 126, 0.05);
    border-radius: 8px;
    border: 1px solid rgba(26, 35, 126, 0.1);

    .body--dark & {
      background: rgba(255, 255, 255, 0.05);
      border-color: rgba(255, 255, 255, 0.1);
    }

    .stat-item {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.9rem;
      color: #333;

      .body--dark & {
        color: rgba(255, 255, 255, 0.8);
      }
    }
  }

  .actions {
    margin-top: 1rem;
    animation: fadeIn 0.3s ease-in;
  }
}

@keyframes bounce {
  0%, 20%, 53%, 80%, 100% {
    transform: translate3d(0, 0, 0);
  }
  40%, 43% {
    transform: translate3d(0, -10px, 0);
  }
  70% {
    transform: translate3d(0, -5px, 0);
  }
  90% {
    transform: translate3d(0, -2px, 0);
  }
}

@keyframes fadeInOut {
  0% {
    opacity: 0;
    transform: translateY(10px);
  }
  10%, 90% {
    opacity: 1;
    transform: translateY(0);
  }
  100% {
    opacity: 0;
    transform: translateY(-10px);
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style> 