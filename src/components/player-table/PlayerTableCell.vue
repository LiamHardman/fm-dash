<template>
  <!-- Circular gauge for all per-stat columns -->
  <StatGauge
    v-if="column.isFifaStat || column.isOverallStat"
    :value="typeof displayValue === 'number' ? displayValue : (displayValue !== undefined && displayValue !== null ? Number(displayValue) : null)"
    :max="column.gaugeMax || columnMaxValues[column.name] || 100"
  />

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

  <!-- Name cell: flag inline + personality/media subtitle -->
  <div v-else-if="column.name === 'name'" class="name-cell">
    <div class="name-primary">
      <img
        v-if="player.nationalityIso"
        :src="`https://flagcdn.com/w20/${player.nationalityIso.toLowerCase()}.png`"
        :alt="player.nationality || 'Flag'"
        width="16"
        height="11"
        class="name-flag"
        @error="onFlagError($event, player)"
      />
      <q-icon
        v-else
        name="flag"
        size="14px"
        class="name-flag-placeholder"
      />
      <span class="name-text">{{ player.name || '-' }}</span>
    </div>
    <div
      v-if="player.personality || player.media_handling"
      class="name-meta"
    >
      {{ [player.personality, player.media_handling].filter(Boolean).join(' · ') }}
    </div>
  </div>

  <div v-else-if="column.name === 'nationality_display'"
       class="flex items-center no-wrap nationality-cell">
    <img v-if="player.nationalityIso"
         :src="`https://flagcdn.com/w20/${player.nationalityIso.toLowerCase()}.png`"
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

  <!-- Position cell: position text + best role subtitle -->
  <div v-else-if="column.name === 'position'" class="position-cell">
    <span class="position-text">{{ player.position || '-' }}</span>
    <span v-if="player.bestRoleOverall" class="position-role">{{ player.bestRoleOverall }}</span>
  </div>

  <div v-else-if="column.name === 'club'" class="club-cell">
    <TeamLogo
      v-if="player[column.field] && player[column.field] !== '-'"
      :team-name="player[column.field]"
      :size="28"
      class="club-logo"
    />
    <div class="club-text">
      <span class="club-link"
            @click.stop="$emit('club-click', player)"
            :title="`View ${player[column.field]} team page`">
        {{ player[column.field] !== undefined && player[column.field] !== null
            ? player[column.field]
            : "-" }}
      </span>
      <span
        v-if="player.division"
        class="club-division"
        @click.stop="$emit('division-click', player.division)"
        :title="`View ${player.division} league page`"
      >{{ player.division }}</span>
    </div>
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
import TeamLogo from '../TeamLogo.vue'
import StatGauge from './StatGauge.vue'

export default {
  name: 'PlayerTableCell',
  components: { StatGauge, TeamLogo },
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
    columnMaxValues: {
      type: Object,
      default: () => ({}),
    },
  },
  emits: ['club-click', 'division-click'],
  setup() {
    const $q = useQuasar()
    const isDark = computed(() => $q.dark.isActive)
    return { isDark }
  },
}
</script>

<style lang="scss" scoped>
@import './PlayerTableStyles.scss';

.modern-stat-badge {
  padding: 3px 8px;
  border-radius: 8px;
  font-size: 0.78rem;
  font-weight: 700;
  text-align: center;
  min-width: 34px;
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

/* Name cell */
.name-cell {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.name-primary {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.name-flag {
  flex-shrink: 0;
  width: 16px !important;
  height: 11px !important;
  border-radius: 2px;
  border: 1px solid rgba(0, 0, 0, 0.15);
  object-fit: cover;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);

  .body--dark & {
    border-color: rgba(255, 255, 255, 0.15);
  }
}

.name-flag-placeholder {
  flex-shrink: 0;
  width: 16px;
  color: #9ca3af;

  .body--dark & {
    color: #6b7280;
  }
}

.name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
  min-width: 0;
}

.name-meta {
  font-size: 0.72rem;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;

  .body--dark & {
    color: rgba(255, 255, 255, 0.45);
  }
}

/* Legacy nationality cell (kept in case column is still used elsewhere) */
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
  flex-direction: row;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.club-logo {
  flex-shrink: 0;
}

.club-text {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
  overflow: hidden;
}

.club-cell .club-link {
  cursor: pointer;
  color: inherit;
  text-decoration: none;
  font-weight: 600;
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

/* Position cell */
.position-cell {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.position-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
}

.position-role {
  font-size: 0.72rem;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;

  .body--dark & {
    color: rgba(255, 255, 255, 0.45);
  }
}

.club-division {
  font-size: 0.72rem;
  color: #64748b;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
  cursor: pointer;

  &:hover {
    color: #2e74b5;
    text-decoration: underline;
  }

  .body--dark & {
    color: rgba(255, 255, 255, 0.45);

    &:hover {
      color: #60a5fa;
    }
  }
}
</style>
