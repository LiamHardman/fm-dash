<template>
    <article
        ref="cardRef"
        class="hover-card"
        :class="[`hover-card--${card.tier}`, `hover-card--effect-${card.effect}`]"
        :style="cardStyle"
        tabindex="0"
        @pointermove="handlePointerMove"
        @pointerleave="resetPointer"
        @pointercancel="resetPointer"
        @blur="resetPointer"
    >
        <div class="hover-card__plane hover-card__plane--base"></div>
        <div class="hover-card__plane hover-card__plane--foil"></div>
        <div class="hover-card__plane hover-card__plane--flare"></div>
        <div class="hover-card__plane hover-card__plane--crest" aria-hidden="true">
            {{ card.crest }}
        </div>
        <div class="hover-card__plane hover-card__plane--strip" aria-hidden="true">
            <span>FM24 AUTHENTIC</span>
            <span>{{ card.club }}</span>
            <span>{{ card.series }}</span>
        </div>

        <div class="hover-card__content">
            <header class="hover-card__header">
                <div class="hover-card__rating-block">
                    <div class="hover-card__rating">{{ card.overall }}</div>
                    <div class="hover-card__position">{{ card.position }}</div>
                </div>
                <div class="hover-card__identity">
                    <div class="hover-card__series">{{ card.series }}</div>
                    <div class="hover-card__name">{{ card.name }}</div>
                    <div class="hover-card__vitals">{{ card.vitals }}</div>
                </div>
            </header>

            <section class="hover-card__art" aria-label="Player identity">
                <div class="hover-card__marks">
                    <div class="hover-card__flag" aria-label="Example nation flag">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                    <div class="hover-card__club" :aria-label="`${card.club} crest mark`">
                        {{ card.crest }}
                    </div>
                </div>

                <div class="hover-card__portrait" aria-label="Example player portrait">
                    <q-icon name="person" size="124px" />
                </div>
            </section>

            <footer class="hover-card__footer">
                <div class="hover-card__stats">
                    <div
                        v-for="stat in card.stats"
                        :key="stat.label"
                        class="hover-card__stat"
                    >
                        <span>{{ stat.value }}</span>
                        {{ stat.label }}
                    </div>
                </div>
            </footer>
        </div>
    </article>
</template>

<script>
import { computed, defineComponent, ref } from 'vue'

const resetPointerProperties = (element) => {
  if (!element) return

  const defaults = {
    '--pointer-x': '50%',
    '--pointer-y': '45%',
    '--shift-x': '0px',
    '--shift-y': '0px',
    '--surface-x': '0px',
    '--surface-y': '0px',
    '--surface-soft-x': '0px',
    '--surface-soft-y': '0px',
    '--surface-reverse-x': '0px',
    '--surface-reverse-y': '0px',
    '--surface-hard-x': '0px',
    '--surface-hard-y': '0px',
    '--tilt-x': '0deg',
    '--tilt-y': '0deg',
    '--glare-angle': '128deg',
  }

  Object.entries(defaults).forEach(([key, value]) => {
    element.style.setProperty(key, value)
  })
}

