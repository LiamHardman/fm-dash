<template>
    <section
        class="page-hero"
        :class="{
            'page-hero--rounded': rounded,
            'page-hero--compact': compact,
        }"
    >
        <div
            class="page-hero__container"
            :class="{ 'page-hero__container--split': hasSide }"
        >
            <div class="page-hero__content">
                <div v-if="badge" class="page-hero__badge">
                    <q-icon v-if="icon" :name="icon" size="1.2rem" />
                    <span>{{ badge }}</span>
                </div>
                <h1 class="page-hero__title">
                    {{ title }}
                    <span v-if="highlight" class="page-hero__highlight">{{
                        highlight
                    }}</span>
                </h1>
                <p v-if="subtitle" class="page-hero__subtitle">
                    {{ subtitle }}
                </p>
                <div v-if="$slots.actions" class="page-hero__actions">
                    <slot name="actions" />
                </div>
            </div>
            <div v-if="hasSide" class="page-hero__side">
                <slot name="side" />
            </div>
        </div>
    </section>
</template>

<script>
import { computed, defineComponent } from 'vue'

/**
 * The single hero treatment shared by all pages: one indigo brand gradient
 * with a subtle radial orb. Page identity comes from the badge icon and the
 * content, not from a different gradient per page.
 *
 * Slots:
 *  - actions: buttons rendered under the subtitle
 *  - side: right-hand panel (stats, feature grid…). Use `.hero-stat` /
 *    `.hero-stat-number` / `.hero-stat-label` for pre-styled stat cards.
 */
export default defineComponent({
  name: 'PageHero',
  props: {
    icon: { type: String, default: '' },
    badge: { type: String, default: '' },
    title: { type: String, required: true },
    highlight: { type: String, default: '' },
    subtitle: { type: String, default: '' },
    // Rounded panel variant for heroes rendered inside padded containers.
    rounded: { type: Boolean, default: false },
    // Less vertical padding for utility pages.
    compact: { type: Boolean, default: false },
  },
  setup(_props, { slots }) {
    const hasSide = computed(() => Boolean(slots.side))
    return { hasSide }
  },
})
</script>

<style lang="scss" scoped>
@keyframes page-hero-orb {
    0%,
    100% {
        transform: translate(0, 0) scale(1);
    }
    33% {
        transform: translate(-70px, 55px) scale(1.1);
    }
    66% {
        transform: translate(55px, -35px) scale(0.92);
    }
}

.page-hero {
    background: var(--brand-gradient);
    color: var(--text-on-brand);
    padding: 4rem 0;
    margin-bottom: 2rem;
    position: relative;
    overflow: hidden;

    &::before {
        content: "";
        position: absolute;
        inset: 0;
        background: radial-gradient(
            circle at 30% 20%,
            rgba(255, 255, 255, 0.05) 0%,
            transparent 50%
        );
        pointer-events: none;
    }

    &::after {
        content: "";
        position: absolute;
        width: 640px;
        height: 640px;
        border-radius: 50%;
        background: radial-gradient(
            circle,
            rgba(100, 181, 246, 0.09) 0%,
            transparent 70%
        );
        top: -120px;
        right: 8%;
        pointer-events: none;
        animation: page-hero-orb 14s ease-in-out infinite;
    }

    &--rounded {
        border-radius: var(--radius-lg);
    }

    &--compact {
        padding: 2.5rem 0;
    }

    @media (max-width: $bp-sm) {
        padding: 3rem 0;

        &--compact {
            padding: 2rem 0;
        }
    }
}

.page-hero__container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    justify-content: center;
    align-items: center;
    position: relative;
    z-index: 1;

    &--split {
        justify-content: space-between;
        gap: 3rem;

        .page-hero__content {
            text-align: left;
        }

        @media (max-width: $bp-sm) {
            flex-direction: column;
            gap: 2rem;

            .page-hero__content {
                text-align: center;
            }
        }
    }
}

.page-hero__content {
    text-align: center;
    max-width: 640px;
}

.page-hero__badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.12);
    padding: 0.5rem 1rem;
    border-radius: 20px;
    font-size: 0.875rem;
    font-weight: 500;
    margin-bottom: 1.5rem;
    backdrop-filter: blur(10px);
}

.page-hero__title {
    font-size: clamp(2.25rem, 4vw, 3rem);
    font-weight: 700;
    line-height: 1.1;
    margin: 0 0 1rem 0;

    .page-hero__highlight {
        background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }
}

.page-hero__subtitle {
    font-size: 1.125rem;
    line-height: 1.6;
    margin: 0;
    opacity: 0.9;
}

.page-hero__actions {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
    justify-content: center;
    margin-top: 2rem;

    .page-hero__container--split & {
        justify-content: flex-start;

        @media (max-width: $bp-sm) {
            justify-content: center;
        }
    }
}

.page-hero__side {
    flex-shrink: 0;
}
</style>

<style lang="scss">
// Helper classes for slotted hero content (slot content is compiled in the
// parent's scope, so these need to be unscoped — namespaced to stay safe).
.page-hero .page-hero__side .hero-stats {
    display: flex;
    gap: 1.5rem;
    flex-wrap: wrap;
    justify-content: center;
}

.page-hero .hero-stat {
    background: rgba(255, 255, 255, 0.12);
    padding: 1.5rem;
    border-radius: var(--radius-lg);
    text-align: center;
    min-width: 120px;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);

    .hero-stat-number {
        font-size: 2rem;
        font-weight: 600;
        color: white;
        margin-bottom: 0.5rem;
    }

    .hero-stat-label {
        font-size: 0.875rem;
        color: rgba(255, 255, 255, 0.8);
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }
}
</style>
