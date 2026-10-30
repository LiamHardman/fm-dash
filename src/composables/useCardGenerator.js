// State/logic for the Card Generator dialog (CardGeneratorDialog.vue).
// Selecting a Tier auto-populates stats+Overall via the boost formula; the
// user can freely edit any field afterward as an override; switching Tier
// re-runs the formula and resets to a fresh baseline (discards prior manual
// edits). Overall is a fully independent editable field — editing individual
// stats never auto-recomputes it. Uploaded image/logo are ephemeral object
// URLs, never sent to the backend. See .scratch/card-generator/map.md for
// the full set of decisions this implements.
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { applyTierBoost, CARD_TIERS, statKeysFor, statLabelsFor } from '../utils/cardGeneratorTiers'

export function useCardGenerator(player) {
  const selectedTierKey = ref('gold')
  const rare = ref(false)
  const baseEdition = ref('gold')
  const playstylePlus = ref(player.value?.playstylePlus || '')
  const portraitMode = ref('cutout')
  const exportError = ref('')
  const editableStats = reactive({})
  const editableOverall = ref(0)
  const faceObjectUrl = ref(null)
  const logoObjectUrl = ref(null)
  const exporting = ref(false)

  const selectedTier = computed(
    () => CARD_TIERS.find((t) => t.key === selectedTierKey.value) || CARD_TIERS[0]
  )
  const statKeys = computed(() => statKeysFor(player.value))
  const statLabels = computed(() => statLabelsFor(player.value))

  function resetToTierBaseline() {
    for (const key of Object.keys(editableStats)) delete editableStats[key]
    const { overall, ...stats } = applyTierBoost(player.value, selectedTier.value)
    editableOverall.value = overall
    for (const key of statKeys.value) {
      editableStats[key] = stats[key]
    }
  }

  watch([selectedTierKey, player], resetToTierBaseline, { immediate: true })
  watch(selectedTierKey, () => {
    if (!selectedTier.value.rareToggle) rare.value = false
  })
  watch(player, () => {
    onFaceFileSelected(null)
    onLogoFileSelected(null)
    playstylePlus.value = player.value?.playstylePlus || ''
    exportError.value = ''
  })
  onBeforeUnmount(() => {
    if (faceObjectUrl.value) URL.revokeObjectURL(faceObjectUrl.value)
    if (logoObjectUrl.value) URL.revokeObjectURL(logoObjectUrl.value)
  })

  function setStat(key, value) {
    editableStats[key] = Math.max(0, Math.min(99, Number(value) || 0))
  }

  function setOverall(value) {
    editableOverall.value = Math.max(0, Math.min(99, Number(value) || 0))
  }

  function onFaceFileSelected(file) {
    if (faceObjectUrl.value) URL.revokeObjectURL(faceObjectUrl.value)
    faceObjectUrl.value = file ? URL.createObjectURL(file) : null
  }

  function onLogoFileSelected(file) {
    if (logoObjectUrl.value) URL.revokeObjectURL(logoObjectUrl.value)
    logoObjectUrl.value = file ? URL.createObjectURL(file) : null
  }

  const previewPlayer = computed(() => ({
    ...player.value,
    Overall: editableOverall.value,
    overall: editableOverall.value,
    ...editableStats,
  }))

  async function exportAsPng(card, filenameHint) {
    exportError.value = ''
    exporting.value = true
    try {
      if (!card?.exportPng) throw new Error('The card is not ready to export.')
      const blob = await card.exportPng(2)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${(filenameHint || 'card').replace(/[^a-z0-9-_]+/gi, '_')}.png`
      document.body.appendChild(a)
      a.click()
      a.remove()
      setTimeout(() => URL.revokeObjectURL(url), 10000)
    } catch (error) {
      exportError.value = error.message || 'Unable to export this card.'
    } finally {
      exporting.value = false
    }
  }

  return {
    tiers: CARD_TIERS,
    selectedTierKey,
    selectedTier,
    rare,
    statKeys,
    statLabels,
    editableStats,
    editableOverall,
    setStat,
    setOverall,
    resetToTierBaseline,
    onFaceFileSelected,
    onLogoFileSelected,
    faceObjectUrl,
    logoObjectUrl,
    previewPlayer,
    baseEdition,
    playstylePlus,
    portraitMode,
    exportError,
    exporting,
    exportAsPng,
  }
}
