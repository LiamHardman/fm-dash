<template>
  <q-dialog
    :model-value="show"
    @hide="$emit('close')"
    persistent
    maximized
    transition-show="slide-up"
    transition-hide="slide-down"
  >
    <q-card class="comparison-dialog">
      <!-- Header -->
      <q-card-section class="row items-center q-pb-none card-header">
        <q-icon name="compare_arrows" size="md" class="q-mr-sm" />
        <div class="text-h6">Player Comparison</div>
        <q-space />
        <q-btn
          icon="close"
          flat
          round
          dense
          v-close-popup
          @click="$emit('close')"
        />
      </q-card-section>

      <!-- Empty state -->
      <q-card-section v-if="players.length < 2" class="text-center q-pa-xl">
        <q-icon name="compare_arrows" size="5rem" color="grey-5" class="q-mb-md" />
        <div class="text-h6 text-grey-6">Add at least 2 players to compare</div>
        <div class="text-caption text-grey-5 q-mt-sm">
          Right-click any player in the table and select "Add to Comparison"
        </div>
      </q-card-section>

      <template v-else>
        <!-- Player column headers (sticky) -->
        <q-card-section class="player-columns-header q-pt-sm q-pb-sm">
          <div class="col-row">
            <div class="label-col" />
            <div
              v-for="(player, idx) in players"
              :key="player.name + player.club"
              class="player-col"
            >
              <div class="player-header-card" :style="{ borderTopColor: PLAYER_COLORS[idx] }">
                <div class="player-header-name" :style="{ color: PLAYER_COLORS[idx] }">
                  {{ player.name }}
                </div>
                <div class="text-caption text-grey-6">{{ player.club }}</div>
                <div class="text-caption text-grey-6">{{ player.division }}</div>
                <q-btn
                  icon="remove_circle_outline"
                  flat dense round size="xs"
                  class="q-mt-xs"
                  color="grey-6"
                  @click="$emit('remove-player', player)"
                >
                  <q-tooltip>Remove</q-tooltip>
                </q-btn>
              </div>
            </div>
          </div>
        </q-card-section>

        <q-separator />

        <!-- Scrollable content -->
        <q-card-section class="comp-scroll-body scroll">

          <!-- Radar Chart -->
          <div class="radar-section q-mb-lg">
            <div class="section-heading">
              <q-icon name="radar" size="1rem" class="q-mr-xs" color="primary" />
              Attribute Profile
            </div>
            <div class="radar-wrapper">
              <Radar :data="radarData" :options="radarOptions" style="max-height: 340px;" />
            </div>
          </div>

          <!-- Data Sections -->
          <template v-for="section in sections" :key="section.title">
            <div class="data-section q-mb-md">
              <div class="section-heading q-mb-xs">
                <q-icon :name="section.icon" size="1rem" class="q-mr-xs" color="primary" />
                {{ section.title }}
              </div>
              <div class="section-body">
                <div
                  v-for="row in section.rows"
                  :key="row.label"
                  class="attr-row"
                >
                  <div class="label-col attr-label">{{ row.label }}</div>
                  <div
                    v-for="(_, idx) in players"
                    :key="idx"
                    class="player-col attr-value"
                    :style="cellStyle(row, idx)"
                  >
                    {{ row.displayValues ? row.displayValues[idx] : (row.values[idx] == null ? '—' : row.values[idx]) }}
                  </div>
                </div>
              </div>
            </div>
          </template>

        </q-card-section>
      </template>
    </q-card>
  </q-dialog>
</template>

<script setup>
import {
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  PointElement,
  RadialLinearScale,
  Tooltip,
} from 'chart.js'
import { useQuasar } from 'quasar'
import { computed } from 'vue'
// biome-ignore lint/correctness/noUnusedImports: used in template
import { Radar } from 'vue-chartjs'

ChartJS.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend)

const props = defineProps({
  show: { type: Boolean, default: false },
  players: { type: Array, default: () => [] },
})

