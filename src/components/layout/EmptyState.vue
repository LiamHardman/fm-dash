<template>
    <div class="empty-state">
        <q-icon :name="icon" class="empty-state__icon" />
        <h3 class="empty-state__title">{{ title }}</h3>
        <p v-if="description" class="empty-state__description">{{ description }}</p>
        <div v-if="$slots.actions" class="empty-state__actions">
            <slot name="actions" />
        </div>
    </div>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EmptyState',
  props: {
    icon: {
      type: String,
      default: 'inbox',
    },
    title: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      default: '',
    },
  },
})
</script>

<style lang="scss" scoped>
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 3rem 1.5rem;
    color: var(--text-secondary);

    &__icon {
        font-size: 2.75rem;
        color: var(--text-muted);
        margin-bottom: 0.75rem;
    }

    &__title {
        margin: 0 0 0.35rem 0;
        font-size: 1.05rem;
        font-weight: 600;
        color: var(--text-primary);
    }

    &__description {
        margin: 0;
        max-width: 36rem;
        font-size: 0.9rem;
        line-height: 1.5;
    }

    &__actions {
        margin-top: 1.25rem;
        display: flex;
        gap: 0.75rem;
        flex-wrap: wrap;
        justify-content: center;
    }
}
</style>
