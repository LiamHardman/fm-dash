<template>
  <div class="pros-cons-container">
    <div class="pros-cons-card">
      <div class="pros-section">
        <div class="section-header">
          <q-icon name="thumb_up" color="positive" class="q-mr-sm" />
          <span class="section-title">Pros</span>
        </div>
        <div class="pros-list">
          <div 
            v-for="(pro, index) in topAttributes" 
            :key="`pro-${index}`"
            class="pro-item"
          >
            <span class="category-name">{{ pro.category }}</span>
            <span class="attribute-value">{{ pro.value }}</span>
          </div>
        </div>
      </div>
      
      <div class="cons-section">
        <div class="section-header">
          <q-icon name="thumb_down" color="negative" class="q-mr-sm" />
          <span class="section-title">Cons</span>
        </div>
        <div class="cons-list">
          <div 
            v-for="(con, index) in bottomAttributes" 
            :key="`con-${index}`"
            class="con-item"
          >
            <span class="category-name">{{ con.category }}</span>
            <span class="attribute-value">{{ con.value }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

const performanceStatMap = {
  'Asts/90': 'Assists per 90',
  'Av Rat': 'Average Rating',
  'Blk/90': 'Blocks per 90',
  'Ch C/90': 'Chances Created per 90',
  'Clr/90': 'Clearances per 90',
  'Cr C/90': 'Crosses Completed per 90',
  'Drb/90': 'Dribbles per 90',
  'xA/90': 'Expected Assists per 90',
  'xG/90': 'Expected Goals per 90',
  'Gls/90': 'Goals per 90',
  'Hdrs W/90': 'Headers Won per 90',
  'Int/90': 'Interceptions per 90',
  'K Ps/90': 'Key Passes per 90',
  'Ps C/90': 'Passes Completed per 90',
  'Shot/90': 'Shots per 90',
  'Tck/90': 'Tackles per 90',
  'Poss Won/90': 'Possession Won per 90',
  'ShT/90': 'Shots on Target per 90',
  'Pres C/90': 'Pressures Completed per 90',
  'Poss Lost/90': 'Possession Lost per 90',
  'Pr passes/90': 'Progressive Passes per 90',
  'Conv %': 'Conversion %',
  'Tck R': 'Tackle Ratio %',
  'Pas %': 'Pass Completion %',
  'Cr C/A': 'Cross Completion %',
  Fls: 'Fouls',
  Apps: 'Appearances',
  'NP-xG/90': 'Non-Penalty xG per 90',
  'Ps A/90': 'Pass Attempts per 90',
  Mins: 'Minutes Played',
  'Clean Sheets': 'Clean Sheets',
  FA: 'Fouls Against',
  'CRS A/90': 'Crosses Attempted per 90',
  'Con/90': 'Goals Conceded per 90',
  'Cln/90': 'Clean Sheets per 90',
  'xGP/90': 'Expected Goals Prevented per 90',
  'Sv %': 'Save Percentage',
}

// Performance stat categories for grouping
const performanceStatCategories = {
  Discipline: {
    stats: ['Fls', 'FA'],
    lowerIsBetter: true, // Lower percentile is better for discipline stats
  },
  'Goal Scoring': {
    stats: ['Gls/90', 'xG/90', 'NP-xG/90', 'Conv %'],
    lowerIsBetter: false,
  },
  'Shot Frequency': {
    stats: ['Shot/90', 'ShT/90'],
    lowerIsBetter: false,
  },
  'Ball Progression': {
    stats: [
      'Ps C/90',
      'Ps A/90',
      'Pas %',
      'Pr passes/90',
      'Asts/90',
      'xA/90',
      'K Ps/90',
      'Ch C/90',
      'Cr C/90',
      'CRS A/90',
      'Cr C/A',
    ],
    lowerIsBetter: false,
  },
  Dribbling: {
    stats: ['Drb/90'],
    lowerIsBetter: false,
  },
  Defending: {
    stats: ['Tck/90', 'Int/90', 'Blk/90', 'Clr/90', 'Tck R', 'Hdrs W/90', 'Pres C/90'],
    lowerIsBetter: false,
  },
  'Shot Stopping': {
    stats: ['Sv %', 'xGP/90'],
    lowerIsBetter: false,
  },
  'GK Team Impact': {
    stats: ['Con/90', 'Clean Sheets', 'Cln/90'],
    lowerIsBetter: true, // Lower is better for goals conceded
  },
}

export default defineComponent({
  name: 'ProsCons',
  props: {
    player: {
      type: Object,
      required: true,
    },
    selectedComparisonGroup: {
      type: String,
      default: 'Global',
    },
  },
  setup(props) {
    // Computed properties for pros/cons analysis based on performance percentiles
    const topAttributes = computed(() => {
      if (!props.player?.performancePercentiles) return []

      const percentiles = props.player.performancePercentiles
      const groupPercentiles = percentiles[props.selectedComparisonGroup] || {}

      // Debug logging
      console.log('ProsCons Debug:', {
        hasPercentiles: !!props.player.performancePercentiles,
        selectedGroup: props.selectedComparisonGroup,
        availableGroups: Object.keys(percentiles),
        groupPercentiles: groupPercentiles,
        groupPercentilesKeys: Object.keys(groupPercentiles),
      })

      // Group stats by category and find the best in each category
      const categorizedPros = {}

      Object.entries(performanceStatCategories).forEach(([categoryName, category]) => {
        const categoryStats = category.stats
          .map((statKey) => {
            const percentile = groupPercentiles[statKey]
            if (
              percentile === null ||
              percentile === undefined ||
              Number.isNaN(percentile) ||
              percentile < 0
            ) {
              return null
            }
            return {
              key: statKey,
              percentile: Number(percentile),
              strength: getStrengthLevel(Number(percentile)),
              category: categoryName,
              lowerIsBetter: category.lowerIsBetter,
            }
          })
          .filter((item) => item !== null)
          .filter((item) => {
            if (item.lowerIsBetter) {
              // For discipline stats, lower percentile is better
              return item.strength === 'Very Weak' || item.strength === 'Weak'
            } else {
              // For regular stats, higher percentile is better
              return item.strength === 'Very Strong' || item.strength === 'Strong'
            }
          })
          .sort((a, b) => {
            if (a.lowerIsBetter) {
              return a.percentile - b.percentile // Lower is better
            } else {
              return b.percentile - a.percentile // Higher is better
            }
          })

        if (categoryStats.length > 0) {
          // Take the best stat from each category
          const bestStat = categoryStats[0]
          categorizedPros[categoryName] = {
            key: bestStat.key,
            value: bestStat.strength,
            percentile: bestStat.percentile,
            category: categoryName,
          }
        }
      })

      return Object.values(categorizedPros)
    })

    const bottomAttributes = computed(() => {
      if (!props.player?.performancePercentiles) return []

      const percentiles = props.player.performancePercentiles
      const groupPercentiles = percentiles[props.selectedComparisonGroup] || {}

      // Group stats by category and find the worst in each category
      const categorizedCons = {}

      Object.entries(performanceStatCategories).forEach(([categoryName, category]) => {
        const categoryStats = category.stats
          .map((statKey) => {
            const percentile = groupPercentiles[statKey]
            if (
              percentile === null ||
              percentile === undefined ||
              Number.isNaN(percentile) ||
              percentile < 0
            ) {
              return null
            }
            return {
              key: statKey,
              percentile: Number(percentile),
              strength: getStrengthLevel(Number(percentile)),
              category: categoryName,
              lowerIsBetter: category.lowerIsBetter,
            }
          })
          .filter((item) => item !== null)
          .filter((item) => {
            if (item.lowerIsBetter) {
              // For discipline stats, higher percentile is worse
              return item.strength === 'Very Strong' || item.strength === 'Strong'
            } else {
              // For regular stats, lower percentile is worse
              return item.strength === 'Weak' || item.strength === 'Very Weak'
            }
          })
          .sort((a, b) => {
            if (a.lowerIsBetter) {
              return b.percentile - a.percentile // Higher is worse
            } else {
              return a.percentile - b.percentile // Lower is worse
            }
          })

        if (categoryStats.length > 0) {
          // Take the worst stat from each category
          const worstStat = categoryStats[0]
          categorizedCons[categoryName] = {
            key: worstStat.key,
            value: worstStat.strength,
            percentile: worstStat.percentile,
            category: categoryName,
          }
        }
      })

      return Object.values(categorizedCons)
    })

    // Helper function to determine strength level based on percentile
    const getStrengthLevel = (percentile) => {
      if (percentile >= 89) return 'Very Strong'
      if (percentile >= 74) return 'Strong'
      if (percentile <= 11) return 'Very Weak'
      if (percentile <= 36) return 'Weak'
      return null // For middle range percentiles (37-73)
    }

    return {
      topAttributes,
      bottomAttributes,
      performanceStatMap,
    }
  },
})
</script>

<style lang="scss" scoped>
// Pros/Cons Component Styles
.pros-cons-container {
  margin-top: 40px; // Much more spacing to prevent intersection
  padding: 0 16px;
  
  .pros-cons-card {
    background: rgba(255, 255, 255, 0.95);
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    display: flex; // Side-by-side layout
    
    .body--dark & {
      background: rgba(30, 41, 59, 0.95);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    }
    
    .pros-section,
    .cons-section {
      padding: 16px;
      flex: 1; // Equal width for both sections
      
      .section-header {
        display: flex;
        align-items: center;
        margin-bottom: 12px;
        
        .section-title {
          font-size: 1rem;
          font-weight: 600;
          color: #334155;
          
          .body--dark & {
            color: rgba(255, 255, 255, 0.9);
          }
        }
      }
      
      .pros-list,
      .cons-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }
      
      .pro-item,
      .con-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 8px 12px;
        border-radius: 6px;
        transition: all 0.2s ease;
        
        &:hover {
          transform: translateX(2px);
        }
      }
      
      .pro-item {
        background: rgba(34, 197, 94, 0.1);
        border-left: 3px solid #22c55e;
        
        .body--dark & {
          background: rgba(34, 197, 94, 0.15);
          border-left-color: #34d399;
        }
        
        .attribute-value {
          color: #059669;
          font-weight: 600;
          
          .body--dark & {
            color: #34d399;
          }
        }
      }
      
      .con-item {
        background: rgba(239, 68, 68, 0.1);
        border-left: 3px solid #ef4444;
        
        .body--dark & {
          background: rgba(239, 68, 68, 0.15);
          border-left-color: #f87171;
        }
        
        .attribute-value {
          color: #dc2626;
          font-weight: 600;
          
          .body--dark & {
            color: #f87171;
          }
        }
      }
      
      .category-name {
        font-size: 0.85rem;
        font-weight: 600;
        color: #334155;
        letter-spacing: 0.5px;
        
        .body--dark & {
          color: rgba(255, 255, 255, 0.85);
        }
      }
      
      .attribute-value {
        font-size: 0.9rem;
        font-weight: 700;
      }
    }
    
    .cons-section {
      border-left: 1px solid rgba(0, 0, 0, 0.1); // Changed from border-top to border-left
      
      .body--dark & {
        border-left-color: rgba(255, 255, 255, 0.1);
      }
    }
  }
}

