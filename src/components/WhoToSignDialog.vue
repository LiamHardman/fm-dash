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
                <!-- No API key configured -->
                <div v-if="!hasApiKey" class="empty-state">
                    <q-icon name="vpn_key_off" size="48px" color="grey-6" />
                    <div class="text-subtitle1 q-mt-sm">No OpenAI API key set</div>
                    <div class="text-caption text-grey-6" style="max-width: 420px; text-align: center;">
                        Who to Sign uses your own OpenAI API key — add one in Settings (the gear icon in the
                        header) to use this feature.
                    </div>
                </div>

                <!-- Form -->
                <q-stepper v-else-if="phase === 'form'" v-model="step" flat animated color="primary">
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
                            <div class="main-pick-highlight q-pa-md q-mb-md">
                                <div class="row items-center">
                                    <div class="text-h6">{{ rec.mainPick.name }}</div>
                                    <q-space />
                                    <div class="text-caption">
                                        {{ rec.mainPick.club }} · {{ rec.mainPick.age }}y · {{ rec.mainPick.overall }} OVR
                                    </div>
                                </div>
                                <div class="text-caption q-mb-sm">{{ rec.mainPick.bestRoleOverall }}</div>
                                <ul class="q-mt-none">
                                    <li v-for="(b, i) in rec.mainPick.reasoning" :key="i">{{ b }}</li>
                                </ul>
                                <div class="text-caption text-grey-7">
                                    {{ formatCurrency(rec.mainPick.transferValueAmount, currencySymbol) }} ·
                                    {{ formatCurrency(rec.mainPick.wageAmount, currencySymbol) }}/wk
                                </div>
                            </div>
                            <div v-if="rec.runnersUp.length" class="text-subtitle2 q-mb-sm">Also considered</div>
                            <div class="row q-col-gutter-sm">
                                <div v-for="p in rec.runnersUp" :key="p.uid" class="col-12 col-md-6">
                                    <q-card flat bordered class="q-pa-sm">
                                        <div class="text-weight-medium">
                                            {{ p.name }} <span class="text-caption text-grey-6">({{ p.club }})</span>
                                        </div>
                                        <div class="text-caption">{{ p.reasoning[0] }}</div>
                                    </q-card>
                                </div>
                            </div>
                        </q-tab-panel>
                    </q-tab-panels>
                    <div class="row justify-end q-mt-md">
                        <q-btn flat label="New search" @click="phase = 'form'; step = 1" />
                    </div>
                </div>
            </q-card-section>
        </q-card>
    </q-dialog>
</template>

<script>
import { computed, reactive, ref, watch } from 'vue'
import { useUiStore } from '../stores/uiStore'
import { formatCurrency } from '../utils/currencyUtils'

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
  props: {
    show: Boolean,
    datasetId: { type: String, default: '' },
    teams: { type: Array, default: () => [] },
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
          try {
            const body = await res.json()
            if (body?.error) message = body.error
          } catch (_e) {}
          errorStatus.value = res.status
          errorMessage.value = message
          phase.value = 'error'
          return
        }

        const data = await res.json()
        response.value = data
        activeTab.value = data?.recommendations?.[0]?.position ?? null
        phase.value = 'results'
      } catch (_e) {
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
      teams: computed(() => props.teams),
      attributeOptions: ATTRIBUTE_OPTIONS,
      squadStatusOptions: SQUAD_STATUS_OPTIONS,
      submit,
      formatCurrency,
    }
  },
}
</script>

<style scoped>
.who-to-sign-dialog {
    max-width: 100%;
}
.who-to-sign-body {
    max-width: 900px;
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
.main-pick-highlight {
    border-left: 4px solid var(--q-primary);
    background: rgba(var(--q-primary-rgb, 25, 118, 210), 0.08);
    border-radius: 4px;
}
</style>