export default defineComponent({
  name: 'CardHoverRenderer',
  props: {
    card: {
      type: Object,
      required: true,
    },
  },
  setup(props) {
    const cardRef = ref(null)

    const cardStyle = computed(() => ({
      '--card-bg': props.card.tokens.bg,
      '--card-surface': props.card.tokens.surface,
      '--card-surface-strong': props.card.tokens.surfaceStrong,
      '--card-border': props.card.tokens.border,
      '--card-edge': props.card.tokens.edge,
      '--card-accent': props.card.tokens.accent,
      '--card-accent-strong': props.card.tokens.accentStrong,
      '--card-text': props.card.tokens.text,
      '--card-muted': props.card.tokens.muted,
      '--card-stat': props.card.tokens.stat,
      '--card-shadow': props.card.tokens.shadow,
    }))

    const handlePointerMove = (event) => {
      const card = cardRef.value
      if (!card) return

      const bounds = card.getBoundingClientRect()
      const rawX = (event.clientX - bounds.left) / bounds.width
      const rawY = (event.clientY - bounds.top) / bounds.height
      const pointerX = Math.min(Math.max(rawX, 0), 1)
      const pointerY = Math.min(Math.max(rawY, 0), 1)
      const centeredX = pointerX - 0.5
      const centeredY = pointerY - 0.5
      const surfaceX = (0.5 - pointerX) * 54
      const surfaceY = (0.5 - pointerY) * 44
      const glareAngle = 90 + Math.atan2(centeredY, centeredX) * (180 / Math.PI)

      card.style.setProperty('--pointer-x', `${(pointerX * 100).toFixed(2)}%`)
      card.style.setProperty('--pointer-y', `${(pointerY * 100).toFixed(2)}%`)
      card.style.setProperty('--shift-x', `${(centeredX * 24).toFixed(2)}px`)
      card.style.setProperty('--shift-y', `${(centeredY * 18).toFixed(2)}px`)
      card.style.setProperty('--surface-x', `${surfaceX.toFixed(2)}px`)
      card.style.setProperty('--surface-y', `${surfaceY.toFixed(2)}px`)
      card.style.setProperty('--surface-soft-x', `${(-surfaceX * 0.25).toFixed(2)}px`)
      card.style.setProperty('--surface-soft-y', `${(-surfaceY * 0.25).toFixed(2)}px`)
      card.style.setProperty('--surface-reverse-x', `${(-surfaceX * 0.45).toFixed(2)}px`)
      card.style.setProperty('--surface-reverse-y', `${(-surfaceY * 0.45).toFixed(2)}px`)
      card.style.setProperty('--surface-hard-x', `${(-surfaceX * 0.72).toFixed(2)}px`)
      card.style.setProperty('--surface-hard-y', `${(-surfaceY * 0.72).toFixed(2)}px`)
      card.style.setProperty('--tilt-x', `${(-centeredY * 20).toFixed(2)}deg`)
      card.style.setProperty('--tilt-y', `${(centeredX * 24).toFixed(2)}deg`)
      card.style.setProperty('--glare-angle', `${glareAngle.toFixed(2)}deg`)
    }

    const resetPointer = () => resetPointerProperties(cardRef.value)

    return {
      cardRef,
      cardStyle,
      handlePointerMove,
      resetPointer,
    }
  },
})
</script>

<style lang="scss" scoped>
.hover-card {
    --pointer-x: 50%;
    --pointer-y: 45%;
    --shift-x: 0px;
    --shift-y: 0px;
    --surface-x: 0px;
    --surface-y: 0px;
    --surface-soft-x: 0px;
    --surface-soft-y: 0px;
    --surface-reverse-x: 0px;
    --surface-reverse-y: 0px;
    --surface-hard-x: 0px;
    --surface-hard-y: 0px;
    --tilt-x: 0deg;
    --tilt-y: 0deg;
    --glare-angle: 128deg;

    width: min(100%, 300px);
    aspect-ratio: 2 / 3;
    position: relative;
    overflow: hidden;
    color: var(--card-text);
    border: 2px solid var(--card-border);
    border-radius: 14px;
    background:
        radial-gradient(circle at 50% 8%, rgba(255, 255, 255, 0.16), transparent 8rem),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 58%, #080707 100%);
    box-shadow:
        0 24px 54px rgba(0, 0, 0, 0.42),
        inset 0 0 0 1px rgba(255, 255, 255, 0.18),
        inset 0 0 0 7px rgba(0, 0, 0, 0.16),
        0 0 24px var(--card-shadow);
    cursor: pointer;
    isolation: isolate;
    transform: perspective(960px) translate3d(0, 0, 0) rotateX(0deg) rotateY(0deg);
    transform-style: preserve-3d;
    transition:
        transform 180ms ease,
        box-shadow 180ms ease;
    will-change: transform;

    &:hover,
    &:focus-visible {
        transform:
            perspective(960px)
            translate3d(var(--shift-x), calc(-10px + var(--shift-y)), 0)
            rotateX(var(--tilt-x))
            rotateY(var(--tilt-y));
        box-shadow:
            0 36px 82px rgba(0, 0, 0, 0.54),
            inset 0 0 0 1px rgba(255, 255, 255, 0.28),
            inset 0 0 0 7px rgba(0, 0, 0, 0.12),
            0 0 42px var(--card-shadow);
    }

    &::before {
        content: "";
        position: absolute;
        inset: 10px;
        z-index: 4;
        border: 1px solid color-mix(in srgb, var(--card-edge) 70%, transparent);
        border-radius: 10px;
        box-shadow:
            inset 0 0 0 1px rgba(255, 255, 255, 0.14),
            0 0 18px color-mix(in srgb, var(--card-accent) 24%, transparent);
        pointer-events: none;
    }

    &::after {
        content: "";
        position: absolute;
        left: -18%;
        right: -18%;
        top: 67%;
        z-index: 5;
        height: 3px;
        background: linear-gradient(90deg, transparent, var(--card-edge), var(--card-accent), transparent);
        box-shadow: 0 0 16px color-mix(in srgb, var(--card-accent) 52%, transparent);
        opacity: 0.88;
        pointer-events: none;
        transform: rotate(-2deg) translateZ(22px);
    }
}

