<template>
    <SectionCard title="Jump to" icon="explore">
        <div class="shortcuts-grid">
            <q-btn
                v-for="shortcut in shortcuts"
                :key="shortcut.to"
                unelevated
                no-caps
                align="left"
                class="shortcut-btn"
                :to="shortcut.to"
            >
                <q-icon :name="shortcut.icon" class="q-mr-sm" />
                {{ shortcut.label }}
            </q-btn>
        </div>
    </SectionCard>
</template>

<script>
import { computed, defineComponent } from 'vue'
import { usePlayerStore } from '@/stores/playerStore'
import SectionCard from '../layout/SectionCard.vue'

export default defineComponent({
  name: 'ShortcutsWidget',
  components: { SectionCard },
  setup() {
    const playerStore = usePlayerStore()

    const shortcuts = computed(() => [
      {
        label: 'Players',
        icon: 'groups',
        to: playerStore.currentDatasetId ? `/dataset/${playerStore.currentDatasetId}` : '/upload',
      },
      { label: 'Teams', icon: 'shield', to: '/teams' },
      { label: 'Nations', icon: 'flag', to: '/nations' },
      { label: 'Leagues', icon: 'emoji_events', to: '/leagues' },
      { label: 'Performance', icon: 'leaderboard', to: '/performance' },
      { label: 'Wishlist', icon: 'favorite', to: '/wishlist' },
    ])

    return { shortcuts }
  },
})
</script>

<style lang="scss" scoped>
.shortcuts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: var(--density-gap);
}

.shortcut-btn {
    background: var(--surface-raised);
    color: var(--text-primary);
    border-radius: var(--radius-sm);
    justify-content: flex-start;

    .q-icon {
        color: var(--accent);
    }
}
</style>
