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
                <div v-else-if="phase === 'results'" class="dossier">
                    <div class="eyebrow-row">
                        <div>
                            <div class="unit">Head Scout &middot; Recruitment Report</div>
                            <div class="club">{{ form.team }}</div>
                        </div>
                        <div class="tabs-a">
                            <button
                                v-for="rec in response.recommendations"
                                :key="rec.position"
                                :class="{ active: activeTab === rec.position }"
                                @click="activeTab = rec.position"
                            >{{ rec.position }}</button>
                        </div>
                    </div>

                    <template v-for="rec in response.recommendations" :key="rec.position">
                        <div v-if="activeTab === rec.position">
                            <div v-if="rec.positionRationale" class="position-rationale">{{ rec.positionRationale }}</div>

                            <div class="verdict-row">
                                <div class="verdict-head">Sign <span class="who">{{ rec.mainPick.name }}</span></div>
                                <div class="grade-chip" :class="'grade-' + gradeSlug(rec.mainPick.grade)">
                                    <div class="num">{{ rec.mainPick.grade }}</div>
                                    <div class="lbl">Score<br />{{ rec.mainPick.signingScore }}/100</div>
                                </div>
                            </div>

                            <div class="row q-col-gutter-lg q-mb-md main-pick-row">
                                <div class="col-auto">
                                    <PlayerCards
                                        v-if="fullPlayer(rec.mainPick.uid)"
                                        :player="fullPlayer(rec.mainPick.uid)"
                                        :currency-symbol="currencySymbol"
                                        :nation-flag-url="getFlagUrl(fullPlayer(rec.mainPick.uid).nationalityIso || fullPlayer(rec.mainPick.uid).nationality_fifa_code)"
                                        :club-image-url="getTeamLogoUrl(fullPlayer(rec.mainPick.uid).club)"
                                        :player-face-url="getPlayerFaceUrl(fullPlayer(rec.mainPick.uid).name, fullPlayer(rec.mainPick.uid).club)"
                                        :dataset-id="datasetId"
                                        @click="openDetail(fullPlayer(rec.mainPick.uid))"
                                    />
                                </div>
                                <div class="col budget-strip">
                                    <div class="budget-item">
                                        <div class="lbl">
                                            <span>Transfer fee</span>
                                            <span>{{ formatCurrency(rec.mainPick.transferValueAmount, currencySymbol) }} / {{ formatCurrency((form.maxTransferBudgetM || 0) * 1000000, currencySymbol) }}</span>
                                        </div>
                                        <div class="budget-track"><div class="budget-fill" :style="{ width: budgetPct(rec.mainPick.transferValueAmount, form.maxTransferBudgetM * 1000000) + '%' }"></div></div>
                                    </div>
                                    <div class="budget-item">
                                        <div class="lbl">
                                            <span>Weekly wage</span>
                                            <span>{{ formatCurrency(rec.mainPick.wageAmount, currencySymbol) }} / {{ formatCurrency((form.maxWageBudgetK || 0) * 1000, currencySymbol) }}</span>
                                        </div>
                                        <div class="budget-track"><div class="budget-fill" :style="{ width: budgetPct(rec.mainPick.wageAmount, form.maxWageBudgetK * 1000) + '%' }"></div></div>
                                    </div>

                                    <div class="split-a">
                                        <div class="col-card strengths">
                                            <h3>Strengths</h3>
                                            <ul>
                                                <li v-for="(b, i) in rec.mainPick.strengths" :key="i">{{ b }}</li>
                                            </ul>
                                        </div>
                                        <div class="col-card watch">
                                            <h3>Watch-outs</h3>
                                            <ul>
                                                <li v-for="(b, i) in rec.mainPick.watchOuts" :key="i">{{ b }}</li>
                                            </ul>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div v-if="rec.closestSquadPlayer && fullPlayer(rec.mainPick.uid) && fullPlayer(rec.closestSquadPlayer.uid)" class="compare-a">
                                <h3>Vs. closest current option</h3>
                                <p class="sub">{{ rec.closestSquadPlayer.name }} &middot; {{ rec.position }} &middot; age {{ rec.closestSquadPlayer.age }} &middot; already at the club</p>
                                <div class="compare-legend">
                                    <span><i class="swatch swatch-main"></i>{{ rec.mainPick.name }} (target)</span>
                                    <span><i class="swatch swatch-squad"></i>{{ rec.closestSquadPlayer.name }} (squad)</span>
                                </div>
                                <div v-for="stat in fifaStatKeys" :key="stat.key" class="bar-row">
                                    <div class="k">{{ stat.label }}</div>
                                    <div class="bar-track">
                                        <div class="bar-fill main" :style="{ width: fullPlayer(rec.mainPick.uid)[stat.key] + '%' }"></div>
                                    </div>
                                    <div class="bar-track squad-track">
                                        <div class="bar-fill squad" :style="{ width: fullPlayer(rec.closestSquadPlayer.uid)[stat.key] + '%' }"></div>
                                    </div>
                                </div>
                                <q-btn flat dense size="sm" class="q-mt-sm" icon="compare_arrows" label="Full attribute comparison" @click="openComparison(rec)" />
                            </div>

                            <div v-if="rec.runnersUp.length" class="runners-a">
                                <h3>Also considered for this pick</h3>
                                <div
                                    v-for="(p, i) in rec.runnersUp"
                                    :key="p.uid"
                                    class="runner-row"
                                    @click="fullPlayer(p.uid) && openDetail(fullPlayer(p.uid))"
                                >
                                    <div class="runner-rank">{{ i + 2 }}</div>
                                    <div>
                                        <div class="runner-name">{{ p.name }} <span class="runner-val">&middot; {{ p.club }} &middot; {{ p.shortPositions.join('/') }}</span></div>
                                        <div class="runner-note" v-if="p.strengths && p.strengths[0]">{{ p.strengths[0] }}</div>
                                    </div>
                                    <div class="grade-chip small" :class="'grade-' + gradeSlug(p.grade)">
                                        <div class="num">{{ p.grade }}</div>
                                    </div>
                                    <div class="runner-val">{{ formatCurrency(p.transferValueAmount, currencySymbol) }}</div>
                                </div>
                            </div>

                            <div v-if="consideredFullPlayers(rec).length" class="mini-table">
                                <h3>Full shortlist scouted</h3>
                                <p class="count">{{ consideredFullPlayers(rec).length }} players pulled from the market search for this position &middot; click a row for the full attribute file</p>
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
                        </div>
                    </template>

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

    const consideredFullPlayers = (rec) =>
      (rec.playersConsidered || []).map((p) => fullPlayer(p.uid)).filter((p) => p !== null)

    // Grade is LLM-authored (D/C/B/A/A*) — slug it for a CSS class, not a literal lookup,
    // since "A*" isn't a valid class-name character sequence on its own.
    const gradeSlug = (grade) => (grade === 'A*' ? 'aplus' : String(grade || '').toLowerCase())

    const budgetPct = (used, max) => {
      if (!max || max <= 0) return 0
      return Math.min(100, Math.round((used / max) * 100))
    }

    // FIFA-style aggregate stats for the visual squad-comparison bars only — these are raw
    // Player fields (not LLM-authored text), so they're a different concern from the
    // FM-attribute-only rule that applies to the scout's written strengths/watchOuts.
    const fifaStatKeys = [
      { key: 'pac', label: 'Pace' },
      { key: 'sho', label: 'Sho' },
      { key: 'pas', label: 'Pas' },
      { key: 'dri', label: 'Dri' },
      { key: 'def', label: 'Def' },
      { key: 'phy', label: 'Phy' },
    ]

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
      gradeSlug,
      budgetPct,
      fifaStatKeys,
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
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 64px 16px;
    text-align: center;
}

