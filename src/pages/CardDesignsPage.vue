<template>
    <q-page class="card-designs-page">
        <section class="cards-preview-shell">
            <div class="preview-header">
                <div>
                    <p class="preview-kicker">Bronze non-rare explorations</p>
                    <h1>FM-Dash card surface studies</h1>
                </div>
                <p class="preview-summary">
                    Three CSS-only treatments using one shared player-card
                    hierarchy: rating, position, name, vitals, identity marks,
                    face area, and six footer stats.
                </p>
            </div>

            <div
                class="card-gallery"
                aria-label="Bronze non-rare card design previews"
            >
                <article
                    v-for="design in designs"
                    :key="design.id"
                    class="design-panel"
                >
                    <div
                        class="concept-card"
                        :class="`concept-card--${design.id}`"
                        :style="design.tokens"
                    >
                        <div class="concept-card__texture"></div>
                        <div class="concept-card__content">
                            <header class="concept-card__header">
                                <div class="concept-card__rating-block">
                                    <div class="concept-card__rating">
                                        {{ player.overall }}
                                    </div>
                                    <div class="concept-card__position">
                                        {{ player.position }}
                                    </div>
                                </div>
                                <div class="concept-card__identity">
                                    <div class="concept-card__name">
                                        {{ player.name }}
                                    </div>
                                    <div class="concept-card__vitals">
                                        {{ player.vitals }}
                                    </div>
                                </div>
                            </header>

                            <section
                                class="concept-card__middle"
                                aria-label="Player identity"
                            >
                                <div class="concept-card__marks">
                                    <div
                                        class="concept-card__flag"
                                        aria-label="Example nation flag"
                                    >
                                        <span></span>
                                        <span></span>
                                        <span></span>
                                    </div>
                                    <div
                                        class="concept-card__club"
                                        aria-label="Example club crest"
                                    >
                                        <q-icon name="shield" size="30px" />
                                    </div>
                                </div>

                                <div
                                    class="concept-card__portrait"
                                    aria-label="Example player portrait"
                                >
                                    <q-icon name="person" size="124px" />
                                </div>
                            </section>

                            <footer class="concept-card__footer">
                                <div class="concept-card__stats">
                                    <div
                                        v-for="stat in player.stats"
                                        :key="stat.label"
                                        class="concept-card__stat"
                                    >
                                        <span>{{ stat.value }}</span>
                                        {{ stat.label }}
                                    </div>
                                </div>
                            </footer>
                        </div>
                    </div>

                    <div class="design-notes">
                        <h2>{{ design.name }}</h2>
                        <p>{{ design.description }}</p>
                    </div>
                </article>
            </div>
        </section>
    </q-page>
</template>

<script>
import { defineComponent } from 'vue'

const player = {
  overall: 62,
  position: 'CM',
  name: 'Hartley',
  vitals: '21 | 325K | 2K/wk',
  stats: [
    { label: 'PAC', value: 58 },
    { label: 'SHO', value: 53 },
    { label: 'PAS', value: 59 },
    { label: 'DRI', value: 57 },
    { label: 'DEF', value: 54 },
    { label: 'PHY', value: 60 },
  ],
}

