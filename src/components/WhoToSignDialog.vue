<template>
    <q-dialog
        :model-value="show"
        @hide="$emit('close')"
        persistent
        maximized
        transition-show="slide-up"
        transition-hide="slide-down"
    >
        <q-card class="who-to-sign-dialog">
            <q-card-section class="row items-center q-pb-none card-header">
                <q-icon name="person_search" size="md" class="q-mr-sm" />
                <div class="text-h6">Who to Sign</div>
                <q-space />
                <q-btn icon="close" flat round dense v-close-popup @click="$emit('close')" />
            </q-card-section>

            <q-card-section class="q-pt-md who-to-sign-body">
                <!-- Non-blocking hint: no key saved in this browser. Doesn't block submission —
                     a self-hosted backend may have its own OPENAI_API_KEY fallback configured;
                     the backend is the source of truth and will error clearly if neither exists. -->
                <q-banner v-if="!hasApiKey && phase === 'form'" dense class="bg-grey-3 text-grey-9 q-mb-md">
                    <template v-slot:avatar><q-icon name="vpn_key_off" /></template>
                    No OpenAI API key saved in this browser's Settings. You can still try — this may work if
                    the server has its own key configured — otherwise add your key in Settings (the gear icon
                    in the header).
                </q-banner>

                <!-- Form -->
                <q-stepper v-if="phase === 'form'" v-model="step" flat animated color="primary">
                    <q-step :name="1" title="Basics" icon="account_balance" :done="step > 1">
                        <div class="row q-col-gutter-md">
                            <div class="col-12 col-md-6">
                                <q-select
                                    filled
                                    v-model="form.team"
                                    :options="teams"
                                    label="Which team do you manage?"
                                    :rules="[(val) => !!val || 'Required']"
                                />
                            </div>
                            <div class="col-12 col-md-6">
                                <q-select
                                    filled
                                    v-model="form.squadStatus"
                                    :options="squadStatusOptions"
                                    emit-value
                                    map-options
                                    label="Squad status target"
                                    :rules="[(val) => !!val || 'Required']"
                                />
                            </div>
                            <div class="col-6">
                                <q-input
                                    filled
                                    v-model.number="form.maxTransferBudgetM"
                                    type="number"
                                    :label="`Max transfer budget (${currencySymbol}m)`"
                                    :rules="[(val) => (val !== null && val > 0) || 'Required']"
                                />
                            </div>
                            <div class="col-6">
                                <q-input
                                    filled
                                    v-model.number="form.maxWageBudgetK"
                                    type="number"
                                    :label="`Max wage budget (${currencySymbol}k/wk)`"
                                    :rules="[(val) => (val !== null && val > 0) || 'Required']"
                                />
                            </div>
                        </div>
                    </q-step>

                    <q-step :name="2" title="Positions" icon="grid_view" :done="step > 2">
                        <div class="text-caption text-grey-6 q-mb-md">
                            Optional — skip if you'd rather your scout tell you what the squad needs.
                        </div>
                        <div v-for="(pos, i) in form.positions" :key="i" class="row q-col-gutter-sm q-mb-sm items-center">
                            <div class="col-3">
                                <q-input filled dense v-model="pos.position" label="Position (e.g. ST)" />
                            </div>
                            <div class="col-7">
                                <q-select
                                    filled
                                    dense
                                    multiple
                                    v-model="pos.attributes"
                                    :options="attributeOptions"
                                    label="Up to 5 attributes that matter most"
                                    :max-values="5"
                                    use-chips
                                />
                            </div>
                            <div class="col-2">
                                <q-btn flat round dense icon="delete" @click="form.positions.splice(i, 1)" />
                            </div>
                        </div>
                        <q-btn flat dense icon="add" label="Add position" @click="form.positions.push({ position: '', attributes: [] })" />
                    </q-step>

                    <q-step :name="3" title="Preferences" icon="tune">
                        <q-input filled v-model="form.playerProfile" label="Preferred player profile (optional)" class="q-mb-md" hint="e.g. proven experience, wonderkid, great value" />
                        <q-input filled v-model="form.freeText" type="textarea" autogrow label="Anything else? (optional)" />
                    </q-step>

                    <template #navigation>
                        <q-stepper-navigation class="row justify-between">
                            <q-btn v-if="step > 1" flat label="Back" @click="step--" />
                            <div v-else />
                            <q-btn v-if="step < 3" color="primary" unelevated label="Continue" @click="step++" />
                            <q-btn v-else color="primary" unelevated label="Find Signings" icon="search" @click="submit" />
                        </q-stepper-navigation>
                    </template>
                </q-stepper>

                <!-- Loading -->
                <div v-else-if="phase === 'loading'" class="empty-state">
                    <q-circular-progress indeterminate size="64px" color="primary" class="q-mb-md" />
                    <div class="text-subtitle1">Consulting your head scout…</div>
                    <div class="text-caption text-grey-6">Checking the market against your budget and requirements. This can take up to a minute.</div>
                </div>

                <!-- Error -->
                <div v-else-if="phase === 'error'" class="empty-state">
                    <q-icon name="report_problem" size="48px" color="negative" />
                    <div class="text-subtitle1 q-mt-sm">Scouting failed{{ errorStatus ? ` (${errorStatus})` : '' }}</div>
                    <div class="text-caption text-grey-7 q-mb-md" style="max-width: 480px; text-align: center;">{{ errorMessage }}</div>
                    <q-btn color="primary" label="Back to form" unelevated @click="phase = 'form'; step = 1" />
                </div>

                <!-- Results -->
                <div v-else-if="phase === 'results'">
                    <q-tabs v-model="activeTab" dense class="text-grey" active-color="primary" indicator-color="primary" align="left">
                        <q-tab v-for="rec in response.recommendations" :key="rec.position" :name="rec.position" :label="rec.position" />
                    </q-tabs>
                    <q-separator />
                    <q-tab-panels v-model="activeTab" animated>
                        <q-tab-panel v-for="rec in response.recommendations" :key="rec.position" :name="rec.position">
                            <div v-if="rec.positionRationale" class="text-caption text-grey-6 q-mb-md">{{ rec.positionRationale }}</div>

                            <div class="row q-col-gutter-md q-mb-md cards-row">
                                <div v-for="(p, idx) in topCardPlayers(rec)" :key="p.uid" class="col-auto card-col">
                                    <q-badge v-if="idx === 0" color="primary" class="q-mb-xs">Recommended</q-badge>
                                    <PlayerCards
                                        v-if="fullPlayer(p.uid)"
                                        :player="fullPlayer(p.uid)"
                                        :currency-symbol="currencySymbol"
                                        :nation-flag-url="getFlagUrl(fullPlayer(p.uid).nationalityIso || fullPlayer(p.uid).nationality_fifa_code)"
                                        :club-image-url="getTeamLogoUrl(fullPlayer(p.uid).club)"
                                        :player-face-url="getPlayerFaceUrl(fullPlayer(p.uid).name, fullPlayer(p.uid).club)"
                                        :dataset-id="datasetId"
                                        @click="openDetail(fullPlayer(p.uid))"
                                    />
                                    <ul class="reasoning-list text-caption">
                                        <li v-for="(b, i) in p.reasoning" :key="i">{{ b }}</li>
                                    </ul>
                                </div>
                            </div>

                            <div v-if="rec.closestSquadPlayer" class="row q-mb-md">
                                <q-btn
                                    flat
                                    dense
                                    icon="compare_arrows"
                                    :label="`Compare with ${rec.closestSquadPlayer.name} (current squad)`"
                                    @click="openComparison(rec)"
                                />
                            </div>

                            <div v-if="consideredFullPlayers(rec).length">
                                <div class="text-subtitle2 q-mb-sm">Other players considered</div>
                                <PlayerDataTable
                                    :players="consideredFullPlayers(rec)"
                                    :loading="false"
                                    @player-selected="openDetail"
                                    :currency-symbol="currencySymbol"
                                    :dataset-id="datasetId"
                                    default-sort-field="MBR"
                                    default-sort-direction="desc"
                                />
                            </div>
                        </q-tab-panel>
                    </q-tab-panels>
                    <div class="row justify-end q-mt-md">
                        <q-btn flat label="New search" @click="phase = 'form'; step = 1" />
                    </div>
                </div>
            </q-card-section>
        </q-card>

        <PlayerDetailDialog
            :player="selectedPlayer"
            :show="showDetail"
            @close="showDetail = false"
            :currency-symbol="currencySymbol"
            :dataset-id="datasetId"
        />
        <PlayerComparisonDialog
            :show="showComparison"
            :players="comparisonPlayers"
            @close="showComparison = false"
        />
    </q-dialog>
