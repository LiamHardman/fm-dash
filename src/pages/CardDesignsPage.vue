<template>
    <q-page class="card-designs-page">
        <section class="cards-preview-shell">
            <div class="preview-header">
                <div>
                    <p class="preview-kicker">Card tier explorations</p>
                    <h1>FM-Dash card surface studies</h1>
                </div>
                <p class="preview-summary">
                    CSS-only card treatments grouped by standard rarity,
                    performance events, legacy stories, and match moments. Each
                    set keeps the shared player-card hierarchy: rating,
                    position, name, vitals, identity marks, face area, and six
                    footer stats.
                </p>
            </div>

            <div class="design-sections">
                <section
                    class="design-section design-section--effects"
                    :aria-labelledby="`${effectSection.id}-title`"
                >
                    <div class="design-section__header">
                        <div>
                            <p class="preview-kicker">
                                {{ effectSection.kicker }}
                            </p>
                            <h2 :id="`${effectSection.id}-title`">
                                {{ effectSection.title }}
                            </h2>
                        </div>
                        <p>{{ effectSection.description }}</p>
                    </div>

                    <div
                        class="card-gallery card-gallery--effects"
                        :aria-label="`${effectSection.title} card effect previews`"
                    >
                        <article
                            v-for="(effect, index) in effectSection.designs"
                            :key="effect.id"
                            class="design-panel effect-panel"
                        >
                            <div
                                class="concept-card concept-card--effect-preview"
                                :class="[
                                    `concept-card--${effect.variant}`,
                                    `concept-card--${effect.baseDesignId}`,
                                    `concept-card--effect-${effect.effect}`,
                                ]"
                                :style="effect.tokens"
                                @mousemove="updateEffectCardPointer"
                                @mouseleave="resetEffectCardPointer"
                                @blur.capture="resetEffectCardPointer"
                            >
                                <div class="concept-card__texture"></div>
                                <div class="concept-card__content">
                                    <header class="concept-card__header">
                                        <div class="concept-card__rating-block">
                                            <div class="concept-card__rating">
                                                {{
                                                    effectSection.player
                                                        .overall
                                                }}
                                            </div>
                                            <div class="concept-card__position">
                                                {{
                                                    effectSection.player
                                                        .position
                                                }}
                                            </div>
                                        </div>
                                        <div class="concept-card__identity">
                                            <div class="concept-card__name">
                                                {{ effectSection.player.name }}
                                            </div>
                                            <div class="concept-card__vitals">
                                                {{
                                                    effectSection.player.vitals
                                                }}
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
                                                <q-icon
                                                    name="shield"
                                                    size="30px"
                                                />
                                            </div>
                                        </div>

                                        <div
                                            class="concept-card__portrait"
                                            aria-label="Example player portrait"
                                        >
                                            <q-icon
                                                name="person"
                                                size="124px"
                                            />
                                        </div>
                                    </section>

                                    <footer class="concept-card__footer">
                                        <div class="concept-card__stats">
                                            <div
                                                v-for="stat in effectSection
                                                    .player.stats"
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
                                <p class="design-number">
                                    {{ index + 1 }}
                                </p>
                                <h3>{{ effect.name }}</h3>
                                <p>{{ effect.description }}</p>
                            </div>
                        </article>
                    </div>
                </section>

                <section
                    v-for="section in designSections"
                    :key="section.id"
                    class="design-section"
                    :aria-labelledby="`${section.id}-title`"
                >
                    <div class="design-section__header">
                        <div>
                            <p class="preview-kicker">{{ section.kicker }}</p>
                            <h2 :id="`${section.id}-title`">
                                {{ section.title }}
                            </h2>
                        </div>
                        <p>{{ section.description }}</p>
                    </div>

                    <div
                        class="card-gallery"
                        :aria-label="`${section.title} card design previews`"
                    >
                        <article
                            v-for="(design, index) in section.designs"
                            :key="design.id"
                            class="design-panel"
                        >
                            <div
                                class="concept-card"
                                :class="[
                                    `concept-card--${design.variant}`,
                                    `concept-card--${design.id}`,
                                ]"
                                :style="design.tokens"
                            >
                                <div class="concept-card__texture"></div>
                                <div class="concept-card__content">
                                    <header class="concept-card__header">
                                        <div class="concept-card__rating-block">
                                            <div class="concept-card__rating">
                                                {{ section.player.overall }}
                                            </div>
                                            <div class="concept-card__position">
                                                {{ section.player.position }}
                                            </div>
                                        </div>
                                        <div class="concept-card__identity">
                                            <div class="concept-card__name">
                                                {{ section.player.name }}
                                            </div>
                                            <div class="concept-card__vitals">
                                                {{ section.player.vitals }}
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
                                                <q-icon
                                                    name="shield"
                                                    size="30px"
                                                />
                                            </div>
                                        </div>

                                        <div
                                            class="concept-card__portrait"
                                            aria-label="Example player portrait"
                                        >
                                            <q-icon
                                                name="person"
                                                size="124px"
                                            />
                                        </div>
                                    </section>

                                    <footer class="concept-card__footer">
                                        <div class="concept-card__stats">
                                            <div
                                                v-for="stat in section.player
                                                    .stats"
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
                                <p class="design-number">
                                    {{ index + 1 }}
                                </p>
                                <h3>{{ design.name }}</h3>
                                <p>{{ design.description }}</p>
                            </div>
                        </article>
                    </div>
                </section>
            </div>
        </section>
    </q-page>
</template>

<script>
import { defineComponent } from 'vue'

const bronzeNonRarePlayer = {
  overall: 62,
  position: 'CM',
  name: 'Hartley',
  vitals: '21 | 325K | 2K/wk',
  stats: [
    { label: 'PAC', value: 58 },
    { label: 'DRI', value: 57 },
    { label: 'SHO', value: 53 },
    { label: 'DEF', value: 54 },
    { label: 'PAS', value: 59 },
    { label: 'PHY', value: 60 },
  ],
}

const bronzeRarePlayer = {
  overall: 62,
  position: 'CM',
  name: 'Hartley',
  vitals: '21 | 325K | 2K/wk',
  stats: [
    { label: 'PAC', value: 69 },
    { label: 'DRI', value: 68 },
    { label: 'SHO', value: 57 },
    { label: 'DEF', value: 51 },
    { label: 'PAS', value: 66 },
    { label: 'PHY', value: 64 },
  ],
}

const silverNonRarePlayer = {
  overall: 72,
  position: 'RB',
  name: 'Larsen',
  vitals: '24 | 2.4M | 9K/wk',
  stats: [
    { label: 'PAC', value: 73 },
    { label: 'DRI', value: 70 },
    { label: 'SHO', value: 58 },
    { label: 'DEF', value: 71 },
    { label: 'PAS', value: 68 },
    { label: 'PHY', value: 69 },
  ],
}

const silverRarePlayer = {
  overall: 72,
  position: 'RB',
  name: 'Larsen',
  vitals: '24 | 2.4M | 9K/wk',
  stats: [
    { label: 'PAC', value: 82 },
    { label: 'DRI', value: 79 },
    { label: 'SHO', value: 61 },
    { label: 'DEF', value: 72 },
    { label: 'PAS', value: 76 },
    { label: 'PHY', value: 74 },
  ],
}

const goldNonRarePlayer = {
  overall: 82,
  position: 'CAM',
  name: 'Marquez',
  vitals: '27 | 24M | 68K/wk',
  stats: [
    { label: 'PAC', value: 78 },
    { label: 'DRI', value: 83 },
    { label: 'SHO', value: 80 },
    { label: 'DEF', value: 54 },
    { label: 'PAS', value: 84 },
    { label: 'PHY', value: 72 },
  ],
}

const goldRarePlayer = {
  overall: 82,
  position: 'CAM',
  name: 'Marquez',
  vitals: '27 | 24M | 68K/wk',
  stats: [
    { label: 'PAC', value: 88 },
    { label: 'DRI', value: 89 },
    { label: 'SHO', value: 84 },
    { label: 'DEF', value: 58 },
    { label: 'PAS', value: 87 },
    { label: 'PHY', value: 78 },
  ],
}

const totwBronzePlayer = {
  overall: 64,
  position: 'ST',
  name: 'Owusu',
  vitals: '22 | 450K | 3K/wk',
  stats: [
    { label: 'PAC', value: 76 },
    { label: 'DRI', value: 67 },
    { label: 'SHO', value: 72 },
    { label: 'DEF', value: 38 },
    { label: 'PAS', value: 60 },
    { label: 'PHY', value: 69 },
  ],
}

const totwSilverPlayer = {
  overall: 74,
  position: 'LW',
  name: 'Iversen',
  vitals: '25 | 4.8M | 14K/wk',
  stats: [
    { label: 'PAC', value: 84 },
    { label: 'DRI', value: 79 },
    { label: 'SHO', value: 75 },
    { label: 'DEF', value: 45 },
    { label: 'PAS', value: 73 },
    { label: 'PHY', value: 68 },
  ],
}

const totwGoldPlayer = {
  overall: 84,
  position: 'ST',
  name: 'Bennett',
  vitals: '28 | 38M | 91K/wk',
  stats: [
    { label: 'PAC', value: 86 },
    { label: 'DRI', value: 85 },
    { label: 'SHO', value: 89 },
    { label: 'DEF', value: 44 },
    { label: 'PAS', value: 80 },
    { label: 'PHY', value: 82 },
  ],
}

const totsBronzePlayer = {
  overall: 66,
  position: 'CDM',
  name: 'Nolan',
  vitals: '23 | 700K | 5K/wk',
  stats: [
    { label: 'PAC', value: 68 },
    { label: 'DRI', value: 70 },
    { label: 'SHO', value: 58 },
    { label: 'DEF', value: 74 },
    { label: 'PAS', value: 72 },
    { label: 'PHY', value: 78 },
  ],
}

const totsSilverPlayer = {
  overall: 76,
  position: 'CM',
  name: 'Kovac',
  vitals: '26 | 7.5M | 22K/wk',
  stats: [
    { label: 'PAC', value: 77 },
    { label: 'DRI', value: 80 },
    { label: 'SHO', value: 73 },
    { label: 'DEF', value: 78 },
    { label: 'PAS', value: 82 },
    { label: 'PHY', value: 79 },
  ],
}

const totsGoldPlayer = {
  overall: 88,
  position: 'RW',
  name: 'Valente',
  vitals: '29 | 72M | 145K/wk',
  stats: [
    { label: 'PAC', value: 91 },
    { label: 'DRI', value: 90 },
    { label: 'SHO', value: 87 },
    { label: 'DEF', value: 52 },
    { label: 'PAS', value: 88 },
    { label: 'PHY', value: 80 },
  ],
}

const iconPlayer = {
  overall: 91,
  position: 'CM',
  name: 'Moretti',
  vitals: 'Legend | 0 | 0/wk',
  stats: [
    { label: 'PAC', value: 82 },
    { label: 'DRI', value: 92 },
    { label: 'SHO', value: 86 },
    { label: 'DEF', value: 84 },
    { label: 'PAS', value: 94 },
    { label: 'PHY', value: 83 },
  ],
}

const heroPlayer = {
  overall: 87,
  position: 'CAM',
  name: 'Duarte',
  vitals: 'Hero | 0 | 0/wk',
  stats: [
    { label: 'PAC', value: 84 },
    { label: 'DRI', value: 89 },
    { label: 'SHO', value: 85 },
    { label: 'DEF', value: 58 },
    { label: 'PAS', value: 88 },
    { label: 'PHY', value: 78 },
  ],
}

const manOfTheMatchPlayer = {
  overall: 86,
  position: 'GK',
  name: 'Ramos',
  vitals: '30 | 31M | 88K/wk',
  stats: [
    { label: 'DIV', value: 89 },
    { label: 'HAN', value: 87 },
    { label: 'KIC', value: 82 },
    { label: 'REF', value: 91 },
    { label: 'SPD', value: 58 },
    { label: 'POS', value: 88 },
  ],
}

const bronzeNonRareDesigns = [
  {
    id: 'weathered-alloy',
    variant: 'non-rare',
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
      '--card-cut': '#c2834c',
      '--card-highlight': '#f3e5d6',
      '--card-text': '#f3e5d6',
      '--card-muted': '#c8a98a',
      '--card-stat': '#d28f56',
      '--card-shadow': 'rgba(146, 88, 48, 0.24)',
    },
  },
  {
    id: 'brushed-copper',
    variant: 'non-rare',
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
      '--card-cut': '#d0975d',
      '--card-highlight': '#f4e6d8',
      '--card-text': '#f4e6d8',
      '--card-muted': '#bea07f',
      '--card-stat': '#d79b62',
      '--card-shadow': 'rgba(154, 99, 57, 0.22)',
    },
  },
  {
    id: 'aged-plate',
    variant: 'non-rare',
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
      '--card-cut': '#b5794d',
      '--card-highlight': '#f0e1d1',
      '--card-text': '#f0e1d1',
      '--card-muted': '#b99d80',
      '--card-stat': '#c98755',
      '--card-shadow': 'rgba(111, 78, 49, 0.24)',
    },
  },
]

const bronzeRareDesigns = [
  {
    id: 'polished-bronze',
    variant: 'rare',
    name: 'Polished Bronze',
    description:
      'Polished bronze base with a radial brushed surface, layered rim, and warm cuts that mark it as a standout low-tier card.',
    tokens: {
      '--card-bg': '#20120d',
      '--card-surface': '#7a3f20',
      '--card-surface-soft': '#b66d37',
      '--card-border': '#d38c4c',
      '--card-border-muted': '#5e321e',
      '--card-accent': '#f0a35c',
      '--card-cut': '#ffd09a',
      '--card-highlight': '#fff1d5',
      '--card-text': '#fff0df',
      '--card-muted': '#e3b98f',
      '--card-stat': '#ffc079',
      '--card-shadow': 'rgba(224, 126, 55, 0.34)',
    },
  },
  {
    id: 'copper-starburst',
    variant: 'rare',
    name: 'Copper Starburst',
    description:
      'Darker chocolate bronze with a restrained starburst behind the portrait and small orange-gold sparks near the rating frame.',
    tokens: {
      '--card-bg': '#1a100c',
      '--card-surface': '#603119',
      '--card-surface-soft': '#a45a2b',
      '--card-border': '#c47b3c',
      '--card-border-muted': '#4a2718',
      '--card-accent': '#ef9546',
      '--card-cut': '#ffc46e',
      '--card-highlight': '#ffe6bd',
      '--card-text': '#fff2e3',
      '--card-muted': '#d9ad82',
      '--card-stat': '#ffb361',
      '--card-shadow': 'rgba(230, 117, 42, 0.32)',
    },
  },
  {
    id: 'amber-facets',
    variant: 'rare',
    name: 'Amber Facets',
    description:
      'Angular polished-bronze facets, thin luminous seams, and a sharper stat glow without drifting into silver or gold value cues.',
    tokens: {
      '--card-bg': '#21130f',
      '--card-surface': '#6d351b',
      '--card-surface-soft': '#bd6f34',
      '--card-border': '#d18b48',
      '--card-border-muted': '#56301f',
      '--card-accent': '#f3a65f',
      '--card-cut': '#ffd38d',
      '--card-highlight': '#fff0cc',
      '--card-text': '#fff1df',
      '--card-muted': '#deb087',
      '--card-stat': '#ffbd72',
      '--card-shadow': 'rgba(218, 119, 49, 0.34)',
    },
  },
]

const silverNonRareDesigns = [
  {
    id: 'satin-steel',
    variant: 'non-rare',
    name: 'Satin Steel',
    description:
      'Muted satin-metal body with soft horizontal brushing, graphite lower field, and a quiet steel rim for a stable standard card.',
    tokens: {
      '--card-bg': '#14181d',
      '--card-surface': '#5f6971',
      '--card-surface-soft': '#8b969e',
      '--card-border': '#9ba5ac',
      '--card-border-muted': '#35404a',
      '--card-accent': '#b5bec5',
      '--card-cut': '#c9d0d5',
      '--card-highlight': '#dde4e8',
      '--card-text': '#edf2f5',
      '--card-muted': '#aab5bd',
      '--card-stat': '#cfd8de',
      '--card-shadow': 'rgba(124, 140, 151, 0.18)',
    },
  },
  {
    id: 'graphite-aluminium',
    variant: 'non-rare',
    name: 'Graphite Aluminium',
    description:
      'Brushed aluminium over a graphite plate, using low-sheen blue-grey reflections and restrained edge light.',
    tokens: {
      '--card-bg': '#10151b',
      '--card-surface': '#505c66',
      '--card-surface-soft': '#7e8b95',
      '--card-border': '#8e9aa3',
      '--card-border-muted': '#2e3943',
      '--card-accent': '#aeb9c1',
      '--card-cut': '#c5ced5',
      '--card-highlight': '#dbe3e7',
      '--card-text': '#eaf0f3',
      '--card-muted': '#a5b1ba',
      '--card-stat': '#c8d3da',
      '--card-shadow': 'rgba(106, 124, 138, 0.18)',
    },
  },
  {
    id: 'pale-geometric',
    variant: 'non-rare',
    name: 'Frosted Plate',
    description:
      'Frosted steel panels and understated angular seams give the average-tier card structure without adding rare-card shine.',
    tokens: {
      '--card-bg': '#151a20',
      '--card-surface': '#59646c',
      '--card-surface-soft': '#8f9aa1',
      '--card-border': '#9ea8af',
      '--card-border-muted': '#37414a',
      '--card-accent': '#b9c2c8',
      '--card-cut': '#d0d7dc',
      '--card-highlight': '#e2e8eb',
      '--card-text': '#eef3f5',
      '--card-muted': '#adb8bf',
      '--card-stat': '#d2dbe0',
      '--card-shadow': 'rgba(132, 148, 158, 0.16)',
    },
  },
]

const silverRareDesigns = [
  {
    id: 'polished-silver',
    variant: 'rare',
    name: 'Machined Prism',
    description:
      'High-polish silver with machined triangular plates, a double rim, and bright corner glints while staying out of event-card blue.',
    tokens: {
      '--card-bg': '#10161d',
      '--card-surface': '#8b98a3',
      '--card-surface-soft': '#d4dde3',
      '--card-border': '#eef5f8',
      '--card-border-muted': '#586672',
      '--card-accent': '#f7fbfd',
      '--card-cut': '#ffffff',
      '--card-highlight': '#ffffff',
      '--card-text': '#fbfdff',
      '--card-muted': '#d3e0e8',
      '--card-stat': '#ffffff',
      '--card-shadow': 'rgba(207, 229, 239, 0.36)',
    },
  },
  {
    id: 'crystal-facet',
    variant: 'rare',
    name: 'Crystal Facet',
    description:
      'Diamond-cut geometry, icy white facets, and small cyan glints arranged as a crystalline field behind the portrait.',
    tokens: {
      '--card-bg': '#0f1720',
      '--card-surface': '#74818c',
      '--card-surface-soft': '#c7d5dc',
      '--card-border': '#e6f3f6',
      '--card-border-muted': '#465660',
      '--card-accent': '#d9fbff',
      '--card-cut': '#8decf6',
      '--card-highlight': '#ffffff',
      '--card-text': '#f8fdff',
      '--card-muted': '#cbe3ea',
      '--card-stat': '#ddfbff',
      '--card-shadow': 'rgba(128, 231, 242, 0.3)',
    },
  },
  {
    id: 'prismatic-edge',
    variant: 'rare',
    name: 'Prismatic Lattice',
    description:
      'Bright silver foil with a prismatic lattice, layered inset trim, and controlled glints near the footer divider.',
    tokens: {
      '--card-bg': '#111820',
      '--card-surface': '#82909b',
      '--card-surface-soft': '#dce6eb',
      '--card-border': '#edf6f8',
      '--card-border-muted': '#52616b',
      '--card-accent': '#f5fbff',
      '--card-cut': '#bff8ff',
      '--card-highlight': '#ffffff',
      '--card-text': '#fbfdff',
      '--card-muted': '#d5e2e9',
      '--card-stat': '#f4fcff',
      '--card-shadow': 'rgba(194, 240, 249, 0.34)',
    },
  },
]

