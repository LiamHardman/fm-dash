<template>
  <div class="row q-col-gutter-md">
    <!-- Technical/Goalkeeping Attributes -->
    <div class="col-12 col-md-4">
      <q-card flat bordered class="attribute-card modern-attribute-card full-height-card">
        <q-card-section class="attribute-card-header">
          <div class="attribute-section-title">
            <q-icon :name="isGoalkeeper ? 'sports_soccer' : 'precision_manufacturing'" class="q-mr-sm" />
            {{ isGoalkeeper ? 'Goalkeeping' : 'Technical' }}
          </div>
        </q-card-section>
        
        <q-card-section class="q-pa-md">
          <q-list separator dense class="attribute-list">
            <q-item
              v-for="attrKey in (isGoalkeeper ? attributeCategories.goalkeeping : attributeCategories.technical)"
              :key="attrKey"
              class="attribute-list-item modern-attribute-item"
            >
              <q-item-section>
                <q-item-label lines="1" class="attribute-name">
                  {{ attributeFullNameMap[attrKey] || attrKey }}
                </q-item-label>
                <q-tooltip
                  :class="$q.dark.isActive ? 'bg-grey-7 text-white' : 'bg-white text-dark'"
                  :delay="500"
                  max-width="300px"
                  class="modern-tooltip"
                >
                  <div class="tooltip-header">
                    {{ attributeFullNameMap[attrKey] || attrKey }}
                  </div>
                  <div class="tooltip-description">
                    {{ attributeDescriptions[attrKey] || 'No description available' }}
                  </div>
                </q-tooltip>
              </q-item-section>
              <q-item-section side>
                <span
                  :class="[
                    'attribute-value modern-attribute-value',
                    getUnifiedRatingClass(player.attributes[attrKey], 20)
                  ]"
                >
                  {{ getDisplayAttribute(attrKey) }}
                </span>
              </q-item-section>
            </q-item>
            
            <q-item
              v-if="!(isGoalkeeper ? attributeCategories.goalkeeping : attributeCategories.technical).length"
              class="no-attributes-item"
            >
              <q-item-section class="text-center q-py-md">
                <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                No {{ isGoalkeeper ? "goalkeeping" : "technical" }} attributes.
              </q-item-section>
            </q-item>
          </q-list>
        </q-card-section>
      </q-card>
    </div>
    
    <!-- Mental Attributes -->
    <div class="col-12 col-md-4">
      <q-card flat bordered class="attribute-card modern-attribute-card full-height-card">
        <q-card-section class="attribute-card-header">
          <div class="attribute-section-title">
            <q-icon name="psychology" class="q-mr-sm" />
            Mental
          </div>
        </q-card-section>
        
        <q-card-section class="q-pa-md">
          <q-list separator dense class="attribute-list">
            <q-item
              v-for="attrKey in attributeCategories.mental"
              :key="attrKey"
              class="attribute-list-item modern-attribute-item"
            >
              <q-item-section>
                <q-item-label lines="1" class="attribute-name">
                  {{ attributeFullNameMap[attrKey] || attrKey }}
                </q-item-label>
                <q-tooltip
                  :class="$q.dark.isActive ? 'bg-grey-7 text-white' : 'bg-white text-dark'"
                  :delay="500"
                  max-width="300px"
                  class="modern-tooltip"
                >
                  <div class="tooltip-header">
                    {{ attributeFullNameMap[attrKey] || attrKey }}
                  </div>
                  <div class="tooltip-description">
                    {{ attributeDescriptions[attrKey] || 'No description available' }}
                  </div>
                </q-tooltip>
              </q-item-section>
              <q-item-section side>
                <span
                  :class="[
                    'attribute-value modern-attribute-value',
                    getUnifiedRatingClass(player.attributes[attrKey], 20)
                  ]"
                >
                  {{ getDisplayAttribute(attrKey) }}
                </span>
              </q-item-section>
            </q-item>
            
            <q-item v-if="!attributeCategories.mental.length" class="no-attributes-item">
              <q-item-section class="text-center q-py-md">
                <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                No mental attributes.
              </q-item-section>
            </q-item>
          </q-list>
        </q-card-section>
      </q-card>
    </div>
    
    <!-- Physical Attributes & Role Ratings -->
    <div class="col-12 col-md-4">
      <div class="row q-col-gutter-md">
        <!-- Physical Attributes -->
        <div class="col-12">
          <q-card flat bordered class="attribute-card modern-attribute-card">
            <q-card-section class="attribute-card-header">
              <div class="attribute-section-title">
                <q-icon name="fitness_center" class="q-mr-sm" />
                Physical
              </div>
            </q-card-section>
            
            <q-card-section class="q-pa-md">
              <q-list separator dense class="attribute-list">
                <q-item
                  v-for="attrKey in attributeCategories.physical"
                  :key="attrKey"
                  class="attribute-list-item modern-attribute-item"
                >
                  <q-item-section>
                    <q-item-label lines="1" class="attribute-name">
                      {{ attributeFullNameMap[attrKey] || attrKey }}
                    </q-item-label>
                    <q-tooltip
                      :class="$q.dark.isActive ? 'bg-grey-7 text-white' : 'bg-white text-dark'"
                      :delay="500"
                      max-width="300px"
                      class="modern-tooltip"
                    >
                      <div class="tooltip-header">
                        {{ attributeFullNameMap[attrKey] || attrKey }}
                      </div>
                      <div class="tooltip-description">
                        {{ attributeDescriptions[attrKey] || 'No description available' }}
                      </div>
                    </q-tooltip>
                  </q-item-section>
                  <q-item-section side>
                    <span
                      :class="[
                        'attribute-value modern-attribute-value',
                        getUnifiedRatingClass(player.attributes[attrKey], 20)
                      ]"
                    >
                      {{ getDisplayAttribute(attrKey) }}
                    </span>
                  </q-item-section>
                </q-item>
                
                <q-item v-if="!attributeCategories.physical.length" class="no-attributes-item">
                  <q-item-section class="text-center q-py-md">
                    <q-icon name="info_outline" size="sm" class="q-mr-sm" />
                    No physical attributes.
                  </q-item-section>
                </q-item>
              </q-list>
            </q-card-section>
          </q-card>
        </div>
        
        <!-- Best Roles -->
        <div class="col-12" v-if="player.roleSpecificOveralls && player.roleSpecificOveralls.length > 0">
          <q-card flat bordered class="attribute-card modern-attribute-card role-ratings-card">
            <q-card-section class="attribute-card-header">
              <div class="attribute-section-title">
                <q-icon name="star" class="q-mr-sm" />
                Best Roles
              </div>
            </q-card-section>
            
            <q-card-section class="q-pa-md">
              <q-list separator dense class="role-specific-ratings-list">
                <q-item
                  v-for="roleOverall in sortedRoleSpecificOveralls"
                  :key="`role-${roleOverall.roleName}-${roleOverall.score}`"
                  :class="{
                    'best-role-highlight': roleOverall.score === player.Overall,
                  }"
                  class="attribute-list-item modern-attribute-item role-item"
                >
                  <q-item-section>
                    <q-item-label lines="1" class="attribute-name role-name" :title="roleOverall.roleName">
                      {{ roleOverall.roleName }}
                    </q-item-label>
                  </q-item-section>
                  <q-item-section side>
                    <span
                      :class="[
                        'attribute-value modern-attribute-value',
                        getUnifiedRatingClass(roleOverall.score, 100)
                      ]"
                    >
                      {{ roleOverall.score }}
                    </span>
                  </q-item-section>
                </q-item>
              </q-list>
            </q-card-section>
          </q-card>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { defineComponent, computed } from 'vue'