</template>

<script>
import { computed, reactive, ref, watch } from 'vue'
import { useUiStore } from '../stores/uiStore'
import { formatCurrency } from '../utils/currencyUtils'
import { getFlagUrl, getPlayerFaceUrl, getTeamLogoUrl } from '../utils/imageOptimization'
import PlayerCards from './PlayerCards.vue'
import PlayerComparisonDialog from './PlayerComparisonDialog.vue'
import PlayerDataTable from './PlayerDataTable.vue'
import PlayerDetailDialog from './PlayerDetailDialog.vue'

const ATTRIBUTE_OPTIONS = [
  'Acceleration',
  'Pace',
  'Strength',
  'Stamina',
  'Natural Fitness',
  'Balance',
  'Jumping Reach',
  'Agility',
  'Aggression',
  'Anticipation',
  'Bravery',
  'Composure',
  'Concentration',
  'Decisions',
  'Determination',
  'Flair',
  'Leadership',
  'Off The Ball',
  'Positioning',
  'Team Work',
  'Vision',
  'Work Rate',
  'Corners',
  'Crossing',
  'Dribbling',
  'Finishing',
  'First Touch',
  'Free Kick Taking',
  'Heading',
  'Long Shots',
  'Long Throws',
  'Marking',
  'Passing',
  'Penalty Taking',
  'Tackling',
  'Technique',
  'Aerial Reach',
  'Command Of Area',
  'Communication',
  'Eccentricity',
  'Handling',
  'Kicking',
  'One On Ones',
  'Reflexes',
  'Rushing Out (Tendency)',
  'Throwing',
  'Punching',
]

