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
            animation-speed="200"
          />
          <div class="progress-text">
            {{ Math.round(progressValue * 100) }}%
          </div>
        </div>

        <!-- Fun loading message with fixed height -->
        <div class="loading-message">
          <h4 class="message-title">{{ currentMessage.title }}</h4>
          <p class="message-subtitle">{{ currentMessage.subtitle }}</p>
        </div>

        <!-- Stats display with fixed height -->
        <div class="upload-stats" :class="{ 'has-content': uploadStats.filename }">
          <div class="stat-item" v-if="uploadStats.filename">
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
        <div class="actions">
          <q-btn
            v-if="showCancelButton"
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
    const dotIndex = ref(0)
    const showCancelButton = ref(true)
    const currentIconIndex = ref(0)
    const startTime = ref(null)

    // Fun loading messages inspired by games
    const loadingMessages = [
      // Upload stage messages (0-70%)
      {
        title: "Uploading Player Data",
        subtitle: "Transferring your save file to our servers",
        stage: "upload"
      },
      {
        title: "Reticulating Splines", 
        subtitle: "The classic... organizing tactical formations",
        stage: "upload"
      },
      {
        title: "Calculating Player DNA",
        subtitle: "Analyzing genetic predisposition for late winners", 
        stage: "upload"
      },
      {
        title: "Quantifying Bottle Jobs",
        subtitle: "Measuring tendency to lose when it matters most",
        stage: "upload"
      },
      {
        title: "Digitizing Championship Dreams",
        subtitle: "Converting hopes into 1s and 0s",
        stage: "upload"
      },
      {
        title: "Extracting Tactical Genius",
        subtitle: "Locating your 4-4-2 diamond formation data",
        stage: "upload"
      },
      {
        title: "Uploading Transfer Budget Fantasies",
        subtitle: "Processing unrealistic spending expectations",
        stage: "upload"
      },
      {
        title: "Compressing Squad Rotation Strategies",
        subtitle: "Squeezing 25 players into 11 positions",
        stage: "upload"
      },
      // Processing stage messages (70-80%)
      {
        title: "Processing Player Database",
        subtitle: "Parsing player attributes and statistics",
        stage: "processing"
      },
      {
        title: "Consulting Team Talks Database",
        subtitle: "Loading passionate speeches and motivational clichés",
        stage: "processing"
      },
      {
        title: "Simulating Transfer Negotiations",
        subtitle: "Adding unrealistic agent demands",
        stage: "processing"
      },
      {
        title: "Calculating Injury Probability Matrix",
        subtitle: "Programming your best player to get injured in January",
        stage: "processing"
      },
      {
        title: "Installing Referee Bias Algorithms",
        subtitle: "Ensuring controversial decisions at crucial moments",
        stage: "processing"
      },
      {
        title: "Generating Press Conference Scripts",
        subtitle: "Preparing 47 variations of 'I'm happy with the performance'",
        stage: "processing"
      },
      {
        title: "Optimizing Formation Confusion",
        subtitle: "Making sure 3-5-2 looks exactly like 5-3-2",
        stage: "processing"
      },
      {
        title: "Compiling Excuses Database",
        subtitle: "Loading reasons why it's never your fault",
        stage: "processing"
      },
      {
        title: "Processing Board Expectations",
        subtitle: "Multiplying unrealistic demands by coefficient of impossibility",
        stage: "processing"
      },
      // Data fetching stage messages (80-95%)
      {
        title: "Organizing Squad Data",
        subtitle: "Sorting players by potential and current ability",
        stage: "fetching"
      },
      {
        title: "Calibrating Wonderkid Sensors",
        subtitle: "Detecting future Ballon d'Or winners",
        stage: "fetching"
      },
      {
        title: "Loading Match Engine Physics",
        subtitle: "Implementing impossible last-minute goals",
        stage: "fetching"
      },
      {
        title: "Synchronizing Striker Finishing",
        subtitle: "Ensuring they miss from 2 yards when you need a goal",
        stage: "fetching"
      },
      {
        title: "Randomizing Set Piece Accuracy",
        subtitle: "Making corners as effective as throwing paper planes",
        stage: "fetching"
      },
      {
        title: "Indexing Youth Intake Quality",
        subtitle: "Preparing to disappoint you with 2-star potential players",
        stage: "fetching"
      },
      {
        title: "Configuring VAR Incompetence",
        subtitle: "Training virtual assistants to make questionable calls",
        stage: "fetching"
      },
      {
        title: "Loading Weather Impact Systems",
        subtitle: "Ensuring rain ruins your tiki-taka masterpiece",
        stage: "fetching"
      },
      {
        title: "Assembling Loan Army Statistics",
        subtitle: "Cataloging players you'll never see again",
        stage: "fetching"
      },
      {
        title: "Buffering Transfer Market Chaos",
        subtitle: "Preparing unrealistic valuations for average players",
        stage: "fetching"
      },
      // Finalizing stage messages (95-100%)
      {
        title: "Finalizing Dataset",
        subtitle: "Preparing your player database for analysis",
        stage: "finalizing"
      },
      {
        title: "Generating Media Questions",
        subtitle: "Preparing repetitive press conference topics",
        stage: "finalizing"
      },
      {
        title: "Optimizing Squad Harmony",
        subtitle: "Balancing egos and personality clashes",
        stage: "finalizing"
      },
      {
        title: "Calibrating Last-Minute Drama",
        subtitle: "Ensuring maximum stress during crucial matches",
        stage: "finalizing"
      },
      {
        title: "Finalizing Tactical Flexibility",
        subtitle: "Making every formation feel slightly wrong",
        stage: "finalizing"
      },
      {
        title: "Polishing Championship Bottling",
        subtitle: "Perfecting the art of losing when you're ahead",
        stage: "finalizing"
      },
      {
        title: "Completing Scouting Network",
        subtitle: "Ensuring recommended players are always overpriced",
        stage: "finalizing"
      },
      {
        title: "Finalizing Goalkeeper AI",
        subtitle: "Programming them to parry into dangerous areas",
        stage: "finalizing"
      },
      {
        title: "Wrapping Up Wage Structure",
        subtitle: "Making sure backup players earn more than starters",
        stage: "finalizing"
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

    // Get current stage based on progress
    const getCurrentStage = () => {
      const progress = props.progress
      if (progress < 70) return "upload"
      if (progress < 80) return "processing" 
      if (progress < 95) return "fetching"
      return "finalizing"
    }

    const previousStage = ref(getCurrentStage())

    // Randomize array function
    const shuffleArray = (array) => {
      const shuffled = [...array]
      for (let i = shuffled.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
      }
      return shuffled
    }

    // Get messages for current stage (randomized)
    const getStageMessages = () => {
      const currentStage = getCurrentStage()
      const stageMessages = loadingMessages.filter(msg => msg.stage === currentStage)
      return shuffleArray(stageMessages)
    }

    // Get a random message from current stage
    const getRandomMessage = () => {
      const stageMessages = getStageMessages()
      return stageMessages[Math.floor(Math.random() * stageMessages.length)]
    }

    const currentMessage = ref(getRandomMessage())
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
      currentMessage.value = getRandomMessage()
      
      // Change message every 3 seconds
      messageInterval = setInterval(() => {
        currentMessage.value = getRandomMessage()
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

    // Watch for stage changes and reset message index
    watch(() => props.progress, () => {
      const currentStage = getCurrentStage()
      if (currentStage !== previousStage.value) {
        currentMessage.value = getRandomMessage()
        previousStage.value = currentStage
      }
    })

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
@use "sass:color";
@import '../quasar-variables.scss';

.upload-loader-dialog {
  .upload-loader-card {
    width: 500px;
    min-width: 500px;
    max-width: 500px;
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

  .loader-content {
    padding: 2rem;
    height: auto;
    min-height: 400px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  .loader-icon-container {
    position: relative;
    margin-bottom: 2rem;
    height: 100px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

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

  .progress-container {
    margin-bottom: 2rem;
    height: 40px;
    display: flex;
    flex-direction: column;
    justify-content: center;

    .progress-bar {
      margin-bottom: 0.5rem;
      transition: all 0.2s ease;
    }

    .progress-text {
      font-size: 0.9rem;
      color: $secondary;
      font-weight: 500;
      text-align: center;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }
  }

  .loading-message {
    margin-bottom: 2rem;
    height: 80px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    text-align: center;

    .message-title {
      font-size: 1.4rem;
      font-weight: 600;
      color: $primary;
      margin: 0 0 0.5rem 0;
      animation: fadeInOut 3s ease-in-out;
      line-height: 1.2;
      height: 1.8rem;
      display: flex;
      align-items: center;
      justify-content: center;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }

    .message-subtitle {
      font-size: 1rem;
      color: $secondary;
      margin: 0;
      animation: fadeInOut 3s ease-in-out;
      line-height: 1.3;
      height: 2.6rem;
      display: flex;
      align-items: center;
      justify-content: center;
      text-align: center;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }
  }

  .upload-stats {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 2rem;
    padding: 1rem;
    background: rgba(46, 116, 181, 0.05);
    border-radius: 8px;
    border: 1px solid rgba(46, 116, 181, 0.1);
    min-height: 100px;
    opacity: 0;
    transform: translateY(10px);
    transition: all 0.3s ease;

    &.has-content {
      opacity: 1;
      transform: translateY(0);
    }

    .body--dark & {
      background: rgba(46, 116, 181, 0.1);
      border: 1px solid rgba(46, 116, 181, 0.2);
    }

    .stat-item {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.9rem;
      color: $dark;
      min-height: 24px;

      .body--dark & {
        color: color.adjust($primary, $lightness: 15%);
      }
    }
  }

  .actions {
    margin-top: 1rem;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    
    .q-btn {
      animation: fadeIn 0.3s ease-in;
    }
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