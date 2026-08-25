<template>
  <div class="pagination-container">
    <q-pagination v-model="scope.pagination.page"
                  :max="pagesNumber"
                  :max-pages="maxPagesToShow"
                  boundary-links
                  direction-links
                  @update:model-value="$emit('page-change', $event)"
                  color="primary"
                  active-color="primary"
                  text-color="primary"
                  active-text-color="white" />
    
    <q-space />
    
    <span class="q-ml-md text-caption pagination-info">
      {{ paginationStartRow }} - {{ paginationEndRow }} of {{ paginationTotalRows }}
      <span v-if="isSliced" class="text-italic q-ml-xs">
        (from {{ totalSortedCount }} total sorted)
      </span>
    </span>
  </div>
</template>

<script>
export default {
  name: 'PlayerTablePagination',
  props: {
    scope: {
      type: Object,
      required: true,
    },
    pagesNumber: {
      type: Number,
      required: true,
    },
    maxPagesToShow: {
      type: Number,
      default: 7,
    },
    paginationStartRow: {
      type: Number,
      required: true,
    },
    paginationEndRow: {
      type: Number,
      required: true,
    },
    paginationTotalRows: {
      type: Number,
      required: true,
    },
    totalSortedCount: {
      type: Number,
      default: 0,
    },
    isSliced: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['page-change'],
}
</script>

<style lang="scss" scoped>
.pagination-container {
  display: flex;
  align-items: center;
  width: 100%;
}

.pagination-info {
  color: var(--text-secondary) !important;
  font-weight: 500;
}
</style> 