const goldNonRareDesigns = [
  {
    id: 'antique-foil',
    variant: 'non-rare',
    name: 'Antique Foil',
    description:
      'Muted antique-gold foil with a soft black underplate, shallow brushing, and a quiet rim for premium standard cards without rare-card shine.',
    tokens: {
      '--card-bg': '#17130b',
      '--card-surface': '#776029',
      '--card-surface-soft': '#a68d4a',
      '--card-border': '#b39a58',
      '--card-border-muted': '#443719',
      '--card-accent': '#c8ae65',
      '--card-cut': '#d5bc75',
      '--card-highlight': '#ead79d',
      '--card-text': '#f3e6bd',
      '--card-muted': '#cab378',
      '--card-stat': '#d0ae58',
      '--card-shadow': 'rgba(176, 143, 62, 0.16)',
    },
  },
  {
    id: 'satin-sovereign',
    variant: 'non-rare',
    name: 'Satin Sovereign',
    description:
      'Low-lustre satin plating with olive-gold shadows, broad diagonal flow, and deliberately reduced glow around the footer stats.',
    tokens: {
      '--card-bg': '#17140d',
      '--card-surface': '#6e5520',
      '--card-surface-soft': '#aa914d',
      '--card-border': '#aa9252',
      '--card-border-muted': '#3f3218',
      '--card-accent': '#c6ad67',
      '--card-cut': '#d9c27a',
      '--card-highlight': '#ead9a5',
      '--card-text': '#f2e4bc',
      '--card-muted': '#c7af73',
      '--card-stat': '#cfad59',
      '--card-shadow': 'rgba(164, 132, 56, 0.15)',
    },
  },
  {
    id: 'ochre-panel',
    variant: 'non-rare',
    name: 'Ochre Panel',
    description:
      'Darker ochre metal panels with a smoked lower field, stamped seams, and minimal shimmer so the card reads as gold but still restrained.',
    tokens: {
      '--card-bg': '#15110a',
      '--card-surface': '#634a1b',
      '--card-surface-soft': '#927438',
      '--card-border': '#9e854b',
      '--card-border-muted': '#382b14',
      '--card-accent': '#bca15d',
      '--card-cut': '#cbb16d',
      '--card-highlight': '#e6d196',
      '--card-text': '#efe1b8',
      '--card-muted': '#bea66c',
      '--card-stat': '#c8a751',
      '--card-shadow': 'rgba(142, 110, 45, 0.14)',
    },
  },
]

const goldRareDesigns = [
  {
    id: 'royal-foil-burst',
    variant: 'rare',
    name: 'Swept Foil Fan',
    description:
      'Reflective gold foil pulled into a swept fan of raised ribs, with bright ridge highlights behind the portrait and calmer metal below the stats.',
    tokens: {
      '--card-bg': '#171006',
      '--card-surface': '#a97812',
      '--card-surface-soft': '#f0c95d',
      '--card-border': '#ffe08a',
      '--card-border-muted': '#684911',
      '--card-accent': '#ffd667',
      '--card-cut': '#fff0a8',
      '--card-highlight': '#fff9dc',
      '--card-text': '#fff7d7',
      '--card-muted': '#f0d58b',
      '--card-stat': '#ffe17a',
      '--card-shadow': 'rgba(255, 199, 76, 0.38)',
    },
  },
  {
    id: 'jewel-trim',
    variant: 'rare',
    name: 'Folded Lattice',
    description:
      'High-polish gold with folded triangular planes, inset ridge lines, and alternating bright and shadowed facets across the upper body.',
    tokens: {
      '--card-bg': '#140f08',
      '--card-surface': '#b88616',
      '--card-surface-soft': '#f4d06a',
      '--card-border': '#ffe79a',
      '--card-border-muted': '#704d12',
      '--card-accent': '#ffdc74',
      '--card-cut': '#fff6c2',
      '--card-highlight': '#fffdf0',
      '--card-text': '#fff8dc',
      '--card-muted': '#f1d890',
      '--card-stat': '#ffe682',
      '--card-shadow': 'rgba(255, 211, 94, 0.4)',
    },
  },
  {
    id: 'crown-facets',
    variant: 'rare',
    name: 'Embossed Relief',
    description:
      'Embossed abstract relief shapes over yellow-gold foil, using raised shadow cuts and fine speckle so it feels special without becoming an Icon card.',
    tokens: {
      '--card-bg': '#181006',
      '--card-surface': '#a26f10',
      '--card-surface-soft': '#e6ba4b',
      '--card-border': '#f6d77a',
      '--card-border-muted': '#64430d',
      '--card-accent': '#ffd45f',
      '--card-cut': '#ffec9a',
      '--card-highlight': '#fff8ce',
      '--card-text': '#fff5ce',
      '--card-muted': '#ebcf81',
      '--card-stat': '#ffdc6d',
      '--card-shadow': 'rgba(244, 190, 64, 0.38)',
    },
  },
]

const goldRareBackgroundDesigns = [
  {
    id: 'sunburst-vault',
    variant: 'rare',
    name: 'Sunburst Vault',
    description:
      'A classic centred foil burst with alternating polished rays, strongest behind the portrait and fading into a calmer lower panel.',
    tokens: {
      '--card-bg': '#160f05',
      '--card-surface': '#a87410',
      '--card-surface-soft': '#f3ca55',
      '--card-border': '#ffe283',
      '--card-border-muted': '#63450f',
      '--card-accent': '#ffd866',
      '--card-cut': '#fff0a2',
      '--card-highlight': '#fff9d8',
      '--card-text': '#fff7d2',
      '--card-muted': '#efd48b',
      '--card-stat': '#ffe17a',
      '--card-shadow': 'rgba(255, 203, 74, 0.4)',
    },
  },
  {
    id: 'damascus-foil',
    variant: 'rare',
    name: 'Damascus Foil',
    description:
      'Layered wave bands mimic forged precious metal, giving the background a premium organic grain instead of hard facets.',
    tokens: {
      '--card-bg': '#130d05',
      '--card-surface': '#9a6810',
      '--card-surface-soft': '#e7bd4f',
      '--card-border': '#f8da7a',
      '--card-border-muted': '#5b3e0d',
      '--card-accent': '#ffcf5b',
      '--card-cut': '#ffef9f',
      '--card-highlight': '#fff7ce',
      '--card-text': '#fff4cc',
      '--card-muted': '#e8c97e',
      '--card-stat': '#ffdb70',
      '--card-shadow': 'rgba(244, 188, 62, 0.38)',
    },
  },
  {
    id: 'black-gold-chevron',
    variant: 'rare',
    name: 'Black-Gold Chevron',
    description:
      'Deep black enamel chevrons interrupt the gold foil for a sharper standard-rare read without becoming a performance card.',
    tokens: {
      '--card-bg': '#0b0905',
      '--card-surface': '#8e6110',
      '--card-surface-soft': '#e9bd46',
      '--card-border': '#ffe07d',
      '--card-border-muted': '#4b350d',
      '--card-accent': '#ffd15f',
      '--card-cut': '#fff1a8',
      '--card-highlight': '#fff8d0',
      '--card-text': '#fff4cb',
      '--card-muted': '#e5c878',
      '--card-stat': '#ffdc6d',
      '--card-shadow': 'rgba(255, 197, 69, 0.38)',
    },
  },
  {
    id: 'quilted-ingot',
    variant: 'rare',
    name: 'Quilted Ingot',
    description:
      'Soft raised lozenges and inset shadows create an embossed luxury texture that stays readable behind the player area.',
    tokens: {
      '--card-bg': '#171005',
      '--card-surface': '#a57111',
      '--card-surface-soft': '#edc65a',
      '--card-border': '#ffdf82',
      '--card-border-muted': '#67470f',
      '--card-accent': '#ffd467',
      '--card-cut': '#fff0a6',
      '--card-highlight': '#fff9d7',
      '--card-text': '#fff5d0',
      '--card-muted': '#eccf84',
      '--card-stat': '#ffe07a',
      '--card-shadow': 'rgba(250, 196, 70, 0.38)',
    },
  },
  {
    id: 'liquid-aurum',
    variant: 'rare',
    name: 'Liquid Aurum',
    description:
      'Molten gold gradients pool across the upper field, creating broad fluid highlights rather than tight metallic grain.',
    tokens: {
      '--card-bg': '#180e04',
      '--card-surface': '#b0710d',
      '--card-surface-soft': '#ffd15f',
      '--card-border': '#ffe48c',
      '--card-border-muted': '#6a430b',
      '--card-accent': '#ffca52',
      '--card-cut': '#fff3ae',
      '--card-highlight': '#fff9d9',
      '--card-text': '#fff6d2',
      '--card-muted': '#f0ce80',
      '--card-stat': '#ffe078',
      '--card-shadow': 'rgba(255, 190, 53, 0.4)',
    },
  },
  {
    id: 'art-deco-ribs',
    variant: 'rare',
    name: 'Art Deco Ribs',
    description:
      'Symmetric stepped ribs and narrow linework give the card a formal premium background without borrowing Icon ivory cues.',
    tokens: {
      '--card-bg': '#140e05',
      '--card-surface': '#9c6b12',
      '--card-surface-soft': '#e9c04f',
      '--card-border': '#ffe187',
      '--card-border-muted': '#5d410e',
      '--card-accent': '#ffd76b',
      '--card-cut': '#fff0aa',
      '--card-highlight': '#fff8d4',
      '--card-text': '#fff5d0',
      '--card-muted': '#e9cc82',
      '--card-stat': '#ffe17b',
      '--card-shadow': 'rgba(250, 199, 77, 0.39)',
    },
  },
  {
    id: 'confetti-foil',
    variant: 'rare',
    name: 'Confetti Foil',
    description:
      'Small scattered foil chips add sparkle and movement while keeping the palette standard gold instead of season-event blue.',
    tokens: {
      '--card-bg': '#160f05',
      '--card-surface': '#a66e0e',
      '--card-surface-soft': '#f2c957',
      '--card-border': '#ffe486',
      '--card-border-muted': '#65450e',
      '--card-accent': '#ffd05e',
      '--card-cut': '#fff2ab',
      '--card-highlight': '#fffbe0',
      '--card-text': '#fff6d1',
      '--card-muted': '#eed084',
      '--card-stat': '#ffe177',
      '--card-shadow': 'rgba(255, 202, 77, 0.4)',
    },
  },
  {
    id: 'circuit-gilt',
    variant: 'rare',
    name: 'Circuit Gilt',
    description:
      'Thin technical trace lines and plated nodes produce a modern analytical gold rare background for data-heavy players.',
    tokens: {
      '--card-bg': '#120d05',
      '--card-surface': '#98680f',
      '--card-surface-soft': '#eac455',
      '--card-border': '#ffe28a',
      '--card-border-muted': '#583e0e',
      '--card-accent': '#ffd96f',
      '--card-cut': '#fff2ad',
      '--card-highlight': '#fff9d9',
      '--card-text': '#fff5d1',
      '--card-muted': '#e9cd84',
      '--card-stat': '#ffe27d',
      '--card-shadow': 'rgba(249, 198, 79, 0.38)',
    },
  },
  {
    id: 'woven-laurel',
    variant: 'rare',
    name: 'Woven Laurel',
    description:
      'A dark woven foil base with subtle laurel arcs gives the rare card a football-award cue without becoming Team of the Season.',
    tokens: {
      '--card-bg': '#120d05',
      '--card-surface': '#8f6412',
      '--card-surface-soft': '#e3bb4f',
      '--card-border': '#ffdf83',
      '--card-border-muted': '#573d0e',
      '--card-accent': '#ffd36a',
      '--card-cut': '#fff1a7',
      '--card-highlight': '#fff8d4',
      '--card-text': '#fff5cf',
      '--card-muted': '#e7c97e',
      '--card-stat': '#ffdf76',
      '--card-shadow': 'rgba(246, 193, 72, 0.38)',
    },
  },
]

const goldRareEffectBaseDesign = goldRareBackgroundDesigns[6]

const cardEffectDesigns = [
  {
    id: 'cursor-prism-holo',
    variant: goldRareEffectBaseDesign.variant,
    baseDesignId: goldRareEffectBaseDesign.id,
    effect: 'prism-holo',
    name: 'Cursor Prism Holo',
    description:
      'A cursor-led rainbow diffraction wash over the same confetti foil base, with the strongest colour bloom following the pointer.',
    tokens: goldRareEffectBaseDesign.tokens,
  },
  {
    id: 'angled-foil-flash',
    variant: goldRareEffectBaseDesign.variant,
    baseDesignId: goldRareEffectBaseDesign.id,
    effect: 'foil-flash',
    name: 'Angled Foil Flash',
    description:
      'Hard metallic bands catch the pointer like pack foil, keeping the flash directional and quick instead of permanently glowing.',
    tokens: goldRareEffectBaseDesign.tokens,
  },
  {
    id: 'liquid-gold-sheen',
    variant: goldRareEffectBaseDesign.variant,
    baseDesignId: goldRareEffectBaseDesign.id,
    effect: 'liquid-gold',
    name: 'Liquid Gold Sheen',
    description:
      'A warm glossy bloom rolls under the cursor with soft molten highlights for a richer shiny-gold rare treatment.',
    tokens: goldRareEffectBaseDesign.tokens,
  },
  {
    id: 'rainbow-etch-holo',
    variant: goldRareEffectBaseDesign.variant,
    baseDesignId: goldRareEffectBaseDesign.id,
    effect: 'rainbow-etch',
    name: 'Rainbow Etch Holo',
    description:
      'Fine etched lines and tiny spectral flecks become visible on hover, suggesting holographic print embedded in the foil.',
    tokens: goldRareEffectBaseDesign.tokens,
  },
  {
    id: 'mirror-gold-glint',
    variant: goldRareEffectBaseDesign.variant,
    baseDesignId: goldRareEffectBaseDesign.id,
    effect: 'mirror-gold',
    name: 'Mirror Gold Glint',
    description:
      'A polished gold mirror pass adds a crisp moving hotspot while preserving the darker lower card for stat readability.',
    tokens: goldRareEffectBaseDesign.tokens,
  },
  {
    id: 'dark-holo-foil',
    variant: goldRareEffectBaseDesign.variant,
    baseDesignId: goldRareEffectBaseDesign.id,
    effect: 'dark-holo',
    name: 'Dark Holo Foil',
    description:
      'A deeper black-gold foil layer with restrained blue and violet refraction, made to feel premium without becoming an event card.',
    tokens: goldRareEffectBaseDesign.tokens,
  },
]

const totwBronzeDesigns = [
  {
    id: 'bronze-week-strike',
    variant: 'totw',
    name: 'Bronze Week Strike',
    description:
      'Near-black performance body with layered bronze strike marks, a warm inner rim, and a compact flare behind the portrait for a weekly breakout.',
    tokens: {
      '--card-bg': '#070605',
      '--card-surface': '#110b08',
      '--card-surface-soft': '#7d4524',
      '--card-border': '#c27b42',
      '--card-border-muted': '#2d1a10',
      '--card-accent': '#e39a55',
      '--card-cut': '#ffbd76',
      '--card-highlight': '#ffe2bd',
      '--card-text': '#fff1df',
      '--card-muted': '#d6ad83',
      '--card-stat': '#ffb366',
      '--card-shadow': 'rgba(226, 132, 61, 0.34)',
    },
  },
  {
    id: 'bronze-week-signal',
    variant: 'totw',
    name: 'Copper Signal',
    description:
      'Copper soundwave bands run through a near-black field, with amber peaks and smoked bronze depth behind the portrait.',
    tokens: {
      '--card-bg': '#060505',
      '--card-surface': '#100a08',
      '--card-surface-soft': '#69391f',
      '--card-border': '#b8733e',
      '--card-border-muted': '#2c1a11',
      '--card-accent': '#df8b49',
      '--card-cut': '#ffc27c',
      '--card-highlight': '#ffe8cc',
      '--card-text': '#fff0e1',
      '--card-muted': '#d2a47c',
      '--card-stat': '#f6a85f',
      '--card-shadow': 'rgba(219, 121, 55, 0.32)',
    },
  },
  {
    id: 'bronze-week-spotlight',
    variant: 'totw',
    name: 'Amber Spotlight',
    description:
      'A black outer shell, bronze spotlight cone, shadowed stadium arcs, and restrained edge sparks make the lower-tier event card read immediately.',
    tokens: {
      '--card-bg': '#070504',
      '--card-surface': '#130b07',
      '--card-surface-soft': '#875126',
      '--card-border': '#ca8343',
      '--card-border-muted': '#301d12',
      '--card-accent': '#eba057',
      '--card-cut': '#ffd08d',
      '--card-highlight': '#fff0d1',
      '--card-text': '#fff3e4',
      '--card-muted': '#deb58b',
      '--card-stat': '#ffba72',
      '--card-shadow': 'rgba(235, 144, 65, 0.34)',
    },
  },
]

const totwSilverDesigns = [
  {
    id: 'silver-week-beam',
    variant: 'totw',
    name: 'Silver Beam',
    description:
      'Cool stadium beams cut across a blackened steel body with a bright silver inner rim and icy rating highlights.',
    tokens: {
      '--card-bg': '#05080c',
      '--card-surface': '#0a1118',
      '--card-surface-soft': '#778691',
      '--card-border': '#d8e3e9',
      '--card-border-muted': '#26313b',
      '--card-accent': '#edf6fb',
      '--card-cut': '#bdf4ff',
      '--card-highlight': '#ffffff',
      '--card-text': '#f7fbff',
      '--card-muted': '#cad7de',
      '--card-stat': '#e8f7ff',
      '--card-shadow': 'rgba(188, 230, 246, 0.32)',
    },
  },
  {
    id: 'silver-week-shards',
    variant: 'totw',
    name: 'Icy Shards',
    description:
      'Angular silver shards sit over a black match-night base, adding stronger event energy without becoming too blue.',
    tokens: {
      '--card-bg': '#05080d',
      '--card-surface': '#0a1119',
      '--card-surface-soft': '#73828e',
      '--card-border': '#cfdce4',
      '--card-border-muted': '#25313b',
      '--card-accent': '#e5f2f8',
      '--card-cut': '#a7ecf6',
      '--card-highlight': '#ffffff',
      '--card-text': '#f5fbff',
      '--card-muted': '#c5d5dd',
      '--card-stat': '#def5fb',
      '--card-shadow': 'rgba(166, 224, 238, 0.3)',
    },
  },
  {
    id: 'silver-week-lights',
    variant: 'totw',
    name: 'Floodlight Steel',
    description:
      'A darker floodlight motif, steel glow, and black rim make the silver weekly standout feel clinical and controlled.',
    tokens: {
      '--card-bg': '#04070b',
      '--card-surface': '#091018',
      '--card-surface-soft': '#7f8f9b',
      '--card-border': '#e3ebef',
      '--card-border-muted': '#28343e',
      '--card-accent': '#f3fbff',
      '--card-cut': '#c8f8ff',
      '--card-highlight': '#ffffff',
      '--card-text': '#f8fdff',
      '--card-muted': '#cedde4',
      '--card-stat': '#f0fbff',
      '--card-shadow': 'rgba(201, 237, 247, 0.32)',
    },
  },
]

