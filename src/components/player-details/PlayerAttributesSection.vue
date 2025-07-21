# src/components/player-details/PlayerAttributesSection.vue
<template>
  <q-card flat bordered :class="qInstance.dark.isActive ? 'bg-grey-9' : 'bg-blue-grey-1'">
    <q-card-section class="attributes-section">
      <div class="text-subtitle1 q-mb-xs attributes-section-title">
        Attributes
      </div>
      
      <!-- Technical Attributes -->
      <div class="attribute-group q-mb-md">
        <div class="text-caption text-bold q-mb-xs attribute-group-title">
          Technical
        </div>
        <div class="row attribute-list">
          <template v-for="(value, attr) in technicalAttributes" :key="attr">
            <div class="col-6 col-sm-4 col-md-3 attribute-item">
              <div class="row no-wrap items-baseline q-gutter-x-xs">
                <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                <div class="col-auto text-weight-bold attribute-value" :class="getColorClass(value)">
                  {{ value }}
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
      
      <!-- Mental Attributes -->
      <div class="attribute-group q-mb-md">
        <div class="text-caption text-bold q-mb-xs attribute-group-title">
          Mental
        </div>
        <div class="row attribute-list">
          <template v-for="(value, attr) in mentalAttributes" :key="attr">
            <div class="col-6 col-sm-4 col-md-3 attribute-item">
              <div class="row no-wrap items-baseline q-gutter-x-xs">
                <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                <div class="col-auto text-weight-bold attribute-value" :class="getColorClass(value)">
                  {{ value }}
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
      
      <!-- Physical Attributes -->
      <div class="attribute-group q-mb-md">
        <div class="text-caption text-bold q-mb-xs attribute-group-title">
          Physical
        </div>
        <div class="row attribute-list">
          <template v-for="(value, attr) in physicalAttributes" :key="attr">
            <div class="col-6 col-sm-4 col-md-3 attribute-item">
              <div class="row no-wrap items-baseline q-gutter-x-xs">
                <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                <div class="col-auto text-weight-bold attribute-value" :class="getColorClass(value)">
                  {{ value }}
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
      
      <!-- Goalkeeping Attributes (conditional) -->
      <div v-if="hasGoalkeeperAttributes" class="attribute-group">
        <div class="text-caption text-bold q-mb-xs attribute-group-title">
          Goalkeeping
        </div>
        <div class="row attribute-list">
          <template v-for="(value, attr) in goalkeeperAttributes" :key="attr">
            <div class="col-6 col-sm-4 col-md-3 attribute-item">
              <div class="row no-wrap items-baseline q-gutter-x-xs">
                <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                <div class="col-auto text-weight-bold attribute-value" :class="getColorClass(value)">
                  {{ value }}
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import { useQuasar } from 'quasar'
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'PlayerAttributesSection',
  props: {
    player: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const qInstance = useQuasar()

    // Grouped attributes by category
    const technicalAttributes = computed(() => {
      const attributes = {}
      const technicalKeys = [
        'crossing',
        'dribbling',
        'finishing',
        'first_touch',
        'free_kick_taking',
        'heading',
        'long_shots',
        'long_throws',
        'marking',
        'passing',
        'penalty_taking',
        'tackling',
        'technique',
        'corners'
      ]

      for (const key of technicalKeys) {
        if (props.player[key] !== undefined) {
          attributes[key] = props.player[key]
        }
      }

      return attributes
    })

    const mentalAttributes = computed(() => {
      const attributes = {}
      const mentalKeys = [
        'aggression',
        'anticipation',
        'bravery',
        'composure',
        'concentration',
        'decisions',
        'determination',
        'flair',
        'leadership',
        'off_the_ball',
        'positioning',
        'teamwork',
        'vision',
        'work_rate'
      ]

      for (const key of mentalKeys) {
        if (props.player[key] !== undefined) {
          attributes[key] = props.player[key]
        }
      }

      return attributes
    })

    const physicalAttributes = computed(() => {
      const attributes = {}
      const physicalKeys = [
        'acceleration',
        'agility',
        'balance',
        'jumping_reach',
        'natural_fitness',
        'pace',
        'stamina',
        'strength'
      ]

      for (const key of physicalKeys) {
        if (props.player[key] !== undefined) {
          attributes[key] = props.player[key]
        }
      }

      return attributes
    })

    const goalkeeperAttributes = computed(() => {
      const attributes = {}
      const gkKeys = [
        'aerial_reach',
        'command_of_area',
        'communication',
        'eccentricity',
        'handling',
        'kicking',
        'one_on_ones',
        'punching',
        'reflexes',
        'rushing_out',
        'tendency_to_punch',
        'throwing'
      ]

      for (const key of gkKeys) {
        if (props.player[key] !== undefined && props.player[key] !== null) {
          attributes[key] = props.player[key]
        }
      }

      return attributes
    })

    const isGoalkeeper = computed(() => {
      if (!props.player) return false
      return (
        props.player.short_positions?.includes('GK') ||
        props.player.position_groups?.includes('Goalkeepers') ||
                  props.player.parsed_positions?.includes('Goalkeeper')
      )
    })

    const hasGoalkeeperAttributes = computed(() => {
      // First check if player is a goalkeeper based on position data
      if (isGoalkeeper.value) return true
      // Then fall back to checking if we have any goalkeeper attributes
      return Object.keys(goalkeeperAttributes.value).length > 0
    })

    const formatAttrName = attr => {
      return attr
        .replace(/_/g, ' ')
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ')
    }

    const getColorClass = value => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue)) return ''

      if (numValue >= 15) return 'text-green-10'
      if (numValue >= 13) return 'text-green-8'
      if (numValue >= 10) return 'text-amber-8'
      if (numValue >= 7) return 'text-orange'
      return 'text-red'
    }

    return {
      qInstance,
      technicalAttributes,
      mentalAttributes,
      physicalAttributes,
      goalkeeperAttributes,
      hasGoalkeeperAttributes,
      formatAttrName,
      getColorClass
    }
  }
})
</script>

<style lang="scss" scoped>
@use "sass:color";

.attributes-section {
  padding: 12px;
}

.attributes-section-title {
  font-weight: 600;
}

.attribute-group-title {
  color: $primary;
  border-bottom: 1px solid $primary;
  padding-bottom: 2px;
  margin-bottom: 8px;
  
  .body--dark & {
    color: color.adjust($primary, $lightness: 20%);
    border-bottom-color: color.adjust($primary, $lightness: 20%);
  }
}

.attribute-item {
  margin-bottom: 6px;
  padding-right: 8px;
}

.attribute-name {
  font-size: 0.85rem;
  color: rgba(0, 0, 0, 0.7);
  
  .body--dark & {
    color: rgba(255, 255, 255, 0.7);
  }
}

.attribute-value {
  font-size: 0.9rem;
}
</style>