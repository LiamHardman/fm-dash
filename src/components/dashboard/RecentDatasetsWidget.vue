<template>
    <SectionCard title="Recent saves" icon="history">
        <q-list v-if="otherEntries.length" separator class="recent-list">
            <q-item
                v-for="entry in otherEntries"
                :key="entry.datasetId"
                clickable
                v-ripple
                class="recent-item"
                @click="openDataset(entry.datasetId)"
            >
                <q-item-section avatar>
                    <q-icon name="folder_open" />
                </q-item-section>
                <q-item-section>
                    <q-item-label class="recent-item__label">{{ entry.label }}</q-item-label>
                    <q-item-label caption>
                        {{ entry.playerCount }} players · {{ formatDate(entry.uploadedAt) }}
                    </q-item-label>
                </q-item-section>
                <q-item-section side>
                    <q-btn
                        flat
                        round
                        dense
                        icon="close"
                        aria-label="Remove from recent saves"
                        @click.stop="remove(entry.datasetId)"
                    />
                </q-item-section>
            </q-item>
        </q-list>
        <EmptyState
            v-else
            icon="folder_open"
            title="No other recent saves"
            description="Saves you upload will show up here so you can jump back into them."
        />
    </SectionCard>
</template>

<script>
import { computed, defineComponent, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '@/stores/playerStore'
import { useRecentDatasetsStore } from '@/stores/recentDatasetsStore'
import EmptyState from '../layout/EmptyState.vue'
import SectionCard from '../layout/SectionCard.vue'

export default defineComponent({
  name: 'RecentDatasetsWidget',
  components: { SectionCard, EmptyState },
  setup() {
    const router = useRouter()
    const playerStore = usePlayerStore()
    const recentDatasetsStore = useRecentDatasetsStore()

    onMounted(() => {
      recentDatasetsStore.load()
    })

    const otherEntries = computed(() =>
      recentDatasetsStore.entries.filter((e) => e.datasetId !== playerStore.currentDatasetId)
    )

    const openDataset = (datasetId) => {
      router.push(`/dataset/${datasetId}`)
    }

    const remove = (datasetId) => {
      recentDatasetsStore.removeDataset(datasetId)
    }

    const formatDate = (timestamp) => {
      if (!timestamp) return ''
      return new Date(timestamp).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
    }

    return { otherEntries, openDataset, remove, formatDate }
  },
})
</script>

<style lang="scss" scoped>
.recent-list {
    margin: 0 -0.5rem;
}

.recent-item {
    border-radius: var(--radius-sm);

    &__label {
        font-weight: 600;
        color: var(--text-primary);
    }
}
</style>
