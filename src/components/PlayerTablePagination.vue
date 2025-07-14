<template>
    <div class="pagination-container">
        <div class="pagination-info">
            <span class="text-body2">
                Showing {{ startItem }}-{{ endItem }} of {{ totalItems }} players
                <span v-if="isSliced" class="text-warning">
                    (limited to {{ maxDisplayItems }} for performance)
                </span>
            </span>
        </div>
        
        <div class="pagination-controls" v-if="totalPages > 1">
            <q-btn
                flat
                dense
                icon="first_page"
                :disable="currentPage === 1"
                @click="$emit('page-change', 1)"
                class="pagination-btn"
            />
            <q-btn
                flat
                dense
                icon="chevron_left"
                :disable="currentPage === 1"
                @click="$emit('page-change', currentPage - 1)"
                class="pagination-btn"
            />
            
            <div class="page-numbers">
                <q-btn
                    v-for="page in visiblePages"
                    :key="page"
                    flat
                    dense
                    :label="page.toString()"
                    :color="page === currentPage ? 'primary' : undefined"
                    :outline="page === currentPage"
                    @click="$emit('page-change', page)"
                    class="page-btn"
                />
            </div>
            
            <q-btn
                flat
                dense
                icon="chevron_right"
                :disable="currentPage === totalPages"
                @click="$emit('page-change', currentPage + 1)"
                class="pagination-btn"
            />
            <q-btn
                flat
                dense
                icon="last_page"
                :disable="currentPage === totalPages"
                @click="$emit('page-change', totalPages)"
                class="pagination-btn"
            />
        </div>
    </div>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'PlayerTablePagination',
  props: {
    currentPage: { type: Number, required: true },
    totalItems: { type: Number, required: true },
    itemsPerPage: { type: Number, required: true },
    maxDisplayItems: { type: Number, default: 1000 },
    isSliced: { type: Boolean, default: false },
    maxPagesToShow: { type: Number, default: 7 }
  },
  emits: ['page-change'],
  setup(props) {
    const totalPages = computed(() => {
      return Math.ceil(props.totalItems / props.itemsPerPage)
    })

    const startItem = computed(() => {
      return props.totalItems === 0 ? 0 : (props.currentPage - 1) * props.itemsPerPage + 1
    })

    const endItem = computed(() => {
      return Math.min(props.currentPage * props.itemsPerPage, props.totalItems)
    })

    const visiblePages = computed(() => {
      const pages = []
      const total = totalPages.value
      const current = props.currentPage
      const maxShow = props.maxPagesToShow

      if (total <= maxShow) {
        // Show all pages if total is less than max
        for (let i = 1; i <= total; i++) {
          pages.push(i)
        }
      } else {
        // Calculate start and end of visible range
        let start = Math.max(1, current - Math.floor(maxShow / 2))
        const end = Math.min(total, start + maxShow - 1)

        // Adjust start if end is at the boundary
        if (end === total) {
          start = Math.max(1, end - maxShow + 1)
        }

        for (let i = start; i <= end; i++) {
          pages.push(i)
        }
      }

      return pages
    })

    return {
      totalPages,
      startItem,
      endItem,
      visiblePages
    }
  }
}
</script>

<style scoped>
.pagination-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    border-top: 1px solid #e0e0e0;
    margin-top: 16px;
}

.pagination-info {
    color: #666;
}

.pagination-controls {
    display: flex;
    align-items: center;
    gap: 4px;
}

.page-numbers {
    display: flex;
    gap: 2px;
    margin: 0 8px;
}

.pagination-btn, .page-btn {
    min-width: 32px;
    height: 32px;
}

.text-warning {
    color: #f57c00;
    font-weight: 500;
}

@media (max-width: 768px) {
    .pagination-container {
        flex-direction: column;
        gap: 12px;
    }
    
    .page-numbers {
        margin: 0 4px;
    }
}
</style>