defineEmits(['close', 'remove-player'])

const $q = useQuasar()

const PLAYER_COLORS = ['#3b82f6', '#ef4444', '#22c55e', '#f97316']
const PLAYER_BG = [
  'rgba(59,130,246,0.12)',
  'rgba(239,68,68,0.12)',
  'rgba(34,197,94,0.12)',
  'rgba(249,115,22,0.12)',
]

const TECH_ATTRS = [
  'Cor',
  'Cro',
  'Dri',
  'Fin',
  'Fir',
  'Fre',
  'Hea',
  'Lon',
  'Mar',
  'Pas',
  'Pen',
  'Tck',
  'Tec',
]
const MENTAL_ATTRS = [
  'Agg',
  'Ant',
  'Bra',
  'Cmp',
  'Cnt',
  'Dec',
  'Det',
  'Fla',
  'Ldr',
  'OtB',
  'Pos',
  'Tea',
  'Vis',
  'Wor',
]
const PHYSICAL_ATTRS = ['Acc', 'Agi', 'Bal', 'Jum', 'Nat', 'Pac', 'Sta', 'Str']

const allGK = computed(
  () => props.players.length > 0 && props.players.every((p) => p.shortPositions?.includes('GK'))
)

const numAttr = (player, key) => {
  const val = player.numericAttributes?.[key]
  return typeof val === 'number' ? val : null
}

const roleScore = (player, roleName) => {
  const entry = player.roleSpecificOveralls?.find(
    (r) => r.role === roleName || r.roleName === roleName
  )
  return entry ? (entry.overall ?? entry.score ?? null) : null
}

const fifaOverall = (p) => {
  if (allGK.value) {
    return Math.round(
      ((p.div || 0) + (p.han || 0) + (p.ref || 0) + (p.kic || 0) + (p.spd || 0) + (p.pos || 0)) / 6
    )
  }
  return Math.round(
    ((p.pac || 0) + (p.sho || 0) + (p.pas || 0) + (p.dri || 0) + (p.def || 0) + (p.phy || 0)) / 6
  )
}

