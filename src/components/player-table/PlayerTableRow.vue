<template>
  <q-tr :props="props"
        @click="$emit('row-click', props.row)"
        @contextmenu="$emit('right-click', $event, props.row)"
        class="cursor-pointer table-row-hover modern-table-row">
    <q-td v-for="col in props.cols"
          :key="col.name"
          :props="props"
          :class="[col.classes, 'table-cell-enhanced']"
          :style="col.style">
      <PlayerTableCell 
        :player="props.row"
        :column="col"
        :display-value="getDisplayValue(props.row, col)"
        :get-unified-rating-class="getUnifiedRatingClass"
        :get-total-stats-rating-class="getTotalStatsRatingClass"
        :get-value-score-class="getValueScoreClass"
        :get-money-class="getMoneyClass"
        :format-display-currency="formatDisplayCurrency"
        :on-flag-error="onFlagError"
        @club-click="$emit('club-click', $event)"
      />
    </q-td>
  </q-tr>
</template>

<script>
import PlayerTableCell from './PlayerTableCell.vue'

export default {
  name: 'PlayerTableRow',
  components: {
    PlayerTableCell,
  },
  props: {
    props: {
      type: Object,
      required: true,
    },
    getDisplayValue: {
      type: Function,
      required: true,
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
  emits: ['row-click', 'right-click', 'club-click'],
}
</script>

<style lang="scss" scoped>
@import './PlayerTableStyles.scss';
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
</style> 