.hover-card__plane {
    position: absolute;
    inset: 0;
    pointer-events: none;
}

.hover-card__plane--base {
    z-index: 1;
    background:
        radial-gradient(ellipse at 70% 26%, rgba(255, 255, 255, 0.12), transparent 11rem),
        repeating-linear-gradient(116deg, transparent 0 18px, rgba(255, 255, 255, 0.045) 18px 19px),
        linear-gradient(var(--glare-angle), transparent 0 28%, rgba(255, 255, 255, 0.14) 42%, transparent 58%);
    opacity: 0.75;
    transform: translate3d(var(--surface-soft-x), var(--surface-soft-y), 18px);
}

.hover-card__plane--foil,
.hover-card__plane--flare {
    z-index: 2;
    opacity: 0;
    mix-blend-mode: screen;
    transition:
        opacity 160ms ease,
        transform 160ms ease;
}

.hover-card__plane--foil {
    transform: translate3d(var(--surface-x), var(--surface-y), 34px);
}

.hover-card__plane--flare {
    z-index: 6;
    border-radius: 12px;
    transform: translate3d(var(--surface-hard-x), var(--surface-hard-y), 64px);
}

.hover-card__plane--crest {
    left: 8%;
    top: 13%;
    z-index: 2;
    display: grid;
    place-items: center;
    color: color-mix(in srgb, var(--card-edge) 36%, transparent);
    font-size: 10rem;
    font-weight: 900;
    line-height: 1;
    opacity: 0.16;
    text-shadow:
        -1px -1px 0 rgba(255, 255, 255, 0.2),
        1px 1px 0 rgba(0, 0, 0, 0.35);
    transform: translate3d(var(--surface-reverse-x), var(--surface-reverse-y), 24px);
}

.hover-card__plane--strip {
    z-index: 7;
    left: auto;
    width: 28px;
    display: grid;
    align-content: center;
    gap: 1rem;
    padding: 0.8rem 0;
    color: rgba(255, 255, 255, 0.76);
    background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.34), transparent 26% 56%, rgba(255, 255, 255, 0.2)),
        repeating-linear-gradient(180deg, rgba(0, 0, 0, 0.34) 0 9px, rgba(255, 255, 255, 0.18) 9px 18px);
    border-left: 1px solid rgba(255, 255, 255, 0.24);
    opacity: 0;
    mix-blend-mode: screen;
    transform: translate3d(var(--surface-hard-x), 0, 70px);

    span {
        display: block;
        font-size: 0.48rem;
        font-weight: 900;
        letter-spacing: 0.08em;
        text-align: center;
        text-transform: uppercase;
        writing-mode: vertical-rl;
    }
}

.hover-card:hover .hover-card__plane--foil,
.hover-card:hover .hover-card__plane--flare,
.hover-card:focus-visible .hover-card__plane--foil,
.hover-card:focus-visible .hover-card__plane--flare {
    opacity: 1;
}

.hover-card__content {
    position: relative;
    z-index: 8;
    height: 100%;
    display: flex;
    flex-direction: column;
    transform: translateZ(84px);
}

.hover-card__header {
    min-height: 30%;
    padding: 1.1rem 1.2rem 0;
    display: grid;
    grid-template-columns: 4rem minmax(0, 1fr);
    gap: 0.72rem;
}

