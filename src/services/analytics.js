/**
 * Google Analytics Service
 * Centralized service for tracking events and page views
 */

// Analytics configuration
const GA_TRACKING_ID = 'G-QYG3QS5C5Y'

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
  if (!isGtagAvailable()) {
    console.warn('Google Analytics not available')
    return
  }

  try {
    window.gtag('config', GA_TRACKING_ID, {
      page_path: pagePath,
      page_title: pageTitle
    })
  } catch (error) {
    console.error('Failed to track page view:', error)
  }
}

/**
 * Track a custom event
 * @param {string} eventName - Name of the event
 * @param {Object} parameters - Event parameters
 */
export const trackEvent = (eventName, parameters = {}) => {
  if (!isGtagAvailable()) {
    console.warn('Google Analytics not available')
    return
  }

  try {
    window.gtag('event', eventName, {
      ...parameters,
      timestamp: new Date().toISOString()
    })
  } catch (error) {
    console.error('Failed to track event:', error)
  }
}

/**
 * Predefined event tracking functions
 */
export const analytics = {
  // User interactions
  shareDataset: (datasetId) => {
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

  useWonderkids: (filters) => {
    trackEvent('use_wonderkids', {
      filters_applied: Object.keys(filters).length,
      event_category: 'feature_usage'
    })
  },

  useBargainHunter: (filters) => {
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