// State/logic for the Card Generator dialog (CardGeneratorDialog.vue).
// Selecting a Tier auto-populates stats+Overall via the boost formula; the
// user can freely edit any field afterward as an override; switching Tier
// re-runs the formula and resets to a fresh baseline (discards prior manual
// edits). Overall is a fully independent editable field — editing individual
// stats never auto-recomputes it. Uploaded image/logo are ephemeral object
// URLs, never sent to the backend. See .scratch/card-generator/map.md for
// the full set of decisions this implements.
import { computed, reactive, ref, watch } from 'vue'
import { applyTierBoost, CARD_TIERS, statKeysFor, statLabelsFor } from '../utils/cardGeneratorTiers'

export function useCardGenerator(player) {
  const selectedTierKey = ref('gold')
  const rare = ref(false)
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
    const { overall, ...stats } = applyTierBoost(player.value, selectedTier.value)
    editableOverall.value = overall
    for (const key of statKeys.value) {
      editableStats[key] = stats[key]
    }
  }

  watch(selectedTierKey, resetToTierBaseline, { immediate: true })

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

  const cardDesignOverride = computed(() => {
    if (selectedTier.value.design) return selectedTier.value.design
    if (!selectedTier.value.rareToggle) return null
    // Bronze/Silver/Gold don't carry their own design override — they mirror
    // PlayerCards.vue's own designByTypeAndRarity lookup so the Rare toggle
    // here maps to the same finishes it would auto-pick from real stats.
    const rareDesigns = {
      bronze: 'card-design--amber-facets',
      silver: 'card-design--polished-silver',
      gold: 'card-design--confetti-foil',
    }
    const nonRareDesigns = {
      bronze: 'card-design--aged-plate',
      silver: 'card-design--pale-geometric',
      gold: 'card-design--antique-foil',
    }
    return (rare.value ? rareDesigns : nonRareDesigns)[selectedTier.value.key] || null
  })

  // Per .scratch/card-generator/issues/02-png-export-html2canvas-approach.md:
  // html2canvas-pro over vanilla html2canvas, useCORS for the flagcdn.com
  // nation-flag fallback, explicit scale rather than devicePixelRatio for
  // deterministic output quality.
  async function exportAsPng(cardEl, filenameHint) {
    if (!cardEl) return
    exporting.value = true
    try {
      const { default: html2canvas } = await import('html2canvas-pro')
      const canvas = await html2canvas(cardEl, { scale: 3, useCORS: true, backgroundColor: null })
      const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/png'))
      if (!blob) return
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${(filenameHint || 'card').replace(/[^a-z0-9-_]+/gi, '_')}.png`
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
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
    cardDesignOverride,
    exporting,
    exportAsPng,
  }
}