.hover-card__rating-block {
    position: relative;

    &::before {
        content: "";
        position: absolute;
        inset: -0.28rem -0.18rem 0 -0.42rem;
        border-top: 1px solid color-mix(in srgb, var(--card-edge) 70%, transparent);
        border-left: 2px solid var(--card-edge);
        border-radius: 8px 0 0 0;
    }
}

.hover-card__rating {
    font-size: 3.55rem;
    font-weight: 950;
    line-height: 0.86;
    text-shadow:
        0 2px 10px rgba(0, 0, 0, 0.5),
        0 0 16px color-mix(in srgb, var(--card-accent) 24%, transparent);
}

.hover-card__position {
    margin-top: 0.26rem;
    color: var(--card-muted);
    font-size: 1.1rem;
    font-weight: 900;
}

.hover-card__identity {
    min-width: 0;
    padding-top: 0.2rem;
    text-align: center;
}

.hover-card__series {
    color: var(--card-muted);
    font-size: 0.58rem;
    font-weight: 900;
    letter-spacing: 0.12em;
    text-transform: uppercase;
}

.hover-card__name {
    margin-top: 0.16rem;
    overflow: hidden;
    color: var(--card-text);
    font-size: 1.55rem;
    font-weight: 950;
    line-height: 1.08;
    text-overflow: ellipsis;
    text-transform: uppercase;
    white-space: nowrap;
}

.hover-card__vitals {
    display: inline-block;
    max-width: 100%;
    margin-top: 0.52rem;
    padding: 0.22rem 0.5rem;
    overflow: hidden;
    color: #17110a;
    background: linear-gradient(180deg, var(--card-accent), var(--card-edge));
    border: 1px solid rgba(255, 255, 255, 0.34);
    border-radius: 4px;
    font-size: 0.72rem;
    font-weight: 900;
    line-height: 1.12;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.hover-card__art {
    position: relative;
    flex: 1;
    min-height: 0;
}

.hover-card__marks {
    position: absolute;
    top: 0.3rem;
    left: 1.35rem;
    z-index: 2;
    display: grid;
    gap: 0.55rem;
}

.hover-card__flag,
.hover-card__club {
    width: 42px;
    height: 30px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.24);
    border-radius: 4px;
    background: rgba(12, 10, 9, 0.62);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.28);
}

.hover-card__flag {
    display: grid;
    grid-template-columns: repeat(3, 1fr);

    span:nth-child(1) {
        background: #15458a;
    }

    span:nth-child(2) {
        background: #f4f0e8;
    }

    span:nth-child(3) {
        background: #c22b30;
    }
}

.hover-card__club {
    height: 42px;
    display: grid;
    place-items: center;
    color: var(--card-accent);
    font-weight: 950;
}

.hover-card__portrait {
    position: absolute;
    left: calc(50% + 34px);
    top: -0.2rem;
    width: 58%;
    height: 77%;
    display: grid;
    place-items: center;
    color: color-mix(in srgb, var(--card-text) 78%, transparent);
    background:
        radial-gradient(ellipse at 50% 72%, rgba(0, 0, 0, 0.46), transparent 56%),
        radial-gradient(ellipse at 48% 28%, color-mix(in srgb, var(--card-accent) 22%, transparent), transparent 52%),
        linear-gradient(180deg, rgba(255, 255, 255, 0.08), transparent);
    clip-path: polygon(8% 0, 100% 0, 100% 100%, 0 100%, 0 14%);
    transform: translate3d(-50%, 0, 34px);

    .q-icon {
        filter: drop-shadow(0 16px 18px rgba(0, 0, 0, 0.42));
    }
}

.hover-card__footer {
    position: relative;
    z-index: 3;
    padding: 0 1.75rem 1.45rem;
}

.hover-card__stats {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 5.7rem));
    justify-content: center;
    column-gap: 1.15rem;
    row-gap: 0.52rem;
}

