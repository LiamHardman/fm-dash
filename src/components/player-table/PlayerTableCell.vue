<template>
  <span v-if="column.isFifaStat || column.isOverallStat" 
        :class="[
          column.name === 'TotalStats'
            ? getTotalStatsRatingClass(displayValue)
            : getUnifiedRatingClass(displayValue, 100),
          'attribute-value fifa-stat-value modern-stat-badge'
        ]">
    {{ displayValue !== undefined ? displayValue : "-" }}
  </span>

  <span v-else-if="column.isValueScore"
        :class="[getValueScoreClass(player.valueScore), 'attribute-value value-score-value modern-stat-badge']">
    {{ player.valueScore !== undefined && player.valueScore !== null
        ? Math.round(player.valueScore)
        : "-" }}
  </span>

  <span v-else-if="column.name === 'transfer_value' || column.name === 'wage'"
        :class="[getMoneyClass(player[column.sortField || column.field]), 'money-value']">
    {{ formatDisplayCurrency(
        player[column.sortField || column.field],
        player[column.field]
      ) }}
  </span>

  <div v-else-if="column.name === 'nationality_display'" 
       class="flex items-center no-wrap nationality-cell">
    <img v-if="player.nationality_iso"
         :src="`https://flagcdn.com/w20/${player.nationality_iso.toLowerCase()}.png`"
         :alt="player.nationality || 'Flag'"
         width="20"
         height="13"
         class="nationality-flag flex-shrink-0"
         @error="onFlagError($event, player)" />
    <q-icon v-else
            name="flag"
            size="xs"
            :color="isDark ? 'grey-6' : 'grey-7'"
            class="nationality-flag-placeholder flex-shrink-0" />
    <span class="nationality-text">{{ player.nationality || "-" }}</span>
  </div>

  <div v-else-if="column.name === 'club'" class="club-cell">
    <span class="club-link"
          @click.stop="$emit('club-click', player)"
          :title="`View ${player[column.field]} team page`">
      {{ player[column.field] !== undefined && player[column.field] !== null
          ? player[column.field]
          : "-" }}
    </span>
  </div>

  <span v-else>
    {{ player[column.field] !== undefined && player[column.field] !== null
        ? player[column.field]
        : "-" }}
  </span>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed } from 'vue'

export default {
  name: 'PlayerTableCell',
  props: {
    player: {
      type: Object,
      required: true,
    },
    column: {
      type: Object,
      required: true,
    },
    displayValue: {
      type: [String, Number],
      default: undefined,
    },
    getUnifiedRatingClass: {
      type: Function,
      required: true,
    },
    getTotalStatsRatingClass: {
      type: Function,
      required: true,
    },
    getValueScoreClass: {
      type: Function,
      required: true,
    },
    getMoneyClass: {
      type: Function,
      required: true,
    },
    formatDisplayCurrency: {
      type: Function,
      required: true,
    },
    onFlagError: {
      type: Function,
      required: true,
    },
  },
  emits: ['club-click'],
  setup() {
    const $q = useQuasar()

    const isDark = computed(() => $q.dark.isActive)

    return {
      isDark,
    }
  },
}
</script>

<style lang="scss" scoped>
@import './PlayerTableStyles.scss';
.modern-stat-badge {
  padding: 4px 10px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 700;
  text-align: center;
  min-width: 36px;
  display: inline-block;
  border: 1px solid transparent;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  
  .body--dark & {
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  }
}

.money-value {
  font-weight: 500;
}

.money-uniform {
  color: #334155;
  .body--dark & {
    color: rgba(255, 255, 255, 0.85);
  }
}

.money-na {
  color: #9ca3af;
  .body--dark & {
    color: #6b7280;
  }
}

.nationality-flag {
  border: 1px solid rgba(0, 0, 0, 0.15);
  object-fit: cover;
  margin-right: 8px;
  width: 20px !important;
  height: 13px !important;
  flex-shrink: 0;
  border-radius: 3px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);

  .body--dark & {
    border: 1px solid rgba(255, 255, 255, 0.15);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  }
}

.nationality-flag-placeholder {
  margin-right: 8px;
  width: 20px;
  height: 13px;
  flex-shrink: 0;
  color: #9ca3af;
  
  .body--dark & {
    color: #6b7280;
  }
}

.nationality-cell {
  width: 100%;
  overflow: hidden;
  
  .nationality-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
    font-weight: 500;
  }
}

.club-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.club-cell .club-link {
  cursor: pointer;
  color: inherit;
  text-decoration: none;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.club-cell .club-link:hover {
  text-decoration: underline;
}

.body--dark .club-cell .club-link:hover {
  color: #81C784;
}

.body--light .club-cell .club-link:hover {
  color: #2E7D32;
}
</style> 