<template>
  <div class="pagination-container">
    <q-pagination
      v-model="currentPage"
      :max="pagesNumber"
      :max-pages="maxPagesToShow"
      boundary-links
      direction-links
      @update:model-value="onPageChange"
      color="primary"
      active-color="primary"
      text-color="primary"
      active-text-color="white"
    />
    <q-space />
    <span
      class="q-ml-md text-caption pagination-info"
    >
      {{ paginationStartRow }} - {{ paginationEndRow }} of
      {{ paginationTotalRows }}
      <span v-if="isSliced" class="text-italic q-ml-xs"
        >(from {{ totalSortedCount }} total sorted)</span
      >
    </span>
  </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'PlayerTablePagination',
  props: {
    sortedPlayers: { type: Array, required: true },
    pagination: { type: Object, required: true },
    isSliced: { type: Boolean, default: false },
    totalSortedCount: { type: Number, default: 0 },
    maxPagesToShow: { type: Number, default: 7 },
  },
  emits: ['page-change'],

  setup(props, { emit }) {
    const currentPage = computed({
      get: () => props.pagination.page,
      set: (value) => emit('page-change', value),
    })

    const pagesNumber = computed(() => {
      if (
        !props.sortedPlayers ||
        props.sortedPlayers.length === 0 ||
        props.pagination.rowsPerPage === 0
      ) {
        return 1
      }
      return Math.ceil(props.sortedPlayers.length / props.pagination.rowsPerPage)
    })

    const paginationTotalRows = computed(() => props.sortedPlayers.length)

    const paginationStartRow = computed(() => {
      if (paginationTotalRows.value === 0) return 0
      return (props.pagination.page - 1) * props.pagination.rowsPerPage + 1
    })

    const paginationEndRow = computed(() => {
      if (paginationTotalRows.value === 0) return 0
      if (props.pagination.rowsPerPage === 0) return paginationTotalRows.value
      return Math.min(
        props.pagination.page * props.pagination.rowsPerPage,
        paginationTotalRows.value
      )
    })

    const onPageChange = (newPage) => {
      emit('page-change', newPage)
    }

    return {
      currentPage,
      pagesNumber,
      paginationTotalRows,
      paginationStartRow,
      paginationEndRow,
      onPageChange,
    }
  },
}
</script>

<style lang="scss" scoped>
.pagination-info {
  color: #64748b !important;
  font-weight: 500;
  
  .body--dark & {
    color: rgba(255, 255, 255, 0.7) !important;
  }
}
</style> 