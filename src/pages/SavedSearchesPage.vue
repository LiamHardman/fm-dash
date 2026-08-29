<template>
  <q-page class="saved-searches-page">
    <div class="page-container">
      <PageHeader title="Saved searches" subtitle="Reusable, private filter recipes that apply to any Dataset." icon="bookmarks">
        <template #actions><q-btn unelevated color="primary" icon="search" label="Browse players" :to="currentDatasetId ? `/dataset/${currentDatasetId}` : '/upload'" /></template>
      </PageHeader>

      <q-banner v-if="draftFilters" rounded class="q-mb-md bg-blue-1 text-blue-10">
        This shared recipe is a draft. Choose a Dataset before applying it, then save it if you want to keep it.
        <template #action><q-btn flat label="Apply to current Dataset" :disable="!currentDatasetId" @click="applyDraft" /></template>
      </q-banner>

      <EmptyState v-if="!searchStore.document.searches.length && !draftFilters" icon="bookmarks" title="No saved searches" description="Save a player search to reuse its criteria later." />
      <div v-else class="search-list">
        <q-card v-for="search in searchStore.document.searches" :key="search.id" flat bordered class="search-card">
          <q-card-section class="row items-center q-col-gutter-md">
            <div class="col">
              <div class="text-h6">{{ search.name }}</div>
              <div class="text-caption text-grey-7">{{ describeSavedSearch(search.filters).slice(0, 3).join(' · ') || 'Broad result set' }}</div>
            </div>
            <div class="col-auto row q-gutter-xs">
              <q-btn flat dense icon="play_arrow" label="Apply" :disable="!currentDatasetId" @click="applySearch(search)" />
              <q-btn flat dense icon="link" label="Copy recipe" @click="copyRecipe(search.filters)" />
              <q-btn flat dense icon="delete_outline" color="negative" aria-label="Delete search" @click="removeSearch(search)" />
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<script>
import { computed, defineComponent, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import EmptyState from '../components/layout/EmptyState.vue'
import PageHeader from '../components/layout/PageHeader.vue'
import { usePlayerStore } from '../stores/playerStore'
import { useSavedSearchStore } from '../stores/savedSearchStore'
import { decodeSearchRecipe, describeSavedSearch, encodeSearchRecipe } from '../utils/savedSearch'

export default defineComponent({
  name: 'SavedSearchesPage',
  components: { EmptyState, PageHeader },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const playerStore = usePlayerStore()
    const searchStore = useSavedSearchStore()
    const currentDatasetId = computed(() => playerStore.currentDatasetId)
    const draftFilters = computed(() =>
      typeof route.query.recipe === 'string' ? decodeSearchRecipe(route.query.recipe) : null
    )
    onMounted(() => searchStore.load())
    const applySearch = (search) => {
      if (currentDatasetId.value)
        router.push({
          path: `/dataset/${currentDatasetId.value}`,
          query: { savedSearchId: search.id },
        })
    }
    const applyDraft = () => {
      if (currentDatasetId.value && draftFilters.value)
        router.push({
          path: `/dataset/${currentDatasetId.value}`,
          query: { recipe: route.query.recipe },
        })
    }
    const copyRecipe = async (filters) => {
      await navigator.clipboard.writeText(
        `${window.location.origin}/saved-searches?recipe=${encodeSearchRecipe(filters)}`
      )
    }
    const removeSearch = (search) => searchStore.remove(search)
    return {
      searchStore,
      currentDatasetId,
      draftFilters,
      describeSavedSearch,
      applySearch,
      applyDraft,
      copyRecipe,
      removeSearch,
    }
  },
})
</script>

<style scoped>
.saved-searches-page { min-height: 100vh; }
.page-container { max-width: var(--content-max-width); margin: 0 auto; padding: var(--page-gutter); }
.search-list { display: grid; gap: 0.75rem; }
.search-card { background: var(--surface-card); }
@media (max-width: 600px) { .page-container { padding: var(--page-gutter-sm); } }
</style>
