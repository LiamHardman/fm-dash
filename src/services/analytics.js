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
      sendPageView: false, // We'll handle page views manually
      customMap: {
        customParameter1: 'datasetId',
      },
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
      page_title: pageTitle,
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
      timestamp: new Date().toISOString(),
    })
  } catch (_error) {}
}

/**
 * Predefined event tracking functions
 */
export const analytics = {
  // User interactions
  shareDataset: (datasetId) => {
    trackEvent('share_dataset', {
      datasetId: datasetId,
      eventCategory: 'engagement',
    })
  },

  uploadDataset: (fileSize, playerCount) => {
    trackEvent('upload_dataset', {
      fileSize: fileSize,
      playerCount: playerCount,
      eventCategory: 'user_action',
    })
  },

  viewPlayerDetails: (playerId, playerName) => {
    trackEvent('view_player_details', {
      playerId: playerId,
      playerName: playerName,
      eventCategory: 'content_interaction',
    })
  },

  useFilter: (filterType, filterValue) => {
    trackEvent('use_filter', {
      filterType: filterType,
      filterValue: filterValue,
      eventCategory: 'user_interaction',
    })
  },

  navigateToTeamView: (teamName, datasetId) => {
    trackEvent('navigate_team_view', {
      teamName: teamName,
      datasetId: datasetId,
      eventCategory: 'navigation',
    })
  },

  searchPlayers: (searchTerm, resultCount) => {
    trackEvent('search_players', {
      searchTerm: searchTerm,
      resultCount: resultCount,
      eventCategory: 'search',
    })
  },

  downloadData: (dataType, format) => {
    trackEvent('download_data', {
      dataType: dataType,
      format: format,
      eventCategory: 'export',
    })
  },

  // Feature usage
  useWishlist: (action, playerId) => {
    trackEvent('wishlist_action', {
      action: action,
      playerId: playerId,
      eventCategory: 'feature_usage',
    })
  },

  useUpgradeFinder: (position, filters) => {
    trackEvent('use_upgrade_finder', {
      position: position,
      filtersApplied: Object.keys(filters).length,
      eventCategory: 'feature_usage',
    })
  },

  useWonderkids: (filters) => {
    trackEvent('use_wonderkids', {
      filtersApplied: Object.keys(filters).length,
      eventCategory: 'feature_usage',
    })
  },

  useBargainHunter: (filters) => {
    trackEvent('use_bargain_hunter', {
      filtersApplied: Object.keys(filters).length,
      eventCategory: 'feature_usage',
    })
  },

  // Page visits
  visitPage: (pageName, additionalData = {}) => {
    trackEvent('page_visit', {
      pageName: pageName,
      eventCategory: 'navigation',
      ...additionalData,
    })
  },

  // Error tracking
  trackError: (errorType, errorMessage, context = {}) => {
    trackEvent('error_occurred', {
      errorType: errorType,
      errorMessage: errorMessage,
      context: JSON.stringify(context),
      eventCategory: 'error',
    })
  },
}

export default analytics
