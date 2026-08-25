<template>
    <section class="section-card" :class="{ 'section-card--flat': flat }">
        <header v-if="title || $slots.header || $slots.actions" class="section-card__header">
            <slot name="header">
                <q-icon v-if="icon" :name="icon" class="section-card__icon" />
                <h2 v-if="title" class="section-card__title">{{ title }}</h2>
            </slot>
            <div class="section-card__spacer" />
            <div v-if="$slots.actions" class="section-card__actions">
                <slot name="actions" />
            </div>
        </header>
        <div class="section-card__body">
            <slot />
        </div>
    </section>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'SectionCard',
  props: {
    title: {
      type: String,
      default: '',
    },
    icon: {
      type: String,
      default: '',
    },
    flat: {
      type: Boolean,
      default: false,
    },
  },
})
</script>

<style lang="scss" scoped>
.section-card {
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-1);
    padding: var(--density-card-padding);

    &--flat {
        background: transparent;
        border: none;
        box-shadow: none;
        padding: 0;
    }

    &__header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin-bottom: var(--density-gap);
        flex-wrap: wrap;
    }

    &__icon {
        font-size: 1.15rem;
        color: var(--accent);
    }

    &__title {
        margin: 0;
        font-size: 1rem;
        font-weight: 600;
        color: var(--text-primary);
    }

    &__spacer {
        flex: 1 1 auto;
    }

    &__actions {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    &__body {
        color: var(--text-primary);
    }
}
</style>