const totwGoldDesigns = [
  {
    id: 'gold-week-headline',
    variant: 'totw',
    name: 'Headline Rays',
    description:
      'Near-black foil body with deeper headline gold rays, a glowing inner rim, and a bright but controlled stat divider.',
    tokens: {
      '--card-bg': '#060503',
      '--card-surface': '#100b04',
      '--card-surface-soft': '#a87515',
      '--card-border': '#ffd46b',
      '--card-border-muted': '#3b2a0b',
      '--card-accent': '#ffcc58',
      '--card-cut': '#fff0a5',
      '--card-highlight': '#fff8d8',
      '--card-text': '#fff5d1',
      '--card-muted': '#e9c978',
      '--card-stat': '#ffdd75',
      '--card-shadow': 'rgba(255, 195, 65, 0.4)',
    },
  },
  {
    id: 'gold-week-flash',
    variant: 'totw',
    name: 'Foil Flash',
    description:
      'Diagonal gold foil flashes and black-on-black paneling travel through the portrait field while the lower stat area remains readable.',
    tokens: {
      '--card-bg': '#060503',
      '--card-surface': '#110c04',
      '--card-surface-soft': '#9d6810',
      '--card-border': '#f5c95f',
      '--card-border-muted': '#372609',
      '--card-accent': '#ffc852',
      '--card-cut': '#ffe892',
      '--card-highlight': '#fff6cd',
      '--card-text': '#fff3cb',
      '--card-muted': '#e5c36f',
      '--card-stat': '#ffd96f',
      '--card-shadow': 'rgba(246, 184, 55, 0.38)',
    },
  },
  {
    id: 'gold-week-strokes',
    variant: 'totw',
    name: 'Matchday Strokes',
    description:
      'Energetic gold brush strokes, dark stadium bands, and spotlight flecks give the top field movement without overpowering the player data.',
    tokens: {
      '--card-bg': '#060503',
      '--card-surface': '#120c04',
      '--card-surface-soft': '#b27c18',
      '--card-border': '#ffda73',
      '--card-border-muted': '#3d2b0a',
      '--card-accent': '#ffd05d',
      '--card-cut': '#fff1a8',
      '--card-highlight': '#fff9dd',
      '--card-text': '#fff6d3',
      '--card-muted': '#edcd7b',
      '--card-stat': '#ffe27b',
      '--card-shadow': 'rgba(255, 202, 76, 0.4)',
    },
  },
]

const totsBronzeDesigns = [
  {
    id: 'bronze-season-laurel',
    variant: 'tots',
    name: 'Bronze Laurel',
    description:
      'Bronze achievement card with dark seasonal depth, bronze laurel geometry, ceremonial trim, and a quieter teal glow.',
    tokens: {
      '--card-bg': '#091417',
      '--card-surface': '#6f4528',
      '--card-surface-soft': '#b8743f',
      '--card-border': '#d09357',
      '--card-border-muted': '#3b2518',
      '--card-accent': '#2fc4b4',
      '--card-cut': '#f2b06b',
      '--card-highlight': '#ffe4c2',
      '--card-text': '#fff0df',
      '--card-muted': '#cbb191',
      '--card-stat': '#ffc37c',
      '--card-shadow': 'rgba(42, 190, 174, 0.26)',
    },
  },
  {
    id: 'bronze-season-ribbon',
    variant: 'tots',
    name: 'Teal Ribbon',
    description:
      'Flowing teal ribbon forms move through an aged bronze body and frame for a season-long recognition card.',
    tokens: {
      '--card-bg': '#081316',
      '--card-surface': '#684025',
      '--card-surface-soft': '#a96837',
      '--card-border': '#cf8b4f',
      '--card-border-muted': '#392417',
      '--card-accent': '#35bfa9',
      '--card-cut': '#efab67',
      '--card-highlight': '#ffe2be',
      '--card-text': '#fff0de',
      '--card-muted': '#c8ad8b',
      '--card-stat': '#ffbc73',
      '--card-shadow': 'rgba(52, 184, 161, 0.25)',
    },
  },
  {
    id: 'bronze-season-engraved',
    variant: 'tots',
    name: 'Engraved Trophy',
    description:
      'Trophy-engraved bronze plates dominate the seasonal base, with small emerald points of celebration.',
    tokens: {
      '--card-bg': '#091312',
      '--card-surface': '#73482a',
      '--card-surface-soft': '#bd7842',
      '--card-border': '#d29257',
      '--card-border-muted': '#3e2819',
      '--card-accent': '#49c98e',
      '--card-cut': '#f3b371',
      '--card-highlight': '#ffe7c8',
      '--card-text': '#fff2e2',
      '--card-muted': '#cfb392',
      '--card-stat': '#ffc27a',
      '--card-shadow': 'rgba(73, 201, 142, 0.24)',
    },
  },
]

const totsSilverDesigns = [
  {
    id: 'silver-season-crest',
    variant: 'tots',
    name: 'Silver Crest',
    description:
      'Polished silver body, deep navy undertone, and a restrained crest pattern create a composed season card.',
    tokens: {
      '--card-bg': '#08131c',
      '--card-surface': '#7d8b94',
      '--card-surface-soft': '#d5dee3',
      '--card-border': '#eef6f8',
      '--card-border-muted': '#3d4b56',
      '--card-accent': '#4dbfe7',
      '--card-cut': '#f5ffff',
      '--card-highlight': '#ffffff',
      '--card-text': '#f8fdff',
      '--card-muted': '#d2e3ea',
      '--card-stat': '#effbff',
      '--card-shadow': 'rgba(70, 187, 226, 0.28)',
    },
  },
  {
    id: 'silver-season-wave',
    variant: 'tots',
    name: 'Royal Wave',
    description:
      'Silver trophy metal carries elegant royal-blue wave geometry, preserving the tier read without the drama of gold season cards.',
    tokens: {
      '--card-bg': '#08111c',
      '--card-surface': '#74838e',
      '--card-surface-soft': '#cad5dc',
      '--card-border': '#e5eef2',
      '--card-border-muted': '#3a4a58',
      '--card-accent': '#4d8dff',
      '--card-cut': '#f1fbff',
      '--card-highlight': '#ffffff',
      '--card-text': '#f7fcff',
      '--card-muted': '#cedde7',
      '--card-stat': '#eaf8ff',
      '--card-shadow': 'rgba(78, 141, 255, 0.28)',
    },
  },
  {
    id: 'silver-season-laurel',
    variant: 'tots',
    name: 'Composed Laurel',
    description:
      'Layered laurel arcs, clean white-blue glow, and a polished silver body keep the season card formal and legible.',
    tokens: {
      '--card-bg': '#07121a',
      '--card-surface': '#84929b',
      '--card-surface-soft': '#dbe5ea',
      '--card-border': '#f0f7fa',
      '--card-border-muted': '#40505a',
      '--card-accent': '#6fe0ee',
      '--card-cut': '#fbffff',
      '--card-highlight': '#ffffff',
      '--card-text': '#f8fdff',
      '--card-muted': '#d5e5eb',
      '--card-stat': '#f0fdff',
      '--card-shadow': 'rgba(111, 224, 238, 0.28)',
    },
  },
]

const totsGoldDesigns = [
  {
    id: 'gold-season-crown',
    variant: 'tots',
    name: 'Season Crown',
    description:
      'Deep navy trophy foil with crown-like gold geometry, layered celebration rays, royal-blue depth, and the strongest season-card edge glow.',
    tokens: {
      '--card-bg': '#070f1b',
      '--card-surface': '#17275a',
      '--card-surface-soft': '#e5b437',
      '--card-border': '#ffe58c',
      '--card-border-muted': '#2a3762',
      '--card-accent': '#4c7dff',
      '--card-cut': '#fff1a6',
      '--card-highlight': '#fff9dc',
      '--card-text': '#fff7d6',
      '--card-muted': '#ead287',
      '--card-stat': '#ffe789',
      '--card-shadow': 'rgba(255, 210, 86, 0.42)',
    },
  },
  {
    id: 'gold-season-celebration',
    variant: 'tots',
    name: 'Celebration Lights',
    description:
      'Stadium celebration lighting, rich gold trim, confetti-like foil points, and a blue-black body signal season-defining excellence.',
    tokens: {
      '--card-bg': '#060d18',
      '--card-surface': '#17244c',
      '--card-surface-soft': '#dba62d',
      '--card-border': '#ffe07c',
      '--card-border-muted': '#25365c',
      '--card-accent': '#5d8dff',
      '--card-cut': '#ffeca1',
      '--card-highlight': '#fff8d5',
      '--card-text': '#fff5cf',
      '--card-muted': '#e5cb7d',
      '--card-stat': '#ffe27c',
      '--card-shadow': 'rgba(255, 202, 76, 0.4)',
    },
  },
  {
    id: 'gold-season-trophy',
    variant: 'tots',
    name: 'Trophy Foil',
    description:
      'Ornate trophy-foil arcs, starburst facets, and luminous gold edges carry the premium season mood while preserving dark stat contrast.',
    tokens: {
      '--card-bg': '#070e17',
      '--card-surface': '#182652',
      '--card-surface-soft': '#edbb3f',
      '--card-border': '#ffe895',
      '--card-border-muted': '#263251',
      '--card-accent': '#6f95ff',
      '--card-cut': '#fff2a9',
      '--card-highlight': '#fff9df',
      '--card-text': '#fff7d8',
      '--card-muted': '#ead087',
      '--card-stat': '#ffe98d',
      '--card-shadow': 'rgba(255, 216, 96, 0.42)',
    },
  },
]

const iconDesigns = [
  {
    id: 'ivory-museum',
    variant: 'icon',
    name: 'Ivory Museum',
    description:
      'Pearl ivory body, antique-gold trim, and a museum-frame silhouette separate legendary cards from standard gold.',
    tokens: {
      '--card-bg': '#f0e3c8',
      '--card-surface': '#fbf2df',
      '--card-surface-soft': '#d8bd77',
      '--card-border': '#b9903b',
      '--card-border-muted': '#8b743c',
      '--card-accent': '#b88c35',
      '--card-cut': '#f4d98b',
      '--card-highlight': '#fffaf0',
      '--card-text': '#3d2b12',
      '--card-muted': '#6e5a31',
      '--card-stat': '#9f7627',
      '--card-shadow': 'rgba(240, 218, 166, 0.36)',
    },
  },
  {
    id: 'pearl-relief',
    variant: 'icon',
    name: 'Pearl Relief',
    description:
      'Soft porcelain panels and raised heritage relief shapes create a timeless card with restrained gold accents.',
    tokens: {
      '--card-bg': '#eee0c2',
      '--card-surface': '#fff6e5',
      '--card-surface-soft': '#dbc486',
      '--card-border': '#c29b45',
      '--card-border-muted': '#947a3d',
      '--card-accent': '#c0943c',
      '--card-cut': '#f7df93',
      '--card-highlight': '#fffdf5',
      '--card-text': '#382710',
      '--card-muted': '#6a5530',
      '--card-stat': '#a77b27',
      '--card-shadow': 'rgba(238, 214, 158, 0.38)',
    },
  },
  {
    id: 'parchment-frame',
    variant: 'icon',
    name: 'Parchment Frame',
    description:
      'Aged parchment grain, charcoal insets, and elegant gold linework give the card a historic presentation.',
    tokens: {
      '--card-bg': '#e9d8b4',
      '--card-surface': '#fbefd8',
      '--card-surface-soft': '#cfb16a',
      '--card-border': '#ad8432',
      '--card-border-muted': '#806734',
      '--card-accent': '#aa7c29',
      '--card-cut': '#efd178',
      '--card-highlight': '#fff7e8',
      '--card-text': '#35240f',
      '--card-muted': '#66502a',
      '--card-stat': '#956b21',
      '--card-shadow': 'rgba(224, 196, 130, 0.34)',
    },
  },
]

const heroDesigns = [
  {
    id: 'mural-burst',
    variant: 'hero',
    name: 'Mural Burst',
    description:
      'Deep neutral poster base with bold club-adaptable rays, halftone texture, and a badge-like nameplate.',
    tokens: {
      '--card-bg': '#101024',
      '--card-surface': '#1c2151',
      '--card-surface-soft': '#d63f70',
      '--card-border': '#f0c35b',
      '--card-border-muted': '#293060',
      '--card-accent': '#27c1ff',
      '--card-cut': '#ffdf71',
      '--card-highlight': '#fff1bd',
      '--card-text': '#fff2db',
      '--card-muted': '#cfd5ff',
      '--card-stat': '#ffdc6c',
      '--card-shadow': 'rgba(39, 193, 255, 0.34)',
    },
  },
  {
    id: 'scarf-stripes',
    variant: 'hero',
    name: 'Scarf Stripes',
    description:
      'Supporter-scarf striping, enamel trim, and punchy accent channels make the card emotional without becoming classic like Icon.',
    tokens: {
      '--card-bg': '#0e1424',
      '--card-surface': '#172d44',
      '--card-surface-soft': '#e34c42',
      '--card-border': '#f4d060',
      '--card-border-muted': '#28405d',
      '--card-accent': '#40d6a4',
      '--card-cut': '#ffe27a',
      '--card-highlight': '#fff4c7',
      '--card-text': '#fff3df',
      '--card-muted': '#c5e2df',
      '--card-stat': '#ffe06f',
      '--card-shadow': 'rgba(64, 214, 164, 0.32)',
    },
  },
  {
    id: 'badge-poster',
    variant: 'hero',
    name: 'Badge Poster',
    description:
      'Graphic poster blocks, bold frame steps, and adaptable accent panels create a fan-favourite story card.',
    tokens: {
      '--card-bg': '#121024',
      '--card-surface': '#26204b',
      '--card-surface-soft': '#d94f2f',
      '--card-border': '#f2bf4d',
      '--card-border-muted': '#342b62',
      '--card-accent': '#66d3ff',
      '--card-cut': '#ffdc74',
      '--card-highlight': '#fff2c2',
      '--card-text': '#fff2df',
      '--card-muted': '#d0d6ff',
      '--card-stat': '#ffda68',
      '--card-shadow': 'rgba(102, 211, 255, 0.32)',
    },
  },
]

const manOfTheMatchDesigns = [
  {
    id: 'orange-impact',
    variant: 'match-event',
    name: 'Orange Impact',
    description:
      'Vivid orange impact lines explode from a black match-night base, with a hot rim and protected dark stat zone.',
    tokens: {
      '--card-bg': '#160905',
      '--card-surface': '#2a0f07',
      '--card-surface-soft': '#f36a1d',
      '--card-border': '#ff8c2a',
      '--card-border-muted': '#66220b',
      '--card-accent': '#ff7a1a',
      '--card-cut': '#ffd09a',
      '--card-highlight': '#fff0d9',
      '--card-text': '#fff3e8',
      '--card-muted': '#ffc49a',
      '--card-stat': '#ffb05f',
      '--card-shadow': 'rgba(255, 106, 28, 0.44)',
    },
  },
  {
    id: 'match-night-flare',
    variant: 'match-event',
    name: 'Match-Night Flare',
    description:
      'Hot foil flare, deep red-orange shadows, and white edge flashes make the one-match dominance feel immediate.',
    tokens: {
      '--card-bg': '#150706',
      '--card-surface': '#2d0d08',
      '--card-surface-soft': '#e64e19',
      '--card-border': '#ff7f26',
      '--card-border-muted': '#5c1c0c',
      '--card-accent': '#ff6d17',
      '--card-cut': '#ffc487',
      '--card-highlight': '#fff1de',
      '--card-text': '#fff2e8',
      '--card-muted': '#ffbd91',
      '--card-stat': '#ffa654',
      '--card-shadow': 'rgba(255, 92, 24, 0.42)',
    },
  },
  {
    id: 'ember-spotlight',
    variant: 'match-event',
    name: 'Ember Spotlight',
    description:
      'Abstract ember gradients and a sharp orange spotlight frame the portrait while avoiding bright orange behind white text.',
    tokens: {
      '--card-bg': '#130806',
      '--card-surface': '#2a0e08',
      '--card-surface-soft': '#ff741f',
      '--card-border': '#ff9733',
      '--card-border-muted': '#64230c',
      '--card-accent': '#ff7f22',
      '--card-cut': '#ffd39d',
      '--card-highlight': '#fff4e5',
      '--card-text': '#fff4ea',
      '--card-muted': '#ffc69f',
      '--card-stat': '#ffb86d',
      '--card-shadow': 'rgba(255, 118, 31, 0.44)',
    },
  },
]

