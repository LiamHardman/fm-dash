<template>
  <q-menu ref="contextMenu"
          touch-position 
          context-menu
          :offset="[10, 10]">
    <q-list dense style="min-width: 180px">
      <q-item clickable 
              v-close-popup 
              @click="$emit('add-to-wishlist')"
              v-if="contextMenuPlayer && !isPlayerInWishlist">
        <q-item-section avatar>
          <q-icon name="favorite_border" color="positive" />
        </q-item-section>
        <q-item-section>Add to Wishlist</q-item-section>
      </q-item>
      
      <q-item clickable 
              v-close-popup 
              @click="$emit('remove-from-wishlist')"
              v-if="contextMenuPlayer && isPlayerInWishlist">
        <q-item-section avatar>
          <q-icon name="favorite" color="negative" />
        </q-item-section>
        <q-item-section>Remove from Wishlist</q-item-section>
      </q-item>
      
      <q-separator />

      <q-item clickable
              v-close-popup
              @click="$emit('add-to-comparison')"
              v-if="contextMenuPlayer && !isPlayerInComparison">
        <q-item-section avatar>
          <q-icon name="compare_arrows" color="primary" />
        </q-item-section>
        <q-item-section>Add to Comparison</q-item-section>
      </q-item>

      <q-item clickable
              v-close-popup
              @click="$emit('remove-from-comparison')"
              v-if="contextMenuPlayer && isPlayerInComparison">
        <q-item-section avatar>
          <q-icon name="compare_arrows" color="warning" />
        </q-item-section>
        <q-item-section>Remove from Comparison</q-item-section>
      </q-item>

      <q-separator />

      <q-item clickable
              v-close-popup
              @click="$emit('player-details')"
              v-if="contextMenuPlayer">
        <q-item-section avatar>
          <q-icon name="info" color="info" />
        </q-item-section>
        <q-item-section>View Details</q-item-section>
      </q-item>
    </q-list>
  </q-menu>
</template>

<script>
export default {
  name: 'PlayerTableContextMenu',
  props: {
    contextMenuPlayer: {
      type: Object,
      default: null,
    },
    isPlayerInWishlist: {
      type: Boolean,
      default: false,
    },
    isPlayerInComparison: {
      type: Boolean,
      default: false,
    },
  },
  emits: [
    'add-to-wishlist',
    'remove-from-wishlist',
    'player-details',
    'add-to-comparison',
    'remove-from-comparison',
  ],
}
</script>

<style scoped>
/* Context menu styles are minimal since Quasar handles most styling */
</style> 