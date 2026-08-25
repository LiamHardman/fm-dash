<template>
  <div class="filter-bar">
    <div class="division-filter-container">
      <div class="division-filter-header">
        <q-btn
          @click="selectAllDivisions()"
          size="sm"
          color="primary"
          icon="select_all"
          label="Select All"
          dense
          outline
          class="filter-action-btn"
        />
        <q-btn
          @click="clearAllDivisions()"
          size="sm"
          color="negative"
          icon="clear_all"
          label="Clear All"
          dense
          outline
          class="filter-action-btn"
        />
      </div>
      <q-select
        v-model="selectedDivisionsModel"
        :options="divisionOptions"
        label="Filter by Division"
        dense
        outlined
        multiple
        :use-chips="selectedDivisionsModel.length <= 5"
        use-input
        @filter="filterDivisionsFn"
        class="division-filter"
        :display-value="selectedDivisionsDisplayText"
      >
        <template v-slot:no-option>
          <q-item>
            <q-item-section class="text-grey">No divisions found</q-item-section>
          </q-item>
        </template>
      </q-select>
    </div>

    <div class="position-filter-container">
      <div class="position-filter-header">
        <q-btn
          @click="selectAllPositions()"
          size="sm"
          color="primary"
          icon="select_all"
          label="Select All"
          dense
          outline
          class="filter-action-btn"
        />
        <q-btn
          @click="clearAllPositions()"
          size="sm"
          color="negative"
          icon="clear_all"
          label="Clear All"
          dense
          outline
          class="filter-action-btn"
        />
      </div>
      <q-select
        v-model="selectedPositionsModel"
        :options="positionOptions"
        label="Filter by Position"
        dense
        outlined
        multiple
        :use-chips="selectedPositionsModel.length <= 5"
        use-input
        @filter="filterPositionsFn"
        class="position-filter"
        :display-value="selectedPositionsDisplayText"
        emit-value
        map-options
      >
        <template v-slot:no-option>
          <q-item>
            <q-item-section class="text-grey">No positions found</q-item-section>
          </q-item>
        </template>
      </q-select>
    </div>

    <div class="minutes-filter">
      <div class="slider-label">Minimum Minutes Played</div>
      <q-slider
        v-model="sliderValueModel"
        :min="0"
        :max="maxMinutes"
        :step="50"
        label
        :label-value="`${sliderValueModel}+ mins`"
        label-always
        class="q-mt-sm"
        color="primary"
      />
    </div>

    <div class="overall-filter">
      <div class="slider-label">Minimum Overall Rating</div>
      <q-slider
        v-model="overallSliderValueModel"
        :min="0"
        :max="maxOverall"
        :step="1"
        label
        :label-value="`${overallSliderValueModel}+ OVR`"
        label-always
        class="q-mt-sm"
        color="secondary"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const emit = defineEmits([
  'update:selectedDivisions',
  'update:selectedPositions',
  'update:sliderValue',
  'update:overallSliderValue',
])

const props = defineProps({
  // models
  selectedDivisions: { type: Array, required: true },
  selectedPositions: { type: Array, required: true },
  sliderValue: { type: Number, required: true },
  overallSliderValue: { type: Number, required: true },
  // options and labels
  divisionOptions: { type: Array, required: true },
  positionOptions: { type: Array, required: true },
  selectedDivisionsDisplayText: { type: String, required: true },
  selectedPositionsDisplayText: { type: String, required: true },
  // bounds
  maxMinutes: { type: Number, required: true },
  maxOverall: { type: Number, required: true },
  // callbacks
  filterDivisionsFn: { type: Function, required: true },
  filterPositionsFn: { type: Function, required: true },
  selectAllDivisions: { type: Function, required: true },
  clearAllDivisions: { type: Function, required: true },
  selectAllPositions: { type: Function, required: true },
  clearAllPositions: { type: Function, required: true },
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const selectedDivisionsModel = computed({
  get: () => props.selectedDivisions,
  set: (val) => emit('update:selectedDivisions', val),
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const selectedPositionsModel = computed({
  get: () => props.selectedPositions,
  set: (val) => emit('update:selectedPositions', val),
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const sliderValueModel = computed({
  get: () => props.sliderValue,
  set: (val) => emit('update:sliderValue', val),
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const overallSliderValueModel = computed({
  get: () => props.overallSliderValue,
  set: (val) => emit('update:overallSliderValue', val),
})

const {
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  filterDivisionsFn,
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  filterPositionsFn,
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  selectAllDivisions,
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  clearAllDivisions,
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  selectAllPositions,
  // biome-ignore lint/correctness/noUnusedVariables: used in template
  clearAllPositions,
} = props
</script>

<style lang="scss" scoped>
.filter-bar {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: var(--section-gap);
  align-items: start;

  @media (max-width: 1200px) {
    grid-template-columns: 1fr 1fr;
  }

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.division-filter-container,
.position-filter-container {
  width: 100%;
}

.division-filter-header,
.position-filter-header {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.filter-action-btn {
  font-size: 0.8rem;
}

.minutes-filter,
.overall-filter {
  .slider-label {
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-secondary);
    margin-bottom: 0.5rem;
  }
}

:deep(.q-item--disabled) {
  font-weight: 600;
  color: var(--text-secondary) !important;
  text-align: center;
  font-size: 0.8rem;
  pointer-events: none;
}
</style>
