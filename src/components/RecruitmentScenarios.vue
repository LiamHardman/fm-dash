<template>
  <q-expansion-item label="Recruitment scenarios" icon="compare_arrows" class="q-my-md" bordered>
    <q-card flat><q-card-section>
      <p>Compare signings and sales using this Shortlist's planning currency and wage period. Budgets are the additional amount you can spend; sale proceeds and wage savings increase it.</p>
      <q-banner v-if="!hasManagedClub" class="bg-orange-1 text-orange-10">Load a Dataset and configure its Managed club to create a scenario.</q-banner>
      <div class="row q-gutter-sm q-mb-md"><q-input v-model="newName" outlined dense label="Scenario name" maxlength="100" /><q-btn label="Create scenario" color="primary" :disable="!hasManagedClub || !newName.trim()" @click="create" /></div>
      <q-card v-for="scenario in scenarios" :key="scenario.id" flat bordered class="q-mb-md">
        <q-card-section>
          <div class="row items-center justify-between"><strong>{{ scenario.name }}</strong><q-btn flat icon="delete" aria-label="Delete scenario" @click="remove(scenario)" /></div>
          <q-banner v-if="scenario.datasetId !== playerStore.currentDatasetId" class="q-my-sm">Load this scenario's source Dataset to evaluate squad coverage.<q-btn flat label="Open Dataset" :to="`/dataset/${scenario.datasetId}`" /></q-banner>
          <div class="row q-col-gutter-sm q-mt-sm">
            <q-input class="col-12 col-sm-6" v-model.number="scenario.feeBudget" type="number" min="0" outlined dense :label="`Available transfer budget (${currency})`" @change="persist" />
            <q-input class="col-12 col-sm-6" v-model.number="scenario.wageBudget" type="number" min="0" outlined dense :label="`Additional wage budget (${currency}/${period})`" @change="persist" />
          </div>
          <q-select class="q-mt-sm" outlined multiple use-chips emit-value map-options label="Planned signings" :model-value="scenario.signings.map(referenceKey)" :options="signingOptions" @update:model-value="setSignings(scenario, $event)" />
          <div v-for="sale in scenario.sales" :key="referenceKey(sale.playerRef)" class="row q-col-gutter-sm q-mt-sm">
            <span class="col-12">Sell {{ saleName(sale) }}</span>
            <q-input class="col" v-model.number="sale.fee" type="number" min="0" outlined dense :label="`Expected fee (${currency})`" @change="persist" />
            <q-input class="col" v-model.number="sale.wage" type="number" min="0" outlined dense :label="`Wage savings (${currency}/${period})`" @change="persist" />
            <q-btn flat icon="close" aria-label="Remove planned sale" @click="scenario.sales = scenario.sales.filter(s => s !== sale); persist()" />
          </div>
          <q-select v-if="scenario.datasetId === playerStore.currentDatasetId" class="q-mt-sm" outlined dense label="Add a planned sale from Managed club" :model-value="null" :options="saleOptions(scenario)" emit-value map-options @update:model-value="addSale(scenario, $event)" />
          <div class="q-mt-md" v-if="result(scenario).missingTargets">{{ result(scenario).missingTargets }} signing targets were removed from the Shortlist. Remove or restore these references before using the totals.</div>
          <div v-else class="row q-gutter-md q-mt-md">
            <span :class="result(scenario).feeRemaining < 0 ? 'text-negative' : ''">Transfer budget remaining: {{ money(result(scenario).feeRemaining) }}</span>
            <span :class="result(scenario).wageRemaining < 0 ? 'text-negative' : ''">Wage budget remaining: {{ money(result(scenario).wageRemaining) }}/{{ period }}</span>
          </div>
          <div v-if="scenario.datasetId === playerStore.currentDatasetId && managedClub" class="q-mt-sm">
            <span v-if="result(scenario).unresolvedPlayers">Coverage is incomplete: {{ result(scenario).unresolvedPlayers }} Player references unavailable.</span>
            <q-chip v-for="entry in result(scenario).coverage" :key="entry.position">{{ entry.position }}: {{ entry.count }}</q-chip>
            <p class="text-caption">Players eligible for multiple positions contribute to each position.</p>
          </div>
        </q-card-section>
      </q-card>
    </q-card-section></q-card>
  </q-expansion-item>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { apiUrl } from '../utils/apiEndpoint'