const designSections = [
  {
    id: 'bronze-non-rare',
    kicker: 'Standard bronze',
    title: 'Bronze Non-Rare',
    description:
      'Grounded, lower-shine bronze concepts for modest players: metal grain, restrained diagonal paneling, and simple single-rim reads.',
    player: bronzeNonRarePlayer,
    designs: bronzeNonRareDesigns,
  },
  {
    id: 'bronze-rare',
    kicker: 'Hidden-gem bronze',
    title: 'Bronze Rare',
    description:
      'Polished bronze concepts for low-tier players with standout attributes: layered rims, warm luminous cuts, and stronger edge light.',
    player: bronzeRarePlayer,
    designs: bronzeRareDesigns,
  },
  {
    id: 'silver-non-rare',
    kicker: 'Standard silver',
    title: 'Silver Non-Rare',
    description:
      'Balanced silver concepts for average-rated players: satin steel, clean brushed texture, graphite depth, and controlled reflection.',
    player: silverNonRarePlayer,
    designs: silverNonRareDesigns,
  },
  {
    id: 'silver-rare',
    kicker: 'Standout silver',
    title: 'Silver Rare',
    description:
      'Sharper silver concepts for average-rated players with stronger attributes: polished rims, crystalline facets, and cool glints.',
    player: silverRarePlayer,
    designs: silverRareDesigns,
  },
  {
    id: 'gold-non-rare',
    kicker: 'Standard gold',
    title: 'Gold Non-Rare',
    description:
      'Muted premium gold concepts for above-average players: antique foil, satin plating, ochre shadows, low glow, and restrained metal texture.',
    player: goldNonRarePlayer,
    designs: goldNonRareDesigns,
  },
  {
    id: 'gold-rare',
    kicker: 'Elite standard gold',
    title: 'Gold Rare',
    description:
      'Richer rare gold concepts for above-average players with exceptional attributes: swept foil ribs, folded lattice facets, embossed relief, and stronger rating frames.',
    player: goldRarePlayer,
    designs: goldRareDesigns,
  },
  {
    id: 'totw-bronze',
    kicker: 'Weekly performance bronze',
    title: 'Team of the Week Bronze',
    description:
      'Bronze performance concepts for a lower-rated weekly breakout: black outer rims, bronze event energy, and sharp but modest glow.',
    player: totwBronzePlayer,
    designs: totwBronzeDesigns,
  },
  {
    id: 'totw-silver',
    kicker: 'Weekly performance silver',
    title: 'Team of the Week Silver',
    description:
      'Silver performance concepts for a clinical weekly standout: dark match-night bodies, cool beams, shards, and bright inner rims.',
    player: totwSilverPlayer,
    designs: totwSilverDesigns,
  },
  {
    id: 'totw-gold',
    kicker: 'Weekly performance gold',
    title: 'Team of the Week Gold',
    description:
      'Gold performance concepts for headline weekly form: black foil, golden spotlight energy, and controlled event shimmer.',
    player: totwGoldPlayer,
    designs: totwGoldDesigns,
  },
  {
    id: 'tots-bronze',
    kicker: 'Season achievement bronze',
    title: 'Team of the Season Bronze',
    description:
      'Bronze season concepts with ceremonial navy bases, bronze trim, teal accents, and achievement-led patterns.',
    player: totsBronzePlayer,
    designs: totsBronzeDesigns,
  },
  {
    id: 'tots-silver',
    kicker: 'Season achievement silver',
    title: 'Team of the Season Silver',
    description:
      'Prestige silver season concepts using trophy metal, royal-blue depth, laurel geometry, and composed white-blue glow.',
    player: totsSilverPlayer,
    designs: totsSilverDesigns,
  },
  {
    id: 'tots-gold',
    kicker: 'Season achievement gold',
    title: 'Team of the Season Gold',
    description:
      'Season-defining gold concepts with ornate trophy foil, royal-blue structure, crown geometry, and luminous ceremonial edges.',
    player: totsGoldPlayer,
    designs: totsGoldDesigns,
  },
  {
    id: 'icon',
    kicker: 'Legacy legend',
    title: 'Icon',
    description:
      'Classic legacy concepts built from ivory, pearl, parchment, and antique-gold frames rather than black event styling.',
    player: iconPlayer,
    designs: iconDesigns,
  },
  {
    id: 'hero',
    kicker: 'Story favourite',
    title: 'Hero',
    description:
      'Graphic story-card concepts with mural energy, scarf-inspired patterns, adaptable club-color slots, and badge-like frames.',
    player: heroPlayer,
    designs: heroDesigns,
  },
  {
    id: 'man-of-the-match',
    kicker: 'Single-match event',
    title: 'Man of the Match',
    description:
      'Explosive match-event concepts with vivid orange, black match-night depth, hot foil energy, and protected text contrast.',
    player: manOfTheMatchPlayer,
    designs: manOfTheMatchDesigns,
  },
  {
    id: 'gold-rare-background-lab',
    kicker: 'Background exploration',
    title: 'Gold Rare Background Lab',
    description:
      'Nine additional gold rare background studies using the same card structure, data, borders, and typography so the surface direction can be compared directly.',
    player: goldRarePlayer,
    designs: goldRareBackgroundDesigns,
  },
]

const effectSection = {
  id: 'gold-rare-hover-effects',
  kicker: 'Interaction exploration',
  title: 'Gold Rare Hover Effects',
  description:
    'Hover studies using Gold Rare Background Lab design 7 as the fixed base card. Each preview tracks pointer position for physical card movement under a static light source, then layers a different holo, foil, or shiny-gold treatment on top.',
  player: goldRarePlayer,
  designs: cardEffectDesigns,
}

const resetEffectCardPointer = (event) => {
  const card = event.currentTarget

  card.style.setProperty('--shift-x', '0px')
  card.style.setProperty('--shift-y', '0px')
  card.style.setProperty('--surface-x', '0px')
  card.style.setProperty('--surface-y', '0px')
  card.style.setProperty('--surface-soft-x', '0px')
  card.style.setProperty('--surface-soft-y', '0px')
  card.style.setProperty('--surface-reverse-x', '0px')
  card.style.setProperty('--surface-reverse-y', '0px')
  card.style.setProperty('--surface-hard-x', '0px')
  card.style.setProperty('--surface-hard-y', '0px')
  card.style.setProperty('--tilt-x', '0deg')
  card.style.setProperty('--tilt-y', '0deg')
}

const updateEffectCardPointer = (event) => {
  const card = event.currentTarget
  const bounds = card.getBoundingClientRect()
  const pointerX = (event.clientX - bounds.left) / bounds.width
  const pointerY = (event.clientY - bounds.top) / bounds.height
  const clampedX = Math.min(Math.max(pointerX, 0), 1)
  const clampedY = Math.min(Math.max(pointerY, 0), 1)
  const surfaceX = (0.5 - clampedX) * 34
  const surfaceY = (0.5 - clampedY) * 28

  card.style.setProperty('--shift-x', `${((clampedX - 0.5) * 16).toFixed(2)}px`)
  card.style.setProperty('--shift-y', `${((clampedY - 0.5) * 12).toFixed(2)}px`)
  card.style.setProperty('--surface-x', `${surfaceX.toFixed(2)}px`)
  card.style.setProperty('--surface-y', `${surfaceY.toFixed(2)}px`)
  card.style.setProperty('--surface-soft-x', `${(-surfaceX * 0.28).toFixed(2)}px`)
  card.style.setProperty('--surface-soft-y', `${(-surfaceY * 0.28).toFixed(2)}px`)
  card.style.setProperty('--surface-reverse-x', `${(-surfaceX * 0.4).toFixed(2)}px`)
  card.style.setProperty('--surface-reverse-y', `${(-surfaceY * 0.4).toFixed(2)}px`)
  card.style.setProperty('--surface-hard-x', `${(-surfaceX * 0.55).toFixed(2)}px`)
  card.style.setProperty('--surface-hard-y', `${(-surfaceY * 0.55).toFixed(2)}px`)
  card.style.setProperty('--tilt-x', `${((0.5 - clampedY) * 14).toFixed(2)}deg`)
  card.style.setProperty('--tilt-y', `${((clampedX - 0.5) * 16).toFixed(2)}deg`)
}

