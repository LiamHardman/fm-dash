import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

// localStorage-backed library of custom "attribute weight" profiles -- alternate
// recipes for how the FIFA-style category stats (PAC, SHO, PAS, DRI, DEF, PHY, and
// the goalkeeper categories) are derived from a player's raw 1-20 FM attributes.
// Only the profile *library* and which one is active live here, in the browser;
// the weights of whichever profile is active are pushed to the backend (which is
// what actually recomputes the stats -- see AttributeWeightsDialog.vue) so every
// page that reads player data reflects the chosen weights.
const STORAGE_KEY_PROFILES = 'attributeWeightProfiles'
const STORAGE_KEY_ACTIVE = 'activeAttributeWeightProfileId'

// Mirrors defaultAttributeWeightsGo in src/api/config.go. Used as a fallback
// baseline before the live config has loaded from the backend, and as the
// target for "Reset to default".
export const DEFAULT_ATTRIBUTE_WEIGHTS = {
  PAC: { Acc: 12, Pac: 12, Agi: 5 },
  SHO: { Fin: 8, Lon: 6, Pen: 4, Hea: 5, Cmp: 6, Tec: 5, Ant: 4, Dec: 4, Fla: 3 },
  PAS_standard: { Pas: 8, Cro: 6, Fre: 4, Vis: 7, Tec: 5, Tea: 4, Dec: 4, Cor: 3, Fir: 4, OtB: 3 },
  PAS_no_set_pieces: { Pas: 8, Vis: 7, Tec: 5, Tea: 4, Dec: 4, Fir: 4, OtB: 3 },
  PAS_no_off_ball: { Pas: 8, Cro: 6, Fre: 4, Vis: 7, Tec: 5, Tea: 4, Dec: 4, Fir: 4 },
  DRI: { Dri: 8, Fir: 7, Tec: 6, Fla: 5, Cmp: 4, OtB: 3 },
  DEF: { Mar: 8, Tck: 8, Hea: 6, Ant: 7, Cnt: 6, Pos: 7, Dec: 5, Cmp: 4, Bra: 5, Agg: 4, Wor: 4 },
  PHY: { Str: 8, Sta: 7, Nat: 6, Jum: 5, Bal: 4, Agg: 5, Bra: 4, Wor: 4 },
  GK: { Han: 20, Ref: 20, Cmd: 15, Aer: 15, '1v1': 10, Kic: 5, TRO: 5, Com: 3, Thr: 3, Ecc: 1 },
  DIV: { Aer: 8, Ref: 7, Agi: 6, '1v1': 7, Han: 5 },
  HAN: { Han: 10, Cmd: 7, Cmp: 5, Cnt: 4 },
  REF: { Ref: 10, Ant: 6, Cnt: 5, '1v1': 5 },
  KIC: { Kic: 8, Thr: 6, Tec: 5, Vis: 4, Pas: 3 },
  SPD: { Acc: 8, Pac: 8, TRO: 6 },
  POS: { Pos: 8, Cmd: 7, Ant: 6, Dec: 5, TRO: 4, Cnt: 4, Com: 3 },
}

// Full attribute names, keyed by the FM abbreviation used throughout the weight maps.
export const ATTRIBUTE_LABELS = {
  Acc: 'Acceleration',
  Pac: 'Pace',
  Agi: 'Agility',
  Fin: 'Finishing',
  Lon: 'Long Shots',
  Pen: 'Penalty Taking',
  Hea: 'Heading',
  Cmp: 'Composure',
  Tec: 'Technique',
  Ant: 'Anticipation',
  Dec: 'Decisions',
  Fla: 'Flair',
  Pas: 'Passing',
  Cro: 'Crossing',
  Fre: 'Free Kick Taking',
  Vis: 'Vision',
  Tea: 'Teamwork',
  Cor: 'Corners',
  Fir: 'First Touch',
  OtB: 'Off The Ball',
  Dri: 'Dribbling',
  Mar: 'Marking',
  Tck: 'Tackling',
  Cnt: 'Concentration',
  Pos: 'Positioning',
  Bra: 'Bravery',
  Agg: 'Aggression',
  Wor: 'Work Rate',
  Str: 'Strength',
  Sta: 'Stamina',
  Nat: 'Natural Fitness',
  Jum: 'Jumping Reach',
  Bal: 'Balance',
  Han: 'Handling',
  Ref: 'Reflexes',
  Cmd: 'Command of Area',
  Aer: 'Aerial Reach',
  '1v1': 'One on Ones',
  Kic: 'Kicking',
  TRO: 'Rushing Out (Tendency)',
  Com: 'Communication',
  Thr: 'Throwing',
  Ecc: 'Eccentricity',
}