.hover-card__stat {
    display: flex;
    align-items: baseline;
    gap: 0.45rem;
    color: color-mix(in srgb, var(--card-text) 86%, transparent);
    font-size: 1.05rem;
    font-weight: 800;
    line-height: 1.1;

    span {
        min-width: 2ch;
        color: var(--card-stat);
        font-size: 1.28rem;
        font-weight: 950;
        text-align: right;
        text-shadow: 0 0 12px color-mix(in srgb, var(--card-accent) 28%, transparent);
    }
}

.hover-card--bronze {
    .hover-card__rating-block::before {
        opacity: 0.55;
    }
}

.hover-card--silver {
    .hover-card__vitals {
        color: #111722;
    }
}

.hover-card--effect-simple-foil {
    .hover-card__plane--foil {
        background:
            radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(255, 255, 240, 0.72), rgba(255, 234, 160, 0.22) 18%, transparent 38%),
            repeating-linear-gradient(62deg, rgba(255, 255, 255, 0.18) 0 2px, transparent 2px 11px, rgba(0, 0, 0, 0.18) 11px 13px, transparent 13px 26px);
    }

    .hover-card__plane--flare {
        background: linear-gradient(var(--glare-angle), transparent 0 34%, rgba(255, 255, 245, 0.62) 46%, rgba(255, 220, 120, 0.28) 52%, transparent 64%);
        opacity: 0;
    }

    &:hover .hover-card__plane--flare,
    &:focus-visible .hover-card__plane--flare {
        opacity: 0.54;
    }
}

.hover-card--effect-simple-holo {
    .hover-card__plane--foil {
        background:
            radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(255, 255, 255, 0.66), rgba(92, 229, 255, 0.3) 14%, rgba(255, 91, 221, 0.24) 28%, transparent 48%),
            conic-gradient(from var(--glare-angle) at var(--pointer-x) var(--pointer-y), rgba(255, 72, 143, 0.32), rgba(255, 228, 76, 0.26), rgba(86, 255, 178, 0.24), rgba(80, 200, 255, 0.3), rgba(181, 103, 255, 0.28), rgba(255, 72, 143, 0.32));
        mix-blend-mode: color-dodge;
    }

    .hover-card__plane--flare {
        background: repeating-linear-gradient(116deg, transparent 0 10px, rgba(255, 255, 255, 0.14) 10px 11px, transparent 11px 22px);
        opacity: 0;
    }

    &:hover .hover-card__plane--flare,
    &:focus-visible .hover-card__plane--flare {
        opacity: 0.58;
    }
}

.hover-card--effect-mirror-gold {
    .hover-card__plane--foil {
        background:
            radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(255, 248, 190, 0.78), rgba(255, 191, 41, 0.28) 16%, transparent 42%),
            linear-gradient(124deg, transparent 0 18%, rgba(255, 251, 226, 0.72) 30%, rgba(195, 122, 14, 0.34) 39%, transparent 51% 100%),
            repeating-linear-gradient(34deg, rgba(255, 225, 92, 0.12) 0 2px, transparent 2px 16px);
    }

    .hover-card__plane--flare {
        background:
            radial-gradient(circle at 22% 19%, rgba(255, 255, 230, 0.46) 0 2px, transparent 3px),
            radial-gradient(circle at 74% 26%, rgba(255, 237, 148, 0.36) 0 2px, transparent 3px),
            radial-gradient(circle at 81% 62%, rgba(255, 255, 232, 0.4) 0 1px, transparent 3px),
            linear-gradient(var(--glare-angle), transparent 0 36%, rgba(255, 255, 234, 0.7) 44%, rgba(255, 205, 60, 0.28) 51%, transparent 60%);
        background-size: auto, auto, auto, 100% 100%;
        opacity: 0;
    }

    &:hover .hover-card__plane--flare,
    &:focus-visible .hover-card__plane--flare {
        opacity: 0.64;
    }
}

