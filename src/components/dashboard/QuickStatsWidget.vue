<template>
    <SectionCard title="Dataset overview" icon="query_stats">
        <div class="quick-stats-grid">
            <StatTile label="Players" :value="playerCount" icon="groups" />
            <StatTile label="Clubs" :value="clubCount" icon="shield" />
            <StatTile label="Nations" :value="nationCount" icon="flag" />
            <StatTile label="Leagues" :value="divisionCount" icon="emoji_events" />
        </div>
    </SectionCard>
</template>

<script>
import { computed, defineComponent } from 'vue'
import { usePlayerStore } from '@/stores/playerStore'
import SectionCard from '../layout/SectionCard.vue'
import StatTile from '../layout/StatTile.vue'

export default defineComponent({
  name: 'QuickStatsWidget',
  components: { SectionCard, StatTile },
  setup() {
    const playerStore = usePlayerStore()

    const playerCount = computed(() => playerStore.allPlayers?.length || 0)
    const clubCount = computed(() => playerStore.uniqueClubs?.length || 0)
    const nationCount = computed(() => playerStore.uniqueNationalities?.length || 0)
    const divisionCount = computed(() => playerStore.uniqueDivisions?.length || 0)

    return { playerCount, clubCount, nationCount, divisionCount }
  },
})
</script>

<style lang="scss" scoped>
.quick-stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: var(--density-gap);
}
</style>
