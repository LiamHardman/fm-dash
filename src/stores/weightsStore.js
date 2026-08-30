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

// Built-in weight sets covering common scouting philosophies. Unlike user-saved
// profiles these are read-only in the editor (edits must go through "Save As");
// see AttributeWeightsDialog.vue.
export const PRESET_WEIGHT_PROFILES = [
  {
    id: 'preset-physical-powerhouse',
    name: 'Physical Powerhouse',
    description:
      'Leans on raw athleticism (Acceleration, Pace, Strength, Stamina, Jumping, Aggression) across every category. Good for scouting direct/physical systems.',
    weights: {
      PAC: { Acc: 14, Pac: 14, Agi: 4 },
      SHO: { Fin: 6, Lon: 6, Pen: 3, Hea: 9, Cmp: 4, Tec: 3, Ant: 3, Dec: 3, Fla: 2 },
      PAS_standard: {
        Pas: 7,
        Cro: 6,
        Fre: 3,
        Vis: 5,
        Tec: 3,
        Tea: 5,
        Dec: 3,
        Cor: 2,
        Fir: 3,
        OtB: 4,
      },
      PAS_no_set_pieces: { Pas: 7, Vis: 5, Tec: 3, Tea: 5, Dec: 3, Fir: 3, OtB: 4 },
      PAS_no_off_ball: { Pas: 7, Cro: 6, Fre: 3, Vis: 5, Tec: 3, Tea: 5, Dec: 3, Fir: 3 },
      DRI: { Dri: 8, Fir: 5, Tec: 4, Fla: 3, Cmp: 3, OtB: 3 },
      DEF: {
        Mar: 8,
        Tck: 9,
        Hea: 8,
        Ant: 5,
        Cnt: 5,
        Pos: 6,
        Dec: 4,
        Cmp: 3,
        Bra: 6,
        Agg: 7,
        Wor: 6,
      },
      PHY: { Str: 11, Sta: 9, Nat: 8, Jum: 7, Bal: 6, Agg: 6, Bra: 5, Wor: 6 },
      GK: { Han: 18, Ref: 17, Cmd: 17, Aer: 20, '1v1': 8, Kic: 4, TRO: 4, Com: 3, Thr: 3, Ecc: 1 },
      DIV: { Aer: 10, Ref: 6, Agi: 5, '1v1': 6, Han: 6 },
      HAN: { Han: 11, Cmd: 9, Cmp: 3, Cnt: 3 },
      REF: { Ref: 9, Ant: 4, Cnt: 4, '1v1': 4 },
      KIC: { Kic: 8, Thr: 6, Tec: 5, Vis: 4, Pas: 3 },
      SPD: { Acc: 10, Pac: 10, TRO: 6 },
      POS: { Pos: 7, Cmd: 9, Ant: 4, Dec: 4, TRO: 4, Cnt: 3, Com: 3 },
    },
  },
  {
    id: 'preset-technical-maestro',
    name: 'Technical Maestro',
    description:
      'Leans on technique, first touch, dribbling, vision and passing precision; downweights physical attributes. Good for possession-based/technical scouting.',
    weights: {
      PAC: { Acc: 10, Pac: 9, Agi: 8 },
      SHO: { Fin: 9, Lon: 7, Pen: 5, Hea: 3, Cmp: 7, Tec: 8, Ant: 4, Dec: 5, Fla: 5 },
      PAS_standard: {
        Pas: 10,
        Cro: 7,
        Fre: 6,
        Vis: 9,
        Tec: 8,
        Tea: 3,
        Dec: 5,
        Cor: 5,
        Fir: 6,
        OtB: 3,
      },
      PAS_no_set_pieces: { Pas: 10, Vis: 9, Tec: 8, Tea: 3, Dec: 5, Fir: 6, OtB: 3 },
      PAS_no_off_ball: { Pas: 10, Cro: 7, Fre: 6, Vis: 9, Tec: 8, Tea: 3, Dec: 5, Fir: 6 },
      DRI: { Dri: 10, Fir: 9, Tec: 9, Fla: 8, Cmp: 5, OtB: 3 },
      DEF: {
        Mar: 7,
        Tck: 6,
        Hea: 3,
        Ant: 8,
        Cnt: 6,
        Pos: 8,
        Dec: 6,
        Cmp: 5,
        Bra: 3,
        Agg: 2,
        Wor: 3,
      },
      PHY: { Str: 4, Sta: 5, Nat: 4, Jum: 3, Bal: 6, Agg: 3, Bra: 3, Wor: 4 },
      GK: {
        Han: 18,
        Ref: 17,
        Cmd: 12,
        Aer: 10,
        '1v1': 10,
        Kic: 10,
        TRO: 6,
        Com: 4,
        Thr: 4,
        Ecc: 2,
      },
      DIV: { Aer: 5, Ref: 8, Agi: 8, '1v1': 8, Han: 6 },
      HAN: { Han: 9, Cmd: 5, Cmp: 7, Cnt: 5 },
      REF: { Ref: 9, Ant: 7, Cnt: 6, '1v1': 6 },
      KIC: { Kic: 11, Thr: 5, Tec: 7, Vis: 6, Pas: 5 },
      SPD: { Acc: 6, Pac: 6, TRO: 5 },
      POS: { Pos: 7, Cmd: 5, Ant: 7, Dec: 7, TRO: 3, Cnt: 5, Com: 3 },
    },
  },
  {
    id: 'preset-game-intelligence',
    name: 'Game Intelligence',
    description:
      'Leans on anticipation, decisions, concentration, positioning and composure -- the "reads the game well" attributes -- over raw physical or technical execution.',
    weights: {
      PAC: { Acc: 9, Pac: 9, Agi: 6 },
      SHO: { Fin: 6, Lon: 5, Pen: 4, Hea: 4, Cmp: 9, Tec: 4, Ant: 8, Dec: 8, Fla: 2 },
      PAS_standard: {
        Pas: 7,
        Cro: 4,
        Fre: 3,
        Vis: 9,
        Tec: 4,
        Tea: 7,
        Dec: 8,
        Cor: 2,
        Fir: 4,
        OtB: 6,
      },
      PAS_no_set_pieces: { Pas: 7, Vis: 9, Tec: 4, Tea: 7, Dec: 8, Fir: 4, OtB: 6 },
      PAS_no_off_ball: { Pas: 7, Cro: 4, Fre: 3, Vis: 9, Tec: 4, Tea: 7, Dec: 8, Fir: 4 },
      DRI: { Dri: 6, Fir: 6, Tec: 5, Fla: 3, Cmp: 6, OtB: 6 },
      DEF: {
        Mar: 7,
        Tck: 6,
        Hea: 4,
        Ant: 10,
        Cnt: 9,
        Pos: 10,
        Dec: 8,
        Cmp: 5,
        Bra: 4,
        Agg: 3,
        Wor: 4,
      },
      PHY: { Str: 4, Sta: 6, Nat: 5, Jum: 4, Bal: 4, Agg: 4, Bra: 5, Wor: 6 },
      GK: { Han: 16, Ref: 16, Cmd: 17, Aer: 12, '1v1': 9, Kic: 5, TRO: 7, Com: 6, Thr: 3, Ecc: 1 },
      DIV: { Aer: 6, Ref: 6, Agi: 5, '1v1': 6, Han: 5 },
      HAN: { Han: 8, Cmd: 8, Cmp: 7, Cnt: 6 },
      REF: { Ref: 8, Ant: 9, Cnt: 8, '1v1': 4 },
      KIC: { Kic: 6, Thr: 5, Tec: 4, Vis: 7, Pas: 5 },
      SPD: { Acc: 5, Pac: 5, TRO: 9 },
      POS: { Pos: 10, Cmd: 8, Ant: 9, Dec: 8, TRO: 6, Cnt: 6, Com: 4 },
    },
  },
  {
    id: 'preset-set-piece-specialist',
    name: 'Set-Piece Specialist',
    description:
      'Boosts free kicks, corners, penalties, crossing, heading and long shots wherever they appear. Good for squad-building around dead-ball situations.',
    weights: {
      PAC: { Acc: 12, Pac: 12, Agi: 5 },
      SHO: { Fin: 7, Lon: 10, Pen: 9, Hea: 8, Cmp: 6, Tec: 4, Ant: 3, Dec: 3, Fla: 2 },
      PAS_standard: {
        Pas: 6,
        Cro: 10,
        Fre: 9,
        Vis: 5,
        Tec: 4,
        Tea: 3,
        Dec: 3,
        Cor: 9,
        Fir: 3,
        OtB: 2,
      },
      PAS_no_set_pieces: { Pas: 8, Vis: 7, Tec: 5, Tea: 4, Dec: 4, Fir: 4, OtB: 3 },
      PAS_no_off_ball: { Pas: 6, Cro: 10, Fre: 9, Vis: 5, Tec: 4, Tea: 3, Dec: 3, Fir: 3 },
      DRI: { Dri: 8, Fir: 7, Tec: 6, Fla: 5, Cmp: 4, OtB: 3 },
      DEF: {
        Mar: 7,
        Tck: 6,
        Hea: 10,
        Ant: 6,
        Cnt: 5,
        Pos: 6,
        Dec: 4,
        Cmp: 4,
        Bra: 5,
        Agg: 4,
        Wor: 3,
      },
      PHY: { Str: 8, Sta: 6, Nat: 5, Jum: 9, Bal: 4, Agg: 5, Bra: 4, Wor: 3 },
      GK: { Han: 19, Ref: 18, Cmd: 17, Aer: 17, '1v1': 9, Kic: 5, TRO: 5, Com: 3, Thr: 3, Ecc: 1 },
      DIV: { Aer: 9, Ref: 7, Agi: 5, '1v1': 6, Han: 5 },
      HAN: { Han: 10, Cmd: 8, Cmp: 5, Cnt: 4 },
      REF: { Ref: 10, Ant: 6, Cnt: 5, '1v1': 5 },
      KIC: { Kic: 8, Thr: 6, Tec: 5, Vis: 4, Pas: 3 },
      SPD: { Acc: 8, Pac: 8, TRO: 6 },
      POS: { Pos: 8, Cmd: 8, Ant: 6, Dec: 5, TRO: 4, Cnt: 4, Com: 3 },
    },
  },
  {
    id: 'preset-wonderkid-radar',
    name: 'Wonderkid Radar',
    description:
      'Leans on athletic-upside attributes (Acceleration, Pace, Agility, Work Rate, Stamina) and downweights "polish" attributes that tend to develop later with experience.',
    weights: {
      PAC: { Acc: 15, Pac: 15, Agi: 8 },
      SHO: { Fin: 6, Lon: 5, Pen: 3, Hea: 4, Cmp: 3, Tec: 4, Ant: 2, Dec: 2, Fla: 4 },
      PAS_standard: {
        Pas: 6,
        Cro: 5,
        Fre: 2,
        Vis: 5,
        Tec: 4,
        Tea: 5,
        Dec: 2,
        Cor: 2,
        Fir: 4,
        OtB: 6,
      },
      PAS_no_set_pieces: { Pas: 6, Vis: 5, Tec: 4, Tea: 5, Dec: 2, Fir: 4, OtB: 6 },
      PAS_no_off_ball: { Pas: 6, Cro: 5, Fre: 2, Vis: 5, Tec: 4, Tea: 5, Dec: 2, Fir: 4 },
      DRI: { Dri: 10, Fir: 6, Tec: 5, Fla: 6, Cmp: 2, OtB: 5 },
      DEF: {
        Mar: 6,
        Tck: 6,
        Hea: 5,
        Ant: 4,
        Cnt: 4,
        Pos: 4,
        Dec: 3,
        Cmp: 2,
        Bra: 5,
        Agg: 5,
        Wor: 9,
      },
      PHY: { Str: 6, Sta: 10, Nat: 9, Jum: 6, Bal: 7, Agg: 5, Bra: 4, Wor: 9 },
      GK: { Han: 15, Ref: 18, Cmd: 10, Aer: 12, '1v1': 12, Kic: 5, TRO: 4, Com: 2, Thr: 3, Ecc: 2 },
      DIV: { Aer: 7, Ref: 9, Agi: 9, '1v1': 9, Han: 4 },
      HAN: { Han: 9, Cmd: 4, Cmp: 3, Cnt: 3 },
      REF: { Ref: 12, Ant: 4, Cnt: 4, '1v1': 6 },
      KIC: { Kic: 7, Thr: 6, Tec: 4, Vis: 3, Pas: 3 },
      SPD: { Acc: 10, Pac: 10, TRO: 4 },
      POS: { Pos: 5, Cmd: 4, Ant: 4, Dec: 4, TRO: 3, Cnt: 3, Com: 2 },
    },
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
    () =>
      profiles.value.find((p) => p.id === activeProfileId.value) ||
      PRESET_WEIGHT_PROFILES.find((p) => p.id === activeProfileId.value) ||
      null
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
      const isKnownProfile =
        storedActive &&
        (profiles.value.some((p) => p.id === storedActive) ||
          PRESET_WEIGHT_PROFILES.some((p) => p.id === storedActive))
      if (isKnownProfile) {
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
