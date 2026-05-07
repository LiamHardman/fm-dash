<template>
    <q-page class="card-hover-page">
        <section class="card-hover-shell">
            <header class="card-hover-header">
                <div>
                    <p class="card-hover-kicker">Card render lab</p>
                    <h1>Interactive card hover finishes</h1>
                </div>
                <div class="card-hover-summary">
                    <span>{{ hoverCards.length }} render variants</span>
                    <span>CSS surface layers</span>
                    <span>Pointer-driven tilt</span>
                </div>
            </header>

            <section class="featured-render" aria-labelledby="featured-render-title">
                <div class="featured-render__copy">
                    <p class="card-hover-kicker">Featured renderer</p>
                    <h2 id="featured-render-title">{{ featuredCard.effectName }}</h2>
                    <p>{{ featuredCard.description }}</p>
                    <dl class="render-metrics">
                        <div>
                            <dt>Tier</dt>
                            <dd>{{ featuredCard.tierLabel }}</dd>
                        </div>
                        <div>
                            <dt>Finish</dt>
                            <dd>{{ featuredCard.finish }}</dd>
                        </div>
                        <div>
                            <dt>Layers</dt>
                            <dd>{{ featuredCard.layerCount }}</dd>
                        </div>
                    </dl>
                </div>
                <div class="featured-render__stage">
                    <CardHoverRenderer :card="featuredCard" />
                </div>
            </section>

            <section class="hover-card-grid" aria-label="Card hover render variants">
                <article
                    v-for="card in hoverCards"
                    :key="card.id"
                    class="hover-card-panel"
                >
                    <CardHoverRenderer :card="card" />
                    <div class="hover-card-panel__meta">
                        <p>{{ card.tierLabel }}</p>
                        <h3>{{ card.effectName }}</h3>
                        <span>{{ card.finish }}</span>
                    </div>
                </article>
            </section>
        </section>
    </q-page>
</template>

<script>
import { computed, defineComponent } from 'vue'
import CardHoverRenderer from '../components/CardHoverRenderer.vue'

const statSet = {
  winger: [
    { label: 'PAC', value: 91 },
    { label: 'DRI', value: 88 },
    { label: 'SHO', value: 82 },
    { label: 'DEF', value: 49 },
    { label: 'PAS', value: 84 },
    { label: 'PHY', value: 71 },
  ],
  playmaker: [
    { label: 'PAC', value: 74 },
    { label: 'DRI', value: 86 },
    { label: 'SHO', value: 79 },
    { label: 'DEF', value: 67 },
    { label: 'PAS', value: 90 },
    { label: 'PHY', value: 76 },
  ],
  defender: [
    { label: 'PAC', value: 77 },
    { label: 'DRI', value: 72 },
    { label: 'SHO', value: 52 },
    { label: 'DEF', value: 89 },
    { label: 'PAS', value: 76 },
    { label: 'PHY', value: 88 },
  ],
  striker: [
    { label: 'PAC', value: 86 },
    { label: 'DRI', value: 84 },
    { label: 'SHO', value: 91 },
    { label: 'DEF', value: 41 },
    { label: 'PAS', value: 78 },
    { label: 'PHY', value: 83 },
  ],
}

const tokenSets = {
  bronze: {
    bg: '#201715',
    surface: '#6b3d2a',
    surfaceStrong: '#9a6040',
    border: '#8e6748',
    edge: '#cf9467',
    accent: '#d08a53',
    accentStrong: '#f0bd84',
    text: '#f7e8d7',
    muted: '#d4b391',
    stat: '#f0a564',
    shadow: 'rgba(178, 104, 56, 0.34)',
  },
  silver: {
    bg: '#101820',
    surface: '#d8e0e5',
    surfaceStrong: '#f5f8f9',
    border: '#94a6b3',
    edge: '#f3f7f8',
    accent: '#a7c8d9',
    accentStrong: '#ffffff',
    text: '#f7fbff',
    muted: '#c8d9e4',
    stat: '#ffffff',
    shadow: 'rgba(151, 192, 213, 0.36)',
  },
  gold: {
    bg: '#141006',
    surface: '#b98321',
    surfaceStrong: '#f1c55f',
    border: '#916611',
    edge: '#ffe08a',
    accent: '#d99d25',
    accentStrong: '#fff0a9',
    text: '#fff3c2',
    muted: '#ead184',
    stat: '#ffe489',
    shadow: 'rgba(222, 165, 42, 0.42)',
  },
  blackGold: {
    bg: '#090806',
    surface: '#6c4b12',
    surfaceStrong: '#d8a329',
    border: '#181105',
    edge: '#f0ca68',
    accent: '#b77a12',
    accentStrong: '#fff1a8',
    text: '#fff5c9',
    muted: '#d7b858',
    stat: '#ffd65b',
    shadow: 'rgba(255, 190, 50, 0.42)',
  },
  pearl: {
    bg: '#161720',
    surface: '#e5e0d4',
    surfaceStrong: '#fff8e8',
    border: '#b7b1a2',
    edge: '#f3e5c3',
    accent: '#c8e1da',
    accentStrong: '#ffe6f2',
    text: '#fff7e7',
    muted: '#d8d0bc',
    stat: '#fff1cf',
    shadow: 'rgba(214, 226, 218, 0.34)',
  },
}

