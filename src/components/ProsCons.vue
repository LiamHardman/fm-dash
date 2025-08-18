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
            class="attribute-item pro-item"
          >
            <span class="category-name">{{ pro.category }}</span>
            <div class="percentile-display">
              <div class="percentile-bar-container">
                <div class="percentile-bar pro-bar" :style="{ width: pro.percentile + '%' }"></div>
              </div>
              <span class="percentile-value">{{ Math.round(pro.percentile) }}</span>
            </div>
            <span class="attribute-value pro-value">{{ pro.value }}</span>
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
            class="attribute-item con-item"
          >
            <span class="category-name">{{ con.category }}</span>
            <div class="percentile-display">
              <div class="percentile-bar-container">
                <div class="percentile-bar con-bar" :style="{ width: con.percentile + '%' }"></div>
              </div>
              <span class="percentile-value">{{ Math.round(con.percentile) }}</span>
            </div>
            <span class="attribute-value con-value">{{ con.value }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, defineComponent } from 'vue'

// Data maps
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

// --- Start of Changed Section ---

const performanceStatCategories = {
  Fouls: { stats: ['Fls'], lowerIsBetter: true }, // Renamed and isolated Fls
  'Drawing Fouls': { stats: ['FA'], lowerIsBetter: false }, // New category for FA
  'Goal Scoring': { stats: ['Gls/90', 'xG/90', 'NP-xG/90', 'Conv %'], lowerIsBetter: false },
  'Shot Frequency': { stats: ['Shot/90', 'ShT/90'], lowerIsBetter: false },
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
  Dribbling: { stats: ['Drb/90'], lowerIsBetter: false },
  Defending: {
    stats: ['Tck/90', 'Int/90', 'Blk/90', 'Clr/90', 'Tck R', 'Hdrs W/90', 'Pres C/90'],
    lowerIsBetter: false,
  },
  'Shot Stopping': { stats: ['Sv %', 'xGP/90'], lowerIsBetter: false },
  'GK Team Impact': { stats: ['Con/90', 'Clean Sheets', 'Cln/90'], lowerIsBetter: true },
}

export default defineComponent({
  name: 'ProsCons',
  props: {
    player: { type: Object, required: true },
    selectedComparisonGroup: { type: String, default: 'Global' },
  },
  setup(props) {
    // Updated function to return text that reflects the stat's nature
    const getStrengthText = (percentile, lowerIsBetter = false) => {
      if (lowerIsBetter) {
        if (percentile >= 89) return 'Very High' // A con
        if (percentile >= 74) return 'High' // A con
        if (percentile <= 11) return 'Very Low' // A pro
        if (percentile <= 36) return 'Low' // A pro
      } else {
        if (percentile >= 89) return 'Very Strong' // A pro
        if (percentile >= 74) return 'Strong' // A pro
        if (percentile <= 11) return 'Very Weak' // A con
        if (percentile <= 36) return 'Weak' // A con
      }
      return null
    }

    const attributes = computed(() => {
      if (!props.player?.performancePercentiles) return []
      const groupPercentiles =
        props.player.performancePercentiles[props.selectedComparisonGroup] || {}
      const allAttributes = []

      Object.entries(performanceStatCategories).forEach(([categoryName, category]) => {
        const bestStat = category.stats
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
            // Pass lowerIsBetter flag to get correct text
            const strength = getStrengthText(Number(percentile), category.lowerIsBetter)
            if (!strength) return null

            return {
              key: statKey,
              percentile: Number(percentile),
              value: strength,
              category: categoryName,
              lowerIsBetter: category.lowerIsBetter,
            }
          })
          .filter(Boolean)
          .sort((a, b) => {
            const effectivePercentileA = a.lowerIsBetter ? 100 - a.percentile : a.percentile
            const effectivePercentileB = b.lowerIsBetter ? 100 - b.percentile : b.percentile
            return effectivePercentileB - effectivePercentileA
          })[0]

        if (bestStat) {
          allAttributes.push(bestStat)
        }
      })
      return allAttributes
    })

    const topAttributes = computed(() => {
      const pros = ['Very Strong', 'Strong', 'Very Low', 'Low']
      return attributes.value
        .filter((attr) => pros.includes(attr.value))
        .sort((a, b) => {
          const effectivePercentileA = a.lowerIsBetter ? 100 - a.percentile : a.percentile
          const effectivePercentileB = b.lowerIsBetter ? 100 - b.percentile : b.percentile
          return effectivePercentileB - effectivePercentileA // Sort descending by "goodness"
        })
    })

    const bottomAttributes = computed(() => {
      const cons = ['Very Weak', 'Weak', 'Very High', 'High']
      return attributes.value
        .filter((attr) => cons.includes(attr.value))
        .sort((a, b) => {
          const effectivePercentileA = a.lowerIsBetter ? 100 - a.percentile : a.percentile
          const effectivePercentileB = b.lowerIsBetter ? 100 - b.percentile : b.percentile
          return effectivePercentileA - effectivePercentileB // Sort ascending by "goodness" (worst first)
        })
    })

    // --- End of Changed Section ---

    return {
      topAttributes,
      bottomAttributes,
    }
  },
})
</script>


<style lang="scss" scoped>
.pros-cons-container {
  margin-top: 40px;
  padding: 0 16px;

  .pros-cons-card {
    background: rgba(255, 255, 255, 0.95);
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    display: flex;

    .body--dark & {
      background: rgba(30, 41, 59, 0.95);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
    }
  }

  .pros-section,
  .cons-section {
    padding: 16px;
    flex: 1;
    min-width: 0;

    .section-header {
      display: flex;
      align-items: center;
      margin-bottom: 12px;

      .section-title {
        font-size: 1rem;
        font-weight: 600;
        .body--dark & { color: rgba(255, 255, 255, 0.9); }
      }
    }

    .pros-list,
    .cons-list {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
  }

  .attribute-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .pro-item,
  .con-item {
    background: none;
    padding: 0;
  }

  .category-name {
    font-size: 0.85rem;
    font-weight: 600;
    .body--dark & { color: rgba(255, 255, 255, 0.85); }
  }

  .attribute-value {
    font-size: 0.8rem;
    font-weight: 700;
    align-self: flex-end;
  }

  .pro-value {
    color: #059669;
    .body--dark & { color: #34d399; }
  }

  .con-value {
    color: #dc2626;
    .body--dark & { color: #f87171; }
  }

  .percentile-display {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .percentile-bar-container {
    flex-grow: 1;
    height: 6px;
    background-color: #e2e8f0;
    border-radius: 3px;
    overflow: hidden;
    .body--dark & { background-color: #475569; }
  }

  .percentile-bar {
    height: 100%;
    border-radius: 3px;
    transition: width 0.5s ease-out;
  }

  .pro-bar { background-color: #22c55e; }
  .con-bar { background-color: #ef4444; }

  .percentile-value {
    font-size: 0.8rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    min-width: 20px;
    text-align: right;
    .body--dark & { color: #cbd5e1; }
  }

  .cons-section {
    border-left: 1px solid rgba(0, 0, 0, 0.1);
    .body--dark & { border-left-color: rgba(255, 255, 255, 0.1); }
  }
}

// Responsive Design
@media (max-width: 768px) {
  .pros-cons-container .pros-cons-card {
    flex-direction: column;

    .cons-section {
      border-left: none;
      border-top: 1px solid rgba(0, 0, 0, 0.1);
      .body--dark & { border-top-color: rgba(255, 255, 255, 0.1); }
    }
  }
}
</style>