// Responsive Design
@media (max-width: 768px) {
  .pros-cons-container {
    margin-top: 32px;
    padding: 0 12px;
    
    .pros-cons-card {
      flex-direction: column; // Stack vertically on mobile
      
      .pros-section,
      .cons-section {
        padding: 12px;
        flex: none; // Remove flex on mobile
        
        .section-header {
          margin-bottom: 8px;
          
          .section-title {
            font-size: 0.9rem;
          }
        }
        
        .pro-item,
        .con-item {
          padding: 6px 10px;
        }
        
        .attribute-name {
          font-size: 0.8rem;
        }
        
        .attribute-value {
          font-size: 0.85rem;
        }
      }
      
      .cons-section {
        border-left: none; // Remove left border on mobile
        border-top: 1px solid rgba(0, 0, 0, 0.1); // Add top border instead
        
        .body--dark & {
          border-top-color: rgba(255, 255, 255, 0.1);
        }
      }
    }
  }
}

@media (max-width: 480px) {
  .pros-cons-container {
    margin-top: 28px;
    padding: 0 8px;
    
    .pros-cons-card {
      flex-direction: column; // Stack vertically on small mobile
      
      .pros-section,
      .cons-section {
        padding: 10px;
        flex: none; // Remove flex on small mobile
        
        .section-header {
          margin-bottom: 6px;
          
          .section-title {
            font-size: 0.85rem;
          }
        }
        
        .pro-item,
        .con-item {
          padding: 4px 8px;
        }
        
        .attribute-name {
          font-size: 0.75rem;
        }
        
        .attribute-value {
          font-size: 0.8rem;
        }
      }
      
      .cons-section {
        border-left: none; // Remove left border on small mobile
        border-top: 1px solid rgba(0, 0, 0, 0.1); // Add top border instead
        
        .body--dark & {
          border-top-color: rgba(255, 255, 255, 0.1);
        }
      }
    }
  }
}
</style> 