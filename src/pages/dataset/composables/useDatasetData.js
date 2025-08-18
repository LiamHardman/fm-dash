import { computed, ref } from 'vue'
import { usePlayerStore } from '../../../stores/playerStore'
import {
  allRawFmAttributeKeys,
  formatFilterKeyPrefix,
  parseTransferValueRange,
} from '../utils/datasetUtils'

export function useDatasetData() {
  const playerStore = usePlayerStore()

  const pageLoading = ref(true)
  const pageLoadingError = ref('')

  const allPlayersData = computed(() => playerStore.allPlayers)
  const detectedCurrencySymbol = computed(() => playerStore.detectedCurrencySymbol)
  const currentDatasetId = computed(() => playerStore.currentDatasetId)
  const loading = computed(() => playerStore.loading)
  const uniqueClubs = computed(() => playerStore.uniqueClubs)
  const uniqueNationalities = computed(() => playerStore.uniqueNationalities)
  const uniqueDivisions = computed(() => playerStore.uniqueDivisions || [])
  const uniqueMediaHandlings = computed(() => playerStore.uniqueMediaHandlings)
  const uniquePersonalities = computed(() => playerStore.uniquePersonalities)

  const transferValueRangeForFilters = computed(
    () => playerStore.currentDatasetTransferValueRange || { min: 0, max: 100000000 }
  )
  const initialDatasetTransferValueRangeForFilters = computed(
    () => playerStore.initialDatasetTransferValueRange || { min: 0, max: 100000000 }
  )
  const salaryRangeForFilters = computed(() => playerStore.salaryRange || { min: 0, max: 1000000 })

  const currentFilters = ref({
    name: '',
    position: '',
    role: '',
    club: '',
    nationality: [],
    mediaHandling: [],
    personality: [],
    ageRange: { min: playerStore.AGE_SLIDER_MIN_DEFAULT, max: playerStore.AGE_SLIDER_MAX_DEFAULT },
    transferValueRangeLocal: { min: 0, max: 100000000 },
    maxSalary: 1000000,
    minOverall: 0,
    minPAC: 0,
    minSHO: 0,
    minPAS: 0,
    minDRI: 0,
    minDEF: 0,
    minPHY: 0,
    minGK: 0,
    minDIV: 0,
    minHAN: 0,
    minREF: 0,
    minKIC: 0,
    minSPD: 0,
    minPOS: 0,
    continentNationalities: [],
  })

  for (const attrKey of allRawFmAttributeKeys) {
    currentFilters.value[`min${formatFilterKeyPrefix(attrKey)}`] = 0
  }

  const filteredPlayers = computed(() => {
    const players = allPlayersData.value || []
    if (players.length === 0) return []

    return players.filter((player) => {
      if (
        currentFilters.value.name &&
        !player.name.toLowerCase().includes(currentFilters.value.name.toLowerCase())
      )
        return false
      if (currentFilters.value.club && player.club !== currentFilters.value.club) return false

      if (currentFilters.value.position && currentFilters.value.position.length > 0) {
        const playerPositions = player.short_positions || player.shortPositions || []
        const playerPosition = player.position || ''
        const hasPosition =
          currentFilters.value.position.some((pos) => playerPositions.includes(pos)) ||
          (playerPosition &&
            currentFilters.value.position.some((pos) => playerPosition.includes(pos)))
        if (!hasPosition) return false
      }

      if (currentFilters.value.role && currentFilters.value.role.length > 0) {
        const playerRoles = (player.best_roles && player.best_roles.slice(0, 3)) || []
        if (!currentFilters.value.role.some((role) => playerRoles.includes(role))) return false
      }

      if (currentFilters.value.nationality && currentFilters.value.nationality.length > 0) {
        const playerNationality = player.nationality || player.Nationality || ''
        if (!currentFilters.value.nationality.includes(playerNationality)) return false
      }

      if (currentFilters.value.mediaHandling && currentFilters.value.mediaHandling.length > 0) {
        const handling = player.media_handling || player.mediaHandling || ''
        if (!currentFilters.value.mediaHandling.includes(handling)) return false
      }

      if (currentFilters.value.personality && currentFilters.value.personality.length > 0) {
        const personality = player.personality || ''
        if (!currentFilters.value.personality.includes(personality)) return false
      }

      const age = Number.parseInt(player.age || player.Age || 0)
      if (age < currentFilters.value.ageRange.min || age > currentFilters.value.ageRange.max)
        return false

      if (player.salary_numeric) {
        const salary = Number.parseInt(player.salary_numeric)
        if (salary > currentFilters.value.maxSalary) return false
      }

      const transferValueRange = parseTransferValueRange(player.transfer_value)
      const minValue = transferValueRange.min
      const maxValue = transferValueRange.max
      const filterMin = currentFilters.value.transferValueRangeLocal.min
      const filterMax = currentFilters.value.transferValueRangeLocal.max
      if ((minValue !== 0 || maxValue !== 0) && (maxValue < filterMin || minValue > filterMax))
        return false

      const statValueOrZero = (obj, key) => {
        const raw = (obj && obj[key]) ?? null
        const num = Number.parseFloat(String(raw).replace(/,/g, ''))
        return Number.isNaN(num) ? 0 : num
      }

      const overall = statValueOrZero(player, 'Overall')
      if (overall < currentFilters.value.minOverall) return false

      const isGK = (player.short_positions || player.shortPositions || []).includes('GK')
      if (isGK) {
        const keys = ['GK', 'DIV', 'HAN', 'REF', 'KIC', 'SPD', 'POS']
        const mins = ['minGK', 'minDIV', 'minHAN', 'minREF', 'minKIC', 'minSPD', 'minPOS']
        for (let i = 0; i < keys.length; i++) {
          if (statValueOrZero(player, keys[i]) < currentFilters.value[mins[i]]) return false
        }
        return true
      }

      const fifaKeys = ['PAC', 'SHO', 'PAS', 'DRI', 'DEF', 'PHY']
      const mins = ['minPAC', 'minSHO', 'minPAS', 'minDRI', 'minDEF', 'minPHY']
      for (let i = 0; i < fifaKeys.length; i++) {
        if (statValueOrZero(player, fifaKeys[i]) < currentFilters.value[mins[i]]) return false
      }

      return true
    })
  })

  return {
    pageLoading,
    pageLoadingError,
    allPlayersData,
    detectedCurrencySymbol,
    currentDatasetId,
    loading,
    uniqueClubs,
    uniqueNationalities,
    uniqueDivisions,
    uniqueMediaHandlings,
    uniquePersonalities,
    transferValueRangeForFilters,
    initialDatasetTransferValueRangeForFilters,
    salaryRangeForFilters,
    currentFilters,
    filteredPlayers,
  }
}
