<template>
  <!-- Context Menu -->
  <q-menu 
    ref="contextMenu"
    touch-position 
    context-menu
    :offset="[10, 10]"
  >
    <q-list dense style="min-width: 180px">
      <q-item 
        clickable 
        v-close-popup 
        @click="handleAddToWishlist"
        v-if="contextMenuPlayer && !isPlayerInWishlist(contextMenuPlayer)"
      >
        <q-item-section avatar>
          <q-icon name="favorite_border" color="positive" />
        </q-item-section>
        <q-item-section>Add to Wishlist</q-item-section>
      </q-item>
      
      <q-item 
        clickable 
        v-close-popup 
        @click="handleRemoveFromWishlist"
        v-if="contextMenuPlayer && isPlayerInWishlist(contextMenuPlayer)"
      >
        <q-item-section avatar>
          <q-icon name="favorite" color="negative" />
        </q-item-section>
        <q-item-section>Remove from Wishlist</q-item-section>
      </q-item>
      
      <q-separator />
      
      <q-item 
        clickable 
        v-close-popup 
        @click="handlePlayerDetails"
        v-if="contextMenuPlayer"
      >
        <q-item-section avatar>
          <q-icon name="info" color="info" />
        </q-item-section>
        <q-item-section>View Details</q-item-section>
      </q-item>
    </q-list>
  </q-menu>
</template>

<script>
import { useQuasar } from 'quasar'
import { ref } from 'vue'
import { useWishlistStore } from '../../stores/wishlistStore'

export default {
  name: 'PlayerTableContextMenu',
  props: {
    datasetId: { type: String, default: null },
    showWishlistActions: { type: Boolean, default: false },
  },
  emits: ['player-selected', 'remove-from-wishlist'],

  setup(props, { emit }) {
    const $q = useQuasar()
    const wishlistStore = useWishlistStore()
    const contextMenu = ref(null)
    const contextMenuPlayer = ref(null)

    const isPlayerInWishlist = (player) => {
      if (!player || !props.datasetId) return false
      return wishlistStore.isInWishlist(props.datasetId, player)
    }

    const handleAddToWishlist = async () => {
      if (contextMenuPlayer.value && props.datasetId) {
        const success = await wishlistStore.addToWishlist(props.datasetId, contextMenuPlayer.value)
        if (success) {
          $q.notify({
            type: 'positive',
            message: `${contextMenuPlayer.value.name} added to wishlist`,
            position: 'top',
            timeout: 2000,
          })
        } else {
          $q.notify({
            type: 'warning',
            message: `${contextMenuPlayer.value.name} is already in wishlist`,
            position: 'top',
            timeout: 2000,
          })
        }
      }
    }

    const handleRemoveFromWishlist = async () => {
      if (contextMenuPlayer.value && props.datasetId) {
        const success = await wishlistStore.removeFromWishlist(
          props.datasetId,
          contextMenuPlayer.value
        )
        if (success) {
          $q.notify({
            type: 'positive',
            message: `${contextMenuPlayer.value.name} removed from wishlist`,
            position: 'top',
            timeout: 2000,
          })
          if (props.showWishlistActions) {
            emit('remove-from-wishlist', contextMenuPlayer.value)
          }
        }
      }
    }

    const handlePlayerDetails = () => {
      if (contextMenuPlayer.value) {
        emit('player-selected', contextMenuPlayer.value)
      }
    }

    const onRightClick = (event, player) => {
      event.preventDefault()
      contextMenuPlayer.value = player
    }

    const setContextMenuPlayer = (player) => {
      contextMenuPlayer.value = player
    }

    return {
      contextMenu,
      contextMenuPlayer,
      isPlayerInWishlist,
      handleAddToWishlist,
      handleRemoveFromWishlist,
      handlePlayerDetails,
      onRightClick,
      setContextMenuPlayer,
    }
  },
}
</script>

<style scoped>
/* Context menu styles are minimal since Quasar handles most styling */
</style> 