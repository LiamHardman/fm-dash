<template>
  <div class="loading-spinner" :class="{ 'loading-spinner--fullscreen': fullscreen }">
    <div class="loading-spinner__container">
      <q-spinner-dots
        :size="size"
        :color="color"
        class="loading-spinner__icon"
      />
      <div v-if="message" class="loading-spinner__message">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  message: {
    type: String,
    default: 'Loading...'
  },
  size: {
    type: String,
    default: '40px'
  },
  color: {
    type: String,
    default: 'primary'
  },
  fullscreen: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.loading-spinner {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.loading-spinner--fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  z-index: 9999;
}

.loading-spinner__container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.loading-spinner__message {
  font-size: 0.875rem;
  color: var(--q-color-grey-7);
  text-align: center;
}
</style>