export default defineComponent({
  name: 'PlayerAttributesCard',
  props: {
    player: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    // Check if player is a goalkeeper
    const isGoalkeeper = computed(() => {
      if (!props.player?.position) return false
      const position = props.player.position.toLowerCase()
      return position.includes('gk') || position.includes('goalkeeper')
    })
    
    // Attribute categories
    const attributeCategories = computed(() => {
      const allAttributes = props.player?.attributes || {}
      const attributeKeys = Object.keys(allAttributes)
      
      const technical = attributeKeys.filter(key => 
        ['corners', 'crossing', 'dribbling', 'finishing', 'first_touch', 'free_kick_taking', 
         'heading', 'long_shots', 'long_throws', 'marking', 'passing', 'penalty_taking', 
         'tackling', 'technique'].includes(key)
      )
      
      const goalkeeping = attributeKeys.filter(key =>
        ['aerial_reach', 'command_of_area', 'communication', 'eccentricity', 'handling',
         'kicking', 'one_on_ones', 'reflexes', 'rushing_out', 'tendency_to_punch',
         'throwing'].includes(key)
      )
      
      const mental = attributeKeys.filter(key =>
        ['aggression', 'anticipation', 'bravery', 'composure', 'concentration', 'decisions',
         'determination', 'flair', 'leadership', 'off_the_ball', 'positioning', 'teamwork',
         'vision', 'work_rate'].includes(key)
      )
      
      const physical = attributeKeys.filter(key =>
        ['acceleration', 'agility', 'balance', 'jumping_reach', 'natural_fitness',
         'pace', 'stamina', 'strength'].includes(key)
      )
      
      return {
        technical,
        goalkeeping,
        mental,
        physical
      }
    })
    
    // Attribute full name mapping
    const attributeFullNameMap = {
      corners: 'Corners',
      crossing: 'Crossing',
      dribbling: 'Dribbling',
      finishing: 'Finishing',
      first_touch: 'First Touch',
      free_kick_taking: 'Free Kick Taking',
      heading: 'Heading',
      long_shots: 'Long Shots',
      long_throws: 'Long Throws',
      marking: 'Marking',
      passing: 'Passing',
      penalty_taking: 'Penalty Taking',
      tackling: 'Tackling',
      technique: 'Technique',
      aerial_reach: 'Aerial Reach',
      command_of_area: 'Command of Area',
      communication: 'Communication',
      eccentricity: 'Eccentricity',
      handling: 'Handling',
      kicking: 'Kicking',
      one_on_ones: 'One on Ones',
      reflexes: 'Reflexes',
      rushing_out: 'Rushing Out',
      tendency_to_punch: 'Tendency to Punch',
      throwing: 'Throwing',
      aggression: 'Aggression',
      anticipation: 'Anticipation',
      bravery: 'Bravery',
      composure: 'Composure',
      concentration: 'Concentration',
      decisions: 'Decisions',
      determination: 'Determination',
      flair: 'Flair',
      leadership: 'Leadership',
      off_the_ball: 'Off the Ball',
      positioning: 'Positioning',
      teamwork: 'Teamwork',
      vision: 'Vision',
      work_rate: 'Work Rate',
      acceleration: 'Acceleration',
      agility: 'Agility',
      balance: 'Balance',
      jumping_reach: 'Jumping Reach',
      natural_fitness: 'Natural Fitness',
      pace: 'Pace',
      stamina: 'Stamina',
      strength: 'Strength'
    }
    
    // Attribute descriptions
    const attributeDescriptions = {
      corners: 'Ability to deliver accurate corner kicks',
      crossing: 'Ability to deliver accurate crosses from wide positions',
      dribbling: 'Ability to run with the ball and beat opponents',
      finishing: 'Ability to score goals when presented with opportunities',
      first_touch: 'Ability to control the ball with the first touch',
      free_kick_taking: 'Ability to score from direct free kicks',
      heading: 'Ability to direct the ball accurately with the head',
      long_shots: 'Ability to score from distance',
      long_throws: 'Ability to throw the ball long distances accurately',
      marking: 'Ability to track and mark opposing players',
      passing: 'Accuracy and vision when passing the ball',
      penalty_taking: 'Ability to score from penalty kicks',
      tackling: 'Ability to win the ball through tackles',
      technique: 'Technical skill and ball control',
      // Add more descriptions as needed...
    }
    
    // Sorted role specific overalls
    const sortedRoleSpecificOveralls = computed(() => {
      if (!props.player?.roleSpecificOveralls) return []
      
      return [...props.player.roleSpecificOveralls]
        .sort((a, b) => b.score - a.score)
        .slice(0, 10) // Show top 10 roles
    })
    
    // Get display attribute value
    const getDisplayAttribute = (attrKey) => {
      const value = props.player?.attributes?.[attrKey]
      if (value === undefined || value === null) return '-'
      return value
    }
    
    // Get unified rating class
    const getUnifiedRatingClass = (value, maxValue = 20) => {
      if (!value || value === '-') return 'rating-unknown'
      
      const percentage = (value / maxValue) * 100
      
      if (percentage >= 85) return 'rating-excellent'
      if (percentage >= 70) return 'rating-very-good'
      if (percentage >= 55) return 'rating-good'
      if (percentage >= 40) return 'rating-average'
      if (percentage >= 25) return 'rating-poor'
      return 'rating-very-poor'
    }
    
    return {
      isGoalkeeper,
      attributeCategories,
      attributeFullNameMap,
      attributeDescriptions,
      sortedRoleSpecificOveralls,
      getDisplayAttribute,
      getUnifiedRatingClass
    }
  }
})
</script>