/* ---------- Dossier results panel ---------- */
/* Committed to one dark "terminal" tone regardless of app theme, matching how a real
   scouting platform's report view reads — same treatment validated in the mockup review. */
.dossier {
    --d-ground: #0f1620;
    --d-panel: #171f2b;
    --d-panel-2: #1d2634;
    --d-border: #2a3547;
    --d-text: #e9edf3;
    --d-text-dim: #8b96a8;
    --d-accent: #3fb56a;
    --d-risk: #e0a23d;
    --d-blue: #5c8fd6;
    background: var(--d-ground);
    border: 1px solid var(--d-border);
    border-radius: 14px;
    padding: 22px 24px 26px;
    color: var(--d-text);
}
.dossier h3 {
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 12.5px;
    font-weight: 700;
    color: var(--d-text-dim);
    margin: 0 0 12px;
}
.dossier .eyebrow-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    border-bottom: 1px solid var(--d-border);
    padding-bottom: 14px;
    margin-bottom: 18px;
}
.dossier .unit {
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 11px;
    color: var(--d-text-dim);
}
.dossier .club {
    text-transform: uppercase;
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 0.02em;
}
.tabs-a { display: flex; gap: 20px; }
.tabs-a button {
    background: none; border: none; cursor: pointer;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    font-size: 13px;
    font-weight: 700;
    color: var(--d-text-dim);
    padding: 6px 0;
    border-bottom: 2px solid transparent;
}
.tabs-a button.active { color: var(--d-accent); border-color: var(--d-accent); }
.tabs-a button:focus-visible { outline: 2px solid var(--d-accent); outline-offset: 2px; }

