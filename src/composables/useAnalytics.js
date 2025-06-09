import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { analytics, trackPageView, trackEvent } from '../services/analytics'

/**
 * Vue composable for Google Analytics integration
 * @param {Object} options - Configuration options
 * @param {boolean} options.trackPageViews - Whether to automatically track page views
 * @returns {Object} Analytics functions and utilities
 */
export function useAnalytics(options = {}) {
  const { trackPageViews = true } = options
  const route = useRoute()

  // Automatically track page views on route changes
  if (trackPageViews) {
    // Track initial page view
    onMounted(() => {
      trackPageView(route.fullPath, route.meta?.title || document.title)
    })

    // Track subsequent page views
    watch(
      () => route.fullPath,
      (newPath) => {
        trackPageView(newPath, route.meta?.title || document.title)
      }
    )
  }

  return {
    // Direct access to all analytics functions
    ...analytics,
    
    // Additional utility functions
    trackPageView,
    trackEvent,
    
    // Convenience method to track current page manually
    trackCurrentPage: () => {
      trackPageView(route.fullPath, route.meta?.title || document.title)
    },
    
    // Helper to track form submissions
    trackFormSubmission: (formName, formData = {}) => {
      trackEvent('form_submission', {
        form_name: formName,
        form_data: JSON.stringify(formData),
        event_category: 'form_interaction'
      })
    },
    
    // Helper to track button clicks
    trackButtonClick: (buttonName, context = {}) => {
      trackEvent('button_click', {
        button_name: buttonName,
        ...context,
        event_category: 'ui_interaction'
      })
    },
    
    // Helper to track file operations
    trackFileOperation: (operation, fileName, fileSize = null) => {
      trackEvent('file_operation', {
        operation: operation, // 'upload', 'download', 'delete', etc.
        file_name: fileName,
        file_size: fileSize,
        event_category: 'file_interaction'
      })
    }
  }
} 