export default defineComponent({
  name: 'CardDesignsPage',
  setup() {
    return {
      designSections,
      effectSection,
      resetEffectCardPointer,
      updateEffectCardPointer,
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
            rgba(240, 163, 92, 0.18),
            transparent 28rem
        ),
        radial-gradient(
            circle at 82% 22%,
            rgba(171, 204, 222, 0.16),
            transparent 22rem
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

.design-sections {
    display: grid;
    gap: 3.5rem;
}

.design-section {
    display: grid;
    gap: 1.25rem;
}

.design-section__header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(260px, 430px);
    gap: 2rem;
    align-items: end;

    h2 {
        margin: 0;
        color: #fff4e6;
        font-size: 1.45rem;
        line-height: 1.1;
        font-weight: 800;
        letter-spacing: 0;
    }

    p {
        margin: 0;
        color: rgba(246, 238, 230, 0.68);
        font-size: 0.9rem;
        line-height: 1.55;
    }
}

.card-gallery {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1.5rem;
    align-items: start;
}

.card-gallery--effects {
    perspective: 1100px;
}

.design-panel {
    display: grid;
    gap: 1rem;
    justify-items: center;
}

.design-notes {
    width: min(280px, 100%);

    h3 {
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

.design-notes .design-number {
    width: 1.7rem;
    height: 1.7rem;
    display: grid;
    place-items: center;
    margin: 0 0 0.6rem;
    color: #1a100b;
    background: var(--card-accent);
    border-radius: 50%;
    font-size: 0.82rem;
    line-height: 1;
    font-weight: 900;
}

.effect-panel .design-number {
    background:
        radial-gradient(circle at 35% 22%, rgba(255, 255, 255, 0.82), transparent 31%),
        linear-gradient(135deg, #f8dd72, #8fe4ff 44%, #f39cff 67%, #ffc64f);
}

.concept-card {
    width: 280px;
    height: 420px;
    position: relative;
    overflow: hidden;
    color: var(--card-text);
    background:
        radial-gradient(
            circle at 50% 8%,
            rgba(255, 218, 170, 0.18),
            transparent 8rem
        ),
        linear-gradient(155deg, rgba(255, 255, 255, 0.12), transparent 31%),
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
        inset 0 0 0 1px rgba(255, 235, 201, 0.2),
        inset 0 0 0 7px rgba(66, 32, 18, 0.24),
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
            inset 0 0 0 1px rgba(255, 235, 201, 0.28),
            inset 0 0 0 7px rgba(66, 32, 18, 0.2),
            0 0 30px var(--card-shadow);
    }

    &::before {
        content: "";
        position: absolute;
        inset: 10px;
        border: 1px solid rgba(255, 224, 184, 0.36);
        border-radius: 8px;
        box-shadow:
            inset 0 0 0 1px rgba(112, 57, 28, 0.48),
            0 0 18px rgba(255, 161, 81, 0.1);
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
            var(--card-cut),
            var(--card-accent),
            var(--card-cut),
            transparent
        );
        box-shadow: 0 0 14px rgba(255, 172, 92, 0.38);
        opacity: 0.94;
        z-index: 3;
        transform: rotate(-2deg);
    }
}

.concept-card--non-rare {
    background:
        linear-gradient(155deg, rgba(255, 255, 255, 0.08), transparent 31%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            var(--card-bg) 58%,
            #120f0d 100%
        );
    box-shadow:
        0 18px 42px rgba(0, 0, 0, 0.42),
        0 0 18px var(--card-shadow);

    &:hover,
    &:focus-within {
        box-shadow:
            0 26px 56px rgba(0, 0, 0, 0.48),
            0 0 24px var(--card-shadow);
    }

    &::before {
        border-color: rgba(255, 232, 205, 0.18);
        box-shadow: none;
    }

    &::after {
        background: linear-gradient(
            90deg,
            transparent,
            var(--card-accent),
            transparent
        );
        box-shadow: none;
        opacity: 0.9;
    }

    .concept-card__rating-block {
        padding: 0;

        &::before {
            opacity: 0;
        }
    }

    .concept-card__rating {
        text-shadow: 0 2px 9px rgba(0, 0, 0, 0.42);
    }

    .concept-card__vitals {
        border-color: rgba(255, 230, 198, 0.28);
    }

    .concept-card__stat span {
        text-shadow: none;
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

.concept-card--polished-bronze {
    .concept-card__texture {
        background:
            repeating-radial-gradient(
                circle at 50% 12%,
                rgba(255, 229, 196, 0.08) 0 1px,
                transparent 1px 9px
            ),
            repeating-linear-gradient(
                112deg,
                rgba(255, 235, 211, 0.07) 0 1px,
                transparent 1px 12px
            ),
            linear-gradient(
                152deg,
                transparent 0 24%,
                rgba(255, 167, 84, 0.16) 24% 25%,
                transparent 25% 57%,
                rgba(0, 0, 0, 0.32) 57% 68%,
                transparent 68%
            ),
            radial-gradient(
                circle at 70% 18%,
                rgba(255, 180, 96, 0.24),
                transparent 7.5rem
            );
        opacity: 0.82;
    }
}

.concept-card--copper-starburst {
    background:
        radial-gradient(
            circle at 58% 30%,
            rgba(255, 167, 73, 0.18),
            transparent 9rem
        ),
        linear-gradient(110deg, rgba(255, 231, 200, 0.09), transparent 30%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 35%,
            var(--card-bg) 68%
        );

    .concept-card__texture {
        background:
            conic-gradient(
                from 212deg at 56% 32%,
                transparent 0deg,
                rgba(255, 189, 100, 0.18) 14deg,
                transparent 24deg,
                transparent 54deg,
                rgba(255, 142, 54, 0.13) 66deg,
                transparent 78deg,
                transparent 360deg
            ),
            repeating-linear-gradient(
                0deg,
                rgba(255, 238, 220, 0.04) 0 1px,
                transparent 1px 5px
            ),
            repeating-linear-gradient(
                96deg,
                transparent 0 17px,
                rgba(18, 12, 9, 0.24) 17px 18px
            ),
            linear-gradient(
                150deg,
                transparent 0 38%,
                rgba(255, 160, 69, 0.12) 38% 39%,
                rgba(16, 12, 10, 0.46) 39% 62%,
                transparent 62%
            ),
            radial-gradient(
                circle at 32% 7%,
                rgba(255, 196, 118, 0.15),
                transparent 7.5rem
            );
        mix-blend-mode: screen;
        opacity: 0.7;
    }
}

.concept-card--amber-facets {
    background:
        linear-gradient(145deg, rgba(255, 235, 210, 0.12), transparent 26%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 38%,
            var(--card-bg) 100%
        );

    .concept-card__texture {
        background:
            linear-gradient(
                135deg,
                transparent 0 16%,
                rgba(255, 194, 114, 0.18) 16% 17%,
                transparent 17% 42%,
                rgba(0, 0, 0, 0.28) 42% 55%,
                transparent 55%
            ),
            linear-gradient(
                38deg,
                transparent 0 31%,
                rgba(255, 216, 164, 0.16) 31% 32%,
                transparent 32% 64%,
                rgba(230, 130, 58, 0.11) 64% 65%,
                transparent 65%
            ),
            linear-gradient(
                162deg,
                transparent 0 48%,
                rgba(255, 171, 84, 0.17) 48% 49%,
                transparent 49%
            ),
            radial-gradient(
                circle at 68% 20%,
                rgba(255, 185, 99, 0.19),
                transparent 8rem
            ),
            repeating-linear-gradient(
                120deg,
                transparent 0 26px,
                rgba(255, 224, 194, 0.052) 26px 27px
            );
        opacity: 0.78;
    }
}

.concept-card--satin-steel,
.concept-card--graphite-aluminium,
.concept-card--pale-geometric,
.concept-card--polished-silver,
.concept-card--crystal-facet,
.concept-card--prismatic-edge {
    background:
        radial-gradient(
            circle at 50% 8%,
            rgba(240, 248, 252, 0.18),
            transparent 8rem
        ),
        linear-gradient(155deg, rgba(255, 255, 255, 0.16), transparent 31%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            var(--card-bg) 58%,
            #0c1117 100%
        );

    &::before {
        border-color: rgba(240, 248, 252, 0.34);
        box-shadow:
            inset 0 0 0 1px rgba(112, 129, 140, 0.42),
            0 0 18px rgba(214, 235, 244, 0.12);
    }

    &::after {
        background: linear-gradient(
            90deg,
            transparent,
            var(--card-cut),
            var(--card-accent),
            var(--card-cut),
            transparent
        );
        box-shadow: 0 0 14px rgba(210, 237, 247, 0.34);
    }

    .concept-card__rating-block::before {
        border-left-color: var(--card-cut);
        border-top-color: rgba(240, 248, 252, 0.5);
    }

    .concept-card__rating {
        text-shadow:
            0 2px 9px rgba(0, 0, 0, 0.48),
            0 0 14px rgba(214, 235, 244, 0.18);
    }

    .concept-card__vitals {
        color: #10161d;
        background: linear-gradient(180deg, var(--card-accent), var(--card-border-muted));
        border-color: rgba(240, 248, 252, 0.42);
    }

    .concept-card__flag,
    .concept-card__club {
        border-color: rgba(240, 248, 252, 0.28);
        background: rgba(11, 16, 22, 0.62);
    }

    .concept-card__club {
        color: var(--card-accent);
        box-shadow:
            0 7px 16px rgba(0, 0, 0, 0.25),
            inset 0 0 0 1px rgba(230, 242, 248, 0.18);
    }

    .concept-card__portrait {
        color: rgba(240, 248, 252, 0.84);
        background:
            radial-gradient(
                ellipse at 50% 64%,
                rgba(7, 12, 18, 0.46),
                transparent 56%
            ),
            radial-gradient(
                ellipse at 50% 30%,
                rgba(216, 236, 246, 0.14),
                transparent 48%
            ),
            linear-gradient(
                180deg,
                rgba(255, 255, 255, 0.07),
                rgba(255, 255, 255, 0)
            );
    }

    .concept-card__stat span {
        text-shadow: 0 0 12px rgba(220, 244, 252, 0.22);
    }
}

.concept-card--satin-steel,
.concept-card--graphite-aluminium,
.concept-card--pale-geometric {
    &::before {
        border-color: rgba(225, 234, 239, 0.2);
        box-shadow: none;
    }

    &::after {
        background: linear-gradient(
            90deg,
            transparent,
            rgba(210, 219, 224, 0.72),
            transparent
        );
        box-shadow: none;
    }

    .concept-card__rating {
        text-shadow: 0 2px 9px rgba(0, 0, 0, 0.44);
    }

    .concept-card__vitals {
        background: linear-gradient(180deg, var(--card-accent), #66727b);
        border-color: rgba(225, 234, 239, 0.26);
    }

    .concept-card__portrait {
        background:
            radial-gradient(
                ellipse at 50% 64%,
                rgba(7, 12, 18, 0.48),
                transparent 56%
            ),
            radial-gradient(
                ellipse at 50% 30%,
                rgba(192, 207, 216, 0.08),
                transparent 48%
            ),
            linear-gradient(
                180deg,
                rgba(255, 255, 255, 0.04),
                rgba(255, 255, 255, 0)
            );
    }

    .concept-card__stat span {
        text-shadow: none;
    }
}

.concept-card--satin-steel {
    background:
        linear-gradient(120deg, rgba(255, 255, 255, 0.08), transparent 27%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 38%,
            var(--card-bg) 72%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                0deg,
                rgba(255, 255, 255, 0.05) 0 1px,
                transparent 1px 5px
            ),
            repeating-linear-gradient(
                94deg,
                transparent 0 18px,
                rgba(8, 13, 18, 0.28) 18px 19px
            ),
            linear-gradient(
                150deg,
                transparent 0 39%,
                rgba(8, 13, 18, 0.5) 39% 61%,
                transparent 61%
            ),
            radial-gradient(
                circle at 56% 11%,
                rgba(205, 217, 224, 0.08),
                transparent 8.5rem
            );
        opacity: 0.56;
    }
}

.concept-card--graphite-aluminium {
    background:
        linear-gradient(112deg, rgba(255, 255, 255, 0.07), transparent 32%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            #29313a 48%,
            var(--card-bg) 100%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                90deg,
                rgba(240, 248, 252, 0.045) 0 1px,
                transparent 1px 7px
            ),
            repeating-linear-gradient(
                0deg,
                transparent 0 14px,
                rgba(12, 18, 24, 0.22) 14px 15px
            ),
            linear-gradient(
                145deg,
                transparent 0 28%,
                rgba(210, 226, 235, 0.1) 28% 29%,
                transparent 29% 56%,
                rgba(0, 0, 0, 0.25) 56% 70%,
                transparent 70%
            ),
            radial-gradient(
                circle at 74% 20%,
                rgba(184, 201, 211, 0.08),
                transparent 7.5rem
            );
        opacity: 0.58;
    }
}

.concept-card--pale-geometric {
    background:
        linear-gradient(145deg, rgba(255, 255, 255, 0.08), transparent 29%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 42%,
            var(--card-bg) 100%
        );

    .concept-card__texture {
        background:
            linear-gradient(
                136deg,
                transparent 0 17%,
                rgba(244, 250, 253, 0.13) 17% 18%,
                transparent 18% 43%,
                rgba(11, 17, 22, 0.26) 43% 57%,
                transparent 57%
            ),
            linear-gradient(
                42deg,
                transparent 0 33%,
                rgba(235, 244, 248, 0.12) 33% 34%,
                transparent 34% 66%
            ),
            repeating-linear-gradient(
                118deg,
                transparent 0 28px,
                rgba(244, 250, 253, 0.045) 28px 29px
            ),
            radial-gradient(
                circle at 64% 22%,
                rgba(200, 214, 222, 0.08),
                transparent 8rem
            );
        opacity: 0.6;
    }
}

.concept-card--polished-silver {
    background:
        conic-gradient(
            from 228deg at 58% 24%,
            rgba(255, 255, 255, 0) 0deg,
            rgba(255, 255, 255, 0.16) 18deg,
            rgba(87, 103, 115, 0.18) 38deg,
            rgba(255, 255, 255, 0) 58deg,
            rgba(255, 255, 255, 0) 126deg,
            rgba(232, 244, 250, 0.18) 144deg,
            rgba(255, 255, 255, 0) 168deg,
            rgba(255, 255, 255, 0) 360deg
        ),
        radial-gradient(
            circle at 66% 19%,
            rgba(255, 255, 255, 0.24),
            transparent 8rem
        ),
        linear-gradient(112deg, rgba(255, 255, 255, 0.18), transparent 29%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 36%,
            var(--card-bg) 68%
        );

    .concept-card__texture {
        background:
            linear-gradient(
                136deg,
                transparent 0 13%,
                rgba(255, 255, 255, 0.18) 13% 14%,
                transparent 14% 33%,
                rgba(61, 76, 88, 0.3) 33% 48%,
                transparent 48%
            ),
            linear-gradient(
                46deg,
                transparent 0 25%,
                rgba(238, 248, 252, 0.16) 25% 26%,
                transparent 26% 58%,
                rgba(255, 255, 255, 0.14) 58% 59%,
                transparent 59%
            ),
            repeating-radial-gradient(
                circle at 50% 12%,
                rgba(255, 255, 255, 0.09) 0 1px,
                transparent 1px 10px
            ),
            repeating-linear-gradient(
                113deg,
                rgba(255, 255, 255, 0.065) 0 1px,
                transparent 1px 13px
            ),
            linear-gradient(
                154deg,
                transparent 0 24%,
                rgba(255, 255, 255, 0.2) 24% 25%,
                transparent 25% 55%,
                rgba(0, 0, 0, 0.3) 55% 68%,
                transparent 68%
            );
        opacity: 0.8;
    }
}

.concept-card--crystal-facet {
    background:
        conic-gradient(
            from 188deg at 60% 28%,
            rgba(255, 255, 255, 0) 0deg,
            rgba(141, 236, 246, 0.16) 20deg,
            rgba(255, 255, 255, 0) 40deg,
            rgba(255, 255, 255, 0) 92deg,
            rgba(255, 255, 255, 0.18) 116deg,
            rgba(255, 255, 255, 0) 138deg,
            rgba(255, 255, 255, 0) 360deg
        ),
        radial-gradient(
            circle at 58% 31%,
            rgba(141, 236, 246, 0.18),
            transparent 9rem
        ),
        linear-gradient(115deg, rgba(255, 255, 255, 0.16), transparent 30%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 38%,
            var(--card-bg) 71%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                60deg,
                transparent 0 22px,
                rgba(217, 251, 255, 0.08) 22px 23px,
                transparent 23px 44px
            ),
            repeating-linear-gradient(
                120deg,
                transparent 0 24px,
                rgba(255, 255, 255, 0.07) 24px 25px,
                transparent 25px 48px
            ),
            linear-gradient(
                132deg,
                transparent 0 15%,
                rgba(255, 255, 255, 0.2) 15% 16%,
                transparent 16% 38%,
                rgba(141, 236, 246, 0.16) 38% 39%,
                rgba(8, 13, 18, 0.28) 39% 56%,
                transparent 56%
            ),
            linear-gradient(
                38deg,
                transparent 0 31%,
                rgba(238, 251, 255, 0.18) 31% 32%,
                transparent 32% 64%,
                rgba(141, 236, 246, 0.12) 64% 65%,
                transparent 65%
            ),
            radial-gradient(
                circle at 72% 18%,
                rgba(217, 251, 255, 0.2),
                transparent 7rem
            ),
            repeating-linear-gradient(
                118deg,
                transparent 0 24px,
                rgba(236, 250, 255, 0.055) 24px 25px
            );
        opacity: 0.78;
    }
}

.concept-card--prismatic-edge {
    background:
        conic-gradient(
            from 204deg at 62% 23%,
            transparent 0deg,
            rgba(191, 248, 255, 0.2) 18deg,
            transparent 30deg,
            transparent 74deg,
            rgba(255, 255, 255, 0.22) 86deg,
            transparent 96deg,
            transparent 148deg,
            rgba(218, 233, 241, 0.16) 166deg,
            transparent 182deg,
            transparent 360deg
        ),
        linear-gradient(112deg, rgba(255, 255, 255, 0.16), transparent 31%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 36%,
            var(--card-bg) 70%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                45deg,
                transparent 0 20px,
                rgba(191, 248, 255, 0.075) 20px 21px,
                transparent 21px 40px
            ),
            repeating-linear-gradient(
                135deg,
                transparent 0 24px,
                rgba(255, 255, 255, 0.07) 24px 25px,
                transparent 25px 48px
            ),
            linear-gradient(
                150deg,
                transparent 0 22%,
                rgba(191, 248, 255, 0.18) 22% 23%,
                transparent 23% 54%,
                rgba(8, 13, 18, 0.28) 54% 66%,
                transparent 66%
            ),
            linear-gradient(
                28deg,
                transparent 0 42%,
                rgba(255, 255, 255, 0.18) 42% 43%,
                transparent 43%
            ),
            repeating-linear-gradient(
                92deg,
                rgba(255, 255, 255, 0.05) 0 1px,
                transparent 1px 11px
            ),
            radial-gradient(
                circle at 30% 9%,
                rgba(245, 251, 255, 0.17),
                transparent 7.5rem
            );
        opacity: 0.76;
    }
}

.concept-card--antique-foil,
.concept-card--satin-sovereign,
.concept-card--ochre-panel,
.concept-card--royal-foil-burst,
.concept-card--jewel-trim,
.concept-card--crown-facets,
.concept-card--sunburst-vault,
.concept-card--damascus-foil,
.concept-card--black-gold-chevron,
.concept-card--quilted-ingot,
.concept-card--liquid-aurum,
.concept-card--art-deco-ribs,
.concept-card--confetti-foil,
.concept-card--circuit-gilt,
.concept-card--woven-laurel {
    background:
        radial-gradient(
            circle at 50% 8%,
            rgba(226, 190, 102, 0.16),
            transparent 8rem
        ),
        linear-gradient(155deg, rgba(238, 215, 142, 0.1), transparent 31%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            var(--card-bg) 58%,
            #110d07 100%
        );

    &::before {
        border-color: rgba(235, 209, 132, 0.32);
        box-shadow:
            inset 0 0 0 1px rgba(112, 80, 22, 0.5),
            0 0 18px rgba(228, 181, 67, 0.12);
    }

    &::after {
        background: linear-gradient(
            90deg,
            transparent,
            var(--card-cut),
            var(--card-accent),
            var(--card-cut),
            transparent
        );
        box-shadow: 0 0 14px rgba(232, 188, 76, 0.3);
    }

    .concept-card__rating-block::before {
        border-left-color: var(--card-cut);
        border-top-color: rgba(236, 214, 153, 0.48);
    }

    .concept-card__rating {
        text-shadow:
            0 2px 9px rgba(0, 0, 0, 0.52),
            0 0 14px rgba(218, 171, 60, 0.16);
    }

    .concept-card__vitals {
        color: #171006;
        background: linear-gradient(180deg, var(--card-accent), var(--card-border-muted));
        border-color: rgba(238, 216, 154, 0.38);
    }

    .concept-card__flag,
    .concept-card__club {
        border-color: rgba(234, 208, 130, 0.26);
        background: rgba(16, 12, 7, 0.64);
    }

    .concept-card__club {
        color: var(--card-accent);
        box-shadow:
            0 7px 16px rgba(0, 0, 0, 0.25),
            inset 0 0 0 1px rgba(232, 204, 128, 0.18);
    }

    .concept-card__portrait {
        color: rgba(238, 217, 155, 0.84);
        background:
            radial-gradient(
                ellipse at 50% 64%,
                rgba(11, 8, 4, 0.48),
                transparent 56%
            ),
            radial-gradient(
                ellipse at 50% 30%,
                rgba(220, 176, 66, 0.12),
                transparent 48%
            ),
            linear-gradient(
                180deg,
                rgba(255, 255, 255, 0.07),
                rgba(255, 255, 255, 0)
            );
    }

    .concept-card__stat span {
        text-shadow: 0 0 12px rgba(225, 178, 66, 0.2);
    }
}

.concept-card--antique-foil,
.concept-card--satin-sovereign,
.concept-card--ochre-panel {
    border-color: var(--card-border);
    box-shadow:
        0 18px 42px rgba(0, 0, 0, 0.42),
        0 0 13px var(--card-shadow);

    &:hover,
    &:focus-within {
        box-shadow:
            0 26px 56px rgba(0, 0, 0, 0.48),
            0 0 18px var(--card-shadow);
    }

    &::before {
        border-color: rgba(215, 191, 122, 0.18);
        box-shadow: none;
    }

    &::after {
        background: linear-gradient(
            90deg,
            transparent,
            rgba(205, 176, 100, 0.58),
            transparent
        );
        box-shadow: none;
        opacity: 0.68;
    }

    .concept-card__rating {
        text-shadow: 0 2px 9px rgba(0, 0, 0, 0.46);
    }

    .concept-card__vitals {
        background: linear-gradient(180deg, var(--card-accent), #625025);
        border-color: rgba(221, 198, 132, 0.24);
    }

    .concept-card__portrait {
        background:
            radial-gradient(
                ellipse at 50% 64%,
                rgba(11, 8, 4, 0.48),
                transparent 56%
            ),
            radial-gradient(
                ellipse at 50% 30%,
                rgba(190, 151, 57, 0.07),
                transparent 48%
            ),
            linear-gradient(
                180deg,
                rgba(255, 255, 255, 0.04),
                rgba(255, 255, 255, 0)
            );
    }

    .concept-card__stat span {
        text-shadow: none;
    }
}

.concept-card--antique-foil {
    background:
        linear-gradient(114deg, rgba(228, 203, 132, 0.055), transparent 29%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 38%,
            var(--card-bg) 72%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                0deg,
                rgba(230, 208, 148, 0.026) 0 1px,
                transparent 1px 6px
            ),
            repeating-linear-gradient(
                94deg,
                transparent 0 18px,
                rgba(16, 12, 7, 0.34) 18px 19px
            ),
            linear-gradient(
                150deg,
                transparent 0 40%,
                rgba(12, 9, 5, 0.52) 40% 62%,
                transparent 62%
            ),
            radial-gradient(
                circle at 61% 13%,
                rgba(210, 178, 92, 0.08),
                transparent 8rem
            );
        opacity: 0.52;
    }
}

.concept-card--satin-sovereign {
    background:
        linear-gradient(126deg, rgba(225, 205, 139, 0.06), transparent 31%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            #7d652f 42%,
            var(--card-bg) 100%
        );

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                92deg,
                rgba(226, 203, 135, 0.03) 0 1px,
                transparent 1px 10px
            ),
            linear-gradient(
                145deg,
                transparent 0 26%,
                rgba(219, 191, 111, 0.08) 26% 27%,
                transparent 27% 57%,
                rgba(0, 0, 0, 0.34) 57% 70%,
                transparent 70%
            ),
            linear-gradient(
                36deg,
                transparent 0 35%,
                rgba(220, 196, 126, 0.07) 35% 36%,
                transparent 36%
            ),
            radial-gradient(
                circle at 72% 22%,
                rgba(198, 166, 80, 0.07),
                transparent 8rem
            );
        opacity: 0.54;
    }
}

.concept-card--ochre-panel {
    background:
        linear-gradient(145deg, rgba(222, 193, 117, 0.045), transparent 27%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            #3d2f17 50%,
            var(--card-bg) 100%
        );

    .concept-card__texture {
        background:
            linear-gradient(
                134deg,
                transparent 0 18%,
                rgba(208, 179, 102, 0.08) 18% 19%,
                transparent 19% 43%,
                rgba(10, 8, 5, 0.38) 43% 58%,
                transparent 58%
            ),
            linear-gradient(
                45deg,
                transparent 0 34%,
                rgba(195, 160, 76, 0.08) 34% 35%,
                transparent 35% 65%
            ),
            repeating-linear-gradient(
                120deg,
                transparent 0 28px,
                rgba(225, 199, 128, 0.03) 28px 29px
            ),
            radial-gradient(
                circle at 66% 22%,
                rgba(181, 145, 63, 0.06),
                transparent 8rem
            );
        opacity: 0.56;
    }
}

.concept-card--royal-foil-burst {
    background:
        linear-gradient(
            180deg,
            rgba(255, 232, 132, 0.18) 0%,
            rgba(219, 167, 39, 0.08) 43%,
            rgba(33, 22, 6, 0.28) 66%,
            var(--card-bg) 100%
        ),
        conic-gradient(
            from 226deg at 15% 57%,
            transparent 0deg,
            rgba(255, 246, 187, 0.18) 10deg,
            transparent 18deg,
            transparent 34deg,
            rgba(255, 213, 84, 0.19) 44deg,
            transparent 52deg,
            transparent 70deg,
            rgba(255, 238, 150, 0.18) 81deg,
            transparent 90deg,
            transparent 360deg
        ),
        radial-gradient(
            circle at 24% 52%,
            rgba(255, 241, 169, 0.26),
            transparent 11rem
        ),
        linear-gradient(112deg, rgba(255, 246, 205, 0.18), transparent 30%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 36%,
            var(--card-bg) 70%
        );

    .concept-card__texture {
        background:
            repeating-conic-gradient(
                from 232deg at 14% 56%,
                rgba(255, 250, 205, 0.22) 0deg 1deg,
                rgba(107, 75, 18, 0.18) 1deg 2deg,
                transparent 2deg 7deg
            ),
            conic-gradient(
                from 228deg at 14% 56%,
                rgba(255, 255, 224, 0) 0deg,
                rgba(255, 244, 174, 0.28) 24deg,
                rgba(130, 90, 20, 0.18) 46deg,
                rgba(255, 236, 136, 0.2) 70deg,
                transparent 104deg,
                transparent 360deg
            ),
            repeating-linear-gradient(
                86deg,
                rgba(255, 243, 180, 0.04) 0 1px,
                transparent 1px 7px
            ),
            linear-gradient(
                180deg,
                transparent 0 60%,
                rgba(20, 13, 4, 0.32) 60% 100%
            ),
            radial-gradient(
                ellipse at 15% 56%,
                rgba(255, 248, 191, 0.26),
                transparent 12rem
            );
        opacity: 0.82;
    }
}

.concept-card--jewel-trim {
    background:
        linear-gradient(
            140deg,
            rgba(255, 248, 210, 0.22) 0 18%,
            transparent 18% 48%,
            rgba(121, 81, 14, 0.18) 48% 63%,
            transparent 63%
        ),
        linear-gradient(118deg, rgba(255, 248, 213, 0.2), transparent 30%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 37%,
            var(--card-bg) 70%
        );

    .concept-card__texture {
        background:
            linear-gradient(
                132deg,
                transparent 0 12%,
                rgba(255, 255, 230, 0.32) 12% 13%,
                rgba(125, 86, 15, 0.22) 13% 14%,
                transparent 14% 28%,
                rgba(255, 231, 117, 0.22) 28% 44%,
                rgba(111, 75, 13, 0.2) 44% 55%,
                transparent 55%
            ),
            linear-gradient(
                43deg,
                transparent 0 18%,
                rgba(255, 247, 198, 0.24) 18% 19%,
                rgba(109, 75, 16, 0.2) 19% 20%,
                transparent 20% 44%,
                rgba(255, 238, 152, 0.18) 44% 60%,
                transparent 60%
            ),
            linear-gradient(
                21deg,
                transparent 0 36%,
                rgba(92, 63, 12, 0.22) 36% 37%,
                rgba(255, 253, 222, 0.22) 37% 38%,
                transparent 38% 72%,
                rgba(255, 231, 122, 0.2) 72% 84%,
                transparent 84%
            ),
            linear-gradient(
                156deg,
                transparent 0 42%,
                rgba(255, 251, 220, 0.2) 42% 43%,
                transparent 43% 63%,
                rgba(92, 63, 12, 0.2) 63% 64%,
                transparent 64%
            ),
            repeating-linear-gradient(
                112deg,
                rgba(255, 244, 198, 0.052) 0 1px,
                transparent 1px 16px
            ),
            radial-gradient(
                circle at 65% 25%,
                rgba(255, 245, 196, 0.2),
                transparent 8rem
            );
        opacity: 0.84;
    }
}

.concept-card--crown-facets {
    background:
        radial-gradient(
            ellipse at 74% 30%,
            rgba(104, 74, 18, 0.2),
            transparent 9rem
        ),
        radial-gradient(
            circle at 78% 34%,
            rgba(255, 240, 150, 0.2),
            transparent 5rem
        ),
        linear-gradient(115deg, rgba(255, 244, 199, 0.14), transparent 30%),
        linear-gradient(
            180deg,
            var(--card-surface-soft) 0%,
            var(--card-surface) 38%,
            var(--card-bg) 72%
        );

    .concept-card__texture {
        background:
            radial-gradient(
                ellipse at 76% 29%,
                rgba(255, 237, 139, 0.2) 0 9%,
                rgba(91, 65, 16, 0.18) 10% 17%,
                transparent 18%
            ),
            radial-gradient(
                ellipse at 88% 45%,
                rgba(255, 232, 126, 0.18) 0 8%,
                rgba(91, 65, 16, 0.18) 9% 15%,
                transparent 16%
            ),
            radial-gradient(
                ellipse at 63% 46%,
                rgba(255, 224, 108, 0.16) 0 8%,
                rgba(91, 65, 16, 0.18) 9% 15%,
                transparent 16%
            ),
            radial-gradient(
                ellipse at 79% 58%,
                rgba(255, 235, 140, 0.15) 0 8%,
                rgba(91, 65, 16, 0.18) 9% 14%,
                transparent 15%
            ),
            linear-gradient(
                138deg,
                transparent 0 20%,
                rgba(79, 56, 14, 0.26) 20% 21%,
                rgba(255, 246, 194, 0.18) 21% 22%,
                transparent 22% 56%
            ),
            linear-gradient(
                31deg,
                transparent 0 42%,
                rgba(255, 247, 197, 0.16) 42% 43%,
                transparent 43% 64%
            ),
            repeating-radial-gradient(
                circle at 36% 21%,
                rgba(255, 250, 215, 0.045) 0 1px,
                transparent 1px 6px
            ),
            repeating-linear-gradient(
                102deg,
                transparent 0 24px,
                rgba(91, 65, 16, 0.07) 24px 25px
            );
        opacity: 0.84;
    }
}

.concept-card--sunburst-vault {
    background:
        radial-gradient(circle at 50% 43%, rgba(255, 248, 190, 0.26), transparent 10rem),
        conic-gradient(
            from -18deg at 50% 43%,
            rgba(255, 241, 148, 0.23) 0deg 10deg,
            rgba(100, 69, 14, 0.12) 10deg 18deg,
            rgba(255, 205, 67, 0.2) 18deg 29deg,
            transparent 29deg 41deg,
            rgba(255, 248, 196, 0.2) 41deg 52deg,
            rgba(96, 66, 12, 0.16) 52deg 62deg,
            rgba(255, 224, 103, 0.2) 62deg 78deg,
            transparent 78deg 360deg
        ),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 74%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(
                from -18deg at 50% 43%,
                rgba(255, 249, 202, 0.16) 0deg 1deg,
                transparent 1deg 7deg
            ),
            linear-gradient(180deg, transparent 0 58%, rgba(12, 8, 3, 0.44) 58% 100%);
        opacity: 0.84;
    }
}