const designs = [
  {
    id: 'weathered-alloy',
    name: 'Weathered Alloy',
    description:
      'Matte bronze body, softened diagonal panels, low edge light, and fine surface scratches for an entry-tier metal card.',
    tokens: {
      '--card-bg': '#241811',
      '--card-surface': '#5d321c',
      '--card-surface-soft': '#7c4425',
      '--card-border': '#94613b',
      '--card-border-muted': '#4a2a1a',
      '--card-accent': '#c2834c',
      '--card-text': '#f3e5d6',
      '--card-muted': '#c8a98a',
      '--card-stat': '#d28f56',
      '--card-shadow': 'rgba(146, 88, 48, 0.24)',
    },
  },
  {
    id: 'brushed-copper',
    name: 'Brushed Copper',
    description:
      'Linear brushed grain, charcoal lower field, subtle copper divider, and restrained corner highlights.',
    tokens: {
      '--card-bg': '#1d1512',
      '--card-surface': '#714126',
      '--card-surface-soft': '#9b6236',
      '--card-border': '#a87143',
      '--card-border-muted': '#3c261b',
      '--card-accent': '#d0975d',
      '--card-text': '#f4e6d8',
      '--card-muted': '#bea07f',
      '--card-stat': '#d79b62',
      '--card-shadow': 'rgba(154, 99, 57, 0.22)',
    },
  },
  {
    id: 'aged-plate',
    name: 'Aged Plate',
    description:
      'Darker umber plate with stamped geometry, oxidised grooves, and a flatter glow so it stays clearly non-rare.',
    tokens: {
      '--card-bg': '#1c1713',
      '--card-surface': '#4c2b1d',
      '--card-surface-soft': '#6a3b27',
      '--card-border': '#806044',
      '--card-border-muted': '#332319',
      '--card-accent': '#b5794d',
      '--card-text': '#f0e1d1',
      '--card-muted': '#b99d80',
      '--card-stat': '#c98755',
      '--card-shadow': 'rgba(111, 78, 49, 0.24)',
    },
  },
]

export default defineComponent({
  name: 'CardDesignsPage',
  setup() {
    return {
      designs,
      player,
    }
  },
})
</script>

<style lang="scss" scoped>
.card-designs-page {
    min-height: 100vh;
    background:
        radial-gradient(
            circle at 20% 8%,
            rgba(194, 131, 76, 0.18),
            transparent 28rem
        ),
        linear-gradient(135deg, #10141d 0%, #161b24 52%, #0d1118 100%);
    color: #f6eee6;
}

.cards-preview-shell {
    width: min(1180px, calc(100% - 2rem));
    margin: 0 auto;
    padding: 4rem 0 5rem;
}

.preview-header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(260px, 430px);
    gap: 2rem;
    align-items: end;
    margin-bottom: 2.5rem;

    h1 {
        margin: 0;
        font-size: 2.4rem;
        line-height: 1.05;
        font-weight: 800;
        color: #fff8ef;
        letter-spacing: 0;
    }
}

.preview-kicker,
.preview-summary {
    color: rgba(246, 238, 230, 0.72);
    margin: 0;
}

.preview-kicker {
    margin-bottom: 0.65rem;
    font-size: 0.82rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
}

.preview-summary {
    font-size: 0.95rem;
    line-height: 1.6;
}

.card-gallery {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1.5rem;
    align-items: start;
}

.design-panel {
    display: grid;
    gap: 1rem;
    justify-items: center;
}

.design-notes {
    width: min(280px, 100%);

    h2 {
        margin: 0 0 0.35rem;
        font-size: 1rem;
        line-height: 1.2;
        color: #fff4e6;
        font-weight: 750;
    }

    p {
        margin: 0;
        color: rgba(246, 238, 230, 0.68);
        font-size: 0.84rem;
        line-height: 1.45;
    }
}

.concept-card {
    width: 280px;
    height: 420px;
    position: relative;
    overflow: hidden;
    color: var(--card-text);
    background:
        linear-gradient(155deg, rgba(255, 255, 255, 0.08), transparent 31%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            var(--card-bg) 58%,
            #120f0d 100%
        );
    border: 2px solid var(--card-border);
    border-radius: 12px;
    box-shadow:
        0 18px 42px rgba(0, 0, 0, 0.42),
        0 0 18px var(--card-shadow);
    isolation: isolate;
    transition:
        transform 180ms ease,
        box-shadow 180ms ease;

    &:hover,
    &:focus-within {
        transform: translateY(-6px);
        box-shadow:
            0 26px 56px rgba(0, 0, 0, 0.48),
            0 0 24px var(--card-shadow);
    }

    &::before {
        content: "";
        position: absolute;
        inset: 10px;
        border: 1px solid rgba(255, 232, 205, 0.18);
        border-radius: 8px;
        z-index: 1;
        pointer-events: none;
    }

    &::after {
        content: "";
        position: absolute;
        left: -12%;
        right: -12%;
        top: 282px;
        height: 3px;
        background: linear-gradient(
            90deg,
            transparent,
            var(--card-accent),
            transparent
        );
        opacity: 0.9;
        z-index: 3;
        transform: rotate(-2deg);
    }
}

