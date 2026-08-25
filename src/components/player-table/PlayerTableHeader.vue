<template>
  <q-tr :props="props" class="modern-table-header">
    <q-th v-for="col in props.cols"
          :key="col.name"
          :props="props"
          class="text-weight-bold modern-header-cell"
          :class="[
            col.headerClasses,
            { 'active-sort': sortField === col.name },
            { 'cursor-pointer': true },
            { 'sorting-in-progress': isAsyncSorting && sortField === col.name }
          ]"
          :style="col.headerStyle"
          @click="$emit('sort-table', col.name)">
      
      <span v-if="col.name === 'transfer_value' || col.name === 'wage'">
        {{ col.label }} ({{ currencySymbol }})
      </span>
      <span v-else-if="col.name === 'TotalStats'">
        <q-tooltip>Total Stats</q-tooltip>
        {{ col.label }}
      </span>
      <span v-else-if="col.name === 'MBR'">
        <q-tooltip>Moneyball Rating</q-tooltip>
        {{ col.label }}
      </span>
      <span v-else>
        {{ col.label }}
      </span>
      
      <q-icon v-if="sortField === col.name && !isAsyncSorting"
              :name="sortDirection === 'asc' ? 'arrow_upward' : 'arrow_downward'"
              size="xs"
              class="q-ml-xs sort-icon" />
      
      <q-spinner-dots v-if="sortField === col.name && isAsyncSorting"
                      size="xs"
                      color="primary"
                      class="q-ml-xs sorting-spinner" />
    </q-th>
  </q-tr>
</template>

<script>
export default {
  name: 'PlayerTableHeader',
  props: {
    props: {
      type: Object,
      required: true,
    },
    sortField: {
      type: String,
      default: '',
    },
    sortDirection: {
      type: String,
      default: 'asc',
    },
    isAsyncSorting: {
      type: Boolean,
      default: false,
    },
    currencySymbol: {
      type: String,
      default: '$',
    },
  },
  emits: ['sort-table'],
}
</script>

<style lang="scss" scoped>
.modern-table-header {
  background: linear-gradient(180deg, rgba(46, 116, 181, 0.08) 0%, rgba(46, 116, 181, 0.12) 100%);
  
  .body--dark & {
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.12) 100%);
  }
}

.modern-header-cell {
  color: var(--text-primary) !important;
  font-weight: 700 !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 0.8rem !important;
  padding: var(--density-cell-padding) !important;

  &:hover {
    background: rgba(46, 116, 181, 0.15) !important;
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.15) !important;
    }
  }
  
  &.active-sort {
    background: rgba(46, 116, 181, 0.2) !important;
    color: var(--accent) !important;

    .body--dark & {
      background: rgba(255, 255, 255, 0.2) !important;
    }

    .sort-icon {
      color: var(--accent) !important;
    }
  }

  &.sorting-in-progress {
    background: rgba(46, 116, 181, 0.15) !important;

    .body--dark & {
      background: rgba(255, 255, 255, 0.15) !important;
    }

    .sorting-spinner {
      color: var(--accent) !important;
    }
  }
}
</style> 