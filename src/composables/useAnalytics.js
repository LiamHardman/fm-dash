import { onMounted } from 'vue'
import { analytics, trackEvent, trackPageView } from '../services/analytics'

/**
 * Vue composable for Google Analytics integration
 * @param {Object} options - Configuration options
 * @param {boolean} options.trackPageViews - Whether to automatically track page views
 * @returns {Object} Analytics functions and utilities
 */
export function useAnalytics(options = {}) {
  const { trackPageViews = true } = options

  // Track initial page view on mount
  if (trackPageViews) {
    onMounted(() => {
      // Use window.location for initial page view to avoid router dependency
      const currentPath = window.location.pathname + window.location.search
      const currentTitle = document.title
      trackPageView(currentPath, currentTitle)
    })
  }

  return {
    // Direct access to all analytics functions
    ...analytics,

    // Additional utility functions
    trackPageView,
    trackEvent,

    // Convenience method to track current page manually
    trackCurrentPage: () => {
      const currentPath = window.location.pathname + window.location.search
      const currentTitle = document.title
      trackPageView(currentPath, currentTitle)
    },

    // Helper to track form submissions
    trackFormSubmission: (formName, formData = {}) => {
      trackEvent('form_submission', {
        form_name: formName,
        form_data: JSON.stringify(formData),
        event_category: 'form_interaction',
      })
    },

    // Helper to track button clicks
    trackButtonClick: (buttonName, context = {}) => {
      trackEvent('button_click', {
        button_name: buttonName,
        ...context,
        event_category: 'ui_interaction',
      })
    },

    // Helper to track file operations
    trackFileOperation: (operation, fileName, fileSize = null) => {
      trackEvent('file_operation', {
        operation: operation, // 'upload', 'download', 'delete', etc.
        file_name: fileName,
        file_size: fileSize,
        event_category: 'file_interaction',
      })
    },
  }
}
