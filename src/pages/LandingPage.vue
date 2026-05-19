<template>
    <q-page class="landing-page">
        <div class="main-container">
            <!-- Left Content Panel -->
            <div class="content-panel">
                <div class="hero-content">
                    <div class="accent-block">
                        <h1 class="hero-title">
                            Revolutionize Your
                            <span class="gradient-text" :class="{ 'is-fading': isTransitioning }">{{ currentWord }}</span>
                        </h1>
                    </div>
                    <p class="hero-subtitle">
                        Advanced player analysis and team optimization. Upload your data and discover insights that will transform your tactical approach.
                    </p>
                    <div class="hero-actions">
                        <q-btn
                            unelevated
                            color="primary"
                            size="lg"
                            @click="navigateToUpload"
                            class="primary-action-btn"
                            no-caps
                        >
                            <q-icon name="upload" class="q-mr-sm" />
                            Start Analysis
                        </q-btn>
                        <q-btn
                            outline
                            color="primary"
                            size="md"
                            @click="navigateToDemo"
                            class="secondary-action-btn"
                            no-caps
                        >
                            <q-icon name="play_arrow" class="q-mr-xs" />
                            View Demo
                        </q-btn>
                    </div>
                    <div class="hero-meta">
                        Open source &nbsp;·&nbsp; No account needed &nbsp;·&nbsp;
                        <span class="hero-meta-link" @click="openGitHub">
                            <q-icon name="code" size="0.85rem" /> GitHub
                        </span>
                    </div>
                </div>
            </div>

            <!-- Right Features Panel -->
            <div class="features-panel">
                <div class="features-header">
                    <h3 class="features-title">Complete FM24 Analysis Suite</h3>
                </div>

                <div class="feature-rows">
                    <div
                        v-for="(feature, i) in features"
                        :key="feature.name"
                        class="row-item"
                        :style="{ animationDelay: `${0.3 + i * 0.07}s` }"
                    >
                        <q-icon :name="feature.icon" :color="feature.color" size="1.1rem" class="row-icon" />
                        <div class="row-text">
                            <span class="row-name">{{ feature.name }}</span>
                            <span class="row-desc">{{ feature.description }}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import { computed, defineComponent, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUiStore } from '@/stores/uiStore'

const ROTATING_WORDS = ['Scouting Approach', 'Transfer Strategy', 'Tactical Edge']

const FEATURES = [
  {
    name: 'Data Upload',
    icon: 'upload',
    color: 'primary',
    description: 'Upload your FM24 HTML exports for instant analysis.',
  },
  {
    name: 'Player Search',
    icon: 'search',
    color: 'secondary',
    description: 'Advanced filtering by position, age, stats, and more.',
  },
  {
    name: 'Scouting Tools',
    icon: 'stars',
    color: 'accent',
    description: 'Find wonderkids, bargains, and player upgrades.',
  },
  {
    name: 'Nation Analysis',
    icon: 'flag',
    color: 'info',
    description: 'Scout out other national teams and compare squads.',
  },
  {
    name: 'League Analysis',
    icon: 'emoji_events',
    color: 'positive',
    description: 'Browse teams and players by competition.',
  },
  {
    name: 'Top Performers',
    icon: 'leaderboard',
    color: 'warning',
    description: 'See who leads in passes/90, dribbles/90, and other key stats.',
  },
  {
    name: 'Wishlist System',
    icon: 'favorite',
    color: 'negative',
    description: 'Save and track players across datasets.',
  },
  {
    name: 'Team of the Season',
    icon: 'style',
    color: 'purple',
    description: 'Special card designs that turn FM into a trading card game.',
  },
]

