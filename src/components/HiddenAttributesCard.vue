<template>
    <!-- Hidden and personality attributes: save imports only, behind the Settings toggle -->
    <div v-if="visible" class="col-12">
        <q-card flat bordered class="attribute-card modern-attribute-card hidden-attributes-card">
            <q-card-section class="attribute-card-header">
                <div class="attribute-section-title">
                    <q-icon name="visibility" class="q-mr-sm" />
                    Hidden
                </div>
            </q-card-section>
            <q-card-section class="q-pa-md">
                <div v-if="!playerHasValues" class="text-center q-py-md no-hidden-attributes">
                    Not in this import
                </div>
                <template v-else>
                    <div v-for="group in groups" :key="group.name" class="hidden-attribute-group">
                        <div class="hidden-attribute-group-title">{{ group.name }}</div>
                        <q-list separator dense class="attribute-list">
                            <q-item
                                v-for="name in group.attributes"
                                :key="name"
                                class="attribute-list-item modern-attribute-item"
                            >
                                <q-item-section>
                                    <q-item-label lines="1" class="attribute-name">{{ name }}</q-item-label>
                                </q-item-section>
                                <q-item-section side>
                                    <span
                                        :class="['attribute-value modern-attribute-value', colourClass(name)]"
                                        :data-attribute="name"
                                    >
                                        {{ player?.hiddenAttributes?.[name] ?? '-' }}
                                    </span>
                                </q-item-section>
                            </q-item>
                        </q-list>
                    </div>
                </template>
            </q-card-section>
        </q-card>
    </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePlayerStore } from '../stores/playerStore'
import { useUiStore } from '../stores/uiStore'
import { getUnifiedRatingClass } from '../utils/colourSets'
import {
  HIDDEN_ATTRIBUTE_GROUPS,
  hasHiddenAttributes,
  hiddenAttributeColorValue,
} from '../utils/hiddenAttributes.js'

const props = defineProps({
  player: { type: Object, default: null },
})

const uiStore = useUiStore()
const playerStore = usePlayerStore()

// biome-ignore lint/correctness/noUnusedVariables: used in template
const groups = HIDDEN_ATTRIBUTE_GROUPS

// Shown only with the toggle on and a Dataset that carries hidden attributes; CSV
// Datasets never show the card.
// biome-ignore lint/correctness/noUnusedVariables: used in template
const visible = computed(
  () => uiStore.showHiddenAttributes && hasHiddenAttributes(playerStore.allPlayers)
)

// biome-ignore lint/correctness/noUnusedVariables: used in template
const playerHasValues = computed(() => Object.keys(props.player?.hiddenAttributes ?? {}).length > 0)

// Same 1-20 colour scale as the other attributes, inverted where lower is better.
// biome-ignore lint/correctness/noUnusedVariables: used in template
const colourClass = (name) =>
  getUnifiedRatingClass(hiddenAttributeColorValue(name, props.player?.hiddenAttributes?.[name]), 20)
</script>

<style lang="scss" scoped>
// Layout copied from PlayerDetailDialog's scoped attribute-card styles, which don't reach
// this child component; the colour classes come from the global colour sets.
.attribute-card {
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-1);
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
}

.attribute-card-header {
    background: linear-gradient(135deg, var(--accent-soft-strong) 0%, var(--accent-soft) 100%);
    border-radius: 12px 12px 0 0;
    padding: 16px 20px;
}

.attribute-section-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--accent);
    display: flex;
    align-items: center;
}

.modern-attribute-item {
    border-radius: 6px;
    margin: 1px 4px;
    padding: 4px 12px;
    min-height: 32px;

    &:hover {
        background: var(--accent-soft);
    }
}

.attribute-name {
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.modern-attribute-value {
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 700;
    text-align: center;
    min-width: 32px;
    display: inline-block;
}

.hidden-attribute-group + .hidden-attribute-group {
    margin-top: 12px;
}

.hidden-attribute-group-title {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    opacity: 0.7;
    margin: 0 16px 4px;
}

.no-hidden-attributes {
    opacity: 0.7;
}
</style>
