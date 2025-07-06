// Frontend Configuration for v2fmdash
// This file contains runtime configuration for the Vue.js frontend
// Values can be overridden at container startup via environment injection

window.APP_CONFIG = {
  // API Configuration
  api: {
    // Empty string means relative paths (same domain as frontend)
    // Set to full URL for external backend (e.g., "https://api.example.com")
    endpoint: '',
    timeout: 30000, // 30 seconds
  },

  // Application Information
  app: {
    title: 'Football Manager Data Browser',
    version: '1.0.0',
    description: 'A comprehensive platform for analyzing Football Manager player data',
  },

  // Feature Flags
  features: {
    analytics: true,
    export: true,
    comparison: true,
    sharing: false,
    debug: false, // Will be overridden in development
  },

  // UI Configuration
  ui: {
    theme: 'auto', // 'light', 'dark', or 'auto'
    max_players_display: 1000,
    items_per_page: 50,
    auto_save: true,
    show_tooltips: true,
  },

  // Analytics Configuration
  analytics: {
    enabled: true,
    tracking_id: 'G-QYG3QS5C5Y',
    track_page_views: true,
    track_events: true,
    anonymize_ip: true,
  },

  // Upload Configuration
  upload: {
    max_file_size_mb: 50,
    allowed_types: ['.html', '.htm'],
    chunk_size_mb: 5,
    concurrent_uploads: 2,
  },

  // Performance Configuration
  performance: {
    lazy_loading: true,
    virtual_scrolling: true,
    debounce_search_ms: 300,
    cache_ttl_minutes: 15,
  },

  // External Resources
  external: {
    image_api_url: 'https://sortitoutsi.b-cdn.net/uploads',
    help_url: 'https://git.liamhardman.com/liam/v2fmdash',
    issues_url: 'https://git.liamhardman.com/liam/v2fmdash/issues',
  },

  // Development Configuration (overridden in production)
  development: {
    mock_data: false,
    verbose_logging: false,
    show_debug_info: false,
    enable_dev_tools: false,
  },

  // Error Handling
  error_handling: {
    show_stack_traces: false,
    auto_retry_failed_requests: true,
    max_retry_attempts: 3,
    retry_delay_ms: 1000,
  },
}; 