<style lang="scss" scoped>
.attribute-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s ease;
  
  &:hover {
    transform: translateY(-2px);
  }
  
  .body--dark & {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
}

.full-height-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  
  .q-card__section:last-child {
    flex: 1;
  }
}

.attribute-card-header {
  background: linear-gradient(135deg, #1976d2 0%, #1565c0 100%);
  color: white;
  padding: 1rem;
  
  .body--dark & {
    background: linear-gradient(135deg, #424242 0%, #303030 100%);
  }
}

.attribute-section-title {
  font-weight: 600;
  font-size: 1.1rem;
  display: flex;
  align-items: center;
}

.attribute-list-item {
  padding: 0.5rem 0;
  transition: background-color 0.2s ease;
  
  &:hover {
    background: rgba(25, 118, 210, 0.05);
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.03);
    }
  }
}

.attribute-name {
  font-weight: 500;
  font-size: 0.875rem;
}

.attribute-value {
  font-weight: 700;
  font-size: 1rem;
  padding: 4px 8px;
  border-radius: 6px;
  min-width: 40px;
  text-align: center;
  
  &.rating-excellent {
    background: #e8f5e8;
    color: #2e7d32;
    
    .body--dark & {
      background: rgba(76, 175, 80, 0.2);
      color: #81c784;
    }
  }
  
  &.rating-very-good {
    background: #f1f8e9;
    color: #388e3c;
    
    .body--dark & {
      background: rgba(139, 195, 74, 0.2);
      color: #aed581;
    }
  }
  
  &.rating-good {
    background: #fff3e0;
    color: #f57c00;
    
    .body--dark & {
      background: rgba(255, 152, 0, 0.2);
      color: #ffb74d;
    }
  }
  
  &.rating-average {
    background: #fafafa;
    color: #616161;
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.1);
      color: #bdbdbd;
    }
  }
  
  &.rating-poor {
    background: #ffebee;
    color: #d32f2f;
    
    .body--dark & {
      background: rgba(244, 67, 54, 0.2);
      color: #ef5350;
    }
  }
  
  &.rating-very-poor {
    background: #ffebee;
    color: #b71c1c;
    
    .body--dark & {
      background: rgba(244, 67, 54, 0.3);
      color: #f44336;
    }
  }
  
  &.rating-unknown {
    background: #f5f5f5;
    color: #9e9e9e;
    
    .body--dark & {
      background: rgba(255, 255, 255, 0.05);
      color: #757575;
    }
  }
}

.role-ratings-card {
  .attribute-card-header {
    background: linear-gradient(135deg, #7b1fa2 0%, #6a1b9a 100%);
    
    .body--dark & {
      background: linear-gradient(135deg, #512da8 0%, #4527a0 100%);
    }
  }
}

.role-item {
  &.best-role-highlight {
    background: rgba(123, 31, 162, 0.1);
    border-left: 4px solid #7b1fa2;
    
    .body--dark & {
      background: rgba(123, 31, 162, 0.2);
      border-left-color: #ab47bc;
    }
  }
}

.role-name {
  font-weight: 600;
}

.no-attributes-item {
  padding: 1rem;
  color: rgba(0, 0, 0, 0.6);
  
  .body--dark & {
    color: rgba(255, 255, 255, 0.6);
  }
}

.modern-tooltip {
  .tooltip-header {
    font-weight: 600;
    margin-bottom: 0.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
    padding-bottom: 0.25rem;
  }
  
  .tooltip-description {
    font-size: 0.875rem;
    line-height: 1.4;
  }
}
</style>
</code_block_to_apply_changes_from>
</rewritten_file>