.concept-card--damascus-foil {
    background:
        radial-gradient(ellipse at 50% 28%, rgba(255, 236, 143, 0.16), transparent 11rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 40%, var(--card-bg) 74%);

    .concept-card__texture {
        background:
            repeating-radial-gradient(
                ellipse at 18% 22%,
                rgba(255, 246, 178, 0.12) 0 2px,
                rgba(94, 63, 12, 0.16) 2px 4px,
                transparent 4px 12px
            ),
            repeating-radial-gradient(
                ellipse at 78% 34%,
                transparent 0 8px,
                rgba(255, 231, 116, 0.1) 8px 10px,
                rgba(70, 47, 10, 0.15) 10px 12px,
                transparent 12px 22px
            ),
            linear-gradient(180deg, transparent 0 61%, rgba(13, 9, 4, 0.42) 61% 100%);
        opacity: 0.7;
    }
}

.concept-card--black-gold-chevron {
    background:
        linear-gradient(180deg, rgba(255, 232, 128, 0.15), transparent 42%),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 38%, var(--card-bg) 74%);

    .concept-card__texture {
        background:
            repeating-linear-gradient(
                135deg,
                transparent 0 28px,
                rgba(10, 8, 4, 0.48) 28px 46px,
                rgba(255, 238, 142, 0.14) 46px 48px,
                transparent 48px 76px
            ),
            repeating-linear-gradient(
                45deg,
                transparent 0 34px,
                rgba(9, 7, 4, 0.38) 34px 52px,
                rgba(255, 246, 190, 0.1) 52px 54px,
                transparent 54px 88px
            ),
            linear-gradient(180deg, transparent 0 58%, rgba(7, 6, 4, 0.58) 58% 100%);
        opacity: 0.74;
    }
}

.concept-card--quilted-ingot {
    background:
        radial-gradient(circle at 50% 18%, rgba(255, 243, 171, 0.16), transparent 9rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 76%);

    .concept-card__texture {
        background:
            linear-gradient(45deg, rgba(255, 247, 184, 0.12) 25%, transparent 25% 75%, rgba(82, 56, 12, 0.18) 75%),
            linear-gradient(-45deg, rgba(86, 58, 12, 0.18) 25%, transparent 25% 75%, rgba(255, 246, 181, 0.12) 75%),
            repeating-linear-gradient(90deg, rgba(255, 243, 173, 0.035) 0 1px, transparent 1px 18px),
            linear-gradient(180deg, transparent 0 60%, rgba(12, 8, 3, 0.42) 60% 100%);
        background-size:
            56px 56px,
            56px 56px,
            auto,
            auto;
        opacity: 0.66;
    }
}

.concept-card--liquid-aurum {
    background:
        radial-gradient(ellipse at 28% 21%, rgba(255, 245, 174, 0.34), transparent 8rem),
        radial-gradient(ellipse at 74% 36%, rgba(255, 204, 70, 0.28), transparent 9rem),
        radial-gradient(ellipse at 43% 54%, rgba(108, 68, 9, 0.28), transparent 8rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 43%, var(--card-bg) 76%);

    .concept-card__texture {
        background:
            radial-gradient(ellipse at 23% 31%, rgba(255, 252, 205, 0.22) 0 18%, transparent 19%),
            radial-gradient(ellipse at 70% 26%, rgba(255, 235, 127, 0.19) 0 19%, transparent 20%),
            radial-gradient(ellipse at 58% 49%, rgba(87, 54, 8, 0.24) 0 20%, transparent 21%),
            repeating-linear-gradient(96deg, rgba(255, 242, 169, 0.035) 0 1px, transparent 1px 9px),
            linear-gradient(180deg, transparent 0 60%, rgba(13, 8, 3, 0.46) 60% 100%);
        opacity: 0.76;
    }
}

.concept-card--art-deco-ribs {
    background:
        radial-gradient(ellipse at 50% 17%, rgba(255, 242, 161, 0.18), transparent 9rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 40%, var(--card-bg) 74%);

    .concept-card__texture {
        background:
            linear-gradient(90deg, transparent 0 18%, rgba(255, 247, 188, 0.18) 18% 19%, transparent 19% 81%, rgba(255, 247, 188, 0.18) 81% 82%, transparent 82%),
            linear-gradient(90deg, transparent 0 27%, rgba(87, 60, 12, 0.22) 27% 28%, transparent 28% 72%, rgba(87, 60, 12, 0.22) 72% 73%, transparent 73%),
            linear-gradient(90deg, transparent 0 38%, rgba(255, 230, 118, 0.16) 38% 39%, transparent 39% 61%, rgba(255, 230, 118, 0.16) 61% 62%, transparent 62%),
            repeating-linear-gradient(0deg, transparent 0 18px, rgba(255, 244, 174, 0.055) 18px 19px),
            linear-gradient(180deg, transparent 0 61%, rgba(12, 8, 3, 0.44) 61% 100%);
        opacity: 0.78;
    }
}

.concept-card--confetti-foil {
    background:
        radial-gradient(circle at 47% 18%, rgba(255, 244, 177, 0.16), transparent 9rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 40%, var(--card-bg) 74%);

    .concept-card__texture {
        background:
            linear-gradient(24deg, transparent 0 15%, rgba(255, 250, 205, 0.2) 15% 17%, transparent 17% 47%, rgba(93, 64, 13, 0.18) 47% 49%, transparent 49%),
            linear-gradient(111deg, transparent 0 21%, rgba(255, 222, 93, 0.18) 21% 23%, transparent 23% 58%, rgba(255, 248, 194, 0.18) 58% 60%, transparent 60%),
            linear-gradient(152deg, transparent 0 33%, rgba(91, 63, 13, 0.18) 33% 35%, transparent 35% 73%, rgba(255, 236, 126, 0.16) 73% 75%, transparent 75%),
            repeating-linear-gradient(97deg, transparent 0 22px, rgba(255, 246, 187, 0.045) 22px 23px),
            linear-gradient(180deg, transparent 0 60%, rgba(12, 8, 3, 0.44) 60% 100%);
        opacity: 0.8;
    }
}

.concept-card--circuit-gilt {
    background:
        radial-gradient(circle at 64% 24%, rgba(255, 244, 177, 0.15), transparent 9rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 39%, var(--card-bg) 74%);

    .concept-card__texture {
        background:
            radial-gradient(circle at 28% 28%, rgba(255, 246, 184, 0.2) 0 2px, transparent 3px),
            radial-gradient(circle at 68% 36%, rgba(255, 246, 184, 0.18) 0 2px, transparent 3px),
            linear-gradient(90deg, transparent 0 18%, rgba(255, 239, 141, 0.16) 18% 19%, transparent 19% 45%, rgba(86, 59, 12, 0.2) 45% 46%, transparent 46%),
            linear-gradient(0deg, transparent 0 25%, rgba(255, 239, 141, 0.14) 25% 26%, transparent 26% 54%, rgba(86, 59, 12, 0.2) 54% 55%, transparent 55%),
            repeating-linear-gradient(90deg, transparent 0 32px, rgba(255, 245, 180, 0.05) 32px 33px),
            linear-gradient(180deg, transparent 0 60%, rgba(12, 8, 3, 0.46) 60% 100%);
        opacity: 0.78;
    }
}

.concept-card--woven-laurel {
    background:
        radial-gradient(ellipse at 50% 19%, rgba(255, 239, 150, 0.14), transparent 10rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 40%, var(--card-bg) 76%);

    .concept-card__texture {
        background:
            radial-gradient(ellipse at 24% 46%, transparent 0 34%, rgba(255, 236, 126, 0.14) 35% 37%, transparent 38%),
            radial-gradient(ellipse at 76% 46%, transparent 0 34%, rgba(255, 236, 126, 0.14) 35% 37%, transparent 38%),
            repeating-linear-gradient(45deg, rgba(255, 245, 181, 0.045) 0 2px, transparent 2px 9px),
            repeating-linear-gradient(135deg, rgba(72, 50, 12, 0.12) 0 2px, transparent 2px 10px),
            linear-gradient(180deg, transparent 0 60%, rgba(12, 8, 3, 0.48) 60% 100%);
        opacity: 0.7;
    }
}

.concept-card--effect-preview {
    --light-x: 36%;
    --light-y: 18%;
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

    transform: perspective(900px) translate3d(0, 0, 0) rotateX(0deg) rotateY(0deg);
    transform-style: preserve-3d;
    will-change: transform;

    &:hover,
    &:focus-within {
        transform:
            perspective(900px)
            translate3d(var(--shift-x), calc(-8px + var(--shift-y)), 0)
            rotateX(var(--tilt-x))
            rotateY(var(--tilt-y));
        box-shadow:
            0 32px 66px rgba(0, 0, 0, 0.55),
            inset 0 0 0 1px rgba(255, 246, 204, 0.34),
            inset 0 0 0 7px rgba(76, 45, 10, 0.18),
            0 0 36px var(--card-shadow);
    }

    .concept-card__texture::before,
    .concept-card__texture::after,
    .concept-card__content::before {
        content: "";
        position: absolute;
        inset: 0;
        pointer-events: none;
        opacity: 0;
        transition:
            opacity 180ms ease,
            background-position 180ms ease,
            transform 180ms ease;
    }

    .concept-card__texture::before,
    .concept-card__texture::after {
        z-index: 1;
    }

    .concept-card__content::before {
        z-index: 6;
        border-radius: 10px;
        mix-blend-mode: screen;
    }

    &:hover .concept-card__texture::before,
    &:hover .concept-card__texture::after,
    &:hover .concept-card__content::before,
    &:focus-within .concept-card__texture::before,
    &:focus-within .concept-card__texture::after,
    &:focus-within .concept-card__content::before {
        opacity: 1;
    }
}

.concept-card--effect-prism-holo {
    .concept-card__texture::before {
        background:
            radial-gradient(
                circle at var(--light-x) var(--light-y),
                rgba(255, 255, 255, 0.72),
                rgba(128, 241, 255, 0.38) 14%,
                rgba(255, 93, 231, 0.26) 24%,
                transparent 43%
            ),
            conic-gradient(
                from 135deg at var(--light-x) var(--light-y),
                rgba(255, 80, 152, 0.3),
                rgba(255, 222, 84, 0.25),
                rgba(93, 255, 179, 0.24),
                rgba(75, 202, 255, 0.3),
                rgba(174, 106, 255, 0.27),
                rgba(255, 80, 152, 0.3)
            );
        mix-blend-mode: color-dodge;
        transform: translate3d(var(--surface-x), var(--surface-y), 0);
    }

    .concept-card__texture::after {
        background:
            repeating-linear-gradient(
                116deg,
                transparent 0 10px,
                rgba(255, 255, 255, 0.16) 10px 11px,
                transparent 11px 22px
        );
        opacity: 0;
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-reverse-x), var(--surface-reverse-y), 0);
    }

    &:hover .concept-card__texture::after,
    &:focus-within .concept-card__texture::after {
        opacity: 0.72;
    }
}

.concept-card--effect-foil-flash {
    .concept-card__texture::before {
        background:
            radial-gradient(
                ellipse at var(--light-x) var(--light-y),
                rgba(255, 255, 241, 0.72),
                rgba(255, 225, 126, 0.3) 17%,
                transparent 37%
            ),
            repeating-linear-gradient(
                64deg,
                rgba(255, 255, 244, 0.2) 0 2px,
                transparent 2px 11px,
                rgba(81, 55, 10, 0.18) 11px 15px,
                transparent 15px 30px
            );
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-x), var(--surface-y), 0);
    }

    .concept-card__content::before {
        background:
            linear-gradient(
                105deg,
                transparent 0 31%,
                rgba(255, 255, 245, 0.72) 42%,
                rgba(255, 213, 93, 0.34) 48%,
                transparent 59% 100%
            );
        transform: translate3d(var(--surface-hard-x), var(--surface-hard-y), 0);
    }

    &:hover .concept-card__content::before,
    &:focus-within .concept-card__content::before {
        opacity: 0.46;
    }
}

.concept-card--effect-liquid-gold {
    .concept-card__texture::before {
        background:
            radial-gradient(
                ellipse at var(--light-x) var(--light-y),
                rgba(255, 250, 196, 0.82),
                rgba(255, 195, 56, 0.38) 19%,
                rgba(132, 76, 7, 0.18) 39%,
                transparent 58%
            ),
            radial-gradient(circle at 28% 23%, rgba(255, 232, 127, 0.24), transparent 9rem),
            linear-gradient(180deg, rgba(255, 217, 103, 0.12), transparent 62%);
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-x), var(--surface-y), 0);
    }

    .concept-card__texture::after {
        background:
            repeating-radial-gradient(
                ellipse at var(--light-x) var(--light-y),
                rgba(255, 245, 171, 0.18) 0 5px,
                transparent 5px 18px
            );
        mix-blend-mode: soft-light;
        transform: translate3d(var(--surface-soft-x), var(--surface-soft-y), 0);
    }
}

.concept-card--effect-rainbow-etch {
    .concept-card__texture::before {
        background:
            radial-gradient(
                circle at var(--light-x) var(--light-y),
                rgba(255, 255, 255, 0.58),
                rgba(107, 236, 255, 0.25) 13%,
                rgba(255, 111, 218, 0.2) 27%,
                transparent 45%
            ),
            repeating-linear-gradient(
                28deg,
                rgba(255, 110, 203, 0.11) 0 1px,
                transparent 1px 9px,
                rgba(98, 230, 255, 0.1) 9px 10px,
                transparent 10px 19px
            ),
            repeating-linear-gradient(
                118deg,
                transparent 0 17px,
                rgba(255, 248, 196, 0.13) 17px 18px
            );
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-x), var(--surface-y), 0);
    }

    .concept-card__content::before {
        background:
            radial-gradient(circle at 21% 34%, rgba(124, 234, 255, 0.35) 0 1px, transparent 2px),
            radial-gradient(circle at 66% 26%, rgba(255, 129, 215, 0.3) 0 1px, transparent 2px),
            radial-gradient(circle at 79% 58%, rgba(255, 244, 137, 0.34) 0 1px, transparent 2px),
            radial-gradient(circle at 37% 74%, rgba(144, 255, 190, 0.26) 0 1px, transparent 2px);
        background-size: 54px 72px;
        opacity: 0;
        transform: translate3d(var(--surface-reverse-x), var(--surface-reverse-y), 0);
    }

    &:hover .concept-card__content::before,
    &:focus-within .concept-card__content::before {
        opacity: 0.54;
    }
}

.concept-card--effect-mirror-gold {
    .concept-card__texture::before {
        background:
            linear-gradient(
                128deg,
                transparent 0 21%,
                rgba(255, 251, 217, 0.7) 32%,
                rgba(255, 198, 48, 0.34) 39%,
                transparent 51% 100%
            ),
            radial-gradient(
                circle at var(--light-x) var(--light-y),
                rgba(255, 244, 171, 0.48),
                transparent 34%
            );
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-x), var(--surface-y), 0);
    }

    .concept-card__content::before {
        background:
            radial-gradient(
                circle at var(--light-x) var(--light-y),
                rgba(255, 255, 238, 0.78),
                rgba(255, 213, 86, 0.32) 15%,
                transparent 34%
            );
        transform: translate3d(var(--surface-hard-x), var(--surface-hard-y), 0);
    }

    &:hover .concept-card__content::before,
    &:focus-within .concept-card__content::before {
        opacity: 0.5;
    }
}

.concept-card--effect-dark-holo {
    background:
        radial-gradient(circle at 47% 18%, rgba(255, 244, 177, 0.12), transparent 9rem),
        linear-gradient(180deg, #caa13f 0%, #8e6110 36%, #171006 74%);

    .concept-card__texture::before {
        background:
            radial-gradient(
                ellipse at var(--light-x) var(--light-y),
                rgba(255, 255, 255, 0.58),
                rgba(88, 217, 255, 0.25) 16%,
                rgba(178, 103, 255, 0.2) 31%,
                transparent 52%
            ),
            repeating-linear-gradient(
                132deg,
                rgba(6, 5, 6, 0.42) 0 12px,
                rgba(255, 224, 97, 0.11) 12px 14px,
                transparent 14px 28px
            );
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-x), var(--surface-y), 0);
    }

    .concept-card__texture::after {
        background:
            linear-gradient(180deg, rgba(4, 4, 6, 0.18), rgba(4, 4, 6, 0.54)),
            conic-gradient(
                from 215deg at var(--light-x) var(--light-y),
                transparent,
                rgba(77, 231, 255, 0.22),
                rgba(255, 113, 218, 0.18),
                transparent 38%
            );
        mix-blend-mode: screen;
        transform: translate3d(var(--surface-reverse-x), var(--surface-reverse-y), 0);
    }
}

