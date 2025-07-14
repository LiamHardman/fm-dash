<template>
    <tr
        :class="rowClass"
        @click="$emit('player-selected', player)"
        @contextmenu.prevent="$emit('context-menu', player, $event)"
        style="cursor: pointer;"
    >
        <td
            v-for="col in columns"
            :key="col.name"
            :class="getCellClass(col)"
            class="text-center player-cell"
        >
            <span v-if="col.name === 'club' && displayValue(col)" class="club-link">
                <div class="club-cell">
                    <TeamLogo 
                        :team-name="displayValue(col)"
                        :size="18"
                        class="q-mr-xs"
                        v-if="displayValue(col) && displayValue(col) !== '-'"
                    />
                    <a
                        href="#"
                        @click.stop="$emit('team-selected', displayValue(col))"
                        :style="{ color: quasarInstance.dark.isActive ? '#81C784' : '#2E7D32' }"
                    >
                        {{ displayValue(col) }}
                    </a>
                </div>
            </span>
            <span v-else-if="col.type === 'currency'">
                {{ formatCurrency(displayValue(col)) }}
            </span>
            <span v-else-if="col.type === 'rating'" :class="getRatingClass(displayValue(col))">
                {{ displayValue(col) }}
            </span>
            <span v-else>
                {{ displayValue(col) }}
            </span>
        </td>
    </tr>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed } from 'vue'

export default {
  name: 'PlayerTableRow',
  props: {
    player: { type: Object, required: true },
    columns: { type: Array, required: true },
    currencySymbol: { type: String, default: '$' },
    isGoalkeeperView: { type: Boolean, default: false },
    getDisplayValue: { type: Function, required: true },
    formatCurrency: { type: Function, required: true },
    getRatingClass: { type: Function, required: true }
  },
  emits: ['player-selected', 'context-menu', 'team-selected'],
  setup(props) {
    const quasarInstance = useQuasar()

    const rowClass = computed(() => {
      const classes = []

      if (quasarInstance.dark.isActive) {
        classes.push('dark-row')
      }

      // Add any player-specific classes here
      if (props.player.position?.includes('GK')) {
        classes.push('goalkeeper-row')
      }

      return classes.join(' ')
    })

    const getCellClass = col => {
      const classes = ['player-cell']

      if (col.type === 'number' || col.type === 'rating') {
        classes.push('text-right')
      }

      return classes.join(' ')
    }

    const displayValue = col => props.getDisplayValue(props.player, col)

    return {
      quasarInstance,
      rowClass,
      getCellClass,
      displayValue
    }
  }
}
</script>

<style scoped>
.player-cell {
    padding: 8px 4px;
    vertical-align: middle;
}

.club-link a {
    text-decoration: none;
    font-weight: 500;
}

.club-link a:hover {
    text-decoration: underline;
}

.goalkeeper-row {
    background-color: rgba(255, 255, 0, 0.1);
}

.dark-row {
    background-color: rgba(255, 255, 255, 0.05);
}
</style>