export default defineComponent({
  name: 'LandingPage',
  setup() {
    const router = useRouter()
    const uiStore = useUiStore()

    // Rotating headline
    const wordIndex = ref(0)
    const isTransitioning = ref(false)
    const currentWord = computed(() => ROTATING_WORDS[wordIndex.value])
    let wordInterval = null

    onMounted(() => {
      wordInterval = setInterval(() => {
        isTransitioning.value = true
        setTimeout(() => {
          wordIndex.value = (wordIndex.value + 1) % ROTATING_WORDS.length
          isTransitioning.value = false
        }, 350)
      }, 2800)
    })

    onUnmounted(() => {
      clearInterval(wordInterval)
    })

    const navigateToUpload = () => router.push('/upload')
    const navigateToDemo = () => router.push('/dataset/45e277af-1cf1-4688-9874-c59e1f3026ae')
    const showTutorial = () => uiStore.showTutorial()
    const openGitHub = () => window.open('https://github.com/LiamHardman/fm-dash', '_blank')

    return {
      features: FEATURES,
      currentWord,
      isTransitioning,
      navigateToUpload,
      navigateToDemo,
      showTutorial,
      openGitHub,
    }
  },
})
</script>

<style lang="scss" scoped>

// ─── Keyframes ────────────────────────────────────────────────────────────────

@keyframes slideInLeft {
  from { opacity: 0; transform: translateX(-36px); }
  to   { opacity: 1; transform: translateX(0); }
}

@keyframes slideInRight {
  from { opacity: 0; transform: translateX(36px); }
  to   { opacity: 1; transform: translateX(0); }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes bgOrb {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33%       { transform: translate(-70px, 55px) scale(1.1); }
  66%       { transform: translate(55px, -35px) scale(0.92); }
}

@keyframes shimmerSweep {
  from { left: -100%; }
  to   { left: 100%; }
}

// ─── Page ─────────────────────────────────────────────────────────────────────

.landing-page {
    min-height: 100vh;
    height: 100vh;
    overflow: hidden;
    background: linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%);
    color: white;
    position: relative;

    &::before {
        content: "";
        position: absolute;
        inset: 0;
        background: radial-gradient(circle at 30% 20%, rgba(255,255,255,0.05) 0%, transparent 50%);
        pointer-events: none;
        z-index: 0;
    }

    &::after {
        content: "";
        position: absolute;
        width: 640px;
        height: 640px;
        border-radius: 50%;
        background: radial-gradient(circle, rgba(100,181,246,0.09) 0%, transparent 70%);
        top: -120px;
        right: 8%;
        pointer-events: none;
        z-index: 0;
        animation: bgOrb 14s ease-in-out infinite;
    }

    .body--dark & {
        background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
    }
}

// ─── Layout ───────────────────────────────────────────────────────────────────

.main-container {
    display: flex;
    height: 100vh;
    max-width: 1400px;
    margin: 0 auto;
    padding: 3rem;
    gap: 4rem;
    align-items: center;
    position: relative;
    z-index: 1;
}

.content-panel {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem 0;
}

// ─── Hero content ─────────────────────────────────────────────────────────────