.concept-card--totw {
    border-color: #080808;
    background:
        radial-gradient(
            ellipse at 52% 32%,
            rgba(0, 0, 0, 0.28),
            transparent 11rem
        ),
        radial-gradient(
            circle at 50% 18%,
            color-mix(in srgb, var(--card-accent) 26%, transparent),
            transparent 9rem
        ),
        conic-gradient(
            from 220deg at 56% 24%,
            transparent 0deg,
            color-mix(in srgb, var(--card-cut) 16%, transparent) 16deg,
            transparent 34deg,
            transparent 72deg,
            color-mix(in srgb, var(--card-accent) 12%, transparent) 88deg,
            transparent 108deg,
            transparent 360deg
        ),
        linear-gradient(145deg, rgba(255, 255, 255, 0.035), transparent 24%),
        linear-gradient(
            180deg,
            var(--card-surface) 0%,
            var(--card-bg) 54%,
            #030303 100%
        );
    box-shadow:
        0 20px 46px rgba(0, 0, 0, 0.56),
        inset 0 0 0 1px rgba(255, 255, 255, 0.08),
        inset 0 0 0 7px rgba(0, 0, 0, 0.48),
        0 0 27px var(--card-shadow);

    &:hover,
    &:focus-within {
        box-shadow:
            0 29px 60px rgba(0, 0, 0, 0.6),
            inset 0 0 0 1px rgba(255, 255, 255, 0.14),
            inset 0 0 0 7px rgba(0, 0, 0, 0.4),
            0 0 38px var(--card-shadow);
    }

    &::before {
        inset: 9px;
        border-color: var(--card-border);
        box-shadow:
            inset 0 0 0 1px rgba(0, 0, 0, 0.72),
            0 0 20px var(--card-shadow);
    }

    &::after {
        top: 280px;
        height: 4px;
        background: linear-gradient(
            90deg,
            transparent,
            var(--card-cut),
            var(--card-accent),
            var(--card-cut),
            transparent
        );
        box-shadow: 0 0 18px var(--card-shadow);
    }

    .concept-card__rating-block::before {
        border-left-color: var(--card-cut);
        border-top-color: color-mix(in srgb, var(--card-highlight) 52%, transparent);
    }

    .concept-card__vitals {
        color: #080808;
        background: linear-gradient(180deg, var(--card-cut), var(--card-accent));
        border-color: color-mix(in srgb, var(--card-highlight) 42%, transparent);
    }

    .concept-card__flag,
    .concept-card__club {
        background: rgba(5, 5, 5, 0.7);
        border-color: color-mix(in srgb, var(--card-border) 34%, transparent);
    }

    .concept-card__portrait {
        color: color-mix(in srgb, var(--card-highlight) 76%, transparent);
        background:
            radial-gradient(
                ellipse at 50% 64%,
                rgba(0, 0, 0, 0.54),
                transparent 56%
            ),
            radial-gradient(
                ellipse at 52% 28%,
                color-mix(in srgb, var(--card-accent) 18%, transparent),
                transparent 48%
            );
    }
}

.concept-card--bronze-week-strike {
    background:
        radial-gradient(circle at 68% 24%, rgba(255, 163, 80, 0.16), transparent 7rem),
        conic-gradient(from 230deg at 56% 28%, transparent 0deg, rgba(255, 187, 110, 0.2) 18deg, rgba(77, 38, 18, 0.32) 35deg, transparent 56deg, transparent 360deg),
        linear-gradient(135deg, rgba(255, 162, 80, 0.08) 0 14%, transparent 14% 45%, rgba(255, 173, 86, 0.06) 45% 58%, transparent 58%),
        linear-gradient(180deg, #100a07 0%, var(--card-bg) 58%, #030303 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 224deg at 58% 28%, rgba(255, 172, 88, 0.18) 0deg 1deg, transparent 1deg 7deg),
            repeating-linear-gradient(116deg, transparent 0 16px, rgba(255, 154, 77, 0.12) 16px 18px, transparent 18px 34px),
            linear-gradient(142deg, transparent 0 38%, rgba(255, 186, 105, 0.18) 38% 40%, rgba(0, 0, 0, 0.26) 40% 55%, transparent 55%),
            radial-gradient(circle at 68% 34%, rgba(255, 163, 80, 0.18), transparent 8rem);
        opacity: 0.84;
    }
}

.concept-card--bronze-week-signal {
    background:
        radial-gradient(ellipse at 48% 30%, rgba(223, 139, 73, 0.16), transparent 8rem),
        linear-gradient(128deg, rgba(223, 139, 73, 0.06), transparent 26%),
        linear-gradient(180deg, #0e0907 0%, var(--card-bg) 60%, #030303 100%);

    .concept-card__texture {
        background:
            repeating-radial-gradient(ellipse at 50% 34%, rgba(255, 174, 92, 0.18) 0 1px, transparent 1px 8px),
            repeating-radial-gradient(ellipse at 50% 34%, transparent 0 18px, rgba(226, 139, 73, 0.13) 18px 21px, transparent 21px 34px),
            repeating-linear-gradient(90deg, transparent 0 14px, rgba(226, 139, 73, 0.1) 14px 16px, transparent 16px 29px),
            linear-gradient(180deg, transparent 0 58%, rgba(0, 0, 0, 0.46) 58% 100%);
        opacity: 0.78;
    }
}

.concept-card--bronze-week-spotlight {
    background:
        radial-gradient(ellipse at 60% 24%, rgba(255, 177, 91, 0.2), transparent 9rem),
        conic-gradient(from 210deg at 58% 24%, transparent 0deg, rgba(255, 189, 112, 0.22) 18deg, rgba(63, 31, 14, 0.34) 42deg, transparent 68deg, transparent 360deg),
        linear-gradient(152deg, transparent 0 24%, rgba(242, 154, 75, 0.16) 24% 43%, transparent 43%),
        linear-gradient(180deg, #120b07 0%, var(--card-bg) 62%, #030303 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 214deg at 58% 24%, rgba(255, 189, 112, 0.12) 0deg 1deg, transparent 1deg 8deg),
            radial-gradient(circle at 76% 18%, rgba(255, 207, 141, 0.2), transparent 3.5rem),
            repeating-linear-gradient(105deg, transparent 0 22px, rgba(255, 167, 84, 0.08) 22px 24px),
            linear-gradient(180deg, transparent 0 58%, rgba(0, 0, 0, 0.4) 58% 100%);
        opacity: 0.82;
    }
}

.concept-card--silver-week-beam,
.concept-card--silver-week-shards,
.concept-card--silver-week-lights {
    .concept-card__rating {
        text-shadow:
            0 2px 9px rgba(0, 0, 0, 0.52),
            0 0 15px rgba(205, 244, 255, 0.22);
    }
}

.concept-card--silver-week-beam {
    background:
        radial-gradient(circle at 66% 22%, rgba(223, 247, 255, 0.16), transparent 7rem),
        conic-gradient(from 226deg at 54% 20%, transparent 0deg, rgba(238, 250, 255, 0.18) 14deg, rgba(22, 35, 46, 0.44) 30deg, transparent 52deg, transparent 360deg),
        linear-gradient(180deg, #080d13 0%, var(--card-bg) 60%, #020304 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 226deg at 54% 20%, rgba(234, 250, 255, 0.13) 0deg 1deg, transparent 1deg 8deg),
            linear-gradient(117deg, transparent 0 18%, rgba(234, 250, 255, 0.2) 18% 20%, transparent 20% 38%, rgba(184, 236, 248, 0.14) 38% 41%, transparent 41%),
            radial-gradient(circle at 66% 22%, rgba(223, 247, 255, 0.14), transparent 8rem),
            repeating-linear-gradient(96deg, transparent 0 14px, rgba(229, 245, 250, 0.05) 14px 15px);
        opacity: 0.86;
    }
}

.concept-card--silver-week-shards {
    background:
        radial-gradient(ellipse at 58% 26%, rgba(207, 238, 246, 0.14), transparent 8rem),
        linear-gradient(140deg, rgba(255, 255, 255, 0.04), transparent 24%),
        linear-gradient(180deg, #081018 0%, var(--card-bg) 58%, #020304 100%);

    .concept-card__texture {
        background:
            linear-gradient(139deg, transparent 0 13%, rgba(238, 249, 252, 0.22) 13% 29%, transparent 29% 45%, rgba(153, 222, 236, 0.16) 45% 61%, transparent 61%),
            linear-gradient(41deg, transparent 0 34%, rgba(240, 252, 255, 0.17) 34% 36%, rgba(8, 14, 20, 0.38) 36% 52%, transparent 52%),
            repeating-linear-gradient(112deg, transparent 0 22px, rgba(200, 232, 239, 0.08) 22px 24px),
            repeating-linear-gradient(66deg, transparent 0 30px, rgba(255, 255, 255, 0.045) 30px 31px);
        opacity: 0.84;
    }
}

.concept-card--silver-week-lights {
    background:
        radial-gradient(ellipse at 50% 0%, rgba(244, 252, 255, 0.16), transparent 9rem),
        radial-gradient(circle at 72% 24%, rgba(195, 241, 251, 0.12), transparent 6rem),
        linear-gradient(180deg, #070c12 0%, var(--card-bg) 62%, #020304 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 228deg at 50% 0%, rgba(246, 253, 255, 0.12) 0deg 1deg, transparent 1deg 10deg),
            conic-gradient(from 228deg at 50% 0%, transparent 0deg, rgba(246, 253, 255, 0.2) 13deg, transparent 26deg, transparent 44deg, rgba(195, 241, 251, 0.16) 55deg, transparent 68deg, transparent 360deg),
            radial-gradient(ellipse at 50% 0%, rgba(244, 252, 255, 0.16), transparent 10rem),
            linear-gradient(180deg, transparent 0 58%, rgba(0, 0, 0, 0.44) 58% 100%);
        opacity: 0.86;
    }
}

.concept-card--gold-week-headline,
.concept-card--gold-week-flash,
.concept-card--gold-week-strokes {
    .concept-card__rating {
        text-shadow:
            0 2px 9px rgba(0, 0, 0, 0.54),
            0 0 17px rgba(255, 213, 101, 0.24);
    }
}

.concept-card--gold-week-headline {
    background:
        radial-gradient(circle at 54% 28%, rgba(255, 216, 101, 0.16), transparent 8rem),
        conic-gradient(from 220deg at 52% 28%, transparent 0deg, rgba(255, 229, 132, 0.22) 18deg, rgba(61, 43, 10, 0.46) 36deg, transparent 60deg, transparent 360deg),
        linear-gradient(180deg, #0e0a04 0%, var(--card-bg) 60%, #020201 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 220deg at 52% 28%, rgba(255, 229, 132, 0.18) 0deg 1deg, transparent 1deg 7deg),
            conic-gradient(from 218deg at 52% 28%, transparent 0deg, rgba(255, 216, 101, 0.24) 22deg, rgba(0, 0, 0, 0.34) 46deg, transparent 70deg, transparent 360deg),
            repeating-linear-gradient(102deg, transparent 0 18px, rgba(255, 207, 84, 0.08) 18px 19px),
            linear-gradient(180deg, transparent 0 58%, rgba(0, 0, 0, 0.46) 58% 100%);
        opacity: 0.88;
    }
}

.concept-card--gold-week-flash {
    background:
        radial-gradient(ellipse at 62% 24%, rgba(255, 199, 78, 0.16), transparent 8rem),
        linear-gradient(128deg, rgba(255, 236, 161, 0.12) 0 12%, transparent 12% 42%, rgba(255, 199, 78, 0.12) 42% 50%, transparent 50%),
        linear-gradient(180deg, #0f0a04 0%, var(--card-bg) 58%, #020201 100%);

    .concept-card__texture {
        background:
            linear-gradient(128deg, transparent 0 10%, rgba(255, 236, 161, 0.28) 10% 15%, transparent 15% 32%, rgba(255, 199, 78, 0.18) 32% 48%, rgba(0, 0, 0, 0.3) 48% 57%, transparent 57%),
            linear-gradient(40deg, transparent 0 48%, rgba(255, 227, 130, 0.17) 48% 51%, transparent 51%),
            repeating-linear-gradient(96deg, rgba(255, 224, 127, 0.05) 0 1px, transparent 1px 8px),
            repeating-linear-gradient(142deg, transparent 0 26px, rgba(255, 210, 96, 0.08) 26px 28px);
        opacity: 0.88;
    }
}

.concept-card--gold-week-strokes {
    background:
        radial-gradient(circle at 74% 24%, rgba(255, 232, 142, 0.16), transparent 6rem),
        linear-gradient(112deg, transparent 0 14%, rgba(255, 214, 92, 0.16) 14% 21%, transparent 21% 38%, rgba(255, 238, 166, 0.1) 38% 42%, transparent 42%),
        linear-gradient(180deg, #100b04 0%, var(--card-bg) 60%, #020201 100%);

    .concept-card__texture {
        background:
            linear-gradient(112deg, transparent 0 15%, rgba(255, 214, 92, 0.24) 15% 20%, rgba(0, 0, 0, 0.28) 20% 31%, transparent 31% 40%, rgba(255, 238, 166, 0.17) 40% 43%, transparent 43%),
            radial-gradient(circle at 74% 24%, rgba(255, 232, 142, 0.18), transparent 6rem),
            repeating-linear-gradient(118deg, transparent 0 18px, rgba(255, 205, 78, 0.09) 18px 20px),
            repeating-radial-gradient(circle at 72% 22%, rgba(255, 236, 162, 0.08) 0 1px, transparent 1px 7px);
        opacity: 0.88;
    }
}

.concept-card--tots {
    border-color: var(--card-border);
    background:
        radial-gradient(
            circle at 50% 8%,
            color-mix(in srgb, var(--card-surface-soft) 24%, transparent),
            transparent 9rem
        ),
        radial-gradient(
            ellipse at 72% 24%,
            color-mix(in srgb, var(--card-accent) 16%, transparent),
            transparent 8rem
        ),
        conic-gradient(
            from 225deg at 52% 28%,
            transparent 0deg,
            color-mix(in srgb, var(--card-cut) 14%, transparent) 20deg,
            transparent 42deg,
            transparent 360deg
        ),
        linear-gradient(140deg, color-mix(in srgb, var(--card-cut) 12%, transparent), transparent 34%),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 36%, var(--card-bg) 70%, #050a0e 100%);
    box-shadow:
        0 19px 44px rgba(0, 0, 0, 0.48),
        inset 0 0 0 1px color-mix(in srgb, var(--card-highlight) 18%, transparent),
        inset 0 0 0 7px rgba(8, 22, 34, 0.36),
        0 0 28px var(--card-shadow);

    &:hover,
    &:focus-within {
        box-shadow:
            0 28px 58px rgba(0, 0, 0, 0.54),
            inset 0 0 0 1px color-mix(in srgb, var(--card-highlight) 24%, transparent),
            inset 0 0 0 7px rgba(8, 22, 34, 0.3),
            0 0 38px var(--card-shadow);
    }

    &::before {
        inset: 8px;
        border-color: color-mix(in srgb, var(--card-border) 72%, transparent);
        box-shadow:
            inset 0 0 0 1px color-mix(in srgb, var(--card-accent) 22%, transparent),
            0 0 22px var(--card-shadow);
    }

    &::after {
        background: linear-gradient(
            90deg,
            transparent,
            var(--card-accent),
            var(--card-cut),
            var(--card-accent),
            transparent
        );
        box-shadow: 0 0 18px var(--card-shadow);
    }

    .concept-card__vitals {
        color: #071017;
        background: linear-gradient(180deg, var(--card-cut), var(--card-accent));
        border-color: color-mix(in srgb, var(--card-highlight) 38%, transparent);
    }

    .concept-card__portrait {
        color: color-mix(in srgb, var(--card-highlight) 78%, transparent);
        background:
            radial-gradient(ellipse at 50% 64%, rgba(0, 0, 0, 0.48), transparent 56%),
            radial-gradient(ellipse at 50% 28%, color-mix(in srgb, var(--card-accent) 17%, transparent), transparent 48%);
    }
}

.concept-card--bronze-season-laurel,
.concept-card--bronze-season-ribbon,
.concept-card--bronze-season-engraved {
    .concept-card__rating {
        text-shadow:
            0 2px 9px rgba(0, 0, 0, 0.52),
            0 0 15px rgba(242, 176, 107, 0.18);
    }
}

.concept-card--bronze-season-laurel {
    background:
        radial-gradient(circle at 44% 12%, rgba(242, 176, 107, 0.24), transparent 8rem),
        radial-gradient(circle at 70% 24%, rgba(47, 196, 180, 0.14), transparent 7rem),
        linear-gradient(132deg, rgba(242, 176, 107, 0.16), transparent 26%),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 76%);

    .concept-card__texture {
        background:
            repeating-radial-gradient(ellipse at 50% 24%, transparent 0 16px, rgba(242, 176, 107, 0.16) 16px 18px, transparent 18px 34px),
            repeating-linear-gradient(106deg, transparent 0 20px, rgba(255, 217, 181, 0.07) 20px 21px),
            linear-gradient(118deg, transparent 0 26%, rgba(47, 196, 180, 0.16) 26% 29%, rgba(65, 38, 22, 0.24) 29% 42%, transparent 42%),
            radial-gradient(circle at 68% 18%, rgba(47, 196, 180, 0.14), transparent 7rem);
        opacity: 0.78;
    }
}

.concept-card--bronze-season-ribbon {
    background:
        radial-gradient(circle at 36% 12%, rgba(239, 171, 103, 0.22), transparent 8rem),
        linear-gradient(125deg, transparent 0 18%, rgba(53, 191, 169, 0.14) 18% 32%, transparent 32% 54%, rgba(239, 171, 103, 0.16) 54% 62%, transparent 62%),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 44%, var(--card-bg) 76%);

    .concept-card__texture {
        background:
            linear-gradient(125deg, transparent 0 18%, rgba(53, 191, 169, 0.2) 18% 32%, transparent 32% 52%, rgba(239, 171, 103, 0.18) 52% 61%, transparent 61%),
            linear-gradient(36deg, transparent 0 38%, rgba(53, 191, 169, 0.16) 38% 43%, rgba(64, 39, 22, 0.22) 43% 52%, transparent 52%),
            repeating-linear-gradient(100deg, transparent 0 18px, rgba(255, 217, 181, 0.08) 18px 19px);
        opacity: 0.78;
    }
}

.concept-card--bronze-season-engraved {
    background:
        radial-gradient(circle at 42% 14%, rgba(243, 179, 113, 0.24), transparent 8rem),
        radial-gradient(circle at 70% 24%, rgba(73, 201, 142, 0.14), transparent 7rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 76%);

    .concept-card__texture {
        background:
            repeating-linear-gradient(92deg, transparent 0 10px, rgba(255, 220, 184, 0.09) 10px 11px),
            repeating-radial-gradient(circle at 66% 22%, rgba(73, 201, 142, 0.15) 0 1px, transparent 1px 8px),
            repeating-radial-gradient(circle at 34% 22%, rgba(255, 220, 184, 0.07) 0 1px, transparent 1px 7px),
            linear-gradient(145deg, transparent 0 32%, rgba(243, 179, 113, 0.17) 32% 35%, rgba(58, 35, 20, 0.25) 35% 50%, transparent 50%);
        opacity: 0.78;
    }
}

.concept-card--silver-season-crest {
    background:
        radial-gradient(circle at 42% 12%, rgba(245, 255, 255, 0.32), transparent 8rem),
        radial-gradient(circle at 72% 20%, rgba(77, 191, 231, 0.14), transparent 8rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 78%);

    .concept-card__texture {
        background:
            linear-gradient(135deg, transparent 0 15%, rgba(245, 255, 255, 0.22) 15% 18%, transparent 18% 41%, rgba(77, 191, 231, 0.15) 41% 59%, transparent 59%),
            repeating-radial-gradient(ellipse at 50% 24%, transparent 0 18px, rgba(238, 246, 248, 0.14) 18px 20px, transparent 20px 38px),
            repeating-linear-gradient(98deg, rgba(255, 255, 255, 0.06) 0 1px, transparent 1px 9px),
            radial-gradient(circle at 72% 20%, rgba(77, 191, 231, 0.14), transparent 8rem);
        opacity: 0.82;
    }
}

.concept-card--silver-season-wave {
    background:
        radial-gradient(circle at 42% 12%, rgba(241, 251, 255, 0.28), transparent 8rem),
        linear-gradient(126deg, transparent 0 18%, rgba(77, 141, 255, 0.15) 18% 33%, transparent 33% 52%, rgba(241, 251, 255, 0.16) 52% 56%, transparent 56%),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 78%);

    .concept-card__texture {
        background:
            linear-gradient(126deg, transparent 0 18%, rgba(77, 141, 255, 0.2) 18% 34%, transparent 34% 50%, rgba(241, 251, 255, 0.18) 50% 54%, rgba(57, 73, 89, 0.24) 54% 63%, transparent 63%),
            repeating-linear-gradient(107deg, transparent 0 18px, rgba(229, 238, 242, 0.08) 18px 19px),
            repeating-linear-gradient(24deg, rgba(255, 255, 255, 0.045) 0 1px, transparent 1px 12px),
            radial-gradient(circle at 70% 22%, rgba(77, 141, 255, 0.16), transparent 8rem);
        opacity: 0.82;
    }
}

.concept-card--silver-season-laurel {
    background:
        radial-gradient(circle at 42% 12%, rgba(240, 253, 255, 0.3), transparent 8rem),
        radial-gradient(circle at 66% 20%, rgba(111, 224, 238, 0.14), transparent 8rem),
        linear-gradient(180deg, var(--card-surface-soft) 0%, var(--card-surface) 42%, var(--card-bg) 78%);

    .concept-card__texture {
        background:
            repeating-radial-gradient(ellipse at 50% 28%, transparent 0 15px, rgba(240, 253, 255, 0.16) 15px 17px, transparent 17px 34px),
            repeating-linear-gradient(105deg, transparent 0 20px, rgba(255, 255, 255, 0.08) 20px 21px),
            linear-gradient(42deg, transparent 0 34%, rgba(111, 224, 238, 0.16) 34% 39%, rgba(62, 78, 88, 0.22) 39% 49%, transparent 49%),
            radial-gradient(circle at 64% 17%, rgba(111, 224, 238, 0.16), transparent 8rem);
        opacity: 0.82;
    }
}

.concept-card--gold-season-crown,
.concept-card--gold-season-celebration,
.concept-card--gold-season-trophy {
    .concept-card__rating {
        text-shadow:
            0 2px 9px rgba(0, 0, 0, 0.54),
            0 0 18px rgba(255, 218, 105, 0.28);
    }
}

.concept-card--gold-season-crown {
    background:
        radial-gradient(circle at 18% 13%, rgba(255, 225, 125, 0.22), transparent 4.5rem),
        radial-gradient(circle at 80% 17%, rgba(76, 125, 255, 0.22), transparent 6rem),
        radial-gradient(circle at 56% 30%, rgba(255, 225, 125, 0.2), transparent 8.5rem),
        conic-gradient(from 225deg at 50% 28%, transparent 0deg, rgba(255, 225, 125, 0.26) 18deg, rgba(28, 43, 91, 0.35) 38deg, rgba(76, 125, 255, 0.2) 66deg, transparent 92deg, transparent 360deg),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 58%, #050913 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 225deg at 50% 28%, rgba(255, 225, 125, 0.16) 0deg 1deg, transparent 1deg 8deg),
            conic-gradient(from 225deg at 50% 28%, transparent 0deg, rgba(255, 225, 125, 0.22) 18deg, transparent 35deg, transparent 56deg, rgba(76, 125, 255, 0.18) 70deg, transparent 88deg, transparent 360deg),
            linear-gradient(137deg, transparent 0 18%, rgba(255, 241, 166, 0.2) 18% 21%, rgba(17, 29, 62, 0.28) 21% 35%, transparent 35%),
            repeating-linear-gradient(104deg, transparent 0 18px, rgba(255, 225, 125, 0.09) 18px 20px),
            repeating-radial-gradient(circle at 76% 18%, rgba(255, 240, 172, 0.08) 0 1px, transparent 1px 7px);
        opacity: 0.9;
    }
}

