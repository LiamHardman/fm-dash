<template>
    <q-dialog v-model="show">
        <q-card class="customize-card">
            <q-card-section class="row items-center customize-header">
                <q-icon name="tune" class="q-mr-sm" color="primary" />
                <div class="text-subtitle1 text-weight-bold">{{ title }}</div>
                <q-space />
                <q-btn icon="close" flat round dense v-close-popup />
            </q-card-section>

            <q-separator />

            <q-card-section class="customize-body">
                <p v-if="hint" class="customize-hint">{{ hint }}</p>
                <q-list separator>
                    <q-item v-for="(item, idx) in orderedItems" :key="item.id" class="customize-item">
                        <q-item-section avatar>
                            <q-icon :name="item.icon" />
                        </q-item-section>
                        <q-item-section>{{ item.label }}</q-item-section>
                        <q-item-section side>
                            <div class="row items-center q-gutter-xs">
                                <q-btn
                                    dense
                                    flat
                                    round
                                    icon="keyboard_arrow_up"
                                    :disable="idx === 0"
                                    aria-label="Move up"
                                    @click="moveUp(item.id)"
                                />
                                <q-btn
                                    dense
                                    flat
                                    round
                                    icon="keyboard_arrow_down"
                                    :disable="idx === orderedItems.length - 1"
                                    aria-label="Move down"
                                    @click="moveDown(item.id)"
                                />
                                <q-toggle
                                    :model-value="!hiddenSet.has(item.id)"
                                    color="primary"
                                    :aria-label="`Toggle visibility of ${item.label}`"
                                    @update:model-value="(val) => setHidden(item.id, !val)"
                                />
                            </div>
                        </q-item-section>
                    </q-item>
                </q-list>
            </q-card-section>

            <q-card-actions align="right">
                <q-btn flat no-caps label="Reset to default" @click="$emit('reset')" />
                <q-btn unelevated no-caps color="primary" label="Done" v-close-popup />
            </q-card-actions>
        </q-card>
    </q-dialog>
</template>

<script>
import { computed, defineComponent } from 'vue'

export default defineComponent({
  name: 'CustomizeListDialog',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: 'Customize',
    },
    hint: {
      type: String,
      default: '',
    },
    items: {
      type: Array,
      required: true, // [{ id, label, icon }]
    },
    hiddenIds: {
      type: Array,
      default: () => [],
    },
    orderIds: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:modelValue', 'update:hidden', 'update:order', 'reset'],
  setup(props, { emit }) {
    const show = computed({
      get: () => props.modelValue,
      set: (value) => emit('update:modelValue', value),
    })

    const hiddenSet = computed(() => new Set(props.hiddenIds))

    const orderedItems = computed(() => {
      if (!props.orderIds.length) return props.items
      return [...props.items].sort((a, b) => {
        const ai = props.orderIds.indexOf(a.id)
        const bi = props.orderIds.indexOf(b.id)
        return (ai === -1 ? Infinity : ai) - (bi === -1 ? Infinity : bi)
      })
    })

    const setHidden = (id, hidden) => {
      const current = new Set(props.hiddenIds)
      if (hidden) {
        current.add(id)
      } else {
        current.delete(id)
      }
      emit('update:hidden', [...current])
    }

    const currentOrderIds = () =>
      orderedItems.value.length ? orderedItems.value.map((i) => i.id) : props.items.map((i) => i.id)

    const moveUp = (id) => {
      const ids = currentOrderIds()
      const idx = ids.indexOf(id)
      if (idx <= 0) return
      ;[ids[idx - 1], ids[idx]] = [ids[idx], ids[idx - 1]]
      emit('update:order', ids)
    }

    const moveDown = (id) => {
      const ids = currentOrderIds()
      const idx = ids.indexOf(id)
      if (idx === -1 || idx >= ids.length - 1) return
      ;[ids[idx + 1], ids[idx]] = [ids[idx], ids[idx + 1]]
      emit('update:order', ids)
    }

    return {
      show,
      orderedItems,
      hiddenSet,
      setHidden,
      moveUp,
      moveDown,
    }
  },
})
</script>

<style lang="scss" scoped>
.customize-card {
    min-width: 380px;
    max-width: 95vw;
    background: var(--surface-card);
    color: var(--text-primary);
}

.customize-header {
    padding: 1rem 1.25rem;
}

.customize-body {
    max-height: 60vh;
    overflow-y: auto;
}

.customize-hint {
    font-size: 0.82rem;
    color: var(--text-secondary);
    margin: 0 0 0.75rem 0;
}

.customize-item {
    padding-left: 0;
    padding-right: 0;
}
</style>
