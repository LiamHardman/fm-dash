<template>
    <div class="stat-tile">
        <div class="stat-tile__icon-wrap">
            <q-icon :name="icon" class="stat-tile__icon" />
        </div>
        <div class="stat-tile__text">
            <div class="stat-tile__value">{{ value }}</div>
            <div class="stat-tile__label">{{ label }}</div>
        </div>
        <div
            v-if="trend !== null && trend !== undefined && trend !== ''"
            class="stat-tile__trend"
            :class="`stat-tile__trend--${trendDirection}`"
        >
            <q-icon :name="trendIcon" size="0.9rem" />
            {{ trend }}
        </div>
    </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'StatTile',
  props: {
    label: {
      type: String,
      required: true,
    },
    value: {
      type: [String, Number],
      required: true,
    },
    icon: {
      type: String,
      default: 'insights',
    },
    trend: {
      type: [String, Number],
      default: null,
    },
    trendDirection: {
      type: String,
      default: 'neutral', // 'up' | 'down' | 'neutral'
    },
  },
  setup(props) {
    const trendIcon = computed(() => {
      if (props.trendDirection === 'up') return 'arrow_upward'
      if (props.trendDirection === 'down') return 'arrow_downward'
      return 'remove'
    })

    return { trendIcon }
  },
})
</script>

<style lang="scss" scoped>
.stat-tile {
    display: flex;
    align-items: center;
    gap: 0.85rem;
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-1);
    padding: var(--density-card-padding);
    min-width: 0;

    &__icon-wrap {
        flex-shrink: 0;
        width: 2.5rem;
        height: 2.5rem;
        border-radius: var(--radius-sm);
        background: var(--accent-soft);
        display: flex;
        align-items: center;
        justify-content: center;
    }

    &__icon {
        font-size: 1.3rem;
        color: var(--accent);
    }

    &__text {
        min-width: 0;
    }

    &__value {
        font-size: 1.4rem;
        font-weight: 700;
        color: var(--text-primary);
        line-height: 1.2;
    }

    &__label {
        font-size: 0.78rem;
        color: var(--text-secondary);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    &__trend {
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: 0.15rem;
        font-size: 0.78rem;
        font-weight: 600;
        flex-shrink: 0;

        &--up { color: #22c55e; }
        &--down { color: #ef4444; }
        &--neutral { color: var(--text-muted); }
    }
}
</style>
