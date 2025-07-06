/**
 * Analytics Service
 * Centralized service for tracking events and page views
 * Uses Google Analytics for tracking
 */

// Analytics configuration - get from environment variable
const GA_TRACKING_ID = import.meta.env.VITE_GA_TRACKING_ID || __GA_TRACKING_ID__ || 'G-QYG3QS5C5Y'

// // Log the tracking ID in development for debugging
// if (import.meta.env.DEV) {
//   console.log('🔍 Google Analytics Tracking ID:', GA_TRACKING_ID)
// }

/**
 * Initialize Google Analytics by dynamically loading the script
 */
const initializeGA = () => {
  if (typeof window === 'undefined' || !GA_TRACKING_ID) {
    return
  }

  try {
    // Create and inject the gtag script
    const script = document.createElement('script')
    script.async = true
    script.src = `https://www.googletagmanager.com/gtag/js?id=${GA_TRACKING_ID}`
    document.head.appendChild(script)

    // Initialize dataLayer and gtag function
    window.dataLayer = window.dataLayer || []
    window.gtag = (...args) => {
      window.dataLayer.push(args)
    }

    // Initialize with current date and config
    window.gtag('js', new Date())
    window.gtag('config', GA_TRACKING_ID, {
      // Enhanced configuration options
      send_page_view: false, // We'll handle page views manually
      custom_map: {
        custom_parameter_1: 'dataset_id'
      }
    })

    // console.log('✅ Google Analytics initialized with tracking ID:', GA_TRACKING_ID)
  } catch (_error) {}
}

// Auto-initialize when the module is loaded
initializeGA()

/**
 * Check if gtag is available
 * @returns {boolean}
 */
export const isGtagAvailable = () => {
  return typeof window !== 'undefined' && typeof window.gtag === 'function'
}

/**
 * Track a page view
 * @param {string} pagePath - The path of the page
 * @param {string} pageTitle - The title of the page
 */
export const trackPageView = (pagePath, pageTitle = '') => {
  // Google Analytics tracking
  if (!isGtagAvailable()) {
    return
  }

  try {
    window.gtag('config', GA_TRACKING_ID, {
      page_path: pagePath,
      page_title: pageTitle
    })
  } catch (_error) {}
}

/**
 * Track a custom event
 * @param {string} eventName - Name of the event
 * @param {Object} parameters - Event parameters
 */
export const trackEvent = (eventName, parameters = {}) => {
  // Google Analytics tracking
  if (!isGtagAvailable()) {
    return
  }

  try {
    window.gtag('event', eventName, {
      ...parameters,
      timestamp: new Date().toISOString()
    })
  } catch (_error) {}
}

/**
 * Predefined event tracking functions
 */
export const analytics = {
  // User interactions
  shareDataset: datasetId => {
    trackEvent('share_dataset', {
      dataset_id: datasetId,
      event_category: 'engagement'
    })
  },

  uploadDataset: (fileSize, playerCount) => {
    trackEvent('upload_dataset', {
      file_size: fileSize,
      player_count: playerCount,
      event_category: 'user_action'
    })
  },

  viewPlayerDetails: (playerId, playerName) => {
    trackEvent('view_player_details', {
      player_id: playerId,
      player_name: playerName,
      event_category: 'content_interaction'
    })
  },

  useFilter: (filterType, filterValue) => {
    trackEvent('use_filter', {
      filter_type: filterType,
      filter_value: filterValue,
      event_category: 'user_interaction'
    })
  },

  navigateToTeamView: (teamName, datasetId) => {
    trackEvent('navigate_team_view', {
      team_name: teamName,
      dataset_id: datasetId,
      event_category: 'navigation'
    })
  },

  searchPlayers: (searchTerm, resultCount) => {
    trackEvent('search_players', {
      search_term: searchTerm,
      result_count: resultCount,
      event_category: 'search'
    })
  },

  downloadData: (dataType, format) => {
    trackEvent('download_data', {
      data_type: dataType,
      format: format,
      event_category: 'engagement'
    })
  },

  // Feature usage
  useWishlist: (action, playerId) => {
    trackEvent('wishlist_action', {
      action: action, // 'add' or 'remove'
      player_id: playerId,
      event_category: 'feature_usage'
    })
  },

  useUpgradeFinder: (position, filters) => {
    trackEvent('use_upgrade_finder', {
      position: position,
      filters_applied: Object.keys(filters).length,
      event_category: 'feature_usage'
    })
  },

  useWonderkids: filters => {
    trackEvent('use_wonderkids', {
      filters_applied: Object.keys(filters).length,
      event_category: 'feature_usage'
    })
  },

  useBargainHunter: filters => {
    trackEvent('use_bargain_hunter', {
      filters_applied: Object.keys(filters).length,
      event_category: 'feature_usage'
    })
  },

  // Page visits
  visitPage: (pageName, additionalData = {}) => {
    trackEvent('page_visit', {
      page_name: pageName,
      ...additionalData,
      event_category: 'navigation'
    })
  },

  // Error tracking
  trackError: (errorType, errorMessage, context = {}) => {
    trackEvent('error_occurred', {
      error_type: errorType,
      error_message: errorMessage,
      context: JSON.stringify(context),
      event_category: 'error'
    })
  }
}

export default analytics