.concept-card__texture {
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
}

.concept-card--weathered-alloy {
    .concept-card__texture {
        background:
            repeating-linear-gradient(
                100deg,
                rgba(255, 255, 255, 0.055) 0 1px,
                transparent 1px 13px
            ),
            linear-gradient(
                166deg,
                transparent 0 26%,
                rgba(0, 0, 0, 0.28) 26% 39%,
                transparent 39%
            ),
            linear-gradient(
                15deg,
                transparent 0 55%,
                rgba(194, 131, 76, 0.16) 55% 56%,
                transparent 56%
            ),
            radial-gradient(
                circle at 70% 21%,
                rgba(220, 147, 77, 0.16),
                transparent 7rem
            );
        opacity: 0.8;
    }
}

.concept-card--brushed-copper {
    background:
        linear-gradient(110deg, rgba(255, 255, 255, 0.08), transparent 30%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 35%,
            var(--card-bg) 68%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                0deg,
                rgba(255, 238, 220, 0.045) 0 1px,
                transparent 1px 4px
            ),
            repeating-linear-gradient(
                92deg,
                transparent 0 19px,
                rgba(18, 12, 9, 0.3) 19px 20px
            ),
            linear-gradient(
                150deg,
                transparent 0 42%,
                rgba(16, 12, 10, 0.48) 42% 64%,
                transparent 64%
            ),
            radial-gradient(
                circle at 50% 8%,
                rgba(250, 178, 94, 0.12),
                transparent 9rem
            );
        mix-blend-mode: screen;
        opacity: 0.62;
    }
}

.concept-card--aged-plate {
    background:
        linear-gradient(145deg, rgba(255, 235, 210, 0.05), transparent 26%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            #2a1d16 48%,
            var(--card-bg) 100%
        );

    .concept-card__texture {
        background:
            linear-gradient(
                135deg,
                transparent 0 18%,
                rgba(0, 0, 0, 0.25) 18% 19%,
                transparent 19%
            ),
            linear-gradient(
                45deg,
                transparent 0 36%,
                rgba(181, 121, 77, 0.13) 36% 37%,
                transparent 37%
            ),
            repeating-radial-gradient(
                circle at 32% 30%,
                rgba(255, 255, 255, 0.035) 0 1px,
                transparent 1px 9px
            ),
            repeating-linear-gradient(
                120deg,
                transparent 0 26px,
                rgba(255, 224, 194, 0.045) 26px 27px
            );
        opacity: 0.7;
    }
}

.concept-card__content {
    position: relative;
    z-index: 2;
    height: 100%;
    display: flex;
    flex-direction: column;
}

.concept-card__header {
    position: relative;
    min-height: 128px;
    padding: 18px 18px 0;
    display: grid;
    grid-template-columns: 58px minmax(0, 1fr);
    gap: 0.7rem;
}

.concept-card__rating-block {
    text-align: left;
}

.concept-card__rating {
    color: var(--card-text);
    font-size: 3.25rem;
    line-height: 0.88;
    font-weight: 900;
    text-shadow: 0 2px 9px rgba(0, 0, 0, 0.42);
}

.concept-card__position {
    margin-top: 0.2rem;
    padding-left: 0.15rem;
    color: var(--card-muted);
    font-size: 1.15rem;
    line-height: 1;
    font-weight: 800;
}