.hover-card--effect-lenticular-crest {
    .hover-card__plane--foil {
        background:
            repeating-linear-gradient(90deg, transparent 0 8px, rgba(255, 255, 255, 0.32) 8px 10px, transparent 10px 18px),
            radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(255, 255, 255, 0.58), transparent 42%);
        mask-image:
            radial-gradient(circle at 28% 24%, #000 0 9px, transparent 10px),
            radial-gradient(circle at 62% 32%, #000 0 9px, transparent 10px),
            radial-gradient(circle at 43% 54%, #000 0 9px, transparent 10px),
            radial-gradient(circle at 74% 68%, #000 0 9px, transparent 10px),
            radial-gradient(circle at 22% 76%, #000 0 9px, transparent 10px);
        mask-size: 74px 92px;
        mask-position: var(--surface-x) var(--surface-y);
    }

    .hover-card__plane--crest {
        opacity: 0.28;
    }
}

.hover-card--effect-embossed-crest {
    .hover-card__plane--foil {
        background:
            radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(255, 255, 255, 0.42), transparent 34%),
            linear-gradient(var(--glare-angle), rgba(255, 255, 255, 0.2), transparent 38% 62%, rgba(0, 0, 0, 0.34));
        mix-blend-mode: soft-light;
    }

    .hover-card__plane--crest {
        color: rgba(255, 255, 255, 0.18);
        opacity: 0.34;
        text-shadow:
            calc(var(--surface-soft-x) * -0.08) calc(var(--surface-soft-y) * -0.08) 0 rgba(255, 255, 255, 0.38),
            calc(var(--surface-soft-x) * 0.1) calc(var(--surface-soft-y) * 0.1) 0 rgba(0, 0, 0, 0.42);
    }
}

.hover-card--effect-security-strip {
    .hover-card__plane--foil {
        background:
            radial-gradient(circle at var(--pointer-x) var(--pointer-y), rgba(255, 255, 255, 0.5), transparent 38%),
            repeating-linear-gradient(134deg, rgba(255, 255, 255, 0.16) 0 1px, transparent 1px 11px);
    }

    .hover-card__plane--strip {
        opacity: 0.84;
    }
}

.hover-card--effect-stadium-sweep {
    .hover-card__plane--foil {
        background:
            linear-gradient(72deg, transparent 0 18%, rgba(255, 255, 236, 0.52) 31%, transparent 44%),
            linear-gradient(112deg, transparent 0 48%, rgba(255, 240, 168, 0.42) 62%, transparent 76%),
            radial-gradient(ellipse at var(--pointer-x) var(--pointer-y), rgba(255, 255, 255, 0.46), transparent 38%);
    }

    .hover-card__plane--base {
        background:
            repeating-linear-gradient(90deg, rgba(38, 118, 69, 0.08) 0 12px, rgba(16, 54, 34, 0.12) 12px 24px),
            linear-gradient(var(--glare-angle), transparent 0 34%, rgba(255, 255, 255, 0.14) 48%, transparent 62%);
    }
}

.hover-card--effect-liquid-metal {
    .hover-card__plane--foil {
        background:
            radial-gradient(ellipse at var(--pointer-x) var(--pointer-y), rgba(255, 255, 226, 0.78), rgba(255, 190, 42, 0.36) 18%, rgba(123, 66, 7, 0.24) 39%, transparent 58%),
            radial-gradient(ellipse at 42% 28%, rgba(255, 231, 111, 0.24), transparent 7rem),
            repeating-radial-gradient(ellipse at var(--pointer-x) var(--pointer-y), rgba(255, 248, 185, 0.18) 0 5px, transparent 5px 18px);
    }

    .hover-card__plane--flare {
        background: linear-gradient(var(--glare-angle), transparent 0 24%, rgba(255, 255, 235, 0.7) 38%, rgba(255, 200, 63, 0.24) 48%, transparent 66%);
        clip-path: polygon(18% 7%, 100% 0, 100% 62%, 72% 72%, 12% 52%, 0 18%);
        opacity: 0;
    }

    &:hover .hover-card__plane--flare,
    &:focus-visible .hover-card__plane--flare {
        opacity: 0.52;
    }
}

@media (prefers-reduced-motion: reduce) {
    .hover-card,
    .hover-card:hover,
    .hover-card:focus-visible {
        transform: translateY(-4px);
        transition: box-shadow 120ms ease;
    }

    .hover-card__plane {
        transform: none;
    }
}

@media (max-width: 680px) {
    .hover-card {
        width: min(100%, 276px);
    }

    .hover-card__header {
        padding-inline: 1rem;
    }

    .hover-card__footer {
        padding-inline: 1.45rem;
    }
}
</style>
