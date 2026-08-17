<template>
    <q-dialog :model-value="show" persistent @hide="$emit('close')">
        <q-card class="managed-team-dialog">
            <q-card-section class="row items-center q-pb-none card-header">
                <q-icon name="shield" size="md" class="q-mr-sm" />
                <div class="text-h6">Which team do you manage?</div>
                <q-space />
                <q-btn icon="close" flat round dense @click="$emit('skip')" />
            </q-card-section>

            <q-card-section class="q-pt-md">
                <div class="text-caption text-grey-6 q-mb-md">
                    Used to compare players against your own squad (AI Scout Report) and to
                    pre-fill Who to Sign. You can change this any time, or skip it entirely.
                </div>

                <q-select
                    filled
                    v-model="club"
                    :options="clubOptions"
                    label="Club"
                    class="q-mb-md"
                    use-input
                    hide-selected
                    fill-input
                    input-debounce="300"
                    behavior="menu"
                    @filter="filterClubOptions"
                    @update:model-value="onClubChange"
                >
                    <template v-slot:no-option>
                        <q-item><q-item-section class="text-grey">No results</q-item-section></q-item>
                    </template>
                </q-select>
                <q-select
                    filled
                    v-model="division"
                    :options="divisionOptions"
                    label="Division"
                    hint="Auto-filled from your club's players — change if it's wrong"
                />
            </q-card-section>

            <q-card-actions align="right">
                <q-btn flat label="Skip for now" @click="$emit('skip')" />
                <q-btn
                    unelevated
                    color="primary"
                    label="Save"
                    :disable="!club || !division"
                    @click="save"
                />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { computed, defineComponent, ref, watch } from 'vue'

export default defineComponent({
  name: 'ManagedTeamDialog',
  props: {
    show: { type: Boolean, default: false },
    datasetId: { type: String, default: null },
    clubs: { type: Array, default: () => [] },
    players: { type: Array, default: () => [] },
    initialClub: { type: String, default: '' },
    initialDivision: { type: String, default: '' },
  },
  emits: ['close', 'skip', 'saved'],
  setup(props, { emit }) {
    const club = ref(props.initialClub || '')
    const division = ref(props.initialDivision || '')
    const clubOptions = ref(props.clubs)

    function filterClubOptions(val, update, abort) {
      if (val.length < 1 && val !== '') {
        abort()
        return
      }
      update(() => {
        const needle = val.toLowerCase()
        clubOptions.value = props.clubs.filter((c) => c.toLowerCase().indexOf(needle) > -1)
      })
    }

    // Every division any player at the chosen club plays in, most-common first — so the
    // auto-fill lands on the right value even if a club's players disagree on division
    // (data-quality edge case), while still leaving the field editable.
    const divisionOptions = computed(() => {
      if (!club.value) return []
      const counts = new Map()
      for (const p of props.players) {
        if (p.club !== club.value) continue
        const d = p.division
        if (!d || !d.trim()) continue
        counts.set(d, (counts.get(d) || 0) + 1)
      }
      return [...counts.entries()].sort((a, b) => b[1] - a[1]).map(([d]) => d)
    })

    function onClubChange() {
      division.value = divisionOptions.value[0] || ''
    }

    watch(
      () => props.show,
      (visible) => {
        if (visible) {
          club.value = props.initialClub || ''
          division.value = props.initialDivision || ''
        }
      }
    )

    async function save() {
      if (!club.value || !division.value) return
      try {
        const res = await fetch(`/api/managed-team/${props.datasetId}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ club: club.value, division: division.value }),
        })
        if (!res.ok) throw new Error(`Unexpected error (${res.status}).`)
        emit('saved', { club: club.value, division: division.value })
      } catch (e) {
        console.error('[ManagedTeamDialog] failed to save managed team', e)
      }
    }

    return { club, division, divisionOptions, clubOptions, filterClubOptions, onClubChange, save }
  },
})
</script>

<style scoped>
.managed-team-dialog {
    width: 420px;
    max-width: 95vw;
}
.card-header {
    font-weight: 600;
}
</style>
