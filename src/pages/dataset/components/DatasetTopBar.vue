<template>
  <div class="top-bar" v-if="!pageLoading && !pageLoadingError && allPlayersData.length > 0">
    <div class="top-bar-content">
      <div class="dataset-info">
        <div class="dataset-title">
          <q-icon name="analytics" size="1.2rem" class="q-mr-xs" />
          FM Data Hub
        </div>
        <div class="dataset-stats">
          <span class="stat-item">{{ formatNumber(allPlayersData.length) }} Players</span>
          <span class="stat-separator">•</span>
          <span class="stat-item">{{ formatNumber(uniqueClubs.length) }} Clubs</span>
          <span class="stat-separator">•</span>
          <span class="stat-item">{{ formatNumber(uniqueNationalities.length) }} Nations</span>
        </div>
      </div>

      <div class="quick-actions">
        <q-btn
          unelevated
          dense
          icon="find_replace"
          label="Upgrade Finder"
          color="primary"
          @click="$emit('open-upgrade-finder')"
          :disable="allPlayersData.length === 0"
          class="action-btn"
          size="sm"
        />
        <q-btn
          unelevated
          dense
          icon="stars"
          label="Wonderkids"
          color="secondary"
          @click="$emit('open-wonderkids')"
          :disable="allPlayersData.length === 0"
          class="action-btn"
          size="sm"
        />
        <q-btn
          unelevated
          dense
          icon="local_offer"
          label="Bargains"
          color="positive"
          @click="$emit('open-bargains')"
          :disable="allPlayersData.length === 0"
          class="action-btn"
          size="sm"
        />
        <q-btn
          unelevated
          dense
          icon="person_off"
          label="Free Agents"
          color="deep-orange"
          @click="$emit('open-free-agents')"
          :disable="allPlayersData.length === 0"
          class="action-btn"
          size="sm"
        />
        <q-btn
          unelevated
          dense
          icon="download"
          label="Export"
          color="accent"
          @click="$emit('open-export')"
          :disable="loading || !filteredPlayers || filteredPlayers.length === 0"
          class="action-btn"
          size="sm"
        >
          <q-tooltip v-if="filteredPlayers && filteredPlayers.length > 0">
            Export {{ filteredPlayers.length }} filtered players
          </q-tooltip>
          <q-tooltip v-else>
            No players to export
          </q-tooltip>
        </q-btn>
      </div>

      <div class="top-bar-controls">
        <q-btn
          v-if="currentDatasetId"
          flat
          dense
          icon="share"
          @click="$emit('share')"
          class="share-btn"
          size="sm"
        >
          <q-tooltip>Share Dataset</q-tooltip>
        </q-btn>
        <q-btn
          flat
          dense
          :icon="showFilters ? 'filter_list_off' : 'filter_list'"
          @click="$emit('toggle-filters')"
          class="filter-toggle-btn"
          size="sm"
          :color="showFilters ? 'primary' : 'grey-6'"
        >
          <q-tooltip>{{ showFilters ? 'Hide' : 'Show' }} Filters</q-tooltip>
        </q-btn>
      </div>
    </div>
  </div>
</template>

<script setup>
// biome-ignore lint/correctness/noUnusedImports: used in template
import { formatNumber } from '../utils/datasetUtils'

defineProps({
  pageLoading: Boolean,
  pageLoadingError: String,
  allPlayersData: { type: Array, required: true },
  uniqueClubs: { type: Array, required: true },
  uniqueNationalities: { type: Array, required: true },
  loading: Boolean,
  filteredPlayers: { type: Array, required: true },
  currentDatasetId: { type: String, default: '' },
  showFilters: Boolean,
})

defineEmits([
  'open-upgrade-finder',
  'open-wonderkids',
  'open-bargains',
  'open-free-agents',
  'open-export',
  'share',
  'toggle-filters',
])
</script>

<style scoped>
/* Inherit top bar styles from page */
</style>

