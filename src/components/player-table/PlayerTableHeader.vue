<template>
  <q-tr
    :props="props"
    class="modern-table-header"
  >
    <q-th
      v-for="col in props.cols"
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
      @click="sortTable(col.name)"
    >
      <span
        v-if="
          col.name === 'transfer_value' ||
          col.name === 'wage'
        "
      >
        {{ col.label }} ({{ currencySymbol }})
      </span>
      <span v-else>
        {{ col.label }}
      </span>
      <q-icon
        v-if="sortField === col.name && !isAsyncSorting"
        :name="
          sortDirection === 'asc'
            ? 'arrow_upward'
            : 'arrow_downward'
        "
        size="xs"
        class="q-ml-xs sort-icon"
      />
      <!-- Subtle sorting indicator -->
      <q-spinner-dots
        v-if="sortField === col.name && isAsyncSorting"
        size="xs"
        color="primary"
        class="q-ml-xs sorting-spinner"
      />
    </q-th>
  </q-tr>
</template>

<script>
export default {
  name: 'PlayerTableHeader',
  props: {
    props: { type: Object, required: true },
    sortField: { type: String, required: true },
    sortDirection: { type: String, required: true },
    isAsyncSorting: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
  },
  emits: ['sort'],

  setup(props, { emit }) {
    const sortTable = (fieldName) => {
      emit('sort', fieldName)
    }

    return {
      sortTable,
    }
  },
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
  color: #1e293b !important;
  font-weight: 700 !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 0.8rem !important;
  padding: 1rem 0.75rem !important;
  
  .body--dark & {
    color: rgba(255, 255, 255, 0.9) !important;
  }

  &:hover {
    background: rgba(46, 116, 181, 0.15) !important;
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.15) !important;
    }
  }
  
  &.active-sort {
    background: rgba(46, 116, 181, 0.2) !important;
    color: #2e74b5 !important;
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.2) !important;
      color: #60a5fa !important;
    }
    
    .sort-icon {
      color: #2e74b5 !important;
      
      .body--dark & {
        color: #60a5fa !important;
      }
    }
  }
  
  &.sorting-in-progress {
    background: rgba(46, 116, 181, 0.15) !important;
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.15) !important;
    }
    
    .sorting-spinner {
      color: #2e74b5 !important;
      
      .body--dark & {
        color: #60a5fa !important;
      }
    }
  }
}
</style> 