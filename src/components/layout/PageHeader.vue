<template>
    <div class="page-header">
        <q-btn
            v-if="backTo"
            flat
            round
            dense
            icon="arrow_back"
            class="page-header__back"
            :to="backTo"
            aria-label="Back"
        />
        <q-icon v-if="icon" :name="icon" class="page-header__icon" />
        <div class="page-header__text">
            <h1 class="page-header__title">{{ title }}</h1>
            <p v-if="subtitle" class="page-header__subtitle">{{ subtitle }}</p>
        </div>
        <div class="page-header__spacer" />
        <div v-if="$slots.actions" class="page-header__actions">
            <slot name="actions" />
        </div>
    </div>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'PageHeader',
  props: {
    title: {
      type: String,
      required: true,
    },
    subtitle: {
      type: String,
      default: '',
    },
    icon: {
      type: String,
      default: '',
    },
    backTo: {
      type: [String, Object],
      default: null,
    },
  },
})
</script>

<style lang="scss" scoped>
.page-header {
    display: flex;
    align-items: flex-start;
    gap: 0.85rem;
    padding: 0 0 var(--section-gap) 0;
    flex-wrap: wrap;

    &__back {
        color: var(--text-secondary);
        margin-top: 0.1rem;
    }

    &__icon {
        font-size: 1.9rem;
        color: var(--accent);
        margin-top: 0.15rem;
        flex-shrink: 0;
    }

    &__text {
        min-width: 0;
    }

    &__title {
        margin: 0;
        font-size: 1.5rem;
        font-weight: 700;
        color: var(--text-primary);
        line-height: 1.25;
    }

    &__subtitle {
        margin: 0.2rem 0 0 0;
        font-size: 0.9rem;
        color: var(--text-secondary);
        line-height: 1.4;
    }

    &__spacer {
        flex: 1 1 auto;
    }

    &__actions {
        display: flex;
        align-items: center;
        gap: 0.6rem;
        flex-wrap: wrap;
    }
}
</style>