.concept-card__identity {
    min-width: 0;
    padding-top: 0.3rem;
    text-align: center;
}

.concept-card__name {
    color: var(--card-text);
    font-size: 1.5rem;
    line-height: 1.1;
    font-weight: 850;
    text-transform: uppercase;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.concept-card__vitals {
    display: inline-block;
    max-width: 100%;
    margin-top: 0.55rem;
    padding: 0.22rem 0.5rem;
    color: #1a100b;
    background: linear-gradient(180deg, var(--card-accent), #9f6337);
    border: 1px solid rgba(255, 230, 198, 0.28);
    border-radius: 4px;
    font-size: 0.72rem;
    line-height: 1.1;
    font-weight: 800;
    white-space: nowrap;
}

.concept-card__middle {
    position: relative;
    flex: 1;
    min-height: 168px;
}

.concept-card__marks {
    position: absolute;
    left: 20px;
    top: 3px;
    display: grid;
    gap: 0.55rem;
    z-index: 2;
}

.concept-card__flag,
.concept-card__club {
    width: 42px;
    height: 30px;
    overflow: hidden;
    border-radius: 3px;
    border: 1px solid rgba(255, 235, 214, 0.24);
    background: rgba(20, 15, 12, 0.6);
    box-shadow: 0 7px 16px rgba(0, 0, 0, 0.25);
}

.concept-card__flag {
    display: grid;
    grid-template-columns: repeat(3, 1fr);

    span:nth-child(1) {
        background: #113f7b;
    }

    span:nth-child(2) {
        background: #f4f1e8;
    }

    span:nth-child(3) {
        background: #b3222b;
    }
}

.concept-card__club {
    height: 42px;
    display: grid;
    place-items: center;
    color: var(--card-accent);
}

.concept-card__portrait {
    position: absolute;
    left: calc(50% + 34px);
    top: 0px;
    width: 166px;
    height: 156px;
    display: grid;
    place-items: center;
    color: rgba(247, 222, 198, 0.78);
    background:
        radial-gradient(
            ellipse at 50% 64%,
            rgba(15, 11, 9, 0.42),
            transparent 56%
        ),
        linear-gradient(
            180deg,
            rgba(255, 255, 255, 0.06),
            rgba(255, 255, 255, 0)
        );
    clip-path: polygon(8% 0, 100% 0, 100% 100%, 0 100%, 0 14%);
    transform: translateX(-50%);

    .q-icon {
        filter: drop-shadow(0 14px 16px rgba(0, 0, 0, 0.4));
    }
}

.concept-card__footer {
    position: relative;
    z-index: 4;
    padding: 0 28px 20px;
}

.concept-card__stats {
    display: grid;
    grid-template-columns: repeat(2, 88px);
    justify-content: center;
    column-gap: 22px;
    row-gap: 0.54rem;
}

.concept-card__stat {
    display: flex;
    align-items: baseline;
    justify-content: flex-start;
    gap: 0.5rem;
    color: rgba(246, 238, 230, 0.86);
    font-size: 1.12rem;
    line-height: 1.15;
    font-weight: 700;

    span {
        color: var(--card-stat);
        font-size: 1.34rem;
        font-weight: 900;
        min-width: 2ch;
        text-align: right;
    }
}

@media (max-width: 980px) {
    .preview-header,
    .card-gallery {
        grid-template-columns: 1fr;
    }

    .preview-header {
        align-items: start;
    }

    .design-panel {
        grid-template-columns: 280px minmax(0, 1fr);
        justify-items: start;
        align-items: center;
    }

    .design-notes {
        width: auto;
    }
}

@media (max-width: 640px) {
    .cards-preview-shell {
        padding: 2rem 0 3rem;
    }

    .preview-header h1 {
        font-size: 1.9rem;
    }

    .design-panel {
        grid-template-columns: 1fr;
        justify-items: center;
    }

    .design-notes {
        width: min(280px, 100%);
    }
}
</style>
