# src/components/player-details/PlayerRoleRatings.vue
<template>
  <q-card flat bordered :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-blue-grey-1'">
    <q-card-section>
      <div class="text-subtitle1 q-mb-xs">
        Role-Specific Ratings
      </div>
      
      <div class="row q-col-gutter-xs">
        <div
          v-for="(rating, role) in sortedRoleRatings"
          :key="role"
          class="col-6 col-sm-4 col-md-3 q-pa-xs"
        >
          <q-card
            flat
            bordered
            :class="[
              'role-card',
              qInstance.dark.isActive ? 'bg-grey-8' : 'bg-grey-2',
              isBestRole(role) ? (qInstance.dark.isActive ? 'best-role-dark' : 'best-role') : ''
            ]"
          >
            <q-card-section class="q-pa-xs text-center">
              <div class="role-name">{{ role }}</div>
              <div 
                class="text-h6 q-mb-none role-rating" 
                :class="getRatingColorClass(rating)"
              >
                {{ Math.round(rating * 100) / 100 }}
              </div>
            </q-card-section>
          </q-card>
        </div>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'PlayerRoleRatings',
  props: {
    player: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const qInstance = useQuasar()

    const roleRatings = computed(() => {
      if (!props.player || !props.player.roleSpecificOveralls) {
        return {}
      }

      // Handle both array format and object format for compatibility
      if (Array.isArray(props.player.roleSpecificOveralls)) {
        // Convert from array of {roleName, score} to object format
        return props.player.roleSpecificOveralls.reduce((acc, rso) => {
          acc[rso.roleName] = rso.score
          return acc
        }, {})
      }
      // Already in object format
      return props.player.roleSpecificOveralls
    })

    const sortedRoleRatings = computed(() => {
      const entries = Object.entries(roleRatings.value)
      return Object.fromEntries(entries.sort((a, b) => b[1] - a[1]))
    })

    const isBestRole = role => {
      if (!props.player || !props.player.bestRoleOverall) return false
      return props.player.bestRoleOverall === role
    }

    const getRatingColorClass = rating => {
      if (rating >= 16) return 'text-green-10'
      if (rating >= 14) return 'text-green-8'
      if (rating >= 12) return 'text-amber-8'
      if (rating >= 10) return 'text-orange'
      return 'text-red'
    }

    return {
      qInstance,
      sortedRoleRatings,
      isBestRole,
      getRatingColorClass
    }
  }
})
</script>

<style lang="scss" scoped>
@use "sass:color";

.role-card {
  transition: all 0.2s ease;
  
  &.best-role {
    border-color: $positive !important;
    box-shadow: 0 0 4px 0 rgba($positive, 0.6);
  }
  
  &.best-role-dark {
    border-color: color.adjust($positive, $lightness: 15%) !important;
    box-shadow: 0 0 4px 0 rgba(color.adjust($positive, $lightness: 20%), 0.7);
  }
}

.role-name {
  font-size: 0.8rem;
  min-height: 2.4em;
  display: flex;
  align-items: center;
  justify-content: center;
}

.role-rating {
  font-weight: 600;
}
</style>