const makeCard = (overrides) => ({
  club: 'Valencia',
  crest: 'V',
  name: 'Marin',
  position: 'RW',
  overall: 88,
  vitals: '22 | £62M | £95K/wk',
  stats: statSet.winger,
  tier: 'gold',
  tierLabel: 'Gold rare',
  series: 'Rare Finish',
  effect: 'simple-foil',
  effectName: 'Simple Foil',
  finish: 'Brushed metallic wash',
  description:
    'A restrained foil treatment with a hard diagonal glint, fine grain, and readable lower-card data.',
  layerCount: '5 CSS layers',
  tokens: tokenSets.gold,
  ...overrides,
})

const hoverCards = [
  makeCard({
    id: 'simple-foil',
    overall: 82,
    name: 'Bennett',
    position: 'CM',
    club: 'Carlisle',
    crest: 'C',
    tier: 'bronze',
    tierLabel: 'Bronze rare',
    series: 'Rare Finish',
    effect: 'simple-foil',
    effectName: 'Simple Foil',
    finish: 'Brushed metal',
    stats: statSet.playmaker,
    tokens: tokenSets.bronze,
  }),
  makeCard({
    id: 'simple-holo',
    name: 'Araujo',
    club: 'Lisbon',
    crest: 'L',
    effect: 'simple-holo',
    effectName: 'Simple Holo',
    finish: 'Rainbow diffraction',
    series: 'Promo Holo',
    description:
      'A spectral bloom and diagonal diffraction pass sit behind the portrait and border while the text stays calm.',
  }),
  makeCard({
    id: 'mirror-gold',
    overall: 91,
    name: 'Kowalski',
    position: 'ST',
    club: 'Warsaw',
    crest: 'W',
    tier: 'gold',
    tierLabel: 'Elite gold',
    series: 'Major Gold',
    effect: 'mirror-gold',
    effectName: 'CS2 Gold Major',
    finish: 'Black-gold mirror flakes',
    description:
      'A dense black-gold base with fast mirrored highlights and micro flake sparkle for the loudest rare-card tier.',
    layerCount: '7 CSS layers',
    stats: statSet.striker,
    tokens: tokenSets.blackGold,
  }),
  makeCard({
    id: 'lenticular-crest',
    overall: 86,
    name: 'Dubois',
    position: 'CB',
    club: 'Lyon',
    crest: 'L',
    tier: 'silver',
    tierLabel: 'Club identity',
    series: 'Team Foil',
    effect: 'lenticular-crest',
    effectName: 'Lenticular Crest',
    finish: 'Masked crest repeat',
    description:
      'A team-mark layer shifts against the card surface, giving club identity cards a collectible lenticular read.',
    stats: statSet.defender,
    tokens: tokenSets.silver,
  }),
  makeCard({
    id: 'embossed-crest',
    overall: 90,
    name: 'Santos',
    position: 'CAM',
    club: 'Porto',
    crest: 'P',
    tier: 'gold',
    tierLabel: 'Icon companion',
    series: 'Crest Relief',
    effect: 'embossed-crest',
    effectName: 'Embossed Crest',
    finish: 'Raised monochrome mark',
    description:
      'A large crest trades bright colour for relief, shadow, and angled-light detail across the top art area.',
    stats: statSet.playmaker,
    tokens: tokenSets.pearl,
  }),
  makeCard({
    id: 'security-strip',
    overall: 84,
    name: 'Reed',
    position: 'RB',
    club: 'Bristol',
    crest: 'B',
    tier: 'silver',
    tierLabel: 'Limited card',
    series: 'Limited',
    effect: 'security-strip',
    effectName: 'Security Strip',
    finish: 'Banknote foil strip',
    description:
      'A narrow authentication strip adds repeated micro text and alternating cells without disturbing player data.',
    stats: statSet.defender,
    tokens: tokenSets.silver,
  }),
  makeCard({
    id: 'stadium-sweep',
    overall: 87,
    name: 'Okafor',
    position: 'LW',
    club: 'Basel',
    crest: 'B',
    tier: 'gold',
    tierLabel: 'Matchday',
    series: 'TOTW',
    effect: 'stadium-sweep',
    effectName: 'Stadium Sweep',
    finish: 'Floodlight beams',
    description:
      'Wide pitch-light beams sweep through the surface for matchday and team-of-the-week style cards.',
    stats: statSet.winger,
  }),
  makeCard({
    id: 'liquid-metal',
    overall: 92,
    name: 'Ibrahim',
    position: 'ST',
    club: 'Milan',
    crest: 'M',
    tier: 'gold',
    tierLabel: 'Trophy card',
    series: 'Trophy',
    effect: 'liquid-metal',
    effectName: 'Liquid Metal',
    finish: 'Molten gold pool',
    description:
      'A glossy metallic pool stays concentrated in the upper card so the footer stats remain legible.',
    layerCount: '7 CSS layers',
    stats: statSet.striker,
    tokens: tokenSets.blackGold,
  }),
]