.concept-card--gold-season-celebration {
    background:
        radial-gradient(circle at 18% 12%, rgba(255, 217, 111, 0.28), transparent 4rem),
        radial-gradient(circle at 78% 18%, rgba(93, 141, 255, 0.24), transparent 5rem),
        radial-gradient(circle at 58% 32%, rgba(255, 217, 111, 0.18), transparent 9rem),
        conic-gradient(from 228deg at 52% 24%, transparent 0deg, rgba(255, 236, 161, 0.2) 15deg, rgba(93, 141, 255, 0.16) 33deg, transparent 58deg, transparent 360deg),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 58%, #050913 100%);

    .concept-card__texture {
        background:
            repeating-conic-gradient(from 230deg at 50% 18%, rgba(255, 236, 161, 0.16) 0deg 1deg, transparent 1deg 8deg),
            repeating-radial-gradient(circle at 21% 14%, rgba(255, 237, 156, 0.12) 0 1px, transparent 1px 8px),
            linear-gradient(126deg, transparent 0 18%, rgba(255, 224, 124, 0.16) 18% 24%, transparent 24% 48%, rgba(93, 141, 255, 0.14) 48% 62%, transparent 62%),
            radial-gradient(circle at 18% 12%, rgba(255, 217, 111, 0.2), transparent 4rem),
            radial-gradient(circle at 78% 18%, rgba(93, 141, 255, 0.18), transparent 5rem);
        opacity: 0.9;
    }
}

.concept-card--gold-season-trophy {
    background:
        radial-gradient(circle at 50% 18%, rgba(255, 228, 134, 0.24), transparent 8rem),
        radial-gradient(circle at 72% 18%, rgba(111, 149, 255, 0.2), transparent 7rem),
        conic-gradient(from 218deg at 52% 28%, transparent 0deg, rgba(255, 242, 169, 0.24) 18deg, rgba(24, 38, 82, 0.38) 38deg, rgba(111, 149, 255, 0.16) 62deg, transparent 90deg, transparent 360deg),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 58%, #050913 100%);

    .concept-card__texture {
        background:
            repeating-radial-gradient(ellipse at 50% 24%, transparent 0 13px, rgba(255, 228, 134, 0.18) 13px 16px, transparent 16px 31px),
            repeating-conic-gradient(from 218deg at 52% 28%, rgba(255, 242, 169, 0.12) 0deg 1deg, transparent 1deg 8deg),
            linear-gradient(128deg, transparent 0 26%, rgba(255, 242, 169, 0.22) 26% 29%, rgba(24, 38, 82, 0.3) 29% 44%, transparent 44% 56%, rgba(111, 149, 255, 0.16) 56% 69%, transparent 69%),
            radial-gradient(circle at 68% 18%, rgba(255, 228, 134, 0.2), transparent 8rem),
            repeating-linear-gradient(98deg, transparent 0 20px, rgba(255, 231, 142, 0.08) 20px 22px);
        opacity: 0.9;
    }
}

.concept-card--icon {
    border-color: var(--card-border);
    color: var(--card-text);
    background:
        radial-gradient(circle at 50% 8%, rgba(255, 255, 255, 0.4), transparent 8rem),
        linear-gradient(142deg, rgba(184, 140, 53, 0.12), transparent 34%),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 64%, #d9c59e 100%);
    box-shadow:
        0 18px 42px rgba(0, 0, 0, 0.38),
        inset 0 0 0 1px rgba(255, 255, 255, 0.5),
        inset 0 0 0 7px rgba(132, 102, 45, 0.2),
        0 0 22px var(--card-shadow);

    &:hover,
    &:focus-within {
        box-shadow:
            0 26px 56px rgba(0, 0, 0, 0.44),
            inset 0 0 0 1px rgba(255, 255, 255, 0.58),
            inset 0 0 0 7px rgba(132, 102, 45, 0.18),
            0 0 30px var(--card-shadow);
    }

    &::before {
        border-color: rgba(160, 116, 41, 0.54);
        box-shadow:
            inset 0 0 0 1px rgba(255, 255, 255, 0.46),
            0 0 16px rgba(184, 140, 53, 0.16);
    }

    &::after {
        background: linear-gradient(90deg, transparent, var(--card-border), var(--card-cut), var(--card-border), transparent);
        box-shadow: 0 0 10px rgba(184, 140, 53, 0.18);
    }

    .concept-card__rating {
        color: var(--card-text);
        text-shadow: 0 1px 0 rgba(255, 255, 255, 0.35);
    }

    .concept-card__position,
    .concept-card__vitals,
    .concept-card__stat {
        color: var(--card-muted);
    }

    .concept-card__vitals {
        background: linear-gradient(180deg, rgba(255, 249, 232, 0.84), rgba(207, 177, 106, 0.58));
        border-color: rgba(169, 124, 38, 0.34);
    }

    .concept-card__flag,
    .concept-card__club {
        background: rgba(67, 48, 22, 0.1);
        border-color: rgba(156, 113, 38, 0.32);
    }

    .concept-card__club {
        color: var(--card-accent);
    }

    .concept-card__portrait {
        color: rgba(110, 79, 30, 0.46);
        background:
            radial-gradient(ellipse at 50% 64%, rgba(104, 78, 34, 0.12), transparent 56%),
            radial-gradient(ellipse at 50% 30%, rgba(255, 255, 255, 0.42), transparent 48%);
    }
}

.concept-card--ivory-museum {
    .concept-card__texture {
        background:
            repeating-linear-gradient(90deg, rgba(111, 84, 32, 0.04) 0 1px, transparent 1px 12px),
            linear-gradient(145deg, transparent 0 26%, rgba(181, 138, 54, 0.16) 26% 28%, transparent 28% 62%, rgba(181, 138, 54, 0.08) 62% 64%, transparent 64%),
            radial-gradient(circle at 72% 18%, rgba(255, 255, 255, 0.42), transparent 7rem);
        opacity: 0.72;
    }
}

.concept-card--pearl-relief {
    .concept-card__texture {
        background:
            repeating-radial-gradient(ellipse at 50% 24%, transparent 0 18px, rgba(174, 132, 46, 0.08) 18px 20px, transparent 20px 39px),
            linear-gradient(130deg, rgba(255, 255, 255, 0.24), transparent 28%),
            linear-gradient(39deg, transparent 0 40%, rgba(192, 148, 60, 0.12) 40% 42%, transparent 42%);
        opacity: 0.7;
    }
}

.concept-card--parchment-frame {
    .concept-card__texture {
        background:
            repeating-linear-gradient(8deg, rgba(91, 67, 28, 0.035) 0 1px, transparent 1px 7px),
            repeating-linear-gradient(98deg, transparent 0 18px, rgba(119, 87, 32, 0.04) 18px 19px),
            linear-gradient(142deg, transparent 0 34%, rgba(53, 36, 15, 0.1) 34% 44%, transparent 44%);
        opacity: 0.68;
    }
}

.concept-card--hero {
    border-color: var(--card-border);
    background:
        radial-gradient(circle at 50% 18%, color-mix(in srgb, var(--card-accent) 18%, transparent), transparent 8rem),
        linear-gradient(145deg, color-mix(in srgb, var(--card-surface-soft) 16%, transparent), transparent 30%),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 58%, #080914 100%);
    box-shadow:
        0 18px 42px rgba(0, 0, 0, 0.46),
        inset 0 0 0 1px color-mix(in srgb, var(--card-highlight) 18%, transparent),
        inset 0 0 0 7px rgba(10, 12, 30, 0.36),
        0 0 25px var(--card-shadow);

    &::before {
        border-color: color-mix(in srgb, var(--card-border) 70%, transparent);
        box-shadow:
            inset 0 0 0 1px color-mix(in srgb, var(--card-accent) 18%, transparent),
            0 0 18px var(--card-shadow);
    }

    &::after {
        background: linear-gradient(90deg, transparent, var(--card-accent), var(--card-cut), var(--card-accent), transparent);
        box-shadow: 0 0 17px var(--card-shadow);
    }

    .concept-card__vitals {
        color: #0d1020;
        background: linear-gradient(180deg, var(--card-cut), var(--card-accent));
    }

    .concept-card__portrait {
        color: color-mix(in srgb, var(--card-highlight) 76%, transparent);
        background:
            radial-gradient(ellipse at 50% 64%, rgba(0, 0, 0, 0.46), transparent 56%),
            radial-gradient(ellipse at 50% 30%, color-mix(in srgb, var(--card-accent) 16%, transparent), transparent 48%);
    }
}

.concept-card--mural-burst {
    .concept-card__texture {
        background:
            repeating-conic-gradient(from 230deg at 52% 32%, rgba(255, 220, 108, 0.2) 0deg 2deg, transparent 2deg 10deg),
            repeating-radial-gradient(circle at 72% 18%, rgba(255, 255, 255, 0.08) 0 1px, transparent 1px 6px),
            linear-gradient(122deg, transparent 0 34%, rgba(214, 63, 112, 0.2) 34% 50%, transparent 50%);
        opacity: 0.82;
    }
}

.concept-card--scarf-stripes {
    .concept-card__texture {
        background:
            repeating-linear-gradient(116deg, rgba(227, 76, 66, 0.22) 0 12px, transparent 12px 24px, rgba(64, 214, 164, 0.18) 24px 31px, transparent 31px 46px),
            linear-gradient(180deg, transparent 0 58%, rgba(0, 0, 0, 0.34) 58% 100%);
        opacity: 0.68;
    }
}

.concept-card--badge-poster {
    .concept-card__texture {
        background:
            linear-gradient(132deg, rgba(217, 79, 47, 0.22) 0 18%, transparent 18% 44%, rgba(102, 211, 255, 0.16) 44% 57%, transparent 57%),
            linear-gradient(42deg, transparent 0 24%, rgba(255, 220, 116, 0.18) 24% 26%, transparent 26% 64%, rgba(217, 79, 47, 0.16) 64% 76%, transparent 76%),
            repeating-linear-gradient(90deg, transparent 0 18px, rgba(255, 255, 255, 0.05) 18px 19px);
        opacity: 0.78;
    }
}

.concept-card--match-event {
    border-color: var(--card-border);
    background:
        radial-gradient(circle at 54% 24%, rgba(255, 117, 31, 0.24), transparent 9rem),
        linear-gradient(145deg, rgba(255, 196, 135, 0.12), transparent 30%),
        linear-gradient(180deg, var(--card-surface) 0%, var(--card-bg) 58%, #080302 100%);
    box-shadow:
        0 18px 42px rgba(0, 0, 0, 0.5),
        inset 0 0 0 1px rgba(255, 214, 172, 0.18),
        inset 0 0 0 7px rgba(63, 21, 6, 0.34),
        0 0 28px var(--card-shadow);

    &:hover,
    &:focus-within {
        box-shadow:
            0 28px 58px rgba(0, 0, 0, 0.56),
            inset 0 0 0 1px rgba(255, 214, 172, 0.24),
            inset 0 0 0 7px rgba(63, 21, 6, 0.28),
            0 0 40px var(--card-shadow);
    }

    &::before {
        border-color: rgba(255, 151, 51, 0.68);
        box-shadow:
            inset 0 0 0 1px rgba(255, 215, 169, 0.16),
            0 0 22px var(--card-shadow);
    }

    &::after {
        height: 4px;
        background: linear-gradient(90deg, transparent, var(--card-cut), var(--card-accent), var(--card-cut), transparent);
        box-shadow: 0 0 19px var(--card-shadow);
    }

    .concept-card__vitals {
        color: #170704;
        background: linear-gradient(180deg, var(--card-cut), var(--card-accent));
    }

    .concept-card__portrait {
        color: rgba(255, 231, 208, 0.82);
        background:
            radial-gradient(ellipse at 50% 64%, rgba(0, 0, 0, 0.52), transparent 56%),
            radial-gradient(ellipse at 50% 30%, rgba(255, 117, 31, 0.18), transparent 48%);
    }
}

.concept-card--orange-impact {
    .concept-card__texture {
        background:
            repeating-conic-gradient(from 230deg at 58% 28%, rgba(255, 128, 36, 0.24) 0deg 2deg, transparent 2deg 10deg),
            linear-gradient(126deg, transparent 0 24%, rgba(255, 208, 154, 0.2) 24% 26%, transparent 26%),
            linear-gradient(180deg, transparent 0 58%, rgba(0, 0, 0, 0.38) 58% 100%);
        opacity: 0.8;
    }
}

.concept-card--match-night-flare {
    .concept-card__texture {
        background:
            radial-gradient(circle at 70% 18%, rgba(255, 240, 220, 0.22), transparent 4.5rem),
            linear-gradient(135deg, transparent 0 20%, rgba(255, 109, 23, 0.26) 20% 34%, transparent 34% 52%, rgba(255, 196, 135, 0.16) 52% 54%, transparent 54%),
            repeating-linear-gradient(108deg, transparent 0 24px, rgba(255, 126, 38, 0.08) 24px 25px);
        opacity: 0.78;
    }
}

.concept-card--ember-spotlight {
    .concept-card__texture {
        background:
            conic-gradient(from 225deg at 58% 30%, transparent 0deg, rgba(255, 127, 34, 0.32) 26deg, rgba(255, 211, 157, 0.16) 42deg, transparent 64deg, transparent 360deg),
            radial-gradient(circle at 60% 28%, rgba(255, 151, 51, 0.24), transparent 8rem),
            repeating-linear-gradient(96deg, rgba(255, 127, 34, 0.05) 0 1px, transparent 1px 9px);
        opacity: 0.8;
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
    position: relative;
    padding: 0.25rem 0.2rem 0.45rem;
    text-align: left;

    &::before {
        content: "";
        position: absolute;
        inset: -0.28rem -0.12rem 0.1rem -0.38rem;
        border-left: 2px solid var(--card-cut);
        border-top: 1px solid rgba(255, 211, 156, 0.5);
        border-radius: 7px 0 0 0;
        opacity: 0.78;
    }
}

.concept-card__rating {
    color: var(--card-text);
    font-size: 3.25rem;
    line-height: 0.88;
    font-weight: 900;
    text-shadow:
        0 2px 9px rgba(0, 0, 0, 0.48),
        0 0 14px rgba(255, 181, 96, 0.18);
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
    background: linear-gradient(180deg, var(--card-accent), var(--card-border-muted));
    border: 1px solid rgba(255, 230, 198, 0.38);
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
    box-shadow:
        0 7px 16px rgba(0, 0, 0, 0.25),
        inset 0 0 0 1px rgba(255, 208, 154, 0.18);
}

.concept-card__portrait {
    position: absolute;
    left: calc(50% + 34px);
    top: 0px;
    width: 166px;
    height: 156px;
    display: grid;
    place-items: center;
    color: rgba(255, 229, 203, 0.82);
    background:
        radial-gradient(
            ellipse at 50% 64%,
            rgba(15, 11, 9, 0.42),
            transparent 56%
        ),
        radial-gradient(
            ellipse at 50% 30%,
            rgba(255, 172, 85, 0.12),
            transparent 48%
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
        text-shadow: 0 0 12px rgba(255, 181, 96, 0.22);
    }
}

@media (max-width: 980px) {
    .preview-header,
    .design-section__header,
    .card-gallery {
        grid-template-columns: 1fr;
    }

    .preview-header,
    .design-section__header {
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
