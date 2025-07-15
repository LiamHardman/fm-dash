/**
 * Composable for managing dynamic component loading
 * Provides lazy-loaded versions of heavy components with loading states and error boundaries
 */

import ErrorBoundary from '@/components/ErrorBoundary.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { useDynamicLoader } from '@/utils/DynamicComponentLoader.js'

export function useDynamicComponents() {
  const { createAsyncComponent, preloadCriticalComponents } = useDynamicLoader()

  // Heavy components that should be dynamically loaded
  const DynamicPlayerDetailDialog = createAsyncComponent(
    () => import('@/components/PlayerDetailDialog.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 200,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicExportOptionsDialog = createAsyncComponent(
    () => import('@/components/ExportOptionsDialog.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 200,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicScatterPlotCard = createAsyncComponent(
    () => import('@/components/ScatterPlotCard.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 100, // Charts should load faster
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicPlayerDataTable = createAsyncComponent(
    () => import('@/components/PlayerDataTable.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 150,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicPitchDisplay = createAsyncComponent(() => import('@/components/PitchDisplay.vue'), {
    loadingComponent: LoadingSpinner,
    errorComponent: ErrorBoundary,
    delay: 150,
    timeout: 30000,
    retryAttempts: 3
  })

  const DynamicPerformanceMonitor = createAsyncComponent(
    () => import('@/components/PerformanceMonitor.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 200,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicInteractiveUploadLoader = createAsyncComponent(
    () => import('@/components/InteractiveUploadLoader.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 200,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  // Filter components
  const DynamicPlayerFilters = createAsyncComponent(
    () => import('@/components/filters/PlayerFilters.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 150,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  // Player detail components
  const DynamicPlayerAttributesCard = createAsyncComponent(
    () => import('@/components/player-details/PlayerAttributesCard.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 100,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicPlayerProfileCard = createAsyncComponent(
    () => import('@/components/player-details/PlayerProfileCard.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 100,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  const DynamicPerformanceAnalysisCard = createAsyncComponent(
    () => import('@/components/player-details/PerformanceAnalysisCard.vue'),
    {
      loadingComponent: LoadingSpinner,
      errorComponent: ErrorBoundary,
      delay: 150,
      timeout: 30000,
      retryAttempts: 3
    }
  )

  // Preload strategy for critical components
  const preloadStrategy = {
    // High priority - components likely to be used immediately
    critical: [
      {
        loader: () => import('@/components/PlayerDataTable.vue'),
        priority: 9
      },
      {
        loader: () => import('@/components/filters/PlayerFilters.vue'),
        priority: 8
      }
    ],

    // Medium priority - components used in common workflows
    common: [
      {
        loader: () => import('@/components/PlayerDetailDialog.vue'),
        priority: 7
      },
      {
        loader: () => import('@/components/ScatterPlotCard.vue'),
        priority: 6
      },
      {
        loader: () => import('@/components/PitchDisplay.vue'),
        priority: 6
      }
    ],

    // Low priority - components used less frequently
    optional: [
      {
        loader: () => import('@/components/ExportOptionsDialog.vue'),
        priority: 4
      },
      {
        loader: () => import('@/components/PerformanceMonitor.vue'),
        priority: 3
      },
      {
        loader: () => import('@/components/InteractiveUploadLoader.vue'),
        priority: 3
      }
    ]
  }

  // Initialize preloading based on route or user behavior
  const initializePreloading = (route = null) => {
    const allComponents = [
      ...preloadStrategy.critical,
      ...preloadStrategy.common,
      ...preloadStrategy.optional
    ]

    // Route-specific preloading
    if (route) {
      let routeSpecificComponents = []

      switch (route.name) {
        case 'dataset':
          routeSpecificComponents = [
            ...preloadStrategy.critical,
            ...preloadStrategy.common.filter(
              c =>
                c.loader.toString().includes('PlayerDetailDialog') ||
                c.loader.toString().includes('ScatterPlotCard')
            )
          ]
          break
        case 'team-view':
          routeSpecificComponents = [
            ...preloadStrategy.critical.filter(c =>
              c.loader.toString().includes('PlayerDataTable')
            ),
            ...preloadStrategy.common.filter(c => c.loader.toString().includes('PitchDisplay'))
          ]
          break
        case 'performance':
          routeSpecificComponents = [
            ...preloadStrategy.common.filter(c => c.loader.toString().includes('ScatterPlotCard')),
            ...preloadStrategy.optional.filter(c =>
              c.loader.toString().includes('PerformanceMonitor')
            )
          ]
          break
        case 'upload':
          routeSpecificComponents = [
            ...preloadStrategy.optional.filter(c =>
              c.loader.toString().includes('InteractiveUploadLoader')
            )
          ]
          break
        default:
          routeSpecificComponents = preloadStrategy.critical
      }

      preloadCriticalComponents(routeSpecificComponents)
    } else {
      // Default preloading
      preloadCriticalComponents(allComponents)
    }
  }

  return {
    // Dynamic components
    DynamicPlayerDetailDialog,
    DynamicExportOptionsDialog,
    DynamicScatterPlotCard,
    DynamicPlayerDataTable,
    DynamicPitchDisplay,
    DynamicPerformanceMonitor,
    DynamicInteractiveUploadLoader,
    DynamicPlayerFilters,
    DynamicPlayerAttributesCard,
    DynamicPlayerProfileCard,
    DynamicPerformanceAnalysisCard,

    // Preloading utilities
    initializePreloading,
    preloadStrategy
  }
}

// Export individual components for direct use
export const DynamicComponents = {
  PlayerDetailDialog: () => import('@/components/PlayerDetailDialog.vue'),
  ExportOptionsDialog: () => import('@/components/ExportOptionsDialog.vue'),
  ScatterPlotCard: () => import('@/components/ScatterPlotCard.vue'),
  PlayerDataTable: () => import('@/components/PlayerDataTable.vue'),
  PitchDisplay: () => import('@/components/PitchDisplay.vue'),
  PerformanceMonitor: () => import('@/components/PerformanceMonitor.vue'),
  InteractiveUploadLoader: () => import('@/components/InteractiveUploadLoader.vue'),
  PlayerFilters: () => import('@/components/filters/PlayerFilters.vue'),
  PlayerAttributesCard: () => import('@/components/player-details/PlayerAttributesCard.vue'),
  PlayerProfileCard: () => import('@/components/player-details/PlayerProfileCard.vue'),
  PerformanceAnalysisCard: () => import('@/components/player-details/PerformanceAnalysisCard.vue')
}