.position-rationale {
    font-size: 13px;
    color: var(--d-text-dim);
    margin-bottom: 14px;
    line-height: 1.5;
}

.verdict-row {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
    margin-bottom: 16px;
}
.verdict-head {
    text-transform: uppercase;
    font-size: 26px;
    font-weight: 700;
    line-height: 1.05;
}
.verdict-head .who { color: var(--d-accent); }

.grade-chip {
    display: flex; align-items: center; gap: 10px;
    background: var(--d-panel-2);
    border: 1px solid var(--d-border);
    border-radius: 10px;
    padding: 7px 12px;
    white-space: nowrap;
}
.grade-chip.small { padding: 4px 9px; gap: 6px; }
.grade-chip .num {
    font-size: 20px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
}
.grade-chip.small .num { font-size: 15px; }
.grade-chip .lbl { font-size: 10.5px; color: var(--d-text-dim); text-transform: uppercase; letter-spacing: 0.06em; }
.grade-chip.grade-aplus .num, .grade-chip.grade-a .num { color: var(--d-accent); }
.grade-chip.grade-b .num { color: var(--d-blue); }
.grade-chip.grade-c .num { color: var(--d-risk); }
.grade-chip.grade-d .num { color: #d66358; }

.main-pick-row { align-items: flex-start; }
.budget-strip { display: flex; flex-direction: column; gap: 12px; min-width: 260px; }
.budget-item .lbl {
    display: flex; justify-content: space-between; gap: 8px;
    font-size: 11.5px; color: var(--d-text-dim); margin-bottom: 5px;
    font-variant-numeric: tabular-nums;
}
.budget-track { height: 6px; border-radius: 4px; background: var(--d-panel-2); overflow: hidden; }
.budget-fill { height: 100%; background: var(--d-accent); border-radius: 4px; }

.split-a {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
    margin-top: 6px;
}
.col-card {
    background: var(--d-panel);
    border: 1px solid var(--d-border);
    border-radius: 10px;
    padding: 12px 14px;
}
.col-card.strengths h3 { color: var(--d-accent); }
.col-card.watch h3 { color: var(--d-risk); }
.col-card ul { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8px; }
.col-card li {
    font-size: 12.5px;
    line-height: 1.45;
    padding-left: 16px;
    position: relative;
}
.col-card.strengths li::before {
    content: ''; position: absolute; left: 0; top: 5px;
    width: 7px; height: 7px; border-radius: 2px; background: var(--d-accent);
}
.col-card.watch li::before {
    content: ''; position: absolute; left: 0; top: 5px;
    width: 7px; height: 7px; border-radius: 2px; background: var(--d-risk);
}

.compare-a {
    background: var(--d-panel);
    border: 1px solid var(--d-border);
    border-radius: 10px;
    padding: 14px 16px 16px;
    margin-bottom: 22px;
}
.compare-a .sub { font-size: 12px; color: var(--d-text-dim); margin: -6px 0 12px; }
.compare-legend { display: flex; gap: 16px; margin-bottom: 10px; font-size: 12px; color: var(--d-text-dim); }
.compare-legend span { display: inline-flex; align-items: center; gap: 6px; }
.swatch { width: 10px; height: 10px; border-radius: 2px; display: inline-block; }
.swatch-main { background: var(--d-accent); }
.swatch-squad { background: var(--d-blue); }
.bar-row { display: grid; grid-template-columns: 40px 1fr 1fr; gap: 8px; align-items: center; margin-bottom: 7px; }
.bar-row .k { font-size: 11px; color: var(--d-text-dim); text-transform: uppercase; letter-spacing: 0.04em; }
.bar-track { height: 8px; background: var(--d-panel-2); border-radius: 4px; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 4px; }
.bar-fill.main { background: var(--d-accent); }
.bar-fill.squad { background: var(--d-blue); }

.runners-a { margin-bottom: 22px; }
.runner-row {
    display: grid;
    grid-template-columns: 22px 1fr auto auto;
    align-items: center;
    gap: 12px;
    padding: 9px 6px;
    border-bottom: 1px solid var(--d-border);
    cursor: pointer;
    border-radius: 6px;
}
.runner-row:hover { background: var(--d-panel-2); }
.runner-row:last-child { border-bottom: none; }
.runner-rank { color: var(--d-text-dim); font-size: 13px; font-weight: 700; }
.runner-name { font-size: 13px; font-weight: 600; }
.runner-note { font-size: 11.5px; color: var(--d-text-dim); margin-top: 2px; }
.runner-val { font-size: 11.5px; color: var(--d-text-dim); font-variant-numeric: tabular-nums; white-space: nowrap; }

.mini-table h3 { margin-bottom: 4px; }
.mini-table .count { font-size: 12px; color: var(--d-text-dim); margin: 0 0 12px; }
</style>
