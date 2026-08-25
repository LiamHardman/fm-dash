<!-- Card Generator: lets the user pick a card Tier + Rare finish, apply the
     same stat-boost formula TeamViewPage.vue uses for its Icon/Hero/MOTM
     cards, override any individual stat/Overall on top, optionally upload a
     custom player photo/club logo (client-side only, never persisted), and
     export the result as a PNG. Layout/behavior decided via a live prototype
     walkthrough — see .scratch/card-generator/issues/03-card-generator-dialog-ux-layout.md.

     Renders as a plain fixed-position overlay rather than a nested q-dialog
     — stacking a second Quasar dialog's own transition/backdrop machinery on
     top of the already-open PlayerDetailDialog proved fragile in practice
     (the trigger button ended up with a zero-size layout box, clicks got
     silently swallowed). A plain overlay sidesteps that entirely.

     Teleported to <body> deliberately: this component is otherwise mounted
     wherever PlayerDetailDialog sits in the normal component tree, which is
     NOT where Quasar teleports the outer dialog's own content (it teleports
     to <body> too). Two elements at the same z-index resolve stacking ties
     by DOM position among *actual siblings* — comparing a normally-mounted
     element against a teleported one doesn't reliably win that tie, which is
     exactly what caused this panel to render behind the still-open
     PlayerDetailDialog. Teleporting here too makes both real siblings under
     <body>, where "mounted later paints on top" is a real guarantee. -->
<template>
  <Teleport to="body">
  <div v-if="show" class="cardgen-backdrop" @click.self="$emit('close')">
    <div class="cardgen-panel">
      <!-- Dialog chrome: same header convention (icon/title/close) as
           PlayerDetailDialog/UpgradeFinderDialog, in normal flow above the content.
           Colors here are intentionally fixed-dark (not app-theme tokens) — this
           panel is a "card studio" surface, matching the always-dark treatment of
           WhoToSignDialog's results dossier. -->
      <div class="dialog-chrome">
        <div class="dialog-chrome__header">
          <q-icon name="badge" class="dialog-chrome__icon" />
          <div class="dialog-chrome__title">Card Generator</div>
          <q-space />
          <div class="dialog-chrome__actions">
            <q-btn icon="close" flat round dense class="dialog-chrome__close" @click="$emit('close')" />
          </div>
        </div>
      </div>

      <div class="cardgen-body">
        <div class="preview-pane">
          <PlayerCards
            ref="cardRef"
            :player="state.previewPlayer.value"
            :card-design-override="state.cardDesignOverride.value"
            :player-face-url="state.faceObjectUrl.value"
            :club-image-url="state.logoObjectUrl.value"
          />
          <div class="preview-caption">Live preview updates as you edit</div>
        </div>

        <div class="controls-pane">
          <section class="control-section">
            <div class="section-title">Card Type</div>
            <div class="row-inline">
              <q-select
                dense
                outlined
                :model-value="state.selectedTierKey.value"
                @update:model-value="(v) => (state.selectedTierKey.value = v)"
                :options="tierOptions"
                emit-value
                map-options
                label="Tier"
                class="tier-select"
              />
              <q-toggle
                v-if="state.selectedTier.value.rareToggle"
                :model-value="state.rare.value"
                @update:model-value="(v) => (state.rare.value = v)"
                label="Rare"
                color="amber"
              />
            </div>
          </section>

          <section class="control-section">
            <div class="section-title">
              Stats
              <q-btn flat dense size="sm" icon="restart_alt" label="Reset to tier baseline" @click="state.resetToTierBaseline()" />
            </div>
            <div class="stat-grid">
              <div class="stat-field overall-field">
                <label>OVR</label>
                <q-input
                  dense
                  outlined
                  type="number"
                  :model-value="state.editableOverall.value"
                  @update:model-value="(v) => state.setOverall(v)"
                />
              </div>
              <div v-for="key in state.statKeys.value" :key="key" class="stat-field">
                <label>{{ state.statLabels.value[key] }}</label>
                <q-input
                  dense
                  outlined
                  type="number"
                  :model-value="state.editableStats[key]"
                  @update:model-value="(v) => state.setStat(key, v)"
                />
              </div>
            </div>
          </section>

          <section class="control-section">
            <div class="section-title">Images</div>
            <div class="upload-row">
              <q-btn outline dense icon="face" label="Upload player photo" @click="$refs.faceInput.click()" />
              <input ref="faceInput" type="file" accept="image/*" style="display: none" @change="onFace" />
              <span v-if="state.faceObjectUrl.value" class="upload-status">Custom photo set</span>
            </div>
            <div class="upload-row">
              <q-btn outline dense icon="shield" label="Upload club logo" @click="$refs.logoInput.click()" />
              <input ref="logoInput" type="file" accept="image/*" style="display: none" @change="onLogo" />
              <span v-if="state.logoObjectUrl.value" class="upload-status">Custom logo set</span>
            </div>
          </section>

          <section class="control-section export-section">
            <q-btn
              color="primary"
              unelevated
              icon="download"
              label="Export as PNG"
              :loading="state.exporting.value"
              @click="state.exportAsPng(cardRef?.$el, props.player?.name)"
            />
          </section>
        </div>
      </div>
    </div>
  </div>
  </Teleport>
