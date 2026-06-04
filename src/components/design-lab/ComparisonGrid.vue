<template>
  <div style="overflow-x:auto">
    <table class="comparison-table">
      <thead>
        <tr>
          <th class="comp-field-header">Attribute</th>
          <th v-for="p in players" :key="p.uid" class="comp-player-header">
            <div class="comp-player-name">{{ p.name }}</div>
            <div class="comp-player-sub">{{ p.club }}</div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="field in fields" :key="field.key" class="comp-row">
          <td class="comp-field-label">{{ field.label }}</td>
          <td v-for="p in players" :key="p.uid" class="comp-cell"
              :style="cellStyle(p, field)">
            {{ field.format ? field.format(p) : (p[field.key] ?? '—') }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import { defineComponent } from 'vue'

const FIELDS = [
  { label: 'Overall', key: 'Overall', higherBetter: true },
  { label: 'Age', key: 'age', higherBetter: false },
  { label: 'PAC', key: 'pac', higherBetter: true },
  { label: 'SHO', key: 'sho', higherBetter: true },
  { label: 'PAS', key: 'pas', higherBetter: true },
  { label: 'DRI', key: 'dri', higherBetter: true },
  { label: 'DEF', key: 'def', higherBetter: true },
  { label: 'PHY', key: 'phy', higherBetter: true },
  {
    label: 'Transfer value',
    key: 'transferValueAmount',
    higherBetter: false,
    format: (p) => p.transfer_value,
  },
  { label: 'Value score', key: 'valueScore', higherBetter: true },
  { label: 'Wage', key: 'wageAmount', higherBetter: false, format: (p) => p.wage },
  { label: 'Personality', key: 'personality', higherBetter: null },
]

export default defineComponent({
  name: 'ComparisonGrid',
  props: { players: { type: Array, default: () => [] } },
  setup(props) {
    const cellStyle = (player, field) => {
      if (field.higherBetter === null) return {}
      const vals = props.players.map((p) => p[field.key]).filter((v) => v != null)
      if (vals.length < 2) return {}
      const best = field.higherBetter ? Math.max(...vals) : Math.min(...vals)
      const worst = field.higherBetter ? Math.min(...vals) : Math.max(...vals)
      const val = player[field.key]
      if (val == null) return {}
      if (val === best)
        return { background: 'rgba(34,197,94,0.18)', fontWeight: '700', color: '#15803d' }
      if (val === worst) return { background: 'rgba(239,68,68,0.12)', color: '#b91c1c' }
      return {}
    }
    return { fields: FIELDS, cellStyle }
  },
})
</script>

<style scoped>
.comparison-table {
  border-collapse: collapse;
  width: 100%;
}
.comp-field-header {
  text-align: left;
  padding: 8px 14px;
  font-size: 12px;
  font-weight: 700;
  border-bottom: 2px solid #e5e7eb;
  min-width: 110px;
  color: #6b7280;
}
.comp-player-header {
  text-align: center;
  padding: 8px 14px;
  font-size: 12px;
  border-bottom: 2px solid #e5e7eb;
  min-width: 130px;
}
.comp-player-name { font-weight: 700; font-size: 13px; }
.comp-player-sub  { font-size: 10px; color: #9ca3af; font-weight: 400; }
.comp-row { border-bottom: 1px solid #f3f4f6; }
.comp-field-label {
  padding: 6px 14px;
  font-size: 12px;
  color: #6b7280;
  font-weight: 600;
}
.comp-cell {
  padding: 6px 14px;
  text-align: center;
  font-size: 13px;
  transition: background 0.2s;
}
</style>
