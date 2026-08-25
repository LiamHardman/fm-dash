<template>
    <SectionCard title="Wishlist" icon="favorite">
        <template #actions>
            <q-btn flat dense no-caps color="primary" label="View all" to="/wishlist" />
        </template>
        <div v-if="count > 0" class="wishlist-summary">
            <StatTile label="Players saved" :value="count" icon="favorite" />
        </div>
        <EmptyState
            v-else
            icon="favorite_border"
            title="No players wishlisted yet"
            description="Save players from the player table or their detail view to track them here."
        />
    </SectionCard>
</template>

<script>
import { computed, defineComponent } from 'vue'
import { usePlayerStore } from '@/stores/playerStore'
import { useWishlistStore } from '@/stores/wishlistStore'
import EmptyState from '../layout/EmptyState.vue'
import SectionCard from '../layout/SectionCard.vue'
import StatTile from '../layout/StatTile.vue'

export default defineComponent({
  name: 'WishlistWidget',
  components: { SectionCard, StatTile, EmptyState },
  setup() {
    const playerStore = usePlayerStore()
    const wishlistStore = useWishlistStore()

    const count = computed(() => wishlistStore.getWishlistCount(playerStore.currentDatasetId))

    return { count }
  },
})
</script>

<style lang="scss" scoped>
.wishlist-summary {
    display: grid;
    grid-template-columns: minmax(140px, 220px);
}
</style>