</template>

<script>
import { computed, defineComponent, onBeforeUnmount, onMounted, ref } from 'vue'
import { useCardGenerator } from '../composables/useCardGenerator'
import PlayerCards from './PlayerCards.vue'

export default defineComponent({
  name: 'CardGeneratorDialog',
  components: { PlayerCards },
  props: {
    show: { type: Boolean, default: false },
    player: { type: Object, required: true },
  },
  emits: ['close'],
  setup(props, { emit }) {
    const playerRef = computed(() => props.player)
    const state = useCardGenerator(playerRef)
    const cardRef = ref(null)

    const tierOptions = state.tiers.map((t) => ({ label: t.label, value: t.key }))

    function onFace(e) {
      state.onFaceFileSelected(e.target.files?.[0] || null)
    }
    function onLogo(e) {
      state.onLogoFileSelected(e.target.files?.[0] || null)
    }

    function onKeydown(e) {
      const tag = document.activeElement?.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || document.activeElement?.isContentEditable) return
      if (e.key === 'Escape') emit('close')
    }

    onMounted(() => window.addEventListener('keydown', onKeydown))
    onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

    return {
      props,
      state,
      cardRef,
      tierOptions,
      onFace,
      onLogo,
    }
  },
})
</script>

<style scoped>
.cardgen-backdrop {
  position: fixed;
  inset: 0;
  /* Matches Quasar's own dialog/menu z-index baseline (6000) rather than
     exceeding it. Both the already-open outer PlayerDetailDialog and any
     Quasar popup (q-select's dropdown, etc.) opened from *inside* this
     panel sit on that same baseline, so with equal z-index, plain DOM order
     decides stacking: this backdrop — mounted after the outer dialog —
     paints above it, and a freshly-opened popup — mounted even later —
     paints above this panel in turn. Overshooting this value sits above
     Quasar's own popups too, which silently swallows clicks on any dropdown
     opened from within the panel — confirmed the hard way while building
     this. The outer dialog's pointer-events are also disabled while this is
     open (see showCardGenerator in PlayerDetailDialog.vue), as a
     belt-and-suspenders guard against any residual click leakage regardless
     of exact stacking. */
  z-index: 6000;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(3px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}
.cardgen-panel {
  width: 95vw;
  max-width: 1400px;
  height: 90vh;
  max-height: 900px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  border-radius: var(--radius-md);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

/* Dialog chrome, restyled to the panel's fixed-dark "studio" palette rather
   than the app's light/dark-flipping tokens (see comment in the template). */
.dialog-chrome {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}
.dialog-chrome__header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 16px 16px 12px;
}
.dialog-chrome__icon {
  font-size: 1.3rem;
  color: #fff;
  flex-shrink: 0;
}
.dialog-chrome__title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
}
.dialog-chrome__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.dialog-chrome__close {
  color: #fff;
  transition: transform 0.15s ease;
}
.dialog-chrome__close:hover {
  transform: scale(1.08);
}
.cardgen-body {
  flex: 1;
  overflow: auto;
  color: #fff;
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 24px;
  padding: 16px;
  align-items: start;
}
.preview-pane {
  position: sticky;
  top: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.preview-caption {
  font-size: 12px;
  opacity: 0.6;
}
.controls-pane {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.control-section {
  border: 1px solid rgba(128, 128, 128, 0.25);
  border-radius: 8px;
  padding: 12px 16px;
}
.section-title {
  font-weight: 600;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.row-inline {
  display: flex;
  align-items: center;
  gap: 16px;
}
.tier-select {
  min-width: 220px;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}
.stat-field label {
  font-size: 11px;
  opacity: 0.7;
  display: block;
  margin-bottom: 2px;
}
.overall-field {
  grid-column: span 2;
}
.upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.upload-status {
  font-size: 12px;
  opacity: 0.7;
}
.export-section {
  display: flex;
  justify-content: flex-end;
}
</style>
