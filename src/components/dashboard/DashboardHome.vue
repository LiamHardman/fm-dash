<template>
    <div class="dashboard-home">
        <PageHeader title="Dashboard" subtitle="Your loaded save at a glance" icon="dashboard">
            <template #actions>
                <q-btn
                    flat
                    dense
                    no-caps
                    icon="tune"
                    label="Customize"
                    @click="showCustomizeDialog = true"
                />
            </template>
        </PageHeader>

        <div class="dashboard-grid">
            <component :is="widget.component" v-for="widget in visibleWidgets" :key="widget.id" />
        </div>

        <CustomizeListDialog
            v-model="showCustomizeDialog"
            title="Customize dashboard"
            hint="Reorder or hide widgets on your dashboard home."
            :items="widgetMeta"
            :hidden-ids="uiStore.dashboardHiddenWidgets"
            :order-ids="uiStore.dashboardWidgetOrder"
            @update:hidden="uiStore.setDashboardHiddenWidgets"
            @update:order="uiStore.setDashboardWidgetOrder"
            @reset="resetWidgets"
        />
    </div>
</template>

<script>
import { computed, defineComponent, onMounted, ref } from 'vue'
import { usePlayerStore } from '@/stores/playerStore'
import { useUiStore } from '@/stores/uiStore'
import CustomizeListDialog from '../layout/CustomizeListDialog.vue'
import PageHeader from '../layout/PageHeader.vue'
import QuickStatsWidget from './QuickStatsWidget.vue'
import RecentDatasetsWidget from './RecentDatasetsWidget.vue'
import ShortcutsWidget from './ShortcutsWidget.vue'
import WishlistWidget from './WishlistWidget.vue'

const WIDGETS = [
  {
    id: 'quick-stats',
    label: 'Dataset overview',
    icon: 'query_stats',
    component: QuickStatsWidget,
  },
  { id: 'shortcuts', label: 'Jump to', icon: 'explore', component: ShortcutsWidget },
  { id: 'wishlist', label: 'Wishlist', icon: 'favorite', component: WishlistWidget },
  {
    id: 'recent-datasets',
    label: 'Recent saves',
    icon: 'history',
    component: RecentDatasetsWidget,
  },
]

export default defineComponent({
  name: 'DashboardHome',
  components: { PageHeader, CustomizeListDialog },
  setup() {
    const playerStore = usePlayerStore()
    const uiStore = useUiStore()
    const showCustomizeDialog = ref(false)

    // If the dashboard is opened directly (e.g. a reload while on "/") the
    // player store may not have hydrated allPlayers yet even though a
    // dataset id is present -- reuse the same session-cache-aware loader
    // DatasetPage relies on rather than duplicating its logic here.
    onMounted(() => {
      if (playerStore.currentDatasetId && playerStore.allPlayers.length === 0) {
        playerStore.loadFromSessionStorage()
      }
    })

    const widgetMeta = WIDGETS.map(({ id, label, icon }) => ({ id, label, icon }))

    const visibleWidgets = computed(() => {
      const hidden = new Set(uiStore.dashboardHiddenWidgets)
      const order = uiStore.dashboardWidgetOrder
      let widgets = WIDGETS.filter((w) => !hidden.has(w.id))
      if (order.length) {
        widgets = [...widgets].sort((a, b) => {
          const ai = order.indexOf(a.id)
          const bi = order.indexOf(b.id)
          return (ai === -1 ? Infinity : ai) - (bi === -1 ? Infinity : bi)
        })
      }
      return widgets
    })

    const resetWidgets = () => {
      uiStore.setDashboardHiddenWidgets([])
      uiStore.setDashboardWidgetOrder([])
    }

    return { uiStore, widgetMeta, visibleWidgets, showCustomizeDialog, resetWidgets }
  },
})
</script>

<style lang="scss" scoped>
.dashboard-home {
    max-width: var(--content-max-width);
    margin: 0 auto;
    padding: var(--page-gutter);
}

.dashboard-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: var(--section-gap);
    align-items: start;
}

@media (max-width: 768px) {
    .dashboard-home {
        padding: var(--page-gutter-sm);
    }
}
</style>