const topRoles = computed(() => {
  const roleSet = new Set()
  for (const p of props.players) {
    if (p.roleSpecificOveralls) {
      const sorted = [...p.roleSpecificOveralls]
        .sort((a, b) => (b.overall ?? b.score ?? 0) - (a.overall ?? a.score ?? 0))
        .slice(0, 5)
      for (const r of sorted) roleSet.add(r.role ?? r.roleName)
    }
  }
  return [...roleSet].filter(Boolean).slice(0, 8)
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const sections = computed(() => {
  const ps = props.players
  const result = [
    {
      title: 'Overview',
      icon: 'person',
      rows: [
        { label: 'Age', values: ps.map((p) => p.age), higherBetter: false },
        { label: 'Position', values: ps.map((p) => p.position ?? '—'), higherBetter: null },
        { label: 'Nationality', values: ps.map((p) => p.nationality ?? '—'), higherBetter: null },
        { label: 'Overall (FIFA)', values: ps.map(fifaOverall), higherBetter: true },
        { label: 'Personality', values: ps.map((p) => p.personality ?? '—'), higherBetter: null },
      ],
    },
    {
      title: allGK.value ? 'Goalkeeper Stats' : 'FIFA Stats',
      icon: 'sports_soccer',
      rows: allGK.value
        ? [
            { label: 'DIV', values: ps.map((p) => p.div || 0), higherBetter: true },
            { label: 'HAN', values: ps.map((p) => p.han || 0), higherBetter: true },
            { label: 'REF', values: ps.map((p) => p.ref || 0), higherBetter: true },
            { label: 'KIC', values: ps.map((p) => p.kic || 0), higherBetter: true },
            { label: 'SPD', values: ps.map((p) => p.spd || 0), higherBetter: true },
            { label: 'POS', values: ps.map((p) => p.pos || 0), higherBetter: true },
          ]
        : [
            { label: 'PAC', values: ps.map((p) => p.pac || 0), higherBetter: true },
            { label: 'SHO', values: ps.map((p) => p.sho || 0), higherBetter: true },
            { label: 'PAS', values: ps.map((p) => p.pas || 0), higherBetter: true },
            { label: 'DRI', values: ps.map((p) => p.dri || 0), higherBetter: true },
            { label: 'DEF', values: ps.map((p) => p.def || 0), higherBetter: true },
            { label: 'PHY', values: ps.map((p) => p.phy || 0), higherBetter: true },
          ],
    },
    {
      title: 'Technical',
      icon: 'build',
      rows: TECH_ATTRS.map((attr) => ({
        label: attr,
        values: ps.map((p) => numAttr(p, attr)),
        higherBetter: true,
      })),
    },
    {
      title: 'Mental',
      icon: 'psychology',
      rows: MENTAL_ATTRS.map((attr) => ({
        label: attr,
        values: ps.map((p) => numAttr(p, attr)),
        higherBetter: true,
      })),
    },
    {
      title: 'Physical',
      icon: 'fitness_center',
      rows: PHYSICAL_ATTRS.map((attr) => ({
        label: attr,
        values: ps.map((p) => numAttr(p, attr)),
        higherBetter: true,
      })),
    },
    {
      title: 'Financials',
      icon: 'attach_money',
      rows: [
        {
          label: 'Transfer Value',
          values: ps.map((p) => p.transferValueAmount || 0),
          displayValues: ps.map((p) => p.transfer_value || '—'),
          higherBetter: false,
        },
        {
          label: 'Wage',
          values: ps.map((p) => p.wageAmount || 0),
          displayValues: ps.map((p) => p.wage || '—'),
          higherBetter: false,
        },
      ],
    },
  ]

  if (topRoles.value.length > 0) {
    result.splice(5, 0, {
      title: 'Role Suitabilities',
      icon: 'assignment_ind',
      rows: topRoles.value.map((role) => ({
        label: role,
        values: ps.map((p) => roleScore(p, role)),
        higherBetter: true,
      })),
    })
  }

  return result
})

// biome-ignore lint/correctness/noUnusedVariables: used in template
const cellStyle = (row, playerIdx) => {
  if (row.higherBetter == null) return {}
  const numericVals = row.values.filter((v) => v != null && typeof v === 'number')
  if (numericVals.length < 2) return {}
  const best = row.higherBetter ? Math.max(...numericVals) : Math.min(...numericVals)
  const worst = row.higherBetter ? Math.min(...numericVals) : Math.max(...numericVals)
  const val = row.values[playerIdx]
  if (val == null || typeof val !== 'number') return {}
  if (val === best && best !== worst)
    return { background: 'rgba(34,197,94,0.15)', fontWeight: '700', color: '#16a34a' }
  if (val === worst && best !== worst)
    return { background: 'rgba(239,68,68,0.10)', color: '#dc2626' }
  return {}
}

const radarLabels = computed(() =>
  allGK.value
    ? ['DIV', 'HAN', 'REF', 'KIC', 'SPD', 'POS']
    : ['PAC', 'SHO', 'PAS', 'DRI', 'DEF', 'PHY']
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const radarData = computed(() => ({
  labels: radarLabels.value,
  datasets: props.players.map((p, idx) => {
    const values = allGK.value
      ? [p.div || 0, p.han || 0, p.ref || 0, p.kic || 0, p.spd || 0, p.pos || 0]
      : [p.pac || 0, p.sho || 0, p.pas || 0, p.dri || 0, p.def || 0, p.phy || 0]
    return {
      label: p.name,
      data: values,
      borderColor: PLAYER_COLORS[idx],
      backgroundColor: PLAYER_BG[idx],
      borderWidth: 2,
      pointBackgroundColor: PLAYER_COLORS[idx],
      pointRadius: 4,
      pointHoverRadius: 6,
    }
  }),
}))

// biome-ignore lint/correctness/noUnusedVariables: used in template
const radarOptions = computed(() => {
  const dark = $q.dark.isActive
  const gridColor = dark ? 'rgba(255,255,255,0.12)' : 'rgba(0,0,0,0.1)'
  const tickColor = dark ? '#94a3b8' : '#64748b'
  const labelColor = dark ? '#e2e8f0' : '#1e293b'
  return {
    responsive: true,
    maintainAspectRatio: true,
    scales: {
      r: {
        min: 0,
        max: 99,
        ticks: {
          stepSize: 20,
          color: tickColor,
          backdropColor: 'transparent',
          font: { size: 10 },
        },
        grid: { color: gridColor },
        angleLines: { color: gridColor },
        pointLabels: { color: labelColor, font: { size: 13, weight: 'bold' } },
      },
    },
    plugins: {
      legend: {
        labels: { color: labelColor, font: { size: 12 }, boxWidth: 14, padding: 16 },
      },
      tooltip: { mode: 'point' },
    },
  }
})
</script>

<style lang="scss" scoped>
.comparison-dialog {
  border-radius: $border-radius;
  box-shadow: $card-shadow;

  .body--dark & {
    background-color: #1e293b !important;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  // Gradient header matching other dialogs
  .card-header {
    background: linear-gradient(135deg, #2e74b5 0%, #3b82c7 100%);
    color: white;
    padding: 1.5rem;
    border-radius: $border-radius $border-radius 0 0;

    .q-icon { color: rgba(255, 255, 255, 0.9); }
    .text-h6 { font-weight: 600; font-size: 1.25rem; }
    .q-btn {
      color: rgba(255, 255, 255, 0.8);
      &:hover { background-color: rgba(255, 255, 255, 0.1); color: white; }
    }
  }

  // Player column header row
  .player-columns-header {
    background: transparent;

    .body--dark & { background: transparent; }
  }

  .col-row {
    display: flex;
    gap: 10px;
    align-items: stretch;
  }

  .label-col {
    width: 130px;
    min-width: 130px;
    flex-shrink: 0;
  }

  .player-col {
    flex: 1;
    min-width: 0;
  }

  .player-header-card {
    border-top: 3px solid transparent;
    border-radius: 6px;
    padding: 10px 12px;
    text-align: center;
    background: rgba(0, 0, 0, 0.03);

    .body--dark & { background: rgba(255, 255, 255, 0.04); }
  }

  .player-header-name {
    font-size: 14px;
    font-weight: 700;
    line-height: 1.3;
    word-break: break-word;
  }

  // Scrollable body
  .comp-scroll-body {
    height: calc(100vh - 200px);
    overflow-y: auto;
    padding: 16px 20px;
  }

  // Radar
  .radar-section { }
  .radar-wrapper {
    max-width: 460px;
    margin: 0 auto;
  }

  // Section headings
  .section-heading {
    display: flex;
    align-items: center;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: #64748b;
    padding-bottom: 4px;
    border-bottom: 1px solid rgba(100, 116, 139, 0.2);
    margin-bottom: 2px;

    .body--dark & { color: #94a3b8; border-bottom-color: rgba(148, 163, 184, 0.2); }
  }

  // Attribute rows
  .section-body { }

  .attr-row {
    display: flex;
    gap: 10px;
    align-items: center;
    min-height: 30px;
    border-bottom: 1px solid rgba(100, 116, 139, 0.06);

    .body--dark & { border-bottom-color: rgba(148, 163, 184, 0.08); }
  }

  .attr-label {
    font-size: 12px;
    color: #94a3b8;
    font-weight: 500;
    padding: 3px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .attr-value {
    text-align: center;
    font-size: 13px;
    padding: 3px 6px;
    border-radius: 4px;
    transition: background 0.1s;
  }
}
</style>
