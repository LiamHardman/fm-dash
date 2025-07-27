<template>
  <q-tr
    :props="props"
    @click="onRowClick(props.row)"
    @contextmenu="onRightClick($event, props.row)"
    class="cursor-pointer table-row-hover modern-table-row"
  >
    <q-td
      v-for="col in props.cols"
      :key="col.name"
      :props="props"
      :class="[
        col.classes,
        'table-cell-enhanced',
      ]"
      :style="col.style"
    >
      <template v-if="col.isFifaStat || col.isOverallStat">
        <span
          :class="
            getUnifiedRatingClass(
              getDisplayValue(props.row, col),
              100,
            )
          "
          class="attribute-value fifa-stat-value modern-stat-badge"
        >
          {{
            getDisplayValue(props.row, col) !== undefined
              ? getDisplayValue(props.row, col)
              : "-"
          }}
        </span>
      </template>
      <template v-else-if="col.isValueScore">
        <span
          :class="getValueScoreClass(props.row.valueScore)"
          class="attribute-value value-score-value modern-stat-badge"
        >
          {{ 
            props.row.valueScore !== undefined && props.row.valueScore !== null
              ? Math.round(props.row.valueScore)
              : "-"
          }}
        </span>
      </template>
      <template
        v-else-if="
          col.name === 'transfer_value' ||
          col.name === 'wage'
        "
      >
        <span
          :class="
            getMoneyClass(
              props.row[col.sortField || col.field],
            )
          "
          class="money-value"
        >
          {{
            formatDisplayCurrency(
              props.row[col.sortField || col.field],
              props.row[col.field],
            )
          }}
        </span>
      </template>
      <template
        v-else-if="col.name === 'nationality_display'"
      >
        <div class="flex items-center no-wrap nationality-cell">
          <img
            v-if="props.row.nationality_iso"
            :src="`https://flagcdn.com/w20/${props.row.nationality_iso.toLowerCase()}.png`"
            :alt="props.row.nationality || 'Flag'"
            width="20"
            height="13"
            class="nationality-flag flex-shrink-0"
            @error="onFlagError($event, props.row)"
          />
          <q-icon
            v-else
            name="flag"
            size="xs"
            :color="
              qInstance.dark.isActive
                ? 'grey-6'
                : 'grey-7'
            "
            class="nationality-flag-placeholder flex-shrink-0"
          />
          <span class="nationality-text">{{ props.row.nationality || "-" }}</span>
        </div>
      </template>
      <template v-else-if="col.name === 'club'">
        <div class="club-cell">
          <span 
            class="club-link"
            @click.stop="onClubClick(props.row)"
            :title="`View ${props.row[col.field]} team page`"
          >{{
            props.row[col.field] !== undefined &&
            props.row[col.field] !== null
              ? props.row[col.field]
              : "-"
          }}</span>
        </div>
      </template>
      <template v-else>
        <span>{{
          props.row[col.field] !== undefined &&
          props.row[col.field] !== null
            ? props.row[col.field]
            : "-"
        }}</span>
      </template>
    </q-td>
  </q-tr>
</template>

<script>
import { useQuasar } from 'quasar'
import {
  formatDisplayCurrency,
  getDisplayValue,
  getMoneyClass,
  getUnifiedRatingClass,
  getValueScoreClass,
  onFlagError,
} from './PlayerTableUtils'

export default {
  name: 'PlayerTableRow',
  props: {
    props: { type: Object, required: true },
    isGoalkeeperView: { type: Boolean, default: false },
    currencySymbol: { type: String, default: '$' },
    cacheGeneration: { type: Number, default: 0 },
  },
  emits: ['player-selected', 'team-selected', 'context-menu'],

  setup(props, { emit }) {
    const qInstance = useQuasar()

    const onRowClick = (player) => {
      emit('player-selected', player)
    }

    const onClubClick = (player) => {
      if (player.club && player.club.trim() !== '') {
        emit('team-selected', player.club)
      }
    }

    const onRightClick = (event, player) => {
      event.preventDefault()
      emit('context-menu', event, player)
    }

    const getDisplayValueForRow = (player, col) => {
      return getDisplayValue(player, col, props.isGoalkeeperView, props.cacheGeneration)
    }

    const formatDisplayCurrencyForRow = (numericAmount, originalDisplayValue) => {
      return formatDisplayCurrency(numericAmount, props.currencySymbol, originalDisplayValue)
    }

    return {
      qInstance,
      onRowClick,
      onClubClick,
      onRightClick,
      getDisplayValue: getDisplayValueForRow,
      getUnifiedRatingClass,
      getValueScoreClass,
      getMoneyClass,
      formatDisplayCurrency: formatDisplayCurrencyForRow,
      onFlagError,
    }
  },
}
</script>

<style lang="scss" scoped>
.modern-table-row {
  transition: all 0.2s ease;
  
  &:hover {
    background: rgba(46, 116, 181, 0.04) !important;
    box-shadow: 0 2px 8px rgba(46, 116, 181, 0.15);
    transform: translateY(-1px);
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.04) !important;
      box-shadow: 0 2px 8px rgba(255, 255, 255, 0.15);
    }
  }
  
  &:nth-child(even) {
    background: rgba(46, 116, 181, 0.02);
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.02);
    }
    
    &:hover {
      background: rgba(46, 116, 181, 0.06) !important;
      
      .body--dark & {
        background: rgba(255, 255, 255, 0.06) !important;
      }
    }
  }
}

.table-cell-enhanced {
  color: #334155 !important;
  font-weight: 500;
  padding: 0.75rem !important;
  
  .body--dark & {
    color: rgba(255, 255, 255, 0.85) !important;
  }
}

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