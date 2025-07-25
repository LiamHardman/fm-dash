<template>
  <q-page class="processing-status-page">
    <div class="status-container">
      <!-- Header -->
      <div class="status-header">
        <q-icon name="cloud_upload" size="3rem" color="primary" class="header-icon" />
        <h2 class="header-title">Processing Your Dataset</h2>
        <p class="header-subtitle">
          Your file is being processed in the background. This may take a few minutes for large files.
        </p>
      </div>

      <!-- Processing Status Monitor -->
      <ProcessingStatusMonitor
        :dataset-id="datasetId"
        :auto-refresh="true"
        :refresh-interval="5000"
        @completed="handleProcessingCompleted"
      />

      <!-- Additional Info -->
      <div class="info-section">
        <q-card class="info-card" flat bordered>
          <q-card-section>
            <h4 class="info-title">What's happening?</h4>
            <ul class="info-list">
              <li>Parsing player data from your HTML file</li>
              <li>Calculating player attributes and statistics</li>
              <li>Generating performance percentiles</li>
              <li>Preparing the dataset for analysis</li>
            </ul>
          </q-card-section>
        </q-card>
      </div>

      <!-- Back to Upload -->
      <div class="back-section">
        <q-btn
          flat
          color="secondary"
          icon="arrow_back"
          label="Back to Upload"
          @click="router.push('/upload')"
          size="sm"
        />
      </div>
    </div>
  </q-page>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ProcessingStatusMonitor from '../components/ProcessingStatusMonitor.vue'

export default {
  name: 'ProcessingStatusPage',
  components: {
    ProcessingStatusMonitor
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const datasetId = ref('')

    onMounted(() => {
      // Get dataset ID from route params
      datasetId.value = route.params.datasetId
      
      if (!datasetId.value) {
        // Redirect to upload page if no dataset ID
        router.push('/upload')
      }
    })

    const handleProcessingCompleted = (result) => {
      // Redirect to dataset page when processing is complete
      router.push(`/dataset/${result.datasetId}`)
    }

    return {
      datasetId,
      router,
      handleProcessingCompleted
    }
  }
}
</script>

<style lang="scss" scoped>
@use "sass:color";
.processing-status-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #f8f9fc 0%, #ffffff 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;

  .body--dark & {
    background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  }
}

.status-container {
  max-width: 600px;
  width: 100%;
  text-align: center;
}

.status-header {
  margin-bottom: 3rem;

  .header-icon {
    margin-bottom: 1rem;
  }

  .header-title {
    font-size: 2rem;
    font-weight: 600;
    color: $primary;
    margin: 0 0 1rem 0;

    .body--dark & {
      color: color.adjust($primary, $lightness: 15%);
    }
  }

  .header-subtitle {
    font-size: 1.1rem;
    color: $secondary;
    margin: 0;
    line-height: 1.5;

    .body--dark & {
      color: color.adjust($primary, $lightness: 15%);
    }
  }
}

.info-section {
  margin-top: 3rem;

  .info-card {
    background: rgba(255, 255, 255, 0.8);
    border-radius: 12px;
    backdrop-filter: blur(10px);

    .body--dark & {
      background: rgba(30, 41, 59, 0.8);
    }
  }

  .info-title {
    font-size: 1.2rem;
    font-weight: 600;
    color: $primary;
    margin: 0 0 1rem 0;

    .body--dark & {
      color: color.adjust($primary, $lightness: 15%);
    }
  }

  .info-list {
    text-align: left;
    margin: 0;
    padding-left: 1.5rem;
    color: $secondary;
    line-height: 1.6;

    .body--dark & {
      color: color.adjust($primary, $lightness: 15%);
    }

    li {
      margin-bottom: 0.5rem;
    }
  }
}

.back-section {
  margin-top: 2rem;
}
</style> 