<template>
  <div class="error-boundary">
    <div class="error-boundary__container">
      <q-icon 
        name="error_outline" 
        size="48px" 
        color="negative"
        class="error-boundary__icon"
      />
      <h3 class="error-boundary__title">
        {{ title }}
      </h3>
      <p class="error-boundary__message">
        {{ message }}
      </p>
      <div class="error-boundary__actions">
        <q-btn
          v-if="showRetry"
          @click="handleRetry"
          color="primary"
          outline
          label="Try Again"
          icon="refresh"
        />
        <q-btn
          v-if="showDetails && error"
          @click="showErrorDetails = !showErrorDetails"
          color="grey"
          flat
          :label="showErrorDetails ? 'Hide Details' : 'Show Details'"
          :icon="showErrorDetails ? 'expand_less' : 'expand_more'"
        />
      </div>
      <div v-if="showErrorDetails && error" class="error-boundary__details">
        <q-expansion-item
          label="Error Details"
          icon="bug_report"
          header-class="text-negative"
        >
          <div class="error-boundary__error-content">
            <pre>{{ errorDetails }}</pre>
          </div>
        </q-expansion-item>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  error: {
    type: Error,
    default: null
  },
  title: {
    type: String,
    default: 'Something went wrong'
  },
  message: {
    type: String,
    default: 'An error occurred while loading this component. Please try again.'
  },
  showRetry: {
    type: Boolean,
    default: true
  },
  showDetails: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['retry'])

const showErrorDetails = ref(false)

const errorDetails = computed(() => {
  if (!props.error) return 'No error details available'

  return {
    message: props.error.message,
    stack: props.error.stack,
    name: props.error.name,
    timestamp: new Date().toISOString()
  }
})

const handleRetry = () => {
  emit('retry')
}
</script>

<style scoped>
.error-boundary {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  min-height: 200px;
}

.error-boundary__container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  max-width: 500px;
  text-align: center;
}

.error-boundary__icon {
  opacity: 0.7;
}

.error-boundary__title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--q-color-negative);
}

.error-boundary__message {
  margin: 0;
  color: var(--q-color-grey-7);
  line-height: 1.5;
}

.error-boundary__actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: center;
}

.error-boundary__details {
  width: 100%;
  margin-top: 1rem;
}

.error-boundary__error-content {
  background: var(--q-color-grey-1);
  border-radius: 4px;
  padding: 1rem;
  font-family: monospace;
  font-size: 0.75rem;
  text-align: left;
  overflow-x: auto;
  max-height: 300px;
  overflow-y: auto;
}

.error-boundary__error-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>