export default defineComponent({
  name: 'CardHoverPage',
  components: { CardHoverRenderer },
  setup() {
    const featuredCard = computed(() => hoverCards[2])

    return {
      featuredCard,
      hoverCards,
    }
  },
})
</script>

<style lang="scss" scoped>
.card-hover-page {
    min-height: 100vh;
    background:
        radial-gradient(circle at 16% 12%, rgba(232, 184, 83, 0.16), transparent 25rem),
        radial-gradient(circle at 86% 16%, rgba(86, 156, 176, 0.16), transparent 22rem),
        linear-gradient(135deg, #11141b 0%, #1d2027 52%, #121112 100%);
    color: #f6f1e7;
}

.card-hover-shell {
    width: min(100%, 1440px);
    margin: 0 auto;
    padding: clamp(1.25rem, 3vw, 2.5rem);
}

.card-hover-header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1.5rem;
    margin-bottom: 1.5rem;

    h1 {
        margin: 0;
        color: #fff6df;
        font-size: clamp(2rem, 4vw, 4.2rem);
        font-weight: 900;
        line-height: 0.98;
    }
}

.card-hover-kicker {
    margin: 0 0 0.45rem;
    color: #e4be68;
    font-size: 0.76rem;
    font-weight: 900;
    letter-spacing: 0.14em;
    text-transform: uppercase;
}

.card-hover-summary {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.55rem;

    span {
        padding: 0.48rem 0.68rem;
        color: rgba(255, 249, 232, 0.78);
        background: rgba(255, 255, 255, 0.07);
        border: 1px solid rgba(255, 255, 255, 0.12);
        border-radius: 6px;
        font-size: 0.78rem;
        font-weight: 800;
    }
}

.featured-render {
    display: grid;
    grid-template-columns: minmax(260px, 0.85fr) minmax(320px, 1.15fr);
    gap: clamp(1rem, 3vw, 2rem);
    align-items: center;
    margin-bottom: 1.7rem;
    padding: clamp(1rem, 2vw, 1.4rem);
    background:
        linear-gradient(90deg, rgba(255, 255, 255, 0.07), rgba(255, 255, 255, 0.025)),
        rgba(3, 5, 9, 0.32);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
}

.featured-render__copy {
    h2 {
        margin: 0;
        color: #fff2c5;
        font-size: clamp(1.65rem, 3vw, 3rem);
        font-weight: 900;
        line-height: 1.02;
    }

    p:not(.card-hover-kicker) {
        max-width: 34rem;
        margin: 0.85rem 0 0;
        color: rgba(255, 249, 232, 0.75);
        font-size: 1rem;
        line-height: 1.6;
    }
}

.featured-render__stage {
    min-height: 500px;
    display: grid;
    place-items: center;
    perspective: 1200px;
}

.render-metrics {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.65rem;
    max-width: 34rem;
    margin: 1.2rem 0 0;

    div {
        padding: 0.78rem;
        background: rgba(255, 255, 255, 0.065);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 6px;
    }

    dt {
        color: rgba(255, 255, 255, 0.52);
        font-size: 0.68rem;
        font-weight: 900;
        letter-spacing: 0.12em;
        text-transform: uppercase;
    }

    dd {
        margin: 0.28rem 0 0;
        color: #fff5d6;
        font-size: 0.9rem;
        font-weight: 900;
    }
}

.hover-card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(286px, 1fr));
    gap: 1rem;
}

.hover-card-panel {
    min-width: 0;
    display: grid;
    justify-items: center;
    gap: 0.85rem;
    padding: 1rem 0.8rem 0.95rem;
    background: rgba(255, 255, 255, 0.055);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
}

.hover-card-panel__meta {
    width: min(100%, 300px);

    p,
    h3 {
        margin: 0;
    }

    p {
        color: rgba(255, 255, 255, 0.54);
        font-size: 0.68rem;
        font-weight: 900;
        letter-spacing: 0.12em;
        text-transform: uppercase;
    }

    h3 {
        margin-top: 0.18rem;
        color: #fff5d6;
        font-size: 1rem;
        font-weight: 900;
    }

    span {
        display: block;
        margin-top: 0.16rem;
        color: rgba(255, 255, 255, 0.62);
        font-size: 0.86rem;
        font-weight: 700;
    }
}

@media (max-width: 940px) {
    .card-hover-header,
    .featured-render {
        display: block;
    }

    .card-hover-summary {
        justify-content: flex-start;
        margin-top: 1rem;
    }

    .featured-render__stage {
        min-height: 460px;
        margin-top: 1.4rem;
    }
}

@media (max-width: 620px) {
    .card-hover-shell {
        padding: 1rem;
    }

    .render-metrics {
        grid-template-columns: 1fr;
    }

    .featured-render__stage {
        min-height: 430px;
    }

    .hover-card-grid {
        grid-template-columns: 1fr;
    }
}
</style>