import { evaluateScenario, referenceKey } from '../utils/recruitmentScenario'

const props = defineProps({
  store: { type: Object, required: true },
  entries: { type: Array, default: () => [] },
})
const playerStore = usePlayerStore()
const newName = ref('')
const managedClub = ref(null)
const scenarios = computed(() => props.store.activeList?.scenarios || [])
const currency = computed(() => props.store.activeList?.currency || 'USD')
const period = computed(() => props.store.activeList?.wagePeriod || 'week')
const squad = computed(() =>
  playerStore.allPlayers.filter(
    (p) =>
      p.club === managedClub.value?.club &&
      (!managedClub.value?.division || p.division === managedClub.value.division)
  )
)
watch(
  () => playerStore.currentDatasetId,
  async (id, _, onCleanup) => {
    managedClub.value = null
    if (!id) return
    const controller = new AbortController()
    onCleanup(() => controller.abort())
    try {
      const r = await fetch(apiUrl(`/api/managed-team/${id}`), { signal: controller.signal })
      const value = r.ok ? await r.json() : null
      if (!controller.signal.aborted) managedClub.value = value
    } catch {
      /* Setup remains available on the Dataset page. */
    }
  },
  { immediate: true }
)
const hasManagedClub = computed(() => !!managedClub.value && squad.value.length > 0)
const signingOptions = computed(() =>
  props.entries.map((e) => ({
    label: e.player?.name || `Unavailable player ${e.item.playerRef.playerUid}`,
    value: referenceKey(e.item.playerRef),
  }))
)
const money = (value) =>
  new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: currency.value,
    maximumFractionDigits: 0,
  }).format(value)
const persist = () => props.store.updatePlanning({ scenarios: scenarios.value })
async function create() {
  const scenario = {
    id: crypto.randomUUID(),
    name: newName.value.trim(),
    datasetId: playerStore.currentDatasetId,
    feeBudget: 0,
    wageBudget: 0,
    signings: [],
    sales: [],
  }
  await props.store.updatePlanning({ scenarios: [...scenarios.value, scenario] })
  newName.value = ''
}
const remove = (scenario) =>
  props.store.updatePlanning({ scenarios: scenarios.value.filter((s) => s !== scenario) })
function setSignings(scenario, keys) {
  scenario.signings = keys
    .map((key) => props.entries.find((e) => referenceKey(e.item.playerRef) === key)?.item.playerRef)
    .filter(Boolean)
  persist()
}
function saleOptions(scenario) {
  return squad.value
    .filter((p) => !scenario.sales.some((s) => s.playerRef.playerUid === Number(p.uid ?? p.UID)))
    .map((p) => ({ label: p.name, value: Number(p.uid ?? p.UID) }))
}
function addSale(scenario, uid) {
  if (!uid) return
  scenario.sales.push({
    playerRef: { datasetId: scenario.datasetId, playerUid: uid },
    fee: 0,
    wage: 0,
  })
  persist()
}
const saleName = (sale) =>
  (sale.playerRef.datasetId === playerStore.currentDatasetId ? squad.value : []).find(
    (p) => Number(p.uid ?? p.UID) === sale.playerRef.playerUid
  )?.name || `Player ${sale.playerRef.playerUid}`
function result(scenario) {
  return evaluateScenario(
    scenario,
    props.store.activeList.items,
    scenario.datasetId === playerStore.currentDatasetId ? squad.value : [],
    (ref) => props.entries.find((e) => referenceKey(e.item.playerRef) === referenceKey(ref))?.player
  )
}
</script>
