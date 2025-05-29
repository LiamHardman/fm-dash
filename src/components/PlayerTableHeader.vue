<template>
    <thead>
        <tr>
            <th
                v-for="col in columns"
                :key="col.name"
                :class="getHeaderClass(col)"
                class="text-center sortable-header"
                @click="$emit('sort', col.name)"
                style="cursor: pointer; user-select: none;"
            >
                <div class="header-content">
                    <span class="header-label">{{ col.label }}</span>
                    <span v-if="sortField === col.name" class="sort-indicator">
                        <q-icon 
                            :name="sortDirection === 'asc' ? 'keyboard_arrow_up' : 'keyboard_arrow_down'"
                            size="sm"
                            :color="quasarInstance.dark.isActive ? 'grey-4' : 'grey-7'"
                        />
                    </span>
                </div>
            </th>
        </tr>
    </thead>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed } from 'vue'

export default {
  name: 'PlayerTableHeader',
  props: {
    columns: { type: Array, required: true },
    sortField: { type: String, required: true },
    sortDirection: { type: String, required: true }
  },
  emits: ['sort'],
  setup(props) {
    const quasarInstance = useQuasar()

    const getHeaderClass = col => {
      const classes = ['sortable-header']

      if (props.sortField === col.name) {
        classes.push('active-sort')
      }

      if (col.type === 'number' || col.type === 'rating') {
        classes.push('text-right')
      }

      return classes.join(' ')
    }

    return {
      quasarInstance,
      getHeaderClass
    }
  }
}
</script>

<style scoped>
.sortable-header {
    padding: 12px 8px;
    border-bottom: 2px solid #e0e0e0;
    background-color: #fafafa;
    font-weight: 600;
    position: relative;
}

.dark .sortable-header {
    background-color: #424242;
    border-bottom-color: #616161;
    color: #e0e0e0;
}

.sortable-header:hover {
    background-color: #f0f0f0;
}

.dark .sortable-header:hover {
    background-color: #515151;
}

.active-sort {
    background-color: #e3f2fd !important;
}

.dark .active-sort {
    background-color: #1565c0 !important;
}

.header-content {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
}

.header-label {
    flex: 1;
}

.sort-indicator {
    flex-shrink: 0;
}
</style>