// Category display metadata, grouped for the editor UI. `note` explains any special
// handling (e.g. the three passing profiles) in the guidance copy.
export const WEIGHT_CATEGORY_GROUPS = [
  {
    id: 'outfield',
    label: 'Outfield Stats',
    categories: [
      { id: 'PAC', label: 'Pace (PAC)' },
      { id: 'SHO', label: 'Shooting (SHO)' },
      {
        id: 'PAS_standard',
        label: 'Passing — Standard (PAS)',
        note: 'PAS uses whichever of these three passing profiles scores highest for the player.',
      },
      { id: 'PAS_no_set_pieces', label: 'Passing — No Set Pieces (PAS)' },
      { id: 'PAS_no_off_ball', label: 'Passing — No Off-the-Ball (PAS)' },
      { id: 'DRI', label: 'Dribbling (DRI)' },
      { id: 'DEF', label: 'Defending (DEF)' },
      { id: 'PHY', label: 'Physical (PHY)' },
    ],
  },
  {
    id: 'goalkeeping',
    label: 'Goalkeeping Stats',
    categories: [
      { id: 'GK', label: 'Goalkeeping — Overall (GK)' },
      { id: 'DIV', label: 'Diving (DIV)' },
      { id: 'HAN', label: 'Handling (HAN)' },
      { id: 'REF', label: 'Reflexes (REF)' },
      { id: 'KIC', label: 'Kicking (KIC)' },
      { id: 'SPD', label: 'Sweeping Speed (SPD)' },
      { id: 'POS', label: 'Positioning (POS)' },
    ],
  },
]

function generateId() {
  return (
    globalThis.crypto?.randomUUID?.() ||
    `weights-${Date.now()}-${Math.random().toString(36).slice(2)}`
  )
}

function cloneWeights(weights) {
  return JSON.parse(JSON.stringify(weights || {}))
}

export const useWeightsStore = defineStore('weights', () => {
  const profiles = ref([]) // [{ id, name, weights, createdAt, updatedAt }]
  const activeProfileId = ref(null) // null => using the app default weights
  const defaultWeights = ref(cloneWeights(DEFAULT_ATTRIBUTE_WEIGHTS))
  const defaultWeightsLoaded = ref(false)

  const activeProfile = computed(
    () => profiles.value.find((p) => p.id === activeProfileId.value) || null
  )

  function persistProfiles() {
    try {
      localStorage.setItem(STORAGE_KEY_PROFILES, JSON.stringify(profiles.value))
    } catch (_e) {}
  }

  function persistActiveProfileId() {
    try {
      if (activeProfileId.value) {
        localStorage.setItem(STORAGE_KEY_ACTIVE, activeProfileId.value)
      } else {
        localStorage.removeItem(STORAGE_KEY_ACTIVE)
      }
    } catch (_e) {}
  }

  function load() {
    try {
      const stored = localStorage.getItem(STORAGE_KEY_PROFILES)
      profiles.value = stored ? JSON.parse(stored) : []
    } catch (_e) {
      profiles.value = []
    }
    try {
      const storedActive = localStorage.getItem(STORAGE_KEY_ACTIVE)
      if (storedActive && profiles.value.some((p) => p.id === storedActive)) {
        activeProfileId.value = storedActive
      } else {
        activeProfileId.value = null
      }
    } catch (_e) {
      activeProfileId.value = null
    }
  }

  // Records what the backend is actually using right now, e.g. as returned by
  // GET /api/config's attributeWeights field. Only overwrites the fallback
  // baseline before any profile has ever been activated -- once the user has
  // picked a profile, the "default" entry in the dropdown should stay the
  // fixed built-in baseline, not drift with whatever was last active.
  function setDefaultWeights(weights) {
    if (!weights || Object.keys(weights).length === 0) return
    if (!defaultWeightsLoaded.value) {
      defaultWeights.value = cloneWeights(weights)
    }
    defaultWeightsLoaded.value = true
  }

  function createProfile(name, weights) {
    const trimmed = (name || '').trim()
    if (!trimmed) return null
    const profile = {
      id: generateId(),
      name: trimmed,
      weights: cloneWeights(weights),
      createdAt: Date.now(),
      updatedAt: Date.now(),
    }
    profiles.value = [...profiles.value, profile]
    persistProfiles()
    return profile
  }

  function updateProfileWeights(id, weights) {
    const profile = profiles.value.find((p) => p.id === id)
    if (!profile) return
    profile.weights = cloneWeights(weights)
    profile.updatedAt = Date.now()
    persistProfiles()
  }

  function renameProfile(id, name) {
    const profile = profiles.value.find((p) => p.id === id)
    const trimmed = (name || '').trim()
    if (!profile || !trimmed) return
    profile.name = trimmed
    profile.updatedAt = Date.now()
    persistProfiles()
  }

  function deleteProfile(id) {
    profiles.value = profiles.value.filter((p) => p.id !== id)
    persistProfiles()
    if (activeProfileId.value === id) {
      activeProfileId.value = null
      persistActiveProfileId()
    }
  }

  function setActiveProfileId(id) {
    activeProfileId.value = id || null
    persistActiveProfileId()
  }

  return {
    profiles,
    activeProfileId,
    activeProfile,
    defaultWeights,
    load,
    setDefaultWeights,
    createProfile,
    updateProfileWeights,
    renameProfile,
    deleteProfile,
    setActiveProfileId,
  }
})