.hero-content {
    max-width: 500px;

    .accent-block {
        padding-left: 1rem;
        border-left: 3px solid #42a5f5;
        animation: slideInLeft 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
        animation-delay: 0.05s;
    }

    .hero-title {
        font-size: 3rem;
        font-weight: 700;
        line-height: 1.1;
        margin: 0;

        .gradient-text {
            display: block;
            background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            transition: opacity 0.35s ease;

            &.is-fading { opacity: 0; }
        }
    }

    .hero-subtitle {
        font-size: 1.1rem;
        line-height: 1.5;
        margin: 0 0 2rem 0;
        opacity: 0.9;
        font-weight: 300;
        animation: slideInLeft 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
        animation-delay: 0.2s;
    }

    .hero-actions {
        display: flex;
        flex-direction: row;
        gap: 1rem;
        flex-wrap: wrap;
        animation: slideInLeft 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
        animation-delay: 0.35s;

        .primary-action-btn {
            padding: 0.8rem 1.8rem;
            font-weight: 600;
            letter-spacing: 0.5px;
            border-radius: 12px;
            transition: transform 0.3s ease, box-shadow 0.3s ease;
            position: relative;
            overflow: hidden;

            &::after {
                content: "";
                position: absolute;
                top: 0;
                left: -100%;
                width: 60%;
                height: 100%;
                background: linear-gradient(90deg, transparent, rgba(255,255,255,0.18), transparent);
                pointer-events: none;
            }

            &:hover {
                transform: translateY(-2px);
                box-shadow: 0 8px 28px rgba(66, 165, 245, 0.45);

                &::after { animation: shimmerSweep 0.55s ease forwards; }
            }
        }

        .secondary-action-btn {
            padding: 0.8rem 1.5rem;
            font-weight: 600;
            letter-spacing: 0.5px;
            border-radius: 12px;
            border-width: 2px;
            transition: transform 0.3s ease, background 0.3s ease;

            &:hover {
                transform: translateY(-2px);
                background: rgba(255, 255, 255, 0.1);
            }
        }
    }

    .hero-meta {
        font-size: 0.82rem;
        color: rgba(255, 255, 255, 0.4);
        animation: slideInLeft 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
        animation-delay: 0.48s;

        .hero-meta-link {
            color: rgba(100, 181, 246, 0.7);
            cursor: pointer;
            transition: color 0.2s ease;

            &:hover { color: #64b5f6; }
        }
    }
}

// ─── Features panel ───────────────────────────────────────────────────────────

.features-panel {
    flex: 1;
    padding: 2rem 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
}

.features-header {
    margin-bottom: 1.5rem;
    animation: fadeInUp 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
    animation-delay: 0.25s;

    .features-title {
        font-size: 2rem;
        font-weight: 700;
        color: rgba(255, 255, 255, 0.95);
        margin: 0;

        .body--dark & { color: rgba(255, 255, 255, 0.9); }
    }
}

// ─── Stacked rows ─────────────────────────────────────────────────────────────

.feature-rows {
    display: flex;
    flex-direction: column;
}

.row-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.8rem 0.6rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    transition: background 0.2s ease;
    animation: slideInRight 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;

    &:last-child { border-bottom: none; }

    &:hover {
        background: rgba(255, 255, 255, 0.06);
        cursor: default;
    }

    .row-icon {
        flex-shrink: 0;
    }

    .row-text {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
        min-width: 0;
    }

    .row-name {
        font-size: 0.95rem;
        font-weight: 600;
        color: rgba(255, 255, 255, 0.9);
    }

    .row-desc {
        font-size: 0.8rem;
        color: rgba(255, 255, 255, 0.5);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}

// ─── Responsive ───────────────────────────────────────────────────────────────

@media (max-width: 1200px) {
    .main-container { padding: 2rem; gap: 2rem; }
    .hero-content .hero-title { font-size: 2.5rem; }
    .features-header .features-title { font-size: 1.8rem; }
}

@media (max-width: 1024px) {
    .main-container {
        flex-direction: column;
        padding: 2rem;
        gap: 3rem;
        justify-content: center;
    }

    .content-panel,
    .features-panel {
        flex: none;
        height: auto;
        padding: 0;
    }
}

@media (max-width: 768px) {
    .landing-page {
        overflow-y: auto;
        height: auto;
        min-height: 100vh;
    }

    .main-container {
        height: auto;
        padding: 2rem 1rem;
        gap: 2rem;
    }

    .hero-content {
        text-align: center;

        .hero-title { font-size: 2rem; }
        .hero-subtitle { font-size: 1rem; }

        .hero-actions {
            align-items: center;
        }
    }

    .features-header .features-title { font-size: 1.5rem; }

    .row-item .row-desc { white-space: normal; }
}

@media (max-width: 480px) {
    .main-container { padding: 1.5rem; }
    .hero-content .hero-title { font-size: 1.8rem; }
    .hero-content .hero-subtitle { font-size: 0.9rem; margin-bottom: 1.5rem; }
    .features-header .features-title { font-size: 1.3rem; }
}
</style>
