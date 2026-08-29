<template>
  <q-menu ref="contextMenu"
          touch-position 
          context-menu
          :offset="[10, 10]">
    <q-list dense style="min-width: 180px">
      <q-item clickable 
              v-close-popup 
              @click="$emit('add-to-shortlist')"
              v-if="contextMenuPlayer && !isPlayerInShortlist">
        <q-item-section avatar>
          <q-icon name="favorite_border" color="positive" />
        </q-item-section>
        <q-item-section>Add to Shortlist</q-item-section>
      </q-item>
      
      <q-item clickable 
              v-close-popup 
              @click="$emit('remove-from-shortlist')"
              v-if="contextMenuPlayer && isPlayerInShortlist">
        <q-item-section avatar>
          <q-icon name="favorite" color="negative" />
        </q-item-section>
        <q-item-section>Remove from Shortlist</q-item-section>
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
      <q-item clickable
              v-close-popup
              @click="$emit('why-this-player')"
              v-if="contextMenuPlayer && hasMatchFilters">
        <q-item-section avatar>
          <q-icon name="help_outline" color="secondary" />
        </q-item-section>
        <q-item-section>Why this player?</q-item-section>
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
    isPlayerInShortlist: {
      type: Boolean,
      default: false,
    },
    isPlayerInComparison: {
      type: Boolean,
      default: false,
    },
    hasMatchFilters: {
      type: Boolean,
      default: false,
    },
  },
  emits: [
    'add-to-shortlist',
    'remove-from-shortlist',
    'player-details',
    'add-to-comparison',
    'remove-from-comparison',
    'why-this-player',
  ],
}
</script>

<style scoped>
/* Context menu styles are minimal since Quasar handles most styling */
</style>