const SQUAD_STATUS_OPTIONS = [
  { label: 'Star player', value: 'star' },
  { label: 'Regular starter', value: 'regular_starter' },
  { label: 'Rotation option', value: 'rotation' },
]

function defaultForm() {
  return {
    team: null,
    maxTransferBudgetM: null,
    maxWageBudgetK: null,
    squadStatus: null,
    positions: [],
    playerProfile: '',
    freeText: '',
  }
}

export default {
  name: 'WhoToSignDialog',
  components: {
    PlayerCards,
    PlayerComparisonDialog,
    PlayerDataTable,
    PlayerDetailDialog,
  },
  props: {
    show: Boolean,
    datasetId: { type: String, default: '' },
    teams: { type: Array, default: () => [] },
    players: { type: Array, default: () => [] },
    currencySymbol: { type: String, default: '£' },
  },
  emits: ['close'],
  setup(props) {
    const uiStore = useUiStore()

    const hasApiKey = computed(() => !!uiStore.openaiApiKey)
    const phase = ref('form')
    const step = ref(1)
    const form = reactive(defaultForm())
    const response = ref(null)
    const errorStatus = ref(null)
    const errorMessage = ref('')
    const activeTab = ref(null)
    const selectedPlayer = ref(null)
    const showDetail = ref(false)
    const comparisonPlayers = ref([])
    const showComparison = ref(false)

    // The player store keeps uid as a string (protobuf int64 precision safety), but the
    // who-to-sign API returns uid as a plain JSON number — normalize both sides to string
    // so the lookup doesn't silently miss on a type mismatch.
    const playersByUid = computed(() => {
      const map = new Map()
      for (const p of props.players) {
        map.set(String(p.uid), p)
      }
      return map
    })
    const fullPlayer = (uid) => playersByUid.value.get(String(uid)) || null

    // "Top 3" per docs: the main pick plus the next two runners-up.
    const topCardPlayers = (rec) => [rec.mainPick, ...(rec.runnersUp || []).slice(0, 2)]

    const consideredFullPlayers = (rec) =>
      (rec.playersConsidered || []).map((p) => fullPlayer(p.uid)).filter((p) => p !== null)

    const openDetail = (player) => {
      selectedPlayer.value = player
      showDetail.value = true
    }

    const openComparison = (rec) => {
      const mainFull = fullPlayer(rec.mainPick.uid)
      const squadFull = rec.closestSquadPlayer ? fullPlayer(rec.closestSquadPlayer.uid) : null
      comparisonPlayers.value = [mainFull, squadFull].filter((p) => p !== null)
      showComparison.value = true
    }

    watch(
      () => props.show,
      (visible) => {
        if (visible) {
          phase.value = 'form'
          step.value = 1
          Object.assign(form, defaultForm())
        }
      }
    )

    const submit = async () => {
      phase.value = 'loading'
      try {
        const requestBody = {
          team: form.team,
          maxTransferBudget: Math.round((form.maxTransferBudgetM || 0) * 1000000),
          maxWageBudget: Math.round((form.maxWageBudgetK || 0) * 1000),
          squadStatus: form.squadStatus,
          positions: form.positions
            .filter((p) => p.position)
            .map((p) => ({ position: p.position, attributes: p.attributes })),
          playerProfile: form.playerProfile,
          freeText: form.freeText,
        }

        console.log('[WhoToSign] request', requestBody)

        const res = await fetch(`/api/who-to-sign/${props.datasetId}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-OpenAI-Api-Key': uiStore.openaiApiKey,
          },
          body: JSON.stringify(requestBody),
        })

        if (!res.ok) {
          let message = `Unexpected error (${res.status}).`
          let bodyForLog = null
          try {
            bodyForLog = await res.json()
            if (bodyForLog?.error) message = bodyForLog.error
          } catch (_e) {}
          console.error('[WhoToSign] error response', { status: res.status, body: bodyForLog })
          errorStatus.value = res.status
          errorMessage.value = message
          phase.value = 'error'
          return
        }

        const data = await res.json()
        console.log('[WhoToSign] response', data)
        response.value = data
        activeTab.value = data?.recommendations?.[0]?.position ?? null
        phase.value = 'results'
      } catch (e) {
        console.error('[WhoToSign] request threw', e)
        errorStatus.value = null
        errorMessage.value = 'Could not reach the server. Check your connection and try again.'
        phase.value = 'error'
      }
    }

    return {
      hasApiKey,
      phase,
      step,
      form,
      response,
      errorStatus,
      errorMessage,
      activeTab,
      selectedPlayer,
      showDetail,
      comparisonPlayers,
      showComparison,
      teams: computed(() => props.teams),
      attributeOptions: ATTRIBUTE_OPTIONS,
      squadStatusOptions: SQUAD_STATUS_OPTIONS,
      submit,
      formatCurrency,
      getFlagUrl,
      getTeamLogoUrl,
      getPlayerFaceUrl,
      fullPlayer,
      topCardPlayers,
      consideredFullPlayers,
      openDetail,
      openComparison,
    }
  },
}
</script>

<style scoped>
.who-to-sign-dialog {
    max-width: 100%;
}
.who-to-sign-body {
    max-width: 1100px;
    margin: 0 auto;
}
.cards-row {
    flex-wrap: nowrap;
    overflow-x: auto;
    padding-bottom: 8px;
}
.card-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    min-width: 220px;
}
.reasoning-list {
    max-width: 240px;
    margin-top: 8px;
}
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 64px 16px;
    text-align: center;
}
.main-pick-highlight {
    border-left: 4px solid var(--q-primary);
    background: rgba(var(--q-primary-rgb, 25, 118, 210), 0.08);
    border-radius